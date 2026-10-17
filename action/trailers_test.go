package action

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// TestBodyTrailerRefused is T6: a body line that reads as a trailer is
// body-trailer, and a Source whose trailer line holds a blank line fails the
// proof of step 6.
func TestBodyTrailerRefused(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo(agentNm, agent)
	for _, body := range []string{"Some context.\nReviewed: 9f3f design", "reviewed: 9f3f design", "CO-AUTHORED-BY: x <y@z>"} {
		_, err := r.Reaffirm(ctx, []string{"b618"}, Options{Body: body})
		wantRefusal(t, err, "body-trailer")
	}
	if _, err := r.Reaffirm(ctx, []string{"b618"}, Options{Body: "Reviewing: not a trailer key"}); err != nil {
		t.Errorf("a body line of another key: %v", err)
	}
	tr.src.trailer = func(kind, id, gate string) string { return "Reaffirmed: " + id + "\n\nExtra: line" }
	_, err := r.Reaffirm(ctx, []string{"b618"}, Options{})
	var x *ExecError
	if !errors.As(err, &x) || !strings.Contains(x.Stderr, "Git reads the trailers") {
		t.Errorf("a trailer line with a blank line: error %v, want the ExecError of step 6", err)
	}
}

// TestReviewByWrongHand is T7: review 9f31 design as the agent is
// wrong-hand and names the reviewer; as the reviewer by author, and by
// committer alone, it commits.
func TestReviewByWrongHand(t *testing.T) {
	tr := newRepo(t)
	review := func(r *Repo) (*Plan, error) { return r.Review(ctx, []string{"9f31"}, "design", Options{}) }
	_, err := review(tr.repo(agentNm, agent))
	if r := theRefusal(t, err, "wrong-hand"); !strings.Contains(r.Text, ben) || r.Rule != "H2" {
		t.Errorf("the refusal %+v names no reviewer", r)
	}
	byAuthor := &Repo{Dir: tr.dir, Source: tr.src, Env: []string{
		"GIT_AUTHOR_NAME=Ben", "GIT_AUTHOR_EMAIL=" + ben, "GIT_COMMITTER_NAME=Agent", "GIT_COMMITTER_EMAIL=" + agent}}
	byCommitter := &Repo{Dir: tr.dir, Source: tr.src, Env: []string{
		"GIT_AUTHOR_NAME=Agent", "GIT_AUTHOR_EMAIL=" + agent, "GIT_COMMITTER_NAME=Ben", "GIT_COMMITTER_EMAIL=" + ben}}
	for name, r := range map[string]*Repo{"author": byAuthor, "committer": byCommitter} {
		_, res := tr.commit(r, review)
		if got := tr.run("log", "-1", "--format=%ae %ce %(trailers:only,unfold)", res.Commit); got != map[string]string{
			"author": ben + " " + agent + " Reviewed: 9f31 design", "committer": agent + " " + ben + " Reviewed: 9f31 design"}[name] {
			t.Errorf("by %s: the commit reads %q", name, got)
		}
	}
}

// TestEmptyCommits is T17: review, reaffirm and authorise with three tasks
// make one commit with three trailers in the order given and an empty diff;
// one refusal among three makes none.
func TestEmptyCommits(t *testing.T) {
	tr := newRepo(t)
	tr.src.tasks["c07d"].Junctions[3].Reviewer = ben // design
	both := &Repo{Dir: tr.dir, Source: tr.src, Env: []string{
		"GIT_AUTHOR_NAME=Owner", "GIT_AUTHOR_EMAIL=" + owner, "GIT_COMMITTER_NAME=Ben", "GIT_COMMITTER_EMAIL=" + ben}}
	ids := []string{"c07d", "b618", "9f31"}
	cases := []struct {
		name, subject string
		plan          func(*Repo, []string) (*Plan, error)
		trailer       func(string) string
	}{
		{"review", "Accept tasks c07d, b618, 9f31 at design",
			func(r *Repo, ids []string) (*Plan, error) { return r.Review(ctx, ids, "design", Options{}) },
			func(id string) string { return "Reviewed: " + id + " design" }},
		{"reaffirm", "Reaffirm tasks c07d, b618, 9f31",
			func(r *Repo, ids []string) (*Plan, error) { return r.Reaffirm(ctx, ids, Options{}) },
			func(id string) string { return "Reaffirmed: " + id }},
		{"authorise", "Authorise tasks c07d, b618, 9f31",
			func(r *Repo, ids []string) (*Plan, error) { return r.Authorise(ctx, ids, Options{}) },
			func(id string) string { return "Authorised: " + id }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := tr.head()
			p, res := tr.commit(both, func(r *Repo) (*Plan, error) { return c.plan(r, ids) })
			if p.Diff != "" || len(p.Files) != 0 || p.Subject != c.subject {
				t.Errorf("plan: subject %q, diff %q, files %d", p.Subject, p.Diff, len(p.Files))
			}
			if n := tr.run("rev-list", "--count", before+".."+res.Commit); n != "1" {
				t.Errorf("%s commits", n)
			}
			var want []string
			for _, id := range ids {
				want = append(want, c.trailer(id))
			}
			if got := tr.run("log", "-1", "--format=%(trailers:only,unfold)", res.Commit); got != strings.Join(want, "\n") {
				t.Errorf("trailers %q, want %q", got, want)
			}
			if diff := tr.run("diff", "--stat", before, res.Commit); diff != "" {
				t.Errorf("diff %q", diff)
			}
			// One refusal among three makes none.
			head := tr.head()
			among, code := []string{"c07d", "07e0", "9f31"}, "not-leaf"
			if c.name == "authorise" {
				among, code = []string{"c07d", "beef", "9f31"}, "no-task"
			}
			if _, err := c.plan(both, among); !slices.Equal(refusalCodes(t, err), []string{code}) {
				t.Errorf("%v: %v, want one %s", among, err, code)
			}
			if c.name == "authorise" {
				_, err := c.plan(tr.repo("Ada", ada), ids)
				if codes := refusalCodes(t, err); len(codes) != 3 {
					t.Errorf("as ada: %v", codes)
				}
			}
			if tr.head() != head {
				t.Error("a refused plan moved HEAD")
			}
		})
	}
	t.Run("more than four tasks", func(t *testing.T) {
		tasks := make([]Task, 5)
		if got := about(tasks); got != "5 tasks" {
			t.Errorf("about five tasks: %q", got)
		}
	})
	t.Run("a repeated id and no id", func(t *testing.T) {
		_, err := both.Reaffirm(ctx, []string{"b618", "b618"}, Options{})
		wantRefusal(t, err, "bad-field")
		_, err = both.Authorise(ctx, nil, Options{})
		wantRefusal(t, err, "bad-field")
		calls := tr.src.calls["Project"] + tr.src.calls["Task"]
		_, err = both.Review(ctx, []string{"beef", "B618"}, "design", Options{})
		if codes := refusalCodes(t, err); strings.Join(codes, " ") != "no-task no-task" {
			t.Errorf("unknown ids: %v", codes)
		}
		if tr.src.calls["Project"]+tr.src.calls["Task"] != calls {
			t.Error("an unknown id reached the Source")
		}
	})
}

// TestReviewRefusals is the first table of T18.
func TestReviewRefusals(t *testing.T) {
	tr := newRepo(t)
	rows := []struct {
		id, gate, as, code, rule, names string
	}{
		{"b618", "nogate", owner, "unknown-gate", "H1", ""},
		{"b618", "mockup", owner, "gate-not-applicable", "H1", ""},
		{"b618", "undefined", owner, "gate-not-applicable", "H1", ""},
		{"07e0", "design", owner, "not-leaf", "", ""},
		{"b618", "defined", owner, "defined-review", "S11", "authorise"},
		{"c07d", "implementation", owner, "recursive", "", ""},
		{"9f31", "function", owner, "no-reviewer", "", ""},
		{"9f31", "design", ada, "wrong-hand", "H2", ben},
	}
	for _, row := range rows {
		_, err := tr.repo("Hand", row.as).Review(ctx, []string{row.id}, row.gate, Options{})
		r := theRefusal(t, err, row.code)
		if r.Rule != row.rule || !strings.Contains(r.Text, row.names) || r.Gate != row.gate {
			t.Errorf("review %s %s: %+v", row.id, row.gate, r)
		}
	}
}

// TestReaffirmRefusals is the second table of T18.
func TestReaffirmRefusals(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo(agentNm, agent)
	_, err := r.Reaffirm(ctx, []string{"07e0"}, Options{})
	wantRefusal(t, err, "not-leaf")
	tr.src.tasks["9f31"].Status = nil
	_, err = r.Reaffirm(ctx, []string{"9f31"}, Options{})
	wantRefusal(t, err, "no-status")
	tr.src.tasks["c07d"].Status = &Status{Gate: "design"}
	_, err = r.Reaffirm(ctx, []string{"c07d"}, Options{})
	if got := theRefusal(t, err, "recursive"); !strings.Contains(got.Text, "implementation") {
		t.Errorf("the refusal %+v names no next junction", got)
	}
	p, err := r.Reaffirm(ctx, []string{"b618"}, Options{})
	if err != nil || len(p.Warnings) != 0 {
		t.Errorf("reaffirm b618 as its contributor: %v, warnings %v", err, p.Warnings)
	}
}

// TestAuthoriseRefusals is the third table of T18.
func TestAuthoriseRefusals(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo("Owner", owner)
	_, err := tr.repo("Ada", ada).Authorise(ctx, []string{"b618"}, Options{})
	if got := theRefusal(t, err, "wrong-hand"); !strings.Contains(got.Text, ben+", "+owner) {
		t.Errorf("the refusal %+v names no authorities", got)
	}
	if _, err := tr.repo("Ben", ben).Authorise(ctx, []string{"b618"}, Options{}); err != nil {
		t.Errorf("as the nearest authority: %v", err)
	}
	tr.run("checkout", "-q", "-b", "topic")
	_, err = r.Authorise(ctx, []string{"b618"}, Options{})
	if got := theRefusal(t, err, "off-trunk"); !strings.Contains(got.Text, "main") {
		t.Errorf("off the trunk: %+v", got)
	}
	tr.run("checkout", "-q", "main")
	tr.src.project.Trunk = ""
	_, err = r.Authorise(ctx, []string{"b618"}, Options{})
	if got := theRefusal(t, err, "off-trunk"); !strings.Contains(got.Text, "undetermined") {
		t.Errorf("an undetermined trunk: %+v", got)
	}
}
