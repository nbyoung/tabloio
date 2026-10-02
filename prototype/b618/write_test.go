package b618

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testEnv = []string{
	"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
	"GIT_AUTHOR_NAME=Claude Sonnet 5.5", "GIT_AUTHOR_EMAIL=noreply@anthropic.com",
	"GIT_COMMITTER_NAME=Claude Sonnet 5.5", "GIT_COMMITTER_EMAIL=noreply@anthropic.com",
}

var agent = Agent{Model: "claude-sonnet-5-5", Name: "Claude Sonnet 5.5"}

const gatesYAML = `gates:
  - { key: undefined, symbol: ❔, name: Undefined }
  - { key: defined,   symbol: 📝, name: Defined }
  - { key: function,  symbol: ⚙️, name: Functional prototype }
`

// statusYAML has a comment, a blank line, odd spacing and a folded note: the
// parts a re-dump would reformat.
const statusYAML = `# kept verbatim
gate: defined

state:   nominal
note: >
  The prototype is
  next
`

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), testEnv...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixture builds a throwaway repository with a root task 07e0 and leaves.
func fixture(t *testing.T) string {
	dir := t.TempDir()
	run(t, dir, "init", "-q", "-b", "main")
	write(t, dir, ".tableaux/gates.yaml", gatesYAML)
	write(t, dir, ".tableaux/tasks/07e0.yaml", "title: Root\ndescription: r\nassignee: a@example.org\n")
	write(t, dir, ".tableaux/tasks/b618.yaml", "title: Leaf\ndescription: l\nassignee: a@example.org\nparent: { id: \"07e0\", order: 4 }\n")
	write(t, dir, ".tableaux/tasks/c48a.yaml", "title: Leaf\ndescription: l\nassignee: a@example.org\nparent: { id: \"07e0\", order: 2 }\n")
	write(t, dir, ".tableaux/status/b618.yaml", statusYAML)
	run(t, dir, "add", "-A")
	run(t, dir, "commit", "-q", "-m", "Start")
	return dir
}

// apply shows the commit, as the command does, then makes it, and returns
// the shown text.
func apply(t *testing.T, dir string, c Commit) string {
	t.Helper()
	shown, err := c.Show(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Apply(dir, testEnv); err != nil {
		t.Fatal(err)
	}
	return shown
}

func trailers(t *testing.T, dir string) string {
	msg := run(t, dir, "log", "-1", "--format=%B")
	cmd := exec.Command("git", "interpret-trailers", "--parse")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(msg)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestRecordEditsLinesOnly(t *testing.T) {
	dir := fixture(t)
	c, err := Record(dir, Status{ID: "b618", Gate: "function", State: "nominal",
		Note: "Prototype in prototype/b618: commits with trailers", SelfReview: true}, agent)
	if err != nil {
		t.Fatal(err)
	}
	shown := apply(t, dir, c)
	got, _ := os.ReadFile(filepath.Join(dir, statusPath("b618")))
	want := "# kept verbatim\ngate: function\n\nstate:   nominal\nnote: \"Prototype in prototype/b618: commits with trailers\"\n"
	if string(got) != want {
		t.Errorf("status file:\n%q\nwant\n%q", got, want)
	}
	// The diff touches the gate line and the note block only.
	for _, frag := range []string{"-gate: defined", "+gate: function", "+note: \"Prototype", "-  next"} {
		if !strings.Contains(shown, frag) {
			t.Errorf("shown lacks %q:\n%s", frag, shown)
		}
	}
	if strings.Contains(shown, "kept verbatim") && strings.Contains(shown, "-# kept") {
		t.Errorf("comment changed:\n%s", shown)
	}
	wantT := "Reviewed: b618 function\nModel: claude-sonnet-5-5\nCo-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>\n"
	if got := trailers(t, dir); got != wantT {
		t.Errorf("trailers %q want %q", got, wantT)
	}
	if ae := run(t, dir, "log", "-1", "--format=%an <%ae> / %cn <%ce>"); strings.TrimSpace(ae) != "Claude Sonnet 5.5 <noreply@anthropic.com> / Claude Sonnet 5.5 <noreply@anthropic.com>" {
		t.Errorf("identity %q", ae)
	}
}

func TestRecordInsertsInOrderAndRemoves(t *testing.T) {
	dir := fixture(t)
	write(t, dir, statusPath("c48a"), "gate: defined\nstate: nominal\n")
	c, err := Record(dir, Status{ID: "c48a", Gate: "defined", State: "stalled", Reason: "blocked", Note: "waits"}, Agent{})
	if err != nil {
		t.Fatal(err)
	}
	apply(t, dir, c)
	got, _ := os.ReadFile(filepath.Join(dir, statusPath("c48a")))
	if want := "gate: defined\nstate: stalled\nreason: blocked\nnote: waits\n"; string(got) != want {
		t.Errorf("got %q want %q", got, want)
	}
	c, _ = Record(dir, Status{ID: "c48a", Gate: "defined", State: "nominal"}, Agent{})
	apply(t, dir, c)
	got, _ = os.ReadFile(filepath.Join(dir, statusPath("c48a")))
	if want := "gate: defined\nstate: nominal\n"; string(got) != want {
		t.Errorf("got %q want %q", got, want)
	}
	if tr := trailers(t, dir); tr != "" {
		t.Errorf("unexpected trailers %q", tr)
	}
}

func TestEmptyCommitsCarryTrailers(t *testing.T) {
	dir := fixture(t)
	ids := []string{"b618", "c48a"}
	cases := []struct {
		name string
		make func() (Commit, error)
		want string
	}{
		{"authorise", func() (Commit, error) { return Authorise(dir, ids, agent) },
			"Authorised: b618\nAuthorised: c48a\n"},
		{"reaffirm", func() (Commit, error) { return Reaffirm(dir, ids, agent) },
			"Reaffirmed: b618\nReaffirmed: c48a\n"},
		{"review", func() (Commit, error) { return Review(dir, ids, "function", agent) },
			"Reviewed: b618 function\nReviewed: c48a function\n"},
	}
	for _, tc := range cases {
		c, err := tc.make()
		if err != nil {
			t.Fatal(tc.name, err)
		}
		shown := apply(t, dir, c)
		if !strings.Contains(shown, "empty commit") {
			t.Errorf("%s shown: %s", tc.name, shown)
		}
		want := tc.want + "Model: claude-sonnet-5-5\nCo-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>\n"
		if got := trailers(t, dir); got != want {
			t.Errorf("%s trailers %q want %q", tc.name, got, want)
		}
		if files := run(t, dir, "show", "--format=", "--name-only", "HEAD"); files != "" {
			t.Errorf("%s changed files %q", tc.name, files)
		}
	}
	// Git's own query finds the review as the method does.
	q := run(t, dir, "log", "--format=%h", "-E", "--grep=^Reviewed: b618 function$")
	if strings.TrimSpace(q) == "" {
		t.Error("git log --grep does not find the Reviewed trailer")
	}
}

func TestRefusals(t *testing.T) {
	dir := fixture(t)
	if _, err := Review(dir, []string{"b618"}, "nogate", Agent{}); err == nil {
		t.Error("unknown gate accepted")
	}
	if _, err := Authorise(dir, []string{"b618", "zzzz"}, Agent{}); err == nil {
		t.Error("unknown task accepted")
	}
	if _, err := Authorise(dir, nil, Agent{}); err == nil {
		t.Error("no task accepted")
	}
}

func TestProposeAllocatesFreshID(t *testing.T) {
	dir := fixture(t)
	cands := []string{"b618", "0000", "9f31"} // the first exists
	i := 0
	c, id, err := Propose(dir, Proposal{
		Title: "Sync commands", Description: "Fetch and push the trunk, and report what moved on the remote.",
		Assignee: "noreply@anthropic.com", Parent: "07e0", References: []string{"README.md#what-it-does"},
	}, agent, func() string { i++; return cands[i-1] })
	if err != nil {
		t.Fatal(err)
	}
	if id != "0000" {
		t.Fatalf("id %q", id)
	}
	shown := apply(t, dir, c)
	got, _ := os.ReadFile(filepath.Join(dir, taskPath("0000")))
	want := `title: Sync commands
description: >
  Fetch and push the trunk, and report what moved on the remote.
assignee: noreply@anthropic.com
references:
  - { url: README.md#what-it-does }
parent: { id: "07e0", order: 5 }
`
	if string(got) != want {
		t.Errorf("task file:\n%s\nwant\n%s", got, want)
	}
	if !strings.Contains(shown, "+title: Sync commands") {
		t.Errorf("shown:\n%s", shown)
	}
	// A proposal carries no Authorised trailer.
	if tr := trailers(t, dir); strings.Contains(tr, "Authorised") {
		t.Errorf("trailers %q", tr)
	}
}

// TestDemo prints the shown commits for the README: go test -v -run Demo.
func TestDemo(t *testing.T) {
	dir := fixture(t)
	c, _, _ := Propose(dir, Proposal{Title: "Sync commands", Description: "Fetch and push the trunk.",
		Assignee: "noreply@anthropic.com", Parent: "07e0"}, agent, func() string { return "1d2e" })
	t.Log("\n" + apply(t, dir, c))
	c, _ = Record(dir, Status{ID: "b618", Gate: "function", State: "nominal", Note: "Write commands shown", SelfReview: true}, agent)
	t.Log("\n" + apply(t, dir, c))
	c, _ = Authorise(dir, []string{"1d2e", "b618"}, agent)
	t.Log("\n" + apply(t, dir, c))
	t.Log("\n" + run(t, dir, "log", "--format=%h %s%n%(trailers:only,unfold)"))
}
