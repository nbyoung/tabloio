package action

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

// The files of a small project tablo's library reads: the root e4a1, the
// leaf b2c9 with an agent's design junction that a person reviews, and the
// leaf c3d7, whose implementation junction reads the submodule fw and whose
// release does not apply.
const (
	smallVersion = "tableaux: 0.3.1\ntrunk: main\n"
	smallGates   = `gates:
  - { key: undefined,      symbol: ❔, name: Undefined,      criteria: No one has started work on the definition }
  - { key: defined,        symbol: 📝, name: Defined,        criteria: "Title, description, assignee and references exist" }
  - { key: design,         symbol: 📐, name: Design,         criteria: A model and sufficient tests exist }
  - { key: implementation, symbol: 🧱, name: Implementation, criteria: Artifacts exist }
  - { key: release,        symbol: 🚀, name: Release,        criteria: All variants documented and approved }
states:
  - { key: undefined, symbol: ⚪, severity: 0, synopsis: The work has not yet been defined }
  - { key: nominal,   symbol: 🟢, severity: 1, synopsis: The work is proceeding as expected }
  - { key: complete,  symbol: ✅, severity: 0, synopsis: All deliverables satisfy their requirements }
reasons:
  - { key: blocked, symbol: ⛔, synopsis: An external resource is unavailable }
  - { key: review,  symbol: 👓, synopsis: The work waits for its reviewer }
`
	smallRoot = `title: Small
description: The root of a small project.
assignee: olive@example.org
references:
  - { url: README.md }
`
	smallLeaf = `title: Leaf
description: A leaf with an agent's design.
assignee: pat@example.org
references:
  - { url: README.md }
junctions:
  design: { contributor: bot@example.org, model: claude-fable, reviewer: rev@example.org }
parent: { id: "e4a1", order: 1 }
`
	smallSecond = `title: Second leaf
description: A leaf whose implementation lives in a submodule.
assignee: ben@example.org
junctions:
  implementation: { subproject: { url: fw } }
  release:        { applies: false }
parent: { id: "e4a1", order: 2 }
`
)

// smallProject builds the small project at HEAD, with b2c9 at defined.
func smallProject(t *testing.T) *testRepo {
	t.Helper()
	tr := &testRepo{t: t, dir: t.TempDir()}
	isolate(t)
	tr.run("init", "-q", "-b", "main")
	for path, content := range map[string]string{
		".tableaux/version.yaml": smallVersion, ".tableaux/gates.yaml": smallGates,
		".tableaux/tasks/e4a1.yaml": smallRoot, ".tableaux/tasks/b2c9.yaml": smallLeaf,
		".tableaux/tasks/c3d7.yaml": smallSecond, ".tableaux/status/b2c9.yaml": "gate: defined\nstate: nominal\n",
		"README.md": "# Small\n",
	} {
		tr.write(path, content)
	}
	fw := filepath.Join(tr.dir, "fw")
	tr.runIn(fw, "init", "-q", "-b", "main")
	for path, content := range map[string]string{
		".tableaux/version.yaml": smallVersion, ".tableaux/gates.yaml": smallGates,
		".tableaux/tasks/f1a0.yaml": "title: Firmware\ndescription: The firmware.\nassignee: ben@example.org\n",
	} {
		tr.write("fw/"+path, content)
	}
	tr.runIn(fw, "add", ".")
	tr.runIn(fw, "commit", "-q", "-m", "Plan the firmware")
	tr.run("add", ".")
	tr.run("commit", "-q", "-m", "Plan the small project")
	return tr
}

// TestTabloArguments is T27, read as decision 2 implies: Tablo's calls give
// the facts of A2 to A5 through tablo's library on a repository under
// t.TempDir, and a failure to read becomes an *ExecError.
func TestTabloArguments(t *testing.T) {
	tr := smallProject(t)
	src := &Tablo{Dir: tr.dir}

	t.Run("A3 the gates view", func(t *testing.T) {
		p, err := src.Project(ctx)
		want := Project{
			Gates:   []string{"undefined", "defined", "design", "implementation", "release"},
			States:  []string{"undefined", "nominal", "complete"},
			Reasons: []string{"blocked", "review"},
			Trunk:   "main",
		}
		if err != nil || !reflect.DeepEqual(p, want) {
			t.Errorf("got %+v, %v; want %+v", p, err, want)
		}
	})
	t.Run("A2 tablo.Trailer.Line", func(t *testing.T) {
		for _, c := range [][4]string{
			{"reviewed", "b2c9", "design", "Reviewed: b2c9 design"},
			{"reaffirmed", "b2c9", "", "Reaffirmed: b2c9"},
			{"authorised", "c3d7", "", "Authorised: c3d7"},
		} {
			if got, err := src.Trailer(ctx, c[0], c[1], c[2]); err != nil || got != c[3] {
				t.Errorf("%v: got %q, %v", c, got, err)
			}
		}
		var x *ExecError
		if _, err := src.Trailer(ctx, "reviewed", "b2c9", ""); !errors.As(err, &x) || x.Tool != "tablo" {
			t.Errorf("a reviewed trailer with no gate: %v, want an ExecError", err)
		}
	})
	t.Run("A4 the task view at provenance", func(t *testing.T) {
		got, err := src.Task(ctx, "b2c9")
		if err != nil {
			t.Fatal(err)
		}
		if got.Title != "Leaf" || got.Assignee != "pat@example.org" || len(got.Children) != 0 ||
			!reflect.DeepEqual(got.Authorities, []string{"olive@example.org"}) ||
			!reflect.DeepEqual(got.Status, &Status{Gate: "defined", State: "nominal"}) || len(got.Reviews) != 0 {
			t.Errorf("b2c9: %+v", got)
		}
		if j := got.junction("design"); j == nil || *j != (Junction{Gate: "design", Kind: plain, Contributor: "bot@example.org", Model: "claude-fable", Reviewer: "rev@example.org"}) {
			t.Errorf("b2c9's design junction: %+v", j)
		}
		root, err := src.Task(ctx, "e4a1")
		if err != nil || !reflect.DeepEqual(root.Children, []Child{{ID: "b2c9", Order: 1}, {ID: "c3d7", Order: 2}}) || root.Status != nil {
			t.Errorf("e4a1: %+v, %v", root, err)
		}
		second, err := src.Task(ctx, "c3d7")
		if err != nil {
			t.Fatal(err)
		}
		if j := second.junction("implementation"); j == nil || *j != (Junction{Gate: "implementation", Kind: recursive, URL: "fw", Submodule: true}) {
			t.Errorf("c3d7's implementation junction: %+v", j)
		}
		if j := second.junction("release"); j == nil || j.Kind != notApplicable {
			t.Errorf("c3d7's release junction: %+v", j)
		}
	})
	t.Run("the reviews from the history", func(t *testing.T) {
		tr.runAs(hand("Rev", "rev@example.org"), "commit", "-q", "--allow-empty", "-m", "Accept the design", "-m", "Reviewed: b2c9 design")
		tr.runAs(hand("Pat", "pat@example.org"), "commit", "-q", "--allow-empty", "-m", "Review by the wrong hand", "-m", "Reviewed: b2c9 design")
		got, err := (&Tablo{Dir: tr.dir}).Task(ctx, "b2c9")
		if want := []Review{{Gate: "design", By: "rev@example.org"}}; err != nil || !reflect.DeepEqual(got.Reviews, want) {
			t.Errorf("reviews %+v, %v; want %+v", got.Reviews, err, want)
		}
	})
	t.Run("A5 the snapshot of a bare tree", func(t *testing.T) {
		tree := tr.run("rev-parse", "HEAD^{tree}")
		diags, err := src.Validate(ctx, tree)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range diags {
			if d.Severity == "error" {
				t.Errorf("HEAD's tree: %+v", d)
			}
		}
		tr.write(".tableaux/status/b2c9.yaml", "gate: nogate\nstate: nominal\n")
		tr.run("add", ".tableaux")
		bad := tr.run("write-tree")
		tr.run("reset", "-q", "--hard")
		diags, err = src.Validate(ctx, bad)
		found := false
		for _, d := range diags {
			found = found || d.Severity == "error" && d.Code == "S5" && d.Path == ".tableaux/status/b2c9.yaml" && d.Message != ""
		}
		if err != nil || !found {
			t.Errorf("a tree with an unknown gate: %+v, %v; want an S5 error", diags, err)
		}
	})
	t.Run("a failure to read", func(t *testing.T) {
		var x *ExecError
		if _, err := (&Tablo{Dir: t.TempDir()}).Project(ctx); !errors.As(err, &x) || x.Tool != "tablo" {
			t.Errorf("no repository: %v, want an ExecError", err)
		}
		if _, err := src.Validate(ctx, "0123456789abcdef0123456789abcdef01234567"); !errors.As(err, &x) {
			t.Errorf("no such tree: %v, want an ExecError", err)
		}
	})
}
