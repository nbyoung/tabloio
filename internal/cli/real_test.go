package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// realCorpusEntry is the built corpus entry the tests of the real path read.
const realCorpusEntry = "weather-station"

// corpusCopy copies the built corpus entry into a temporary directory, with
// its .git and its submodule, and returns the copy. It skips without
// TABLEAUX_CORPUS. The built corpus is shared and read-only, so no test runs
// anything in it.
func corpusCopy(t *testing.T) string {
	t.Helper()
	corpus := os.Getenv("TABLEAUX_CORPUS")
	if corpus == "" {
		t.Skip("TABLEAUX_CORPUS names no built corpus")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	if _, err := exec.LookPath("cp"); err != nil {
		t.Skip("cp is not installed")
	}
	src := filepath.Join(corpus, realCorpusEntry)
	if _, err := os.Stat(src); err != nil {
		t.Skipf("the corpus holds no %s: %v", realCorpusEntry, err)
	}
	dst := filepath.Join(t.TempDir(), realCorpusEntry)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cp", "-a", src+"/.", dst).CombinedOutput(); err != nil {
		t.Fatalf("copy the corpus entry: %v\n%s", err, out)
	}
	return dst
}

// fixClock replaces the audit's clock with a fixed date for the test.
func fixClock(t *testing.T) {
	t.Helper()
	old := today
	today = func() string { return "2026-10-16" }
	t.Cleanup(func() { today = old })
}

// runReal calls Run over tabloSource with the viewer set.
func runReal(t *testing.T, argv ...string) (code int, stdout, stderr string) {
	t.Helper()
	t.Setenv("GIT_AUTHOR_EMAIL", "nbyoung@nbyoung.com")
	t.Setenv("GIT_OPTIONAL_LOCKS", "")
	var out, errb bytes.Buffer
	code = Run(context.Background(), argv, &Env{Stdout: &out, Stderr: &errb, Version: "1.2.3", Source: tabloSource{}})
	return code, out.String(), errb.String()
}

// gitOut runs git in dir and returns its output.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

// otherProject holds the tasks and refs the draft's lines name that belong
// to the Tableaux plan, not to the corpus entry, and the one line whose -C
// directory does not exist.
var otherProject = map[string]bool{
	"e9c6": true, "bc63": true, "c2ad": true, "2034": true,
	"3cdae52": true, "dc669dd": true, "0704a09..main": true,
	"sub/dir": true, "subprojects/tablo": true,
	"design,implementation,at-risk_2": true, // a gate key gates.yaml lacks
}

// realLines returns the accepted lines of cases.json that name nothing of
// another project, each pointed at dir, with its --output moved to out.
func realLines(t *testing.T, dir, out string) [][]string {
	t.Helper()
	var lines [][]string
Cases:
	for _, c := range loadCases(t) {
		if c.Invocation == nil {
			continue
		}
		argv := []string{"-C", dir}
		for _, a := range c.Args {
			for _, f := range strings.FieldsFunc(a, func(r rune) bool { return r == '=' }) {
				if otherProject[f] {
					continue Cases
				}
			}
			if a == c.Invocation.Output && a != "" {
				a = filepath.Join(out, a)
			}
			argv = append(argv, a)
		}
		lines = append(lines, argv)
	}
	if len(lines) == 0 {
		t.Fatal("no line of cases.json is left")
	}
	return lines
}

// TestRealPathChangesNothing is T16: the accepted lines of cases.json that
// name no task or ref of another project run over tablo's library on a copy
// of the corpus entry. Git reports no change, a digest of the working tree
// and .git is equal afterwards, and two runs give equal bytes.
func TestRealPathChangesNothing(t *testing.T) {
	dir := corpusCopy(t)
	fixClock(t)
	outDir := t.TempDir()
	lines := realLines(t, dir, outDir)

	if got := gitOut(t, dir, "status", "--porcelain"); got != "" {
		t.Fatalf("the copy starts dirty:\n%s", got)
	}
	before := digest(t, dir, nil)
	first := map[string]string{}
	for _, argv := range lines {
		code, out, errs := runReal(t, argv...)
		if code != ExitOK {
			t.Errorf("%v: exit %d, stderr %q", argv, code, errs)
			continue
		}
		first[strings.Join(argv, " ")] = out + "\x00" + readOutput(argv)
	}
	if got := gitOut(t, dir, "status", "--porcelain"); got != "" {
		t.Errorf("git status --porcelain after the reads:\n%s", got)
	}
	if digest(t, dir, nil) != before {
		t.Error("a read command changed the working tree or .git")
	}
	for _, argv := range lines {
		key := strings.Join(argv, " ")
		want, ok := first[key]
		if !ok {
			continue
		}
		if code, out, _ := runReal(t, argv...); code != ExitOK || out+"\x00"+readOutput(argv) != want {
			t.Errorf("%v: the second run differs from the first", argv)
		}
	}
	if len(first) < 10 {
		t.Errorf("only %d lines ran", len(first))
	}
}

// readOutput returns the file the line's --output or -o names, or "".
func readOutput(argv []string) string {
	for i, a := range argv {
		if (a == "-o" || a == "--output") && i+1 < len(argv) && argv[i+1] != "-" {
			b, _ := os.ReadFile(argv[i+1])
			return string(b)
		}
		if p, ok := strings.CutPrefix(a, "--output="); ok && p != "-" {
			b, _ := os.ReadFile(p)
			return string(b)
		}
	}
	return ""
}

// TestRealMatchesTablo is T17: the JSON of each view's Result.Data equals
// the data of `tablo <view> --json` run with the same options.
func TestRealMatchesTablo(t *testing.T) {
	dir := corpusCopy(t)
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go is not installed")
	}
	fixClock(t)
	t.Setenv("GIT_AUTHOR_EMAIL", "nbyoung@nbyoung.com")
	tablo := filepath.Join(t.TempDir(), "tablo")
	build := exec.Command(goBin, "build", "-o", tablo, "github.com/nbyoung/tablo/cmd/tablo")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build tablo: %v\n%s", err, out)
	}

	const id = "a1c0" // the corpus entry's root task
	lines := [][]string{
		{"gates"}, {"gates", "--task", id},
		{"task", id}, {"task", id, "--person", "nbyoung@nbyoung.com"},
		{"authority"}, {"authority", "--proposed"}, {"authority", "--task", id},
		{"assignment"}, {"assignment", "--person", "nbyoung@nbyoung.com"},
		{"queue"}, {"queue", "--person", "nbyoung@nbyoung.com"}, {"queue", "--task", id},
		{"blockage"}, {"blockage", "--task", id},
		{"tableau"}, {"tableau", "--window", "2"}, {"tableau", "--columns", "design,implementation"},
		{"tableau", "--historical"}, {"tableau", "--person", "nbyoung@nbyoung.com"},
		{"context"}, {"context", "--task", id}, {"context", "--person", "nbyoung@nbyoung.com", "--window", "0"},
		{"history"}, {"history", "--task", id}, {"history", "--ref", "HEAD~1..HEAD"},
		{"audit"}, {"audit", "--stale", "3"}, {"audit", "--task", id},
	}
	var cases [][]string
	for _, l := range lines {
		for _, level := range []string{"glance", "detail", "provenance"} {
			cases = append(cases, append(append([]string(nil), l...), "--level", level))
		}
	}
	src := tabloSource{}
	unread := map[string]bool{}
	for _, l := range cases {
		argv := append([]string{"-C", dir}, l...)
		inv, err := invocationOf(argv)
		if err != nil {
			t.Fatalf("%v: %v", l, err)
		}
		if inv.Viewer {
			inv.Query.Person = "nbyoung@nbyoung.com" // the viewer, as Run resolves it
		}
		res, err := src.View(context.Background(), inv.Query)
		if err != nil {
			t.Errorf("%v: %v", l, err)
			continue
		}
		if res.Data == nil {
			t.Errorf("%v: no data", l)
			continue
		}
		got := normalise(t, mustMarshal(t, res.Data))

		targs := append([]string{"-C", dir}, l...)
		targs = append(targs, "--json")
		if inv.Query.View == "audit" {
			targs = append(targs, "--now", today())
		}
		cmd := exec.Command(tablo, targs...)
		cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
		out, err := cmd.Output()
		if err != nil {
			if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
				t.Errorf("tablo %v: %v", targs, err)
				continue
			}
		}
		var env struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(out, &env); err != nil {
			t.Errorf("tablo %v: %v", targs, err)
			continue
		}
		want := normalise(t, env.Data)
		if err := agree(got, want, "", unread); err != "" {
			t.Errorf("%v: Result.Data and tablo's data differ at %s\n tabloio %s\n tablo   %s", l, err, abbreviate(got), abbreviate(want))
		}
	}
	var keys []string
	for k := range unread {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	t.Logf("keys of tablo's data that the view types do not carry: %s", strings.Join(keys, " "))
}

// agree compares tabloio's view with tablo's data, both decoded generically,
// and returns the path of the first difference, or "". The view types carry
// what the renderers read, with no omitempty, so the comparison reads them
// as a projection of tablo's data: a key the types do not carry is noted in
// unread and left out; a key tablo omits must hold a zero value in the view;
// a null and an empty list are equal.
func agree(got, want any, path string, unread map[string]bool) string {
	switch g := got.(type) {
	case map[string]any:
		w, ok := want.(map[string]any)
		if !ok {
			if isZero(got) && isZero(want) {
				return ""
			}
			return path
		}
		for k, gv := range g {
			wv, present := w[k]
			if !present {
				if !isZero(gv) {
					return path + "/" + k
				}
				continue
			}
			if d := agree(gv, wv, path+"/"+k, unread); d != "" {
				return d
			}
		}
		for k := range w {
			if _, ok := g[k]; !ok {
				unread[strings.TrimPrefix(path+"/"+k, "/")] = true
			}
		}
		return ""
	case []any:
		w, ok := want.([]any)
		if !ok {
			if isZero(got) && isZero(want) {
				return ""
			}
			return path
		}
		if len(g) != len(w) {
			return path
		}
		for i := range g {
			if d := agree(g[i], w[i], path+"/*", unread); d != "" {
				return d + "[" + strconv.Itoa(i) + "]"
			}
		}
		return ""
	}
	if reflect.DeepEqual(got, want) || (isZero(got) && isZero(want)) {
		return ""
	}
	return path
}

// isZero reports whether v is JSON null, an empty list or object, "", 0 or
// false, or a list or object that holds only such values.
func isZero(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case float64:
		return x == 0
	case bool:
		return !x
	case []any:
		for _, e := range x {
			if !isZero(e) {
				return false
			}
		}
		return true
	case map[string]any:
		for _, e := range x {
			if !isZero(e) {
				return false
			}
		}
		return true
	}
	return false
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// normalise decodes JSON into generic values, which compare without regard
// to key order or spacing.
func normalise(t *testing.T, b []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

// abbreviate shortens a value's JSON for an error message.
func abbreviate(v any) string {
	b, _ := json.Marshal(v)
	if len(b) > 400 {
		return string(b[:400]) + "…"
	}
	return string(b)
}

// TestRealMockups is T18, which the integrate gate runs: it needs the corpus
// entry tableaux-tooling, which the built corpus does not hold.
func TestRealMockups(t *testing.T) {
	t.Skip("T18 needs the corpus entry tableaux-tooling, which the built corpus lacks; the integrate gate runs it")
}

// TestRealNotFound: a ref or a task the project lacks is a not-found error
// of tablo, which the command reports as a failed read with exit 3.
func TestRealNotFound(t *testing.T) {
	dir := corpusCopy(t)
	for _, argv := range [][]string{
		{"-C", dir, "tableau", "--ref", "nope"},
		{"-C", dir, "task", "0000"},
		{"-C", dir, "gates", "--task", "0000"},
	} {
		if code, out, errs := runReal(t, argv...); code != ExitFailure || out != "" || !strings.HasPrefix(errs, "tabloio ") {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", argv, code, out, errs)
		}
	}
}
