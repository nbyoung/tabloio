package cli

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type testCase struct {
	Args       []string
	Invocation *struct {
		Query  Query
		Text   bool
		Output string
		Viewer bool
	}
	Error string
}

func loadCases(t *testing.T) []testCase {
	t.Helper()
	var cases []testCase
	if err := json.Unmarshal([]byte(golden(t, "cases.json")), &cases); err != nil {
		t.Fatal(err)
	}
	return cases
}

// invocationOf runs the steps of Run that come before the Source: parse,
// look up, check the options, check the values.
func invocationOf(argv []string) (*invocation, error) {
	a, err := Parse(argv, vocabulary)
	if err != nil {
		return nil, err
	}
	c := lookup(a.Command)
	if c == nil {
		return nil, fmt.Errorf("unknown command %q", a.Command)
	}
	if err := checkApplies(c, a); err != nil {
		return nil, err
	}
	return readInvocation(c, a)
}

// TestCases is T1: the 67 command lines of cases.json build the invocation
// the draft names, or earn the usage error it names.
func TestCases(t *testing.T) {
	cases := loadCases(t)
	accepted, refused := 0, 0
	for _, c := range cases {
		name := strings.Join(c.Args, " ")
		inv, err := invocationOf(c.Args)
		if c.Invocation != nil {
			accepted++
			if err != nil {
				t.Errorf("%s: %v", name, err)
				continue
			}
			if want := (invocation{Query: c.Invocation.Query, Text: c.Invocation.Text, Output: c.Invocation.Output, Viewer: c.Invocation.Viewer}); !reflect.DeepEqual(*inv, want) {
				t.Errorf("%s:\n got %+v\nwant %+v", name, *inv, want)
			}
			continue
		}
		refused++
		if err == nil || !strings.Contains(err.Error(), c.Error) {
			t.Errorf("%s: error %v, want one with %q", name, err, c.Error)
		}
	}
	if accepted != 24 || refused != 43 {
		t.Errorf("cases.json holds %d accepted and %d refused lines, want 24 and 43", accepted, refused)
	}
}

// TestCasesThroughRun sends each refused line of cases.json through Run: a
// usage error, a message that names the command, and no Source call.
func TestCasesThroughRun(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Invocation != nil {
			continue
		}
		src := newSource()
		code, out, errs := run(t, src, c.Args...)
		if code != ExitUsage || out != "" || !strings.Contains(errs, c.Error) || !src.untouched() {
			t.Errorf("%v: exit %d, stdout %q, stderr %q, source touched %t", c.Args, code, out, errs, !src.untouched())
		}
		if !strings.HasPrefix(errs, "tabloio") || !strings.Contains(errs, "\nRun 'tabloio ") {
			t.Errorf("%v: stderr %q has no usage-error shape", c.Args, errs)
		}
	}
}

// TestMockupCommands is T2: the command lines of the ten mockups.
func TestMockupCommands(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(golden(t, "mockup-commands.tsv")), "\n")
	counts := map[string]int{}
	for _, l := range lines[1:] {
		f := strings.Split(l, "\t")
		if len(f) != 5 {
			t.Fatalf("malformed line %q", l)
		}
		file, line, command, verdict, becomes := f[0], f[1], f[2], f[3], f[4]
		where := file + ":" + line
		fragment := !strings.HasPrefix(command, "tabloio ")
		switch verdict {
		case "stands":
			if fragment {
				counts["stands fragment"]++
				if _, err := Parse(strings.Fields(command), vocabulary); err != nil {
					t.Errorf("%s: fragment %q: %v", where, command, err)
				}
				continue
			}
			counts["stands"]++
			if _, err := invocationOf(strings.Fields(command)[1:]); err != nil {
				t.Errorf("%s: %q: %v", where, command, err)
			}
		case "changes":
			if fragment {
				counts["changes fragment"]++
				if _, err := Parse(strings.Fields(command), vocabulary); err == nil || !strings.Contains(err.Error(), "unknown option --for") {
					t.Errorf("%s: fragment %q: error %v, want --for refused", where, command, err)
				}
				if _, err := Parse(strings.Fields(becomes), vocabulary); err != nil {
					t.Errorf("%s: replacement %q: %v", where, becomes, err)
				}
				continue
			}
			counts["changes"]++
			if _, err := invocationOf(strings.Fields(command)[1:]); err == nil || !strings.Contains(err.Error(), "unknown option --for") {
				t.Errorf("%s: %q: error %v, want --for refused", where, command, err)
			}
			if _, err := invocationOf(strings.Fields(becomes)[1:]); err != nil {
				t.Errorf("%s: replacement %q: %v", where, becomes, err)
			}
		case "write command, b618":
			counts["write"]++
			if !strings.HasPrefix(command, "tabloio review ") {
				t.Errorf("%s: %q is no write command of b618", where, command)
			}
		default:
			t.Errorf("%s: verdict %q", where, verdict)
		}
	}
	want := map[string]int{"stands": 67, "stands fragment": 2, "changes": 7, "changes fragment": 2, "write": 2}
	if !reflect.DeepEqual(counts, want) {
		t.Errorf("the mockups hold %v, want %v", counts, want)
	}
}

// TestOptionMatrix is T3: each option applies to exactly the views of the
// vocabulary table, and the window and the historical switch reach the two
// tableaux alone.
func TestOptionMatrix(t *testing.T) {
	fakeTables(t, true)
	applies := map[string][]string{
		"gates":      {"task"},
		"task":       {"person"},
		"authority":  {"task", "person", "proposed"},
		"assignment": {"task", "person"},
		"queue":      {"task", "person", "brief"},
		"blockage":   {"task", "person"},
		"tableau":    {"person", "window", "columns", "historical"},
		"context":    {"task", "person", "window", "columns", "historical"},
		"history":    {"task", "person"},
		"audit":      {"task", "person", "stale"},
	}
	sample := map[string][]string{
		"task":       {"--task", "2034"},
		"person":     {"--person", "a@b.c"},
		"window":     {"--window", "1"},
		"columns":    {"--columns", "design"},
		"historical": {"--historical"},
		"proposed":   {"--proposed"},
		"stale":      {"--stale", "7"},
		"brief":      {"--brief", "c2ad", "validate"},
	}
	options := []string{"task", "person", "window", "columns", "historical", "proposed", "stale", "brief"}
	for _, v := range views {
		if _, ok := applies[v.name]; !ok {
			t.Fatalf("the matrix lacks %s", v.name)
		}
		for _, opt := range options {
			argv := []string{v.name}
			if v.name == "task" {
				argv = append(argv, "e9c6")
			}
			argv = append(argv, sample[opt]...)
			src := newSource()
			code, _, errs := run(t, src, argv...)
			applicable := false
			for _, o := range applies[v.name] {
				applicable = applicable || o == opt
			}
			if applicable {
				if code != ExitOK {
					t.Errorf("%v: exit %d, stderr %q", argv, code, errs)
					continue
				}
				q := src.queries[0]
				if (q.Window != nil) != (opt == "window") || q.Historical != (opt == "historical") {
					t.Errorf("%v: window %v, historical %t reach the query", argv, q.Window, q.Historical)
				}
				continue
			}
			want := "--" + opt + " does not apply to " + v.name
			if code != ExitUsage || !strings.Contains(errs, want) || !src.untouched() {
				t.Errorf("%v: exit %d, stderr %q, want %q", argv, code, errs, want)
			}
		}
	}
}

func ptr(n int) *int { return &n }

// TestQueryReachesTheSource is T4: ten lines, one per view, through Run.
func TestQueryReachesTheSource(t *testing.T) {
	fakeTables(t, true)
	for _, c := range []struct {
		argv []string
		want Query
	}{
		{[]string{"gates", "--ref", "main", "--task", "e9c6", "--level", "detail"},
			Query{View: "gates", Ref: "main", Task: "e9c6", Level: "detail"}},
		{[]string{"task", "e9c6", "--person", "a@b.c", "--ref", "3cdae52"},
			Query{View: "task", Task: "e9c6", Person: "a@b.c", Ref: "3cdae52", Level: "glance"}},
		{[]string{"authority", "--proposed", "--task", "bc86"},
			Query{View: "authority", Task: "bc86", Proposed: true, Level: "glance"}},
		{[]string{"assignment", "--person", "a@b.c"},
			Query{View: "assignment", Person: "a@b.c", Level: "glance"}},
		{[]string{"queue", "--person", "a@b.c", "--brief", "c2ad", "validate"},
			Query{View: "queue", Person: "a@b.c", Brief: &Brief{Task: "c2ad", Gate: "validate"}, Level: "provenance"}},
		{[]string{"blockage", "--task", "2034", "--level", "provenance"},
			Query{View: "blockage", Task: "2034", Level: "provenance"}},
		{[]string{"-C", "sub", "tableau", "--columns", "design,unit", "--historical"},
			Query{Dir: "sub", View: "tableau", Columns: []string{"design", "unit"}, Historical: true, Level: "glance"}},
		{[]string{"context", "--task", "2034", "--window", "0"},
			Query{View: "context", Task: "2034", Window: ptr(0), Level: "glance"}},
		{[]string{"history", "--ref", "0704a09..main", "--person", "a@b.c"},
			Query{View: "history", Ref: "0704a09..main", Person: "a@b.c", Level: "glance"}},
		{[]string{"audit", "--stale", "0", "--task", "e9c6"},
			Query{View: "audit", Task: "e9c6", Stale: ptr(0), Level: "glance"}},
	} {
		src := newSource()
		if code, _, errs := run(t, src, c.argv...); code != ExitOK {
			t.Errorf("%v: exit %d, stderr %q", c.argv, code, errs)
			continue
		}
		if len(src.queries) != 1 || !reflect.DeepEqual(src.queries[0], c.want) {
			t.Errorf("%v:\n got %+v\nwant %+v", c.argv, src.queries, c.want)
		}
	}
}

// TestViewer is T5: the viewer default.
func TestViewer(t *testing.T) {
	fakeTables(t, true)
	t.Run("queue and context ask once and name the viewer", func(t *testing.T) {
		for _, argv := range [][]string{{"queue"}, {"context"}, {"-C", "d", "queue", "--ref", "main"}, {"queue", "--brief", "c2ad", "validate"}} {
			src := newSource()
			if code, _, errs := run(t, src, argv...); code != ExitOK {
				t.Errorf("%v: exit %d, stderr %q", argv, code, errs)
				continue
			}
			if len(src.viewerDirs) != 1 || len(src.queries) != 1 || src.queries[0].Person != "viewer@example.org" {
				t.Errorf("%v: viewer calls %v, queries %+v", argv, src.viewerDirs, src.queries)
			}
		}
		src := newSource()
		run(t, src, "-C", "d", "queue")
		if !reflect.DeepEqual(src.viewerDirs, []string{"d"}) {
			t.Errorf("Viewer got %v, want the -C directory", src.viewerDirs)
		}
	})
	t.Run("no email is a usage error", func(t *testing.T) {
		for _, argv := range [][]string{{"queue"}, {"context"}} {
			src := newSource()
			src.viewer = ""
			code, out, errs := run(t, src, argv...)
			if code != ExitUsage || out != "" || !strings.Contains(errs, "no person: give --person, or set user.email in Git") || len(src.queries) != 0 {
				t.Errorf("%v: exit %d, stdout %q, stderr %q, queries %v", argv, code, out, errs, src.queries)
			}
		}
	})
	t.Run("an error is a failure", func(t *testing.T) {
		src := newSource()
		src.viewerErr = errors.New("git is gone")
		code, out, errs := run(t, src, "queue")
		if code != ExitFailure || out != "" || !strings.Contains(errs, "tabloio queue: git is gone") || len(src.queries) != 0 {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("never asked otherwise", func(t *testing.T) {
		lines := [][]string{{"queue", "--person", "a@b.c"}, {"context", "--task", "2034"}, {"context", "--person", "a@b.c"}}
		for _, v := range views {
			if v.name == "queue" || v.name == "context" {
				continue
			}
			argv := []string{v.name}
			if v.name == "task" {
				argv = append(argv, "e9c6")
			}
			lines = append(lines, argv)
		}
		for _, argv := range lines {
			src := newSource()
			if code, _, errs := run(t, src, argv...); code != ExitOK {
				t.Errorf("%v: exit %d, stderr %q", argv, code, errs)
			}
			if len(src.viewerDirs) != 0 {
				t.Errorf("%v: Viewer called %v", argv, src.viewerDirs)
			}
			if len(src.queries) == 1 && argv[0] != "queue" && argv[0] != "context" && src.queries[0].Person != "" {
				t.Errorf("%v: names the person %q", argv, src.queries[0].Person)
			}
		}
	})
}

// TestExitCodes is T6.
func TestExitCodes(t *testing.T) {
	fakeTables(t, true)
	dir := t.TempDir()
	t.Run("data", func(t *testing.T) {
		code, out, errs := run(t, newSource(), "gates")
		if code != ExitOK || out != "fake markdown gates: data=the data level=glance base=\"\"\n" || errs != "" {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("no data and an error diagnostic", func(t *testing.T) {
		src := newSource()
		src.result = Result{Diagnostics: []Diagnostic{
			{Severity: "error", Code: "E12", Path: ".tableaux/tasks/e9c6.yaml", Line: 4, Col: 2, Message: "no such gate"},
			{Severity: "warning", Code: "W3", Path: "x", Message: "ignored"},
			{Severity: "error", Code: "E1", Message: "no project"},
			{Severity: "error", Code: "E2", Path: "gates.yaml", Line: 9, Message: "bad key"},
		}}
		out := filepath.Join(dir, "S.md")
		code, stdout, errs := run(t, src, "gates", "-o", out)
		want := ".tableaux/tasks/e9c6.yaml:4:2: error E12: no such gate\n" +
			"error E1: no project\n" +
			"gates.yaml:9: error E2: bad key\n"
		if code != ExitInvalid || stdout != "" || errs != want {
			t.Errorf("exit %d, stdout %q, stderr %q, want %q", code, stdout, errs, want)
		}
		if _, err := os.Stat(out); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("a view that is prevented wrote %s: %v", out, err)
		}
	})
	t.Run("no data and no error diagnostic", func(t *testing.T) {
		src := newSource()
		src.result = Result{Diagnostics: []Diagnostic{{Severity: "warning", Code: "W3", Message: "w"}}}
		if code, _, errs := run(t, src, "gates"); code != ExitFailure || !strings.Contains(errs, "no data") {
			t.Errorf("exit %d, stderr %q", code, errs)
		}
	})
	t.Run("a source error", func(t *testing.T) {
		src := newSource()
		src.viewErr = errors.New("unknown ref nope")
		code, out, errs := run(t, src, "gates")
		if code != ExitFailure || out != "" || errs != "tabloio gates: unknown ref nope\n" {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("a renderer error", func(t *testing.T) {
		markdown["gates"] = func(any, string, string) ([]byte, error) { return nil, errors.New("cannot draw") }
		code, out, errs := run(t, newSource(), "gates")
		if code != ExitFailure || out != "" || errs != "tabloio gates: cannot draw\n" {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("an unwritable output", func(t *testing.T) {
		code, _, errs := run(t, newSource(), "queue", "--person", "a@b.c", "-o", filepath.Join(dir, "no", "such", "S.md"))
		if code != ExitFailure || !strings.Contains(errs, "tabloio queue: ") {
			t.Errorf("exit %d, stderr %q", code, errs)
		}
	})
	t.Run("a usage error never calls the source", func(t *testing.T) {
		src := newSource()
		if code, _, _ := run(t, src, "queue", "--task", "zz"); code != ExitUsage || !src.untouched() {
			t.Errorf("exit %d, source touched %t", code, !src.untouched())
		}
	})
}

// TestDiagnosticsBesideData is T7.
func TestDiagnosticsBesideData(t *testing.T) {
	fakeTables(t, true)
	plain := newSource()
	_, want, _ := run(t, plain, "gates")
	src := newSource()
	src.result.Diagnostics = []Diagnostic{
		{Severity: "error", Code: "E1", Path: "a", Line: 1, Col: 1, Message: "broken"},
		{Severity: "warning", Code: "W1", Message: "careful"},
	}
	code, out, errs := run(t, src, "gates")
	if code != ExitOK || out != want || errs != "" {
		t.Errorf("exit %d, stdout %q (want %q), stderr %q", code, out, want, errs)
	}
}

// TestOutput is T8.
func TestOutput(t *testing.T) {
	fakeTables(t, true)
	const want = "fake markdown gates: data=the data level=glance base=%q\n"
	dir := t.TempDir()
	t.Run("standard output", func(t *testing.T) {
		code, out, errs := run(t, newSource(), "gates")
		if code != ExitOK || out != fmt.Sprintf(want, "") || errs != "" {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("a file", func(t *testing.T) {
		for i, flag := range []string{"--output", "-o"} {
			name := fmt.Sprint("f", i, ".md")
			path := filepath.Join(dir, name)
			code, out, errs := run(t, newSource(), "-C", dir, "gates", flag, name)
			b, err := os.ReadFile(path)
			if code != ExitOK || out != "" || errs != "" || err != nil || string(b) != fmt.Sprintf(want, "") {
				t.Errorf("%s: exit %d, stdout %q, stderr %q, file %q, %v", flag, code, out, errs, b, err)
			}
			if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o644 {
				t.Errorf("%s: %v, %v", flag, info, err)
			}
		}
	})
	t.Run("a hyphen is standard output", func(t *testing.T) {
		code, out, _ := run(t, newSource(), "gates", "-o", "-")
		if code != ExitOK || out != fmt.Sprintf(want, "") {
			t.Errorf("exit %d, stdout %q", code, out)
		}
	})
	t.Run("a relative path resolves against -C", func(t *testing.T) {
		project := filepath.Join(dir, "project")
		if err := os.MkdirAll(filepath.Join(project, "docs"), 0o755); err != nil {
			t.Fatal(err)
		}
		src := newSource()
		code, out, errs := run(t, src, "-C", project, "gates", "-o", filepath.Join("docs", "S.md"))
		b, err := os.ReadFile(filepath.Join(project, "docs", "S.md"))
		if code != ExitOK || out != "" || errs != "" || err != nil || string(b) != fmt.Sprintf(want, "..") {
			t.Errorf("exit %d, stdout %q, stderr %q, file %q, %v", code, out, errs, b, err)
		}
		if len(src.queries) != 1 || src.queries[0].Dir != project {
			t.Errorf("queries %+v", src.queries)
		}
	})
	t.Run("-C names the guarded directory", func(t *testing.T) {
		src := newSource()
		code, _, errs := run(t, src, "-C", filepath.Join(dir, ".tableaux"), "gates", "-o", "S.md")
		if code != ExitUsage || !strings.Contains(errs, "does not write inside .tableaux") || !src.untouched() {
			t.Errorf("exit %d, stderr %q", code, errs)
		}
	})
}

// TestWriteIsAtomic is T9: a failure leaves an existing file as it stood and
// no temporary file behind.
func TestWriteIsAtomic(t *testing.T) {
	fakeTables(t, true)
	dir := t.TempDir()
	path := filepath.Join(dir, "S.md")
	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	check := func(t *testing.T) {
		t.Helper()
		if b, _ := os.ReadFile(path); string(b) != "old\n" {
			t.Errorf("the file reads %q, want it as it stood", b)
		}
		if got := entries(t, dir); len(got) != 1 {
			t.Errorf("the directory holds %v, want the file alone", got)
		}
	}
	t.Run("a renderer error", func(t *testing.T) {
		markdown["gates"] = func(any, string, string) ([]byte, error) { return nil, errors.New("cannot draw") }
		if code, _, _ := run(t, newSource(), "gates", "-o", path); code != ExitFailure {
			t.Errorf("exit %d", code)
		}
		check(t)
	})
	t.Run("a read-only directory", func(t *testing.T) {
		if err := os.Chmod(dir, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		if probe, err := os.CreateTemp(dir, "probe-*"); err == nil {
			_ = probe.Close()
			_ = os.Remove(probe.Name())
			t.Skip("the directory stays writable, as it does for root")
		}
		if code, _, _ := run(t, newSource(), "gates", "-o", path); code != ExitFailure {
			t.Errorf("exit %d", code)
		}
		check(t)
	})
}

// TestFormat is T11.
func TestFormat(t *testing.T) {
	t.Run("--text with an empty table", func(t *testing.T) {
		fakeTables(t, true)
		src := newSource()
		code, out, errs := run(t, src, "tableau", "--text")
		if code != ExitUsage || out != "" || !strings.Contains(errs, "the text format does not cover tableau yet") || !src.untouched() {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("--text with an entry", func(t *testing.T) {
		fakeTables(t, true, "tableau")
		code, out, _ := run(t, newSource(), "tableau", "--text")
		if code != ExitOK || out != "fake text tableau: data=the data level=glance base=\"\"\n" {
			t.Errorf("exit %d, stdout %q", code, out)
		}
		if code, _, errs := run(t, newSource(), "gates", "--text"); code != ExitUsage || !strings.Contains(errs, "does not cover gates yet") {
			t.Errorf("a view with no entry: exit %d, stderr %q", code, errs)
		}
	})
	t.Run("--markdown and the default agree", func(t *testing.T) {
		fakeTables(t, true)
		_, a, _ := run(t, newSource(), "gates")
		_, b, _ := run(t, newSource(), "gates", "--markdown")
		if a != b {
			t.Errorf("%q and %q differ", a, b)
		}
	})
	t.Run("--brief selects the brief entry", func(t *testing.T) {
		fakeTables(t, true)
		code, out, _ := run(t, newSource(), "queue", "--person", "a@b.c", "--brief", "c2ad", "validate")
		if code != ExitOK || out != "fake markdown brief: data=the data level=provenance base=\"\"\n" {
			t.Errorf("exit %d, stdout %q", code, out)
		}
	})
	t.Run("--brief with no brief entry", func(t *testing.T) {
		fakeTables(t, false)
		src := newSource()
		code, out, errs := run(t, src, "queue", "--person", "a@b.c", "--brief", "c2ad", "validate")
		if code != ExitUsage || out != "" || !strings.Contains(errs, "the markdown format does not cover brief yet") || !src.untouched() {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
}

// digest hashes every path, mode and content under root except the paths in
// skip, which are keys of the slash path from root.
func digest(t *testing.T, root string, skip map[string]bool) string {
	t.Helper()
	h := sha256.New()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if skip[filepath.ToSlash(rel)] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(h, "%s %v\n", filepath.ToSlash(rel), info.Mode())
		if info.Mode().IsRegular() {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			_, _ = h.Write(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// TestChangesNothingButTheOutput is T12: every accepted line of cases.json
// runs in a project with .tableaux and .git fixtures, and a digest of every
// path, mode and content is equal after, but for the named output.
func TestChangesNothingButTheOutput(t *testing.T) {
	var all []string
	for _, v := range views {
		all = append(all, v.name)
	}
	fakeTables(t, true, all...)
	root := t.TempDir()
	write := func(name, content string) {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".tableaux/gates.yaml", "gates: []\n")
	write(".tableaux/tasks/5ca9.yaml", "title: x\n")
	write(".tableaux/status/5ca9.yaml", "gate: design\n")
	write(".git/HEAD", "ref: refs/heads/main\n")
	write(".git/index", "index\n")
	write(".git/config", "[user]\n\temail = viewer@example.org\n")

	// The lines name their own -C directories; make them first, and send
	// each line to the project through the same prefix.
	var lines [][]string
	skip := map[string]bool{}
	for _, c := range loadCases(t) {
		if c.Invocation == nil {
			continue
		}
		argv := append([]string(nil), c.Args...)
		sub := ""
		for i := 0; i < len(argv)-1; i++ {
			if argv[i] == "-C" {
				sub = argv[i+1]
				argv[i+1] = filepath.Join(root, sub)
			}
		}
		if sub == "" {
			argv = append([]string{"-C", root}, argv...)
		}
		if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
			t.Fatal(err)
		}
		if c.Invocation.Output != "" {
			skip[filepath.ToSlash(filepath.Join(sub, c.Invocation.Output))] = true
		}
		lines = append(lines, argv)
	}
	before := digest(t, root, skip)

	src := newSource()
	for _, argv := range lines {
		if code, _, errs := run(t, src, argv...); code != ExitOK {
			t.Errorf("%v: exit %d, stderr %q", argv, code, errs)
		}
	}
	if len(src.queries) == 0 {
		t.Fatal("no line reached the Source")
	}
	for i, l := range src.locks {
		if l != "0" {
			t.Errorf("View call %d saw GIT_OPTIONAL_LOCKS=%q, want 0", i, l)
		}
	}
	if after := digest(t, root, skip); after != before {
		t.Error("a read command changed the project")
	}
	var written []string
	for p := range skip {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(p))); err == nil {
			written = append(written, p)
		}
	}
	sort.Strings(written)
	if !reflect.DeepEqual(written, []string{"out.txt"}) {
		t.Errorf("the named outputs written: %v, want out.txt", written)
	}
}
