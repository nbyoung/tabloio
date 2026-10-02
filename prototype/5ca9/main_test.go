package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func do(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	args = append(args, "-testdata", "testdata")
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func argvOf(t *testing.T, args ...string) []string {
	t.Helper()
	v := findView(args[0])
	o, err := parse(v, args[1:], &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parse %v: %v", args, err)
	}
	return tabloArgv(v, o)
}

func TestArgvMapsFlags(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{[]string{"tableau"}, []string{"tablo", "--json", "view", "global-tableau"}},
		{[]string{"tableau", "--columns", "design,function", "--historical", "--for", "ben@example.org"},
			[]string{"tablo", "--json", "view", "global-tableau", "--person", "ben@example.org", "--columns", "design,function", "--historical-junctions"}},
		{[]string{"task", "9f31", "--level", "provenance", "-C", "/p", "--ref", "main"},
			[]string{"tablo", "-C", "/p", "--ref", "main", "--json", "view", "task-definition", "--task", "9f31", "--level", "provenance"}},
		{[]string{"task", "--for", "ada@example.org", "9f31"},
			[]string{"tablo", "--json", "view", "task-definition", "--task", "9f31", "--person", "ada@example.org"}},
		{[]string{"queue", "--for", "opus@example.org", "--brief", "9f31:function"},
			[]string{"tablo", "--json", "view", "work-queue", "--person", "opus@example.org", "--brief", "9f31:function"}},
		{[]string{"history", "--ref", "v1.0..v1.1", "--task", "e9c6"},
			[]string{"tablo", "--ref", "v1.0..v1.1", "--json", "view", "history", "--task", "e9c6"}},
		{[]string{"audit", "--stale", "3", "--role", "owner"},
			[]string{"tablo", "--json", "view", "audit", "--stale", "3", "--role", "owner"}},
		{[]string{"authority", "--proposed"}, []string{"tablo", "--json", "view", "authority-delegation", "--proposed"}},
		{[]string{"gates", "--window", "0"}, []string{"tablo", "--json", "view", "gate-definition", "--window", "0"}},
	}
	for _, c := range cases {
		if got := argvOf(t, c.args...); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v\n got %v\nwant %v", c.args, got, c.want)
		}
	}
}

func TestUsageErrors(t *testing.T) {
	bad := [][]string{
		{},
		{"nope"},
		{"task"},                      // id required
		{"task", "zz"},                // not an id
		{"tableau", "--task", "9f31"}, // tableau takes no task
		{"gates", "--for", "a@b.c"},   // gates take no person
		{"tableau", "--window", "1", "--columns", "design"},
		{"tableau", "--markdown", "--text"},
		{"context", "--task", "4e2b", "--for", "a@b.c"},
		{"tableau", "--level", "deep"},
		{"tableau", "--role", "boss"},
		{"tableau", "--ref", "a..b"}, // range: history only
		{"queue", "--brief", "9f31"},
		{"audit", "--for", "ben"},
		{"tableau", "extra"},
	}
	for _, a := range bad {
		if code, _, _ := do(t, a...); code != exitUsage {
			t.Errorf("%v: exit %d, want %d", a, code, exitUsage)
		}
	}
}

func TestEveryViewRendersBothFormats(t *testing.T) {
	for _, v := range views {
		args := []string{v.name}
		if v.pos == "task" {
			args = append(args, "9f31")
		}
		for _, f := range []string{"--markdown", "--text"} {
			code, out, errs := do(t, append(args, f)...)
			if code != exitOK {
				t.Fatalf("%v %s: exit %d: %s", args, f, code, errs)
			}
			if !strings.Contains(out, "Weather") && !strings.Contains(out, "9f31") && !strings.Contains(out, "gate") && !strings.Contains(out, "example.org") {
				t.Errorf("%v %s: output has no project content:\n%s", args, f, out)
			}
			if f == "--text" && strings.Contains(out, "|") {
				t.Errorf("%v: text output holds Markdown pipes", args)
			}
			if f == "--markdown" && strings.Contains(out, "┌") {
				t.Errorf("%v: Markdown output holds box characters", args)
			}
		}
	}
}

func TestMarkdownIsTheDefault(t *testing.T) {
	_, a, _ := do(t, "queue")
	_, b, _ := do(t, "queue", "--markdown")
	if a != b || !strings.HasPrefix(a, "# Queue") {
		t.Errorf("default differs from --markdown:\n%s", a)
	}
}

func TestLevelDeepens(t *testing.T) {
	_, glance, _ := do(t, "gates", "--level", "glance")
	_, detail, _ := do(t, "gates", "--level", "detail")
	if strings.Contains(glance, "criteria") || !strings.Contains(detail, "criteria") {
		t.Error("detail adds the criteria; glance omits them")
	}
	_, owner, _ := do(t, "gates", "--role", "owner")
	if !strings.Contains(owner, "level provenance") {
		t.Error("the owner role opens at provenance")
	}
}

func TestOutputToFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "STATUS.md")
	code, out, _ := do(t, "tableau", "-o", p)
	if code != exitOK || out != "" {
		t.Fatalf("exit %d, stdout %q", code, out)
	}
	b, err := os.ReadFile(p)
	if err != nil || !strings.Contains(string(b), "# Tableau") {
		t.Errorf("file: %v %q", err, b)
	}
}

func TestReadFailureExitCode(t *testing.T) {
	code, _, errs := do(t, "task", "c07d")
	if code != exitRead || !strings.Contains(errs, "no test data") {
		t.Errorf("exit %d: %s", code, errs)
	}
}

func TestExecRunnerPlumbing(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "tablo")
	body := "#!/bin/sh\necho \"$@\" >" + dir + "/args\n" +
		`echo '{"schema":"tablo/1","command":"view","ref":{"name":"main"},"data":{"view":"x","level":"glance","note":"ok"},"diagnostics":[{"severity":"warning","code":"W1","path":"","message":"careful"}]}'` + "\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := run([]string{"queue", "-tablo", script, "--for", "a@b.c", "--ref", "main"}, &out, &errb)
	if code != exitOK || !strings.Contains(out.String(), "note: ok") || !strings.Contains(errb.String(), "warning W1") {
		t.Fatalf("exit %d\n%s\n%s", code, out.String(), errb.String())
	}
	got, _ := os.ReadFile(dir + "/args")
	if want := "--ref main --json view work-queue --person a@b.c\n"; string(got) != want {
		t.Errorf("tablo got %q, want %q", got, want)
	}
	// A failing tablo passes its exit code through.
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho unknown ref >&2\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"queue", "-tablo", script}, &out, &errb); code != 3 {
		t.Errorf("exit %d, want 3", code)
	}
}
