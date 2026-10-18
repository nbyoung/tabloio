package action

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// b618Function is the self-reviewed record of plan.txt.
var b618Function = Recording{ID: "b618", Gate: "function", State: "nominal", Note: "prototype/b618 composes commits: trailers, status and task files as text"}

// sonnet is the agent of plan.txt.
var sonnet = Options{Agent: Agent{Model: "claude-sonnet-5-5", CoAuthor: coAuthor}}

func record(rec Recording, o Options) func(*Repo) (*Plan, error) {
	return func(r *Repo) (*Plan, error) { return r.Record(ctx, rec, o) }
}

// TestIdentityFromEnvironment is T1: with user.email configured and the
// identity in Env, Plan.Author and the commit's %ae and %ce hold the
// environment's address, and the self-review of b618 passes only under it.
func TestIdentityFromEnvironment(t *testing.T) {
	tr := newRepo(t)
	tr.run("config", "user.name", "Norman Young")
	tr.run("config", "user.email", owner)
	_, err := (&Repo{Dir: tr.dir, Source: tr.src}).Record(ctx, b618Function, sonnet)
	if got := theRefusal(t, err, "unreviewed"); got.Gate != "function" || !strings.Contains(got.Text, agent) {
		t.Errorf("under the configuration: %+v", got)
	}
	p, res := tr.commit(tr.repo(agentNm, agent), record(b618Function, sonnet))
	if p.Author != (Identity{agentNm, agent}) || p.Committer != (Identity{agentNm, agent}) {
		t.Errorf("plan: author %v, committer %v", p.Author, p.Committer)
	}
	if got := tr.run("log", "-1", "--format=%ae %ce", res.Commit); got != agent+" "+agent {
		t.Errorf("the commit reads %q", got)
	}
	if p.Trailers[0] != (Trailer{"Reviewed", "b618 function"}) {
		t.Errorf("trailers %v", p.Trailers)
	}
}

// testRecordLines is the half of T2 that commits: a CRLF file and a file
// with no final newline keep their endings, and git diff --numstat counts no
// line beyond the changed fields.
func testRecordLines(t *testing.T) {
	tr := newRepo(t)
	tr.write(".tableaux/status/b618.yaml", strings.ReplaceAll(b618Status, "\n", "\r\n"))
	tr.write(".tableaux/status/9f31.yaml", strings.TrimSuffix(f931Status, "\n"))
	tr.run("commit", "-q", "-am", "Change the endings")
	tr.src.tasks["9f31"].Reviews = []Review{{Gate: "design", By: ben}}
	_, res := tr.commit(tr.repo(agentNm, agent), record(Recording{ID: "b618", Gate: "function", State: "nominal", Note: "Write commands shown"}, sonnet))
	if got := tr.run("diff", "--numstat", res.Commit+"~1", res.Commit); got != "2\t2\t.tableaux/status/b618.yaml" {
		t.Errorf("CRLF: numstat %q", got)
	}
	if got := tr.read(".tableaux/status/b618.yaml"); got != "gate: function\r\nstate: nominal\r\nnote: Write commands shown\r\n" {
		t.Errorf("CRLF: %q", got)
	}
	_, res = tr.commit(tr.repo("Ada", ada), record(Recording{ID: "9f31", Gate: "design", State: "nominal"}, Options{}))
	if got := tr.run("diff", "--numstat", res.Commit+"~1", res.Commit); got != "1\t1\t.tableaux/status/9f31.yaml" {
		t.Errorf("no final newline: numstat %q", got)
	}
	if got := tr.read(".tableaux/status/9f31.yaml"); got != "gate: design\nstate: nominal" {
		t.Errorf("no final newline: %q", got)
	}
}

// testRecordNote is the half of T4 that commits: the note of 2026-10-05
// lands in double quotes.
func testRecordNote(t *testing.T) {
	tr := newRepo(t)
	tr.commit(tr.repo("Ada", ada), record(Recording{ID: "9f31", Note: noteOf20261005}, Options{}))
	if got := tr.read(".tableaux/status/9f31.yaml"); got != f931Status+"note: \""+noteOf20261005+"\"\n" {
		t.Errorf("the file reads %q", got)
	}
}

// TestTrailersShareOneParagraph is T5.
func TestTrailersShareOneParagraph(t *testing.T) {
	tr := newRepo(t)
	p, res := tr.commit(tr.repo(agentNm, agent), record(b618Function, sonnet))
	msg := p.Message()
	if _, after, _ := strings.Cut(msg, "\nReviewed:"); strings.Contains(after, "\n\n") || !strings.HasSuffix(msg, coAuthor+"\n") {
		t.Errorf("message %q", msg)
	}
	body := tr.run("log", "-1", "--format=%B", res.Commit)
	cmd := tr.run("log", "-1", "--format=%(trailers:only,unfold)", res.Commit)
	want := "Reviewed: b618 function\nModel: claude-sonnet-5-5\nCo-Authored-By: " + coAuthor
	if cmd != want || !strings.HasSuffix(body, want) {
		t.Errorf("trailers %q, want %q", cmd, want)
	}
	if got := tr.run("log", "--format=%H", "--grep", "^Reviewed: b618 function$"); got != res.Commit {
		t.Errorf("git log --grep finds %q", got)
	}
}

// TestGateNotApplicable is T8: record c07d --gate mockup and review c07d
// mockup are gate-not-applicable, and nothing is written.
func TestGateNotApplicable(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo("Ada", ada)
	head := tr.head()
	_, err := r.Record(ctx, Recording{ID: "c07d", Gate: "mockup"}, Options{})
	if got := theRefusal(t, err, "gate-not-applicable"); got.Rule != "S7" {
		t.Errorf("record: %+v", got)
	}
	_, err = r.Review(ctx, []string{"c07d"}, "mockup", Options{})
	wantRefusal(t, err, "gate-not-applicable")
	if tr.head() != head || tr.read(".tableaux/status/c07d.yaml") != c07dStatus || tr.run("status", "--porcelain") != "" {
		t.Error("a refusal wrote something")
	}
}

// TestCompleteOnlyAtLastGate is T9.
func TestCompleteOnlyAtLastGate(t *testing.T) {
	tr := newRepo(t)
	tr.src.tasks["9f31"].Reviews = []Review{{Gate: "design", By: ben}}
	r := tr.repo("Ada", ada)
	_, err := r.Record(ctx, Recording{ID: "9f31", Gate: "design", State: "complete"}, Options{})
	wantRefusal(t, err, "complete-early")
	_, err = r.Record(ctx, Recording{ID: "9f31", Gate: "release", State: "nominal"}, Options{})
	wantRefusal(t, err, "last-gate-not-complete")
	tr.commit(r, record(Recording{ID: "9f31", Gate: "release"}, Options{}))
	if got := tr.read(".tableaux/status/9f31.yaml"); got != "gate: release\nstate: complete\n" {
		t.Errorf("the file reads %q", got)
	}
}

// TestRecursiveHoldsOnlyGate is T10: record c07d --gate design writes the
// gate alone and drops three lines; with --state it is recursive-holds-gate.
func TestRecursiveHoldsOnlyGate(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo("Ada", ada)
	_, err := r.Record(ctx, Recording{ID: "c07d", Gate: "design", State: "nominal"}, Options{})
	wantRefusal(t, err, "recursive-holds-gate")
	_, res := tr.commit(r, record(Recording{ID: "c07d", Gate: "design"}, Options{}))
	if got := tr.read(".tableaux/status/c07d.yaml"); got != "gate: design\n" {
		t.Errorf("the file reads %q", got)
	}
	if got := tr.run("diff", "--numstat", res.Commit+"~1", res.Commit); got != "1\t4\t.tableaux/status/c07d.yaml" {
		t.Errorf("numstat %q", got)
	}

	// tablo gives a status before a recursive junction as the gate alone,
	// whatever the file holds: the record at that gate still drops the
	// other fields, and a file of the gate alone is no change.
	tr.write(".tableaux/status/c07d.yaml", "gate: design\nstate: nominal\nnote: Waits\n")
	tr.run("commit", "-q", "-am", "Hold a state before the recursive junction")
	tr.src.tasks["c07d"].Status = &Status{Gate: "design"}
	tr.commit(r, record(Recording{ID: "c07d", Gate: "design"}, Options{}))
	if got := tr.read(".tableaux/status/c07d.yaml"); got != "gate: design\n" {
		t.Errorf("the file reads %q", got)
	}
	_, err = r.Record(ctx, Recording{ID: "c07d", Gate: "design"}, Options{})
	wantRefusal(t, err, "no-change")
}

// TestPinAdvance is T11: with the submodule moved, --pin commits the gitlink
// beside the status, Plan.Pins names both hashes, and the submodule's
// checkout does not move; unmoved it is pin-unchanged; on 9f31 it is
// no-submodule.
func TestPinAdvance(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo("Ada", ada)
	old := tr.run("ls-tree", "--object-only", "HEAD", "sub")
	_, err := r.Record(ctx, Recording{ID: "c07d", Pin: true}, Options{})
	wantRefusal(t, err, "pin-unchanged")
	_, err = r.Record(ctx, Recording{ID: "9f31", Pin: true}, Options{})
	wantRefusal(t, err, "no-submodule")

	tr.runIn(tr.dir+"/sub", "checkout", "-q", "main")
	moved := tr.runIn(tr.dir+"/sub", "rev-parse", "HEAD")
	p, res := tr.commit(r, record(Recording{ID: "c07d", Gate: "design", Pin: true}, Options{}))
	if len(p.Pins) != 1 || p.Pins[0] != (PinChange{"sub", old, moved}) || len(p.Files) != 1 {
		t.Errorf("pins %+v, files %d", p.Pins, len(p.Files))
	}
	if got := tr.run("diff", "--name-only", res.Commit+"~1", res.Commit); got != ".tableaux/status/c07d.yaml\nsub" {
		t.Errorf("the commit changes %q", got)
	}
	if got := tr.run("ls-tree", "--object-only", "HEAD", "sub"); got != moved {
		t.Errorf("the gitlink is %s, want %s", got, moved)
	}
	if got := tr.runIn(tr.dir+"/sub", "rev-parse", "HEAD"); got != moved {
		t.Errorf("the checkout moved to %s", got)
	}

	// A pin alone takes its own subject.
	tr.runIn(tr.dir+"/sub", "commit", "-q", "--allow-empty", "-m", "Third")
	tr.src.tasks["c07d"].Status = &Status{Gate: "design"}
	p, err = r.Record(ctx, Recording{ID: "c07d", Pin: true}, Options{})
	if err != nil || p.Subject != "Advance the pin of sub for mast (c07d)" || len(p.Files) != 0 {
		t.Errorf("a pin alone: %v, %+v", err, p)
	}
}

// TestRecordResolution is T12: a table over the steps of "Resolving a
// record".
func TestRecordResolution(t *testing.T) {
	tr := newRepo(t)
	tr.src.tasks["9f31"].Reviews = []Review{{Gate: "design", By: ben}}
	r := tr.repo("Ada", ada)
	rows := []struct {
		name string
		rec  Recording
		code string // "" for a plan
		file string // the planned file when code is ""
	}{
		{"keep on the same gate", Recording{ID: "9f31", Note: "Bench test"}, "", "gate: function\nstate: nominal\nnote: Bench test\n"},
		{"unknown-gate", Recording{ID: "9f31", Gate: "nogate"}, "unknown-gate", ""},
		{"state-required", Recording{ID: "9f31", Gate: "design"}, "state-required", ""},
		{"state-forbidden at undefined", Recording{ID: "9f31", Gate: "undefined", State: "nominal"}, "state-forbidden", ""},
		{"state-forbidden elsewhere", Recording{ID: "9f31", State: "undefined"}, "state-forbidden", ""},
		{"unknown-state", Recording{ID: "9f31", State: "bogus"}, "unknown-state", ""},
		{"unknown-reason", Recording{ID: "9f31", Reason: "bogus"}, "unknown-reason", ""},
		{"not-leaf", Recording{ID: "07e0", Gate: "design"}, "not-leaf", ""},
		{"no-change", Recording{ID: "9f31", State: "nominal"}, "no-change", ""},
		{"bad-field: a note with a line break", Recording{ID: "9f31", Note: "one\ntwo"}, "bad-field", ""},
		{"bad-field: a blank note", Recording{ID: "9f31", Note: "  "}, "bad-field", ""},
		{"bad-field: reason and no reason", Recording{ID: "9f31", Reason: "blocked", NoReason: true}, "bad-field", ""},
		{"no-task", Recording{ID: "beef"}, "no-task", ""},
	}
	for _, row := range rows {
		p, err := r.Record(ctx, row.rec, Options{})
		if row.code != "" {
			if codes := refusalCodes(t, err); len(codes) == 0 || codes[0] != row.code {
				t.Errorf("%s: refusals %v, want %s", row.name, codes, row.code)
			}
			continue
		}
		if err != nil || len(p.Files) != 1 || string(p.Files[0].New) != row.file {
			t.Errorf("%s: %v, %+v", row.name, err, p)
		}
	}

	t.Run("drop reason and note on a new gate", func(t *testing.T) {
		tr.write(".tableaux/status/9f31.yaml", "gate: function\nstate: at_risk\nreason: blocked\nnote: Waits for the sensor\n")
		tr.run("commit", "-q", "-am", "Hold the barometer")
		tr.src.tasks["9f31"].Status = &Status{Gate: "function", State: "at_risk", Reason: "blocked", Note: "Waits for the sensor"}
		p, err := r.Record(ctx, Recording{ID: "9f31", Gate: "design", State: "nominal"}, Options{})
		if err != nil || string(p.Files[0].New) != "gate: design\nstate: nominal\n" {
			t.Errorf("%v, %+v", err, p)
		}
		p, err = r.Record(ctx, Recording{ID: "9f31", NoReason: true}, Options{})
		if err != nil || string(p.Files[0].New) != "gate: function\nstate: at_risk\nnote: Waits for the sensor\n" ||
			p.Subject != "Update the status of barometer (9f31)" {
			t.Errorf("--no-reason: %v, %+v", err, p)
		}
	})
	t.Run("status-form", func(t *testing.T) {
		tr.write(".tableaux/status/9f31.yaml", "gate: function\nstate: nominal\nowner: ada\n")
		tr.run("commit", "-q", "-am", "Break the file")
		tr.src.tasks["9f31"].Status = &Status{Gate: "function", State: "nominal"}
		_, err := r.Record(ctx, Recording{ID: "9f31", Note: "x"}, Options{})
		wantRefusal(t, err, "status-form")
	})
}

// TestReviewsOwed is T13: passing design on 9f31 as its contributor is
// unreviewed; with a review in Reviews it commits with no Reviewed:; defined
// never owes one.
func TestReviewsOwed(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo("Ada", ada)
	_, err := r.Record(ctx, Recording{ID: "9f31", Gate: "design", State: "nominal"}, Options{})
	if got := theRefusal(t, err, "unreviewed"); got.Gate != "design" || got.Rule != "S11" ||
		!strings.Contains(got.Text, ben) || !strings.Contains(got.Text, "--reason review") {
		t.Errorf("unreviewed: %+v", got)
	}
	tr.src.tasks["9f31"].Reviews = []Review{{Gate: "design", By: ben}}
	p, err := r.Record(ctx, Recording{ID: "9f31", Gate: "design", State: "nominal"}, Options{})
	if err != nil || len(p.Trailers) != 0 {
		t.Errorf("reviewed: %v, %+v", err, p)
	}

	task := tr.src.tasks["9f31"]
	task.Junctions[0].Reviewer = ben // defined
	task.Status = &Status{Gate: "undefined", State: "undefined"}
	p, err = r.Record(ctx, Recording{ID: "9f31", Gate: "defined", State: "nominal"}, Options{})
	if err != nil || len(p.Trailers) != 0 {
		t.Errorf("defined: %v, %+v", err, p)
	}
}

// TestHandOff is T14: --reason review on 9f31 writes the reason, keeps the
// gate and takes the hand-off subject; on a task with no reviewer ahead it
// is no-reviewer.
func TestHandOff(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo("Ada", ada)
	p, _ := tr.commit(r, record(Recording{ID: "9f31", Reason: "review", Note: "The design waits on the branch"}, Options{}))
	if p.Subject != "Hand barometer (9f31) to review" ||
		tr.read(".tableaux/status/9f31.yaml") != "gate: function\nstate: nominal\nreason: review\nnote: The design waits on the branch\n" {
		t.Errorf("subject %q, file %q", p.Subject, tr.read(".tableaux/status/9f31.yaml"))
	}
	tr.src.tasks["9f31"].Status = &Status{Gate: "design", State: "nominal"}
	_, err := r.Record(ctx, Recording{ID: "9f31", Reason: "review"}, Options{})
	wantRefusal(t, err, "no-reviewer")
}

// TestPlanString is T20: the self-reviewed record of b618 prints plan.txt,
// and a warning adds its line.
func TestPlanString(t *testing.T) {
	tr := newRepo(t)
	want, err := os.ReadFile("testdata/plan.txt")
	if err != nil {
		t.Fatal(err)
	}
	r := tr.repo(agentNm, agent)
	p, err := r.Record(ctx, b618Function, sonnet)
	if err != nil {
		t.Fatal(err)
	}
	if got, w := p.String(), strings.Replace(string(want), "<base>", short(p.Base), 1); got != w {
		t.Errorf("got\n%s\nwant\n%s", got, w)
	}
	opus := sonnet
	opus.Agent.Model = "claude-opus-5-5"
	p, err = r.Record(ctx, b618Function, opus)
	if err != nil {
		t.Fatal(err)
	}
	if got := p.String(); !strings.HasSuffix(got, "\nwarning: model-mismatch: the junction at function states the model claude-sonnet, and Model: names claude-opus-5-5\n") {
		t.Errorf("with a warning:\n%s", got)
	}
}

// TestApplyCommitsOnlyThePlan is T21: with another file staged, the commit
// holds the planned paths alone, its tree equals Plan.Tree, and the other
// file stays staged.
func TestApplyCommitsOnlyThePlan(t *testing.T) {
	tr := newRepo(t)
	tr.write("README.md", "# Weather station, staged\n")
	tr.run("add", "README.md")
	p, res := tr.commit(tr.repo(agentNm, agent), record(b618Function, sonnet))
	if got := tr.run("diff", "--name-only", res.Commit+"~1", res.Commit); got != ".tableaux/status/b618.yaml" {
		t.Errorf("the commit changes %q", got)
	}
	if got := tr.run("rev-parse", res.Commit+"^{tree}"); got != p.Tree {
		t.Errorf("tree %s, plan %s", got, p.Tree)
	}
	if got := tr.run("diff", "--cached", "--name-only"); got != "README.md" {
		t.Errorf("staged %q", got)
	}
}

// TestApplyRestoresOnFailure is T22: a failing pre-commit hook, and an empty
// identity under user.useConfigOnly, leave HEAD and the status file as they
// were, and a proposed file gone and unstaged.
func TestApplyRestoresOnFailure(t *testing.T) {
	cases := map[string]func(tr *testRepo) *Repo{
		"a failing hook": func(tr *testRepo) *Repo {
			tr.hook("pre-commit", "echo 'the hook refuses' >&2\nexit 1\n")
			return tr.repo(agentNm, agent)
		},
		"no identity": func(tr *testRepo) *Repo {
			tr.run("config", "user.useConfigOnly", "true")
			return &Repo{Dir: tr.dir, Source: tr.src}
		},
	}
	t.Run("a proposal", testProposeRestored)
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			tr := newRepo(t)
			p, err := tr.repo(agentNm, agent).Record(ctx, b618Function, sonnet)
			if err != nil {
				t.Fatal(err)
			}
			head := tr.head()
			_, err = setup(tr).Apply(ctx, p)
			var x *ExecError
			if !errors.As(err, &x) || x.Stderr == "" {
				t.Fatalf("error %v, want an ExecError with Git's text", err)
			}
			if tr.head() != head || tr.read(".tableaux/status/b618.yaml") != b618Status || tr.run("status", "--porcelain") != "" {
				t.Errorf("HEAD %s, file %q, status %q", tr.head(), tr.read(".tableaux/status/b618.yaml"), tr.run("status", "--porcelain"))
			}
		})
	}
}

// TestInvalidTreeRefused is T24: a Source whose second Validate adds an
// error is invalid; an error both runs share, and an added S11, do not
// refuse.
func TestInvalidTreeRefused(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo(agentNm, agent)
	shared := Diagnostic{Severity: "error", Code: "J8", Path: ".tableaux/tasks/c07d.yaml", Message: "url sub resolves to no project"}
	added := Diagnostic{Severity: "error", Code: "S6", Path: ".tableaux/status/b618.yaml", Message: "state names no state"}
	s11 := Diagnostic{Severity: "error", Code: "S11", Path: ".tableaux/status/b618.yaml", Message: "no review"}
	warning := Diagnostic{Severity: "warning", Code: "H3", Path: ".tableaux/status/b618.yaml", Message: "a warning"}
	trees := map[int]string{}
	tr.src.validate = func(call int, tree string) []Diagnostic {
		trees[call] = tree
		if call%2 == 1 {
			return []Diagnostic{shared}
		}
		if call == 2 {
			return []Diagnostic{shared, added}
		}
		return []Diagnostic{shared, s11, warning}
	}
	_, err := r.Record(ctx, b618Function, sonnet)
	if got := theRefusal(t, err, "invalid"); got.Rule != "S6" || got.Text != added.Path+": "+added.Message {
		t.Errorf("invalid: %+v", got)
	}
	if trees[1] != tr.run("rev-parse", "HEAD^{tree}") || trees[2] == trees[1] {
		t.Errorf("validated %v", trees)
	}
	if _, err := r.Record(ctx, b618Function, sonnet); err != nil {
		t.Errorf("a shared error, an S11 and a warning: %v", err)
	}
}

// TestWarnings is T25: not-contributor, model-mismatch and model-missing,
// each with the commit still made.
func TestWarnings(t *testing.T) {
	cases := []struct {
		code       string
		name, hand string
		rec        Recording
		o          Options
	}{
		{"not-contributor", "Ben", ben, Recording{ID: "9f31", Note: "Bench test"}, Options{}},
		{"model-mismatch", agentNm, agent, Recording{ID: "b618", Gate: "function", State: "nominal"}, Options{Agent: Agent{Model: "claude-opus-5-5"}}},
		{"model-missing", agentNm, agent, Recording{ID: "b618", Gate: "function", State: "nominal"}, Options{}},
	}
	for _, c := range cases {
		t.Run(c.code, func(t *testing.T) {
			tr := newRepo(t)
			p, res := tr.commit(tr.repo(c.name, c.hand), record(c.rec, c.o))
			if len(p.Warnings) != 1 || p.Warnings[0].Code != c.code || p.Warnings[0].Task != c.rec.ID || res.Commit != tr.head() {
				t.Errorf("warnings %+v, commit %s", p.Warnings, res.Commit)
			}
		})
	}
	t.Run("none at a junction the hand contributes to", func(t *testing.T) {
		tr := newRepo(t)
		if p, err := tr.repo(agentNm, agent).Record(ctx, b618Function, sonnet); err != nil || len(p.Warnings) != 0 {
			t.Errorf("%v, %+v", err, p)
		}
	})
}
