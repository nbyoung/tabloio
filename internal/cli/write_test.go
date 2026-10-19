package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The files of a small project that tablo's library reads: the root e4a1
// and the leaf b2c9, whose design junction an agent contributes and a
// person reviews.
var writeProject = map[string]string{
	".tableaux/version.yaml": "tableaux: 0.3.1\ntrunk: main\n",
	".tableaux/gates.yaml": `gates:
  - { key: undefined, symbol: ❔, name: Undefined, criteria: No one has started work on the definition }
  - { key: defined,   symbol: 📝, name: Defined,   criteria: "Title, description, assignee and references exist" }
  - { key: design,    symbol: 📐, name: Design,    criteria: A model and sufficient tests exist }
  - { key: release,   symbol: 🚀, name: Release,   criteria: All variants documented and approved }
states:
  - { key: undefined, symbol: ⚪, severity: 0, synopsis: The work has not yet been defined }
  - { key: nominal,   symbol: 🟢, severity: 1, synopsis: The work is proceeding as expected }
  - { key: complete,  symbol: ✅, severity: 0, synopsis: All deliverables satisfy their requirements }
reasons:
  - { key: review,  symbol: 👓, synopsis: The work waits for its reviewer }
`,
	".tableaux/tasks/e4a1.yaml": "title: Small\ndescription: The root.\nassignee: olive@example.org\nreferences:\n  - { url: README.md }\n",
	".tableaux/tasks/b2c9.yaml": `title: Leaf
description: A leaf with an agent's design.
assignee: pat@example.org
references:
  - { url: README.md }
junctions:
  design: { contributor: bot@example.org, model: claude-fable, reviewer: rev@example.org }
parent: { id: "e4a1", order: 1 }
`,
	".tableaux/status/b2c9.yaml": "gate: defined\nstate: nominal\n",
	"README.md":                  "# Small\n",
}

// writeRepo builds the small project in a repository under t.TempDir, with
// every GIT_* variable of the host cleared, and sets the identity of the
// process to bot@example.org.
func writeRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH")
	}
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
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(k, "Bot")
	}
	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "bot@example.org")
	}
	dir := t.TempDir()
	for path, content := range writeProject {
		full := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitIn(t, dir, "init", "-q", "-b", "main")
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-q", "-m", "Plan the small project")
	return dir
}

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// writeRun runs a write command line with a terminal or none, an answer on
// standard input and an environment for TABLOIO_MODEL.
func writeRun(terminal bool, stdin string, vars map[string]string, argv ...string) (code int, stdout, stderr string) {
	var out, errb bytes.Buffer
	code = Run(context.Background(), argv, &Env{
		Stdout: &out, Stderr: &errb, Version: "1.2.3", Source: newSource(),
		Stdin: strings.NewReader(stdin), Terminal: func() bool { return terminal },
		Getenv: func(k string) string { return vars[k] },
	})
	return code, out.String(), errb.String()
}

// TestWriteUsage is T28.
func TestWriteUsage(t *testing.T) {
	dir := writeRepo(t)
	head := func() string { return gitIn(t, dir, "rev-parse", "HEAD") }
	none := map[string]string{}

	t.Run("mixed option order", func(t *testing.T) {
		code1, out1, err1 := writeRun(false, "", none, "-C", dir, "record", "b2c9", "--reason", "review", "--note", "The design waits", "-n")
		code2, out2, err2 := writeRun(false, "", none, "record", "--note", "The design waits", "-n", "--reason", "review", "b2c9", "-C", dir)
		if code1 != ExitOK || code2 != ExitOK || out1 != out2 || err1 != "" || err2 != "" {
			t.Fatalf("exit %d, %d; stderr %q, %q", code1, code2, err1, err2)
		}
		if !strings.HasPrefix(out1, "record b2c9 on main at ") || !strings.Contains(out1, "\n    Hand leaf (b2c9) to review\n") ||
			!strings.Contains(out1, "+reason: review\n+note: The design waits\n") {
			t.Errorf("the plan:\n%s", out1)
		}
	})
	t.Run("usage errors", func(t *testing.T) {
		for _, argv := range [][]string{
			{"-C", dir, "record", "b2c9", "-dry-run"},
			{"-C", dir, "record", "b2c9", "--ref", "main", "-n"},
			{"-C", dir, "record", "b2c9", "--markdown", "-n"},
			{"-C", dir, "record", "b2c9", "-o", "x", "-n"},
			{"-C", dir, "record", "b2c9", "--level", "glance", "-n"},
			{"-C", dir, "record", "b2c9", "--reason", "review", "--no-reason", "-n"},
			{"-C", dir, "record", "-n"},
			{"-C", dir, "review", "b2c9", "-n"},
			{"-C", dir, "reaffirm", "-n"},
			{"-C", dir, "propose", "--title", "T", "--parent", "e4a1", "-n"},
			{"-C", dir, "propose", "--title", "T", "--description", "D", "--parent", "e4a1", "--order", "x", "-n"},
			{"-C", dir, "authorise", "b2c9", "--gate", "design", "-n"},
		} {
			code, out, errs := writeRun(true, "", none, argv...)
			if code != ExitUsage || out != "" || !strings.Contains(errs, "Run 'tabloio ") {
				t.Errorf("%v: exit %d, stdout %q, stderr %q", argv, code, out, errs)
			}
		}
	})
	t.Run("no terminal and no --yes", func(t *testing.T) {
		code, out, errs := writeRun(false, "y\n", none, "-C", dir, "reaffirm", "b2c9")
		if code != ExitUsage || out != "" || !strings.Contains(errs, "--yes") || !strings.Contains(errs, "--dry-run") {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("n at the prompt", func(t *testing.T) {
		before := head()
		code, out, errs := writeRun(true, "n\n", none, "-C", dir, "reaffirm", "b2c9")
		if code != ExitInvalid || !strings.HasPrefix(out, "reaffirm b2c9 on main") || errs != "Commit? [y/N] not committed\n" || head() != before {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("yes at the prompt", func(t *testing.T) {
		code, out, errs := writeRun(true, "Yes\n", none, "-C", dir, "reaffirm", "b2c9")
		if want := "committed " + head()[:7] + " Reaffirm leaf (b2c9)\n"; code != ExitOK || !strings.HasSuffix(out, want) || errs != "Commit? [y/N] " {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("a refusal on one line", func(t *testing.T) {
		code, out, errs := writeRun(false, "", none, "-C", dir, "review", "b2c9", "nogate", "--yes")
		if code != ExitInvalid || out != "" || errs != "tabloio review: nogate: names no gate of gates.yaml (unknown-gate, H1)\n" {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
		code, _, errs = writeRun(false, "", none, "-C", dir, "review", "b2c9", "design", "--yes")
		if code != ExitInvalid || strings.Count(errs, "\n") != 1 || !strings.Contains(errs, "rev@example.org") {
			t.Errorf("by the wrong hand: exit %d, stderr %q", code, errs)
		}
		code, _, errs = writeRun(false, "", none, "-C", dir, "record", "b2c9", "--gate", "design", "--state", "nominal", "-n")
		if want := "tabloio record: b2c9 design: the reviewer rev@example.org has not accepted design; " +
			"hand the work off with --reason review (unreviewed, S11)\n"; code != ExitInvalid || errs != want {
			t.Errorf("unreviewed: exit %d, stderr %q", code, errs)
		}
	})
	t.Run("TABLOIO_MODEL and --model", func(t *testing.T) {
		vars := map[string]string{"TABLOIO_MODEL": "claude-fable-5-1", "TABLOIO_CO_AUTHOR": "Claude Fable <noreply@anthropic.com>"}
		_, out, _ := writeRun(false, "", vars, "-C", dir, "reaffirm", "b2c9", "-n")
		if !strings.Contains(out, "\n    Reaffirmed: b2c9\n    Model: claude-fable-5-1\n    Co-Authored-By: Claude Fable <noreply@anthropic.com>\n") {
			t.Errorf("from the environment:\n%s", out)
		}
		_, out, _ = writeRun(false, "", vars, "-C", dir, "reaffirm", "b2c9", "-n", "--model", "claude-fable-5-2")
		if !strings.Contains(out, "\n    Model: claude-fable-5-2\n") || strings.Contains(out, "5-1") {
			t.Errorf("--model:\n%s", out)
		}
	})
	t.Run("a commit with --yes, a subject and a body", func(t *testing.T) {
		code, out, errs := writeRun(false, "", none, "-C", dir, "propose", "--title", "Third leaf", "--description", "One more.",
			"--parent", "e4a1", "--id", "d00d", "--reference", "README.md The charter", "--reference", "https://example.org/search?q=gate", "-m", "Add a leaf", "--body", "One.", "--body", "Two.", "-y")
		if code != ExitOK || errs != "" || !strings.HasSuffix(out, "committed "+head()[:7]+" Add a leaf\n") {
			t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
		if got := gitIn(t, dir, "log", "-1", "--format=%B"); got != "Add a leaf\n\nOne.\n\nTwo." {
			t.Errorf("message %q", got)
		}
		if got := gitIn(t, dir, "show", "HEAD:.tableaux/tasks/d00d.yaml"); !strings.Contains(got, "  - { url: README.md, text: The charter }\n  - { url: \"https://example.org/search?q=gate\" }\nparent: { id: \"e4a1\", order: 2 }") {
			t.Errorf("task file\n%s", got)
		}
	})
	t.Run("a failure to read", func(t *testing.T) {
		code, out, errs := writeRun(false, "", none, "-C", t.TempDir(), "reaffirm", "b2c9", "-n")
		if code != ExitFailure || out != "" || !strings.HasPrefix(errs, "tabloio reaffirm: git rev-parse") {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("the mockups' write commands", func(t *testing.T) {
		for _, line := range []string{"review 5471 mockup", "review e9c6 design"} {
			a, err := Parse(strings.Fields(line), vocabulary)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := writeRequest(lookup(a.Command), a, func(string) string { return "" }); err != nil {
				t.Errorf("%s: %v", line, err)
			}
		}
	})
}
