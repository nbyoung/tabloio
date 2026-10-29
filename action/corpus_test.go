package action

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nbyoung/tablo"
)

// corpusRepo copies a built entry of the conformance corpus, with its .git
// and its submodule, into a temporary directory, and clears the host's Git
// identity and configuration. It skips without TABLEAUX_CORPUS. The built
// corpus is shared and read-only, so no test runs anything in it.
func corpusRepo(t *testing.T, name string) *testRepo {
	t.Helper()
	corpus := os.Getenv("TABLEAUX_CORPUS")
	if corpus == "" {
		t.Skip("TABLEAUX_CORPUS names no built corpus")
	}
	for _, tool := range []string{"git", "cp"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	src := filepath.Join(corpus, name)
	if _, err := os.Stat(src); err != nil {
		t.Skipf("the corpus holds no %s: %v", name, err)
	}
	isolate(t)
	dir := filepath.Join(t.TempDir(), name)
	if out, err := exec.Command("cp", "-a", src, dir).CombinedOutput(); err != nil {
		t.Fatalf("copy the corpus entry: %v\n%s", err, out)
	}
	return &testRepo{t: t, dir: dir}
}

// as returns a Repo of the checkout whose Source is tablo's library, and
// whose commits are the hand's.
func (tr *testRepo) as(name, email string) *Repo {
	return &Repo{Dir: tr.dir, Env: hand(name, email)}
}

// treeErrors returns the error diagnostics tablo's validator gives the tree
// of HEAD, each as code, path and message.
func (tr *testRepo) treeErrors() map[string]bool {
	tr.t.Helper()
	diags, err := (&Tablo{Dir: tr.dir}).Validate(ctx, tr.run("rev-parse", "HEAD^{tree}"))
	if err != nil {
		tr.t.Fatal(err)
	}
	out := map[string]bool{}
	for _, d := range diags {
		if d.Severity == "error" {
			out[d.Code+" "+d.Path+" "+d.Message] = true
		}
	}
	return out
}

// events returns the kinds of the events tablo's history view gives a task
// at a commit.
func (tr *testRepo) events(id, commit string) []string {
	tr.t.Helper()
	var env struct {
		Data struct {
			Lines []struct {
				Events []struct{ Kind, Commit string }
			}
		}
	}
	if err := (&Tablo{Dir: tr.dir}).view(ctx, &env, tablo.History, tablo.Params{Task: id, Level: tablo.LevelDetail}); err != nil {
		tr.t.Fatal(err)
	}
	var kinds []string
	for _, l := range env.Data.Lines {
		for _, e := range l.Events {
			if e.Commit == commit {
				kinds = append(kinds, e.Kind)
			}
		}
	}
	return kinds
}

// applied makes a plan of tablo's facts and checks the commit against
// tablo: the validator reports no error the tree of the parent lacks, and
// the history of the task shows an event of the kind at the commit.
func (tr *testRepo) applied(r *Repo, plan func(*Repo) (*Plan, error), id, kind string) (*Plan, Result) {
	tr.t.Helper()
	before := tr.treeErrors()
	p, res := tr.commit(r, plan)
	for e := range tr.treeErrors() {
		if !before[e] {
			tr.t.Errorf("%s: the commit adds the error %s", res.Commit[:7], e)
		}
	}
	found := false
	for _, k := range tr.events(id, res.Commit) {
		found = found || k == kind
	}
	if !found {
		tr.t.Errorf("%s: the history of %s shows no %s event: %v", res.Commit[:7], id, kind, tr.events(id, res.Commit))
	}
	return p, res
}

// TestCorpusEntries is T29: on the corpus entries the design names, each
// refusal fires from tablo's own facts, and after each commit tablo's
// validator reports no new error and tablo's history shows the event. On
// weather-station, every value of scalars.tsv that a note or a reference's
// text may hold reads back through tablo's YAML reader as written (the
// scalar rule, which the unit gate proves against the fixtures alone).
func TestCorpusEntries(t *testing.T) {
	t.Run("weather-station", func(t *testing.T) {
		tr := corpusRepo(t, "weather-station")
		ada := tr.as("Ada Byron", "ada@example.org")
		b, err := os.ReadFile("testdata/scalars.tsv")
		if err != nil {
			t.Fatal(err)
		}
		var values []string
		for _, row := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")[1:] {
			values = append(values, strings.SplitN(row, "\t", 2)[0])
		}

		// The block form: each value as the note of 9f31's status.
		notes := 0
		for _, v := range values {
			rec := Recording{ID: "9f31", Note: v}
			if _, err := ada.Record(ctx, rec, Options{}); err != nil {
				if codes := refusalCodes(t, err); codes[0] != "bad-field" && codes[0] != "no-change" {
					t.Errorf("note %q: %v", v, codes)
				}
				continue
			}
			tr.applied(ada, record(rec, Options{}), "9f31", "status")
			task, err := (&Tablo{Dir: tr.dir}).Task(ctx, "9f31")
			if err != nil {
				t.Fatal(err)
			}
			if task.Status == nil || task.Status.Note != v {
				t.Errorf("note %q reads back as %+v", v, task.Status)
			}
			notes++
		}
		if notes < 30 {
			t.Errorf("only %d of the %d values land as a note", notes, len(values))
		}

		// The flow form: each value as the text of a reference of one
		// proposal under the root.
		refs := []Reference{}
		for _, v := range values {
			if strings.TrimSpace(v) != "" && !strings.ContainsAny(v, "\n\r") {
				refs = append(refs, Reference{URL: "docs/overview.md", Text: v})
			}
		}
		probe := Proposal{Title: "Probe", Description: "A task whose references hold every value.", Parent: "a1c0", References: refs}
		p, _ := tr.applied(ada, propose(probe, Options{}), "a1c0", "task")
		var env struct {
			Data struct{ References []struct{ URL, Text string } }
		}
		if err := (&Tablo{Dir: tr.dir}).view(ctx, &env, tablo.Task, tablo.Params{Task: p.NewID, Level: tablo.LevelDetail}); err != nil {
			t.Fatal(err)
		}
		if len(env.Data.References) != len(refs) {
			t.Fatalf("%d references read back, want %d", len(env.Data.References), len(refs))
		}
		for i, r := range refs {
			if got := env.Data.References[i]; got.URL != r.URL || got.Text != r.Text {
				t.Errorf("reference %q reads back as %+v", r.Text, got)
			}
		}

		// A reaffirmation.
		tr.applied(ada, func(r *Repo) (*Plan, error) { return r.Reaffirm(ctx, []string{"9f31"}, Options{}) }, "9f31", "reaffirmed")
	})

	t.Run("junction-kinds", func(t *testing.T) {
		tr := corpusRepo(t, "junction-kinds")
		ben := tr.as("Ben Okafor", "ben@example.org")
		_, err := ben.Record(ctx, Recording{ID: "a600", Gate: "mockup"}, Options{})
		if r := theRefusal(t, err, "gate-not-applicable"); r.Rule != "S7" {
			t.Errorf("record a600 at mockup: %+v", r)
		}
		_, err = ben.Review(ctx, []string{"a600"}, "mockup", Options{})
		wantRefusal(t, err, "gate-not-applicable")
		_, err = ben.Record(ctx, Recording{ID: "a400", State: "nominal"}, Options{})
		if r := theRefusal(t, err, "recursive-holds-gate"); r.Rule != "S9" {
			t.Errorf("record a400 with a state: %+v", r)
		}
		_, err = ben.Review(ctx, []string{"a400"}, "implementation", Options{})
		wantRefusal(t, err, "recursive")
		_, err = ben.Review(ctx, []string{"a500"}, "unit", Options{})
		wantRefusal(t, err, "recursive")
		tr.applied(ben, record(Recording{ID: "a110", Gate: "implementation", State: "nominal"}, Options{}), "a110", "status")
	})

	t.Run("submodule-subproject", func(t *testing.T) {
		tr := corpusRepo(t, "submodule-subproject")
		ben := tr.as("Ben Okafor", "ben@example.org")
		_, err := ben.Record(ctx, Recording{ID: "c100", Pin: true}, Options{})
		wantRefusal(t, err, "pin-unchanged")
		_, err = ben.Record(ctx, Recording{ID: "c100", State: "nominal"}, Options{})
		wantRefusal(t, err, "recursive-holds-gate")
		_, err = ben.Record(ctx, Recording{ID: "c000", Pin: true}, Options{})
		wantRefusal(t, err, "not-leaf")
		tr.runIn(filepath.Join(tr.dir, "sub"), "checkout", "-q", "7fb517e")
		p, _ := tr.applied(ben, record(Recording{ID: "c100", Pin: true}, Options{}), "c100", "pin")
		if len(p.Pins) != 1 || p.Pins[0].Path != "sub" || !strings.HasPrefix(p.Pins[0].New, "7fb517e") {
			t.Errorf("pins %+v", p.Pins)
		}
	})

	t.Run("status-gate-not-applicable", func(t *testing.T) {
		tr := corpusRepo(t, "status-gate-not-applicable")
		pat := tr.as("Pat Singh", "pat@example.org")
		_, err := pat.Record(ctx, Recording{ID: "b2c9", Gate: "design", State: "nominal"}, Options{})
		wantRefusal(t, err, "gate-not-applicable")
		_, err = pat.Review(ctx, []string{"b2c9"}, "design", Options{})
		wantRefusal(t, err, "gate-not-applicable")
		tr.applied(pat, record(Recording{ID: "b2c9", Gate: "release"}, Options{}), "b2c9", "status")
		if got := tr.read(".tableaux/status/b2c9.yaml"); got != "gate: release\nstate: complete\n" {
			t.Errorf("the file reads %q", got)
		}
	})

	t.Run("status-complete-early", func(t *testing.T) {
		tr := corpusRepo(t, "status-complete-early")
		pat := tr.as("Pat Singh", "pat@example.org")
		_, err := pat.Record(ctx, Recording{ID: "b2c9", State: "complete"}, Options{})
		if r := theRefusal(t, err, "complete-early"); r.Rule != "S12" {
			t.Errorf("complete at defined: %+v", r)
		}
		tr.applied(pat, record(Recording{ID: "b2c9", State: "nominal"}, Options{}), "b2c9", "status")
	})

	t.Run("status-state-with-recursive", func(t *testing.T) {
		tr := corpusRepo(t, "status-state-with-recursive")
		pat := tr.as("Pat Singh", "pat@example.org")
		_, err := pat.Record(ctx, Recording{ID: "b2c9", State: "nominal", Note: "Waits"}, Options{})
		if r := theRefusal(t, err, "recursive-holds-gate"); r.Rule != "S9" {
			t.Errorf("a state before a recursive junction: %+v", r)
		}
		tr.applied(pat, record(Recording{ID: "b2c9", Gate: "defined"}, Options{}), "b2c9", "status")
		if got := tr.read(".tableaux/status/b2c9.yaml"); got != "gate: defined\n" {
			t.Errorf("the file reads %q", got)
		}
	})

	t.Run("status-unreviewed-gate", func(t *testing.T) {
		tr := corpusRepo(t, "status-unreviewed-gate")
		pat, olive := tr.as("Pat Singh", "pat@example.org"), tr.as("Olive Marsh", "olive@example.org")
		tr.applied(pat, record(Recording{ID: "b2c9", Gate: "defined", State: "nominal"}, Options{}), "b2c9", "status")
		design := Recording{ID: "b2c9", Gate: "design", State: "nominal"}
		_, err := pat.Record(ctx, design, Options{})
		if r := theRefusal(t, err, "unreviewed"); r.Rule != "S11" || r.Gate != "design" || !strings.Contains(r.Text, "olive@example.org") {
			t.Errorf("design unreviewed: %+v", r)
		}
		tr.applied(olive, func(r *Repo) (*Plan, error) { return r.Review(ctx, []string{"b2c9"}, "design", Options{}) }, "b2c9", "reviewed")
		p, _ := tr.applied(pat, record(design, Options{}), "b2c9", "status")
		if len(p.Trailers) != 0 {
			t.Errorf("a reviewed gate owes a trailer: %v", p.Trailers)
		}
	})

	t.Run("review-by-non-reviewer", func(t *testing.T) {
		tr := corpusRepo(t, "review-by-non-reviewer")
		review := func(r *Repo) (*Plan, error) { return r.Review(ctx, []string{"b2c9"}, "design", Options{}) }
		_, err := review(tr.as("Pat Singh", "pat@example.org"))
		if r := theRefusal(t, err, "wrong-hand"); r.Rule != "H2" || !strings.Contains(r.Text, "olive@example.org") {
			t.Errorf("a review by pat: %+v", r)
		}
		tr.applied(tr.as("Olive Marsh", "olive@example.org"), review, "b2c9", "reviewed")
	})

	t.Run("junction-review-waived", func(t *testing.T) {
		tr := corpusRepo(t, "junction-review-waived")
		tr.run("reset", "-q", "--hard", tr.run("rev-list", "--max-parents=0", "HEAD")) // W1, "Plan the project"
		bot := tr.as("Corpus Agent", "bot@example.org")
		agent := Options{Agent: Agent{Model: "claude-fable", CoAuthor: "Corpus Agent <bot@example.org>"}}
		p, err := bot.Record(ctx, Recording{ID: "b2c9", Gate: "design", State: "nominal"}, agent)
		if err != nil {
			t.Fatalf("a waived design: %v", err)
		}
		for _, tl := range p.Trailers {
			if tl.Key == "Reviewed" {
				t.Errorf("a waived design owes no review: %v", p.Trailers)
			}
		}
		_, err = bot.Record(ctx, Recording{ID: "c3d8", Gate: "design", State: "nominal"}, agent)
		if r := theRefusal(t, err, "unreviewed"); r.Rule != "S11" || !strings.Contains(r.Text, "pat@example.org") {
			t.Errorf("a stated reviewer survives the waiver: %+v", r)
		}
	})

	t.Run("review-nearest-person", func(t *testing.T) {
		tr := corpusRepo(t, "review-nearest-person")
		bot, olive := tr.as("Corpus Agent", "bot@example.org"), tr.as("Olive Marsh", "olive@example.org")
		agent := Options{Agent: Agent{Model: "claude-fable", CoAuthor: "Corpus Agent <bot@example.org>"}}
		review := func(r *Repo, id, gate string) error {
			_, err := r.Review(ctx, []string{id}, gate, Options{})
			return err
		}
		if r := theRefusal(t, review(bot, "b2c9", "design"), "wrong-hand"); !strings.Contains(r.Text, "olive@example.org") {
			t.Errorf("b2c9 at design: %+v", r)
		}
		_, err := bot.Record(ctx, Recording{ID: "b2c9", Gate: "release"}, agent)
		if r := theRefusal(t, err, "unreviewed"); r.Gate != "release" || !strings.Contains(r.Text, "olive@example.org") {
			t.Errorf("b2c9 at release: %+v", r)
		}
		if r := theRefusal(t, review(olive, "d3a6", "design"), "wrong-hand"); !strings.Contains(r.Text, "pat@example.org") {
			t.Errorf("d3a6 at design: %+v", r)
		}
	})
}
