package cli

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/nbyoung/tablo"
)

// TestVersionJSON is the --json half of T14: one line that decodes, with
// data.tablo as "v" joined to tablo.Version.
func TestVersionJSON(t *testing.T) {
	code, out, errs := run(t, newSource(), "version", "--json")
	if code != ExitOK || errs != "" || strings.Count(out, "\n") != 1 {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errs)
	}
	var got struct {
		Command     string            `json:"command"`
		Data        map[string]string `json:"data"`
		Diagnostics []any             `json:"diagnostics"`
		Schema      string            `json:"schema"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"tableaux": "0.3",
		"tablo":    "v" + tablo.Version,
		"tabloio":  "1.2.3",
	}
	if got.Command != "version" || got.Schema != "tabloio/1" || got.Diagnostics == nil || len(got.Diagnostics) != 0 {
		t.Errorf("envelope %+v", got)
	}
	if got.Data["tablo"] != want["tablo"] || got.Data["tabloio"] != want["tabloio"] {
		t.Errorf("data %v, want %v", got.Data, want)
	}
	if !strings.HasPrefix(got.Data["tableaux"], "0.") {
		t.Errorf("data.tableaux = %q", got.Data["tableaux"])
	}
	if !strings.HasPrefix(out, `{"command":"version","data":{"tableaux":`) || !strings.HasSuffix(out, `,"diagnostics":[],"schema":"tabloio/1"}`+"\n") {
		t.Errorf("key order: %s", out)
	}
	for _, argv := range [][]string{{"version", "--json", "extra"}, {"tableau", "--json"}} {
		if code, _, _ := run(t, newSource(), argv...); code != ExitUsage {
			t.Errorf("%v: exit %d, want %d", argv, code, ExitUsage)
		}
	}
}

// TestUsageFault: a usage error from the library is a usage error of the
// command, exit 2; any other error is a failure, exit 3.
func TestUsageFault(t *testing.T) {
	src := newSource()
	src.viewErr = &UsageFault{Msg: "--stale 0 is below one"}
	code, out, errs := run(t, src, "audit")
	if code != ExitUsage || out != "" || !strings.HasPrefix(errs, "tabloio audit: --stale 0 is below one\nRun 'tabloio audit --help'") {
		t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
	}
}

// TestTabloSourceViewer reads the viewer as tablo's own command does:
// GIT_AUTHOR_EMAIL when set and not empty, else user.email of the
// repository, else "".
func TestTabloSourceViewer(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	// An isolated configuration: no global or system user.email leaks in.
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	src := tabloSource{}
	viewer := func() string {
		t.Helper()
		v, err := src.Viewer(context.Background(), dir)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}

	t.Setenv("GIT_AUTHOR_EMAIL", "")
	if v := viewer(); v != "" {
		t.Errorf("no configuration: viewer %q", v)
	}
	git("config", "user.email", "cfg@example.org")
	if v := viewer(); v != "cfg@example.org" {
		t.Errorf("user.email: viewer %q", v)
	}
	t.Setenv("GIT_AUTHOR_EMAIL", "env@example.org")
	if v := viewer(); v != "env@example.org" {
		t.Errorf("GIT_AUTHOR_EMAIL: viewer %q", v)
	}
	t.Setenv("GIT_AUTHOR_EMAIL", "not an email")
	if v := viewer(); v != "" {
		t.Errorf("a value that is no email: viewer %q", v)
	}
}

// TestTabloSourceErrors: a directory with no repository fails the read
// (exit 3).
func TestTabloSourceErrors(t *testing.T) {
	code, out, errs := runReal(t, "-C", t.TempDir(), "tableau")
	if code != ExitFailure || out != "" || errs == "" {
		t.Errorf("no repository: exit %d, stdout %q, stderr %q", code, out, errs)
	}
}
