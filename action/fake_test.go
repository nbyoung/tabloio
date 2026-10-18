package action

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The hands the tests commit as.
const (
	agent    = "noreply@anthropic.com"
	agentNm  = "Claude Sonnet 5.5"
	owner    = "nbyoung@nbyoung.com"
	ada      = "ada@example.org"
	ben      = "ben@example.org"
	coAuthor = agentNm + " <" + agent + ">"
)

// The status files the repository builder commits.
const (
	b618Status = "gate: defined\nstate: nominal\nnote: The functional prototype is next\n"
	f931Status = "gate: function\nstate: nominal\n"
	c07dStatus = "gate: function\nstate: stalled\nreason: blocked\nnote: Barometer ICs on 14-week backorder\n"
)

// fakeSource answers as tablo would for a small project: the gates of this
// repository's gates.yaml; the root 40e8; its child 07e0; and under 07e0 the
// agent-assigned leaf b618, which reviews itself at function, the leaf 9f31,
// whose design junction a person reviews, and the leaf c07d, whose
// implementation junction is a submodule and whose mockup does not apply.
type fakeSource struct {
	project  Project
	tasks    map[string]*Task
	trailer  func(kind, id, gate string) string // nil composes the line as tablo does
	validate func(call int, tree string) []Diagnostic
	calls    map[string]int // by method
	fail     error          // every call fails with it
}

func newFake() *fakeSource {
	gates := []string{"undefined", "defined", "mockup", "function", "design", "implementation", "unit", "integrate", "validate", "release"}
	plainAll := func(contributor, model, reviewer string) []Junction {
		var js []Junction
		for _, g := range gates[1:] {
			js = append(js, Junction{Gate: g, Kind: plain, Contributor: contributor, Model: model, Reviewer: reviewer})
		}
		return js
	}
	set := func(js []Junction, j Junction) {
		for i := range js {
			if js[i].Gate == j.Gate {
				js[i] = j
			}
		}
	}
	b618 := plainAll(agent, "claude-sonnet", "")
	set(b618, Junction{Gate: "defined", Kind: plain, Contributor: agent, Model: "claude-haiku"})
	set(b618, Junction{Gate: "mockup", Kind: notApplicable})
	set(b618, Junction{Gate: "design", Kind: plain, Contributor: agent, Model: "claude-opus", Reviewer: owner})
	set(b618, Junction{Gate: "validate", Kind: plain, Contributor: agent, Model: "claude-sonnet", Reviewer: owner})
	set(b618, Junction{Gate: "release", Kind: plain, Contributor: owner})
	f931 := plainAll(ada, "", "")
	set(f931, Junction{Gate: "design", Kind: plain, Contributor: ada, Reviewer: ben})
	c07d := plainAll(ada, "", "")
	set(c07d, Junction{Gate: "mockup", Kind: notApplicable})
	set(c07d, Junction{Gate: "implementation", Kind: recursive, URL: "sub", Submodule: true})
	authorities := []string{ben, owner}
	return &fakeSource{
		project: Project{
			Gates:   gates,
			States:  []string{"undefined", "nominal", "at_risk", "stalled", "complete"},
			Reasons: []string{"overloaded", "blocked", "review"},
			Trunk:   "main",
		},
		tasks: map[string]*Task{
			"40e8": {Title: "Weather station", Assignee: owner, Children: []Child{{ID: "07e0", Order: 1}},
				Authorities: []string{owner}, Junctions: plainAll(owner, "", "")},
			"07e0": {Title: "Sensors", Assignee: ben, Children: []Child{{ID: "c07d", Order: 1}, {ID: "9f31", Order: 2}, {ID: "b618"}},
				Authorities: []string{owner}, Junctions: plainAll(ben, "", "")},
			"b618": {Title: "Write commands", Assignee: agent, Authorities: authorities, Junctions: b618,
				Status: &Status{Gate: "defined", State: "nominal", Note: "The functional prototype is next"}},
			"9f31": {Title: "Barometer", Assignee: ada, Authorities: authorities, Junctions: f931,
				Status: &Status{Gate: "function", State: "nominal"}},
			"c07d": {Title: "Mast", Assignee: ada, Authorities: authorities, Junctions: c07d,
				Status: &Status{Gate: "function", State: "stalled", Reason: "blocked", Note: "Barometer ICs on 14-week backorder"}},
		},
		calls: map[string]int{},
	}
}

func (f *fakeSource) Project(context.Context) (Project, error) {
	f.calls["Project"]++
	return f.project, f.fail
}

func (f *fakeSource) Task(_ context.Context, id string) (Task, error) {
	f.calls["Task"]++
	if f.fail != nil {
		return Task{}, f.fail
	}
	t, ok := f.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("the fake has no task %s", id)
	}
	out := *t
	out.ID = id
	return out, nil
}

func (f *fakeSource) Trailer(_ context.Context, kind, id, gate string) (string, error) {
	f.calls["Trailer"]++
	if f.trailer != nil {
		return f.trailer(kind, id, gate), nil
	}
	switch kind {
	case "reviewed":
		return "Reviewed: " + id + " " + gate, nil
	case "reaffirmed":
		return "Reaffirmed: " + id, nil
	}
	return "Authorised: " + id, nil
}

func (f *fakeSource) Validate(_ context.Context, tree string) ([]Diagnostic, error) {
	f.calls["Validate"]++
	if f.validate == nil {
		return nil, f.fail
	}
	return f.validate(f.calls["Validate"], tree), f.fail
}

// isolate clears the GIT_* variables of the process for the test, and points
// HOME, the configuration and TMPDIR into a temporary directory, so that no
// identity or setting of the host reaches Git.
func isolate(t *testing.T) {
	t.Helper()
	for _, kv := range os.Environ() {
		if k, _, _ := strings.Cut(kv, "="); strings.HasPrefix(k, "GIT_") {
			t.Setenv(k, "")
			_ = os.Unsetenv(k)
		}
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("TMPDIR", t.TempDir())
}

// hand returns the environment of an identity, author and committer alike.
func hand(name, email string) []string {
	return []string{"GIT_AUTHOR_NAME=" + name, "GIT_AUTHOR_EMAIL=" + email, "GIT_COMMITTER_NAME=" + name, "GIT_COMMITTER_EMAIL=" + email}
}

// testRepo is a checkout under t.TempDir and the fake that answers for it.
type testRepo struct {
	t   *testing.T
	dir string
	src *fakeSource
}

// newRepo builds a repository whose HEAD holds the fake's project: the task
// files, the three status files and, at sub, a submodule with two commits
// whose gitlink names the first.
func newRepo(t *testing.T) *testRepo {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH")
	}
	isolate(t)
	tr := &testRepo{t: t, dir: t.TempDir(), src: newFake()}
	tr.run("init", "-q", "-b", "main")
	tr.write(".tableaux/gates.yaml", "gates: []\n")
	for id, task := range tr.src.tasks {
		tr.write(".tableaux/tasks/"+id+".yaml", "title: "+task.Title+"\n")
	}
	tr.write(".tableaux/status/b618.yaml", b618Status)
	tr.write(".tableaux/status/9f31.yaml", f931Status)
	tr.write(".tableaux/status/c07d.yaml", c07dStatus)
	tr.write("README.md", "# Weather station\n")
	sub := filepath.Join(tr.dir, "sub")
	tr.runIn(sub, "init", "-q", "-b", "main")
	for _, msg := range []string{"First", "Second"} {
		tr.runIn(sub, "commit", "-q", "--allow-empty", "-m", msg)
	}
	first := tr.runIn(sub, "rev-parse", "HEAD~1")
	tr.runIn(sub, "checkout", "-q", first)
	tr.run("add", ".")
	tr.run("commit", "-q", "-m", "Start the weather station")
	return tr
}

// repo returns a Repo over the checkout that commits as the hand.
func (tr *testRepo) repo(name, email string) *Repo {
	return &Repo{Dir: tr.dir, Source: tr.src, Env: hand(name, email)}
}

// run runs git in the checkout as the builder and returns its output
// without the final newline.
func (tr *testRepo) run(args ...string) string { return tr.runIn(tr.dir, args...) }

func (tr *testRepo) runIn(dir string, args ...string) string {
	tr.t.Helper()
	return tr.runEnv(dir, hand("Builder", "builder@example.org"), args...)
}

// runAs runs git in the checkout as another hand.
func (tr *testRepo) runAs(env []string, args ...string) string {
	tr.t.Helper()
	return tr.runEnv(tr.dir, env, args...)
}

func (tr *testRepo) runEnv(dir string, env []string, args ...string) string {
	tr.t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		tr.t.Fatal(err)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		tr.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimRight(string(out), "\n")
}

// write writes a file of the checkout.
func (tr *testRepo) write(path, content string) {
	tr.t.Helper()
	full := filepath.Join(tr.dir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		tr.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		tr.t.Fatal(err)
	}
}

// read returns a file of the checkout, or "<absent>".
func (tr *testRepo) read(path string) string {
	b, err := os.ReadFile(filepath.Join(tr.dir, filepath.FromSlash(path)))
	if err != nil {
		return "<absent>"
	}
	return string(b)
}

// hook installs an executable hook script.
func (tr *testRepo) hook(name, script string) {
	tr.t.Helper()
	path := filepath.Join(tr.dir, ".git", "hooks", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		tr.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		tr.t.Fatal(err)
	}
}

// head returns the full hash of HEAD.
func (tr *testRepo) head() string { return tr.run("rev-parse", "HEAD") }

// commit plans with plan and applies the plan as the same Repo; it fails
// the test on an error of either.
func (tr *testRepo) commit(r *Repo, plan func(*Repo) (*Plan, error)) (*Plan, Result) {
	tr.t.Helper()
	p, err := plan(r)
	if err != nil {
		tr.t.Fatalf("plan: %v", err)
	}
	res, err := r.Apply(context.Background(), p)
	if err != nil {
		tr.t.Fatalf("apply: %v", err)
	}
	return p, res
}

// refusalCodes returns the codes of a plan function's error, or fails the
// test when the error is no Refusals.
func refusalCodes(t *testing.T, err error) []string {
	t.Helper()
	var rs Refusals
	if !errors.As(err, &rs) {
		t.Fatalf("error %v (%T), want Refusals", err, err)
	}
	var codes []string
	for _, r := range rs {
		codes = append(codes, r.Code)
	}
	return codes
}

// wantRefusal checks that err is Refusals holding exactly one refusal of the
// code.
func wantRefusal(t *testing.T, err error, code string) {
	t.Helper()
	_ = theRefusal(t, err, code)
}

// theRefusal checks as wantRefusal does and returns the refusal.
func theRefusal(t *testing.T, err error, code string) Refusal {
	t.Helper()
	codes := refusalCodes(t, err)
	if len(codes) != 1 || codes[0] != code {
		t.Fatalf("refusals %v, want [%s]: %v", codes, code, err)
	}
	var rs Refusals
	errors.As(err, &rs)
	return rs[0]
}

var ctx = context.Background()
