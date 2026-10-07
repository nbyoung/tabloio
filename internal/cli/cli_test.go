package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// fakeSource is a Source that records what it is asked and returns what the
// test sets.
type fakeSource struct {
	queries    []Query
	locks      []string // GIT_OPTIONAL_LOCKS at each View call
	viewerDirs []string // the dir of each Viewer call
	viewer     string
	viewerErr  error
	result     Result
	viewErr    error
}

func newSource() *fakeSource {
	return &fakeSource{viewer: "viewer@example.org", result: Result{Data: "the data"}}
}

func (f *fakeSource) View(_ context.Context, q Query) (Result, error) {
	f.queries = append(f.queries, q)
	f.locks = append(f.locks, os.Getenv("GIT_OPTIONAL_LOCKS"))
	return f.result, f.viewErr
}

func (f *fakeSource) Viewer(_ context.Context, dir string) (string, error) {
	f.viewerDirs = append(f.viewerDirs, dir)
	return f.viewer, f.viewerErr
}

// untouched reports whether the Source was never asked anything.
func (f *fakeSource) untouched() bool { return len(f.queries) == 0 && len(f.viewerDirs) == 0 }

// fakeRender returns bytes that spell its arguments.
func fakeRender(name string) renderFunc {
	return func(data any, level, linkBase string) ([]byte, error) {
		return []byte(fmt.Sprintf("fake %s: data=%v level=%s base=%q\n", name, data, level, linkBase)), nil
	}
}

// fakeTables replaces the two tables with fakes for the test's length: the
// ten views and brief in markdown, and the given names in text.
func fakeTables(t *testing.T, withBrief bool, textNames ...string) {
	t.Helper()
	oldMarkdown, oldText := markdown, text
	t.Cleanup(func() { markdown, text = oldMarkdown, oldText })
	markdown, text = map[string]renderFunc{}, map[string]renderFunc{}
	for _, v := range views {
		markdown[v.name] = fakeRender("markdown " + v.name)
	}
	if withBrief {
		markdown["brief"] = fakeRender("markdown brief")
	}
	for _, n := range textNames {
		text[n] = fakeRender("text " + n)
	}
}

// run calls Run with the source and returns the exit code and both outputs.
// It keeps the process environment as it found it.
func run(t *testing.T, src Source, argv ...string) (code int, stdout, stderr string) {
	t.Helper()
	t.Setenv("GIT_OPTIONAL_LOCKS", "")
	var out, errb bytes.Buffer
	code = Run(context.Background(), argv, &Env{Stdout: &out, Stderr: &errb, Version: "1.2.3", Source: src})
	return code, out.String(), errb.String()
}

func golden(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// ownOptions is the "Options beyond the common ones" column of design 5ca9's
// command table, spelled as a help line spells them.
var ownOptions = map[string][]string{
	"gates":      {"--task <id>"},
	"task":       {"--person <email>"},
	"authority":  {"--task <id>", "--person <email>", "--proposed"},
	"assignment": {"--task <id>", "--person <email>"},
	"queue":      {"--task <id>", "--person <email>", "--brief <id> <gate>"},
	"blockage":   {"--task <id>", "--person <email>"},
	"tableau":    {"--person <email>", "--window <n>", "--columns <gates>", "--historical"},
	"context":    {"--task <id>", "--person <email>", "--window <n>", "--columns <gates>", "--historical"},
	"history":    {"--task <id>", "--person <email>"},
	"audit":      {"--task <id>", "--person <email>", "--stale <days>"},
	"version":    nil,
	"help":       nil,
}

var commonOptions = []string{"--ref <ref>", "--level <level>", "--markdown", "--text", "-o, --output <file>"}

var helpLine = regexp.MustCompile(`^  (.+?) {2,}\S`)

// TestHelp is T13.
func TestHelp(t *testing.T) {
	usage := golden(t, "usage.txt")
	tableau := golden(t, "usage-tableau.txt")

	t.Run("help command", func(t *testing.T) {
		for _, argv := range [][]string{{"help"}, {"--help"}, {"-h"}, {"-C", "d", "--help"}} {
			code, out, errs := run(t, newSource(), argv...)
			if code != ExitOK || out != usage || errs != "" {
				t.Errorf("%v: exit %d, stderr %q, stdout\n%s", argv, code, errs, out)
			}
		}
	})
	t.Run("no command", func(t *testing.T) {
		code, out, errs := run(t, newSource())
		if code != ExitUsage || out != "" || errs != usage {
			t.Errorf("exit %d, stdout %q, stderr\n%s", code, out, errs)
		}
	})
	t.Run("one command", func(t *testing.T) {
		for _, argv := range [][]string{
			{"tableau", "--help"}, {"--help", "tableau"}, {"tableau", "-h"}, {"help", "tableau"},
			{"tableau", "--bogus", "--help"}, {"tableau", "--level", "nope", "--task", "x", "--help"},
		} {
			code, out, errs := run(t, newSource(), argv...)
			if code != ExitOK || out != tableau || errs != "" {
				t.Errorf("%v: exit %d, stderr %q, stdout\n%s", argv, code, errs, out)
			}
		}
	})
	t.Run("help after the double hyphen is an operand", func(t *testing.T) {
		code, out, errs := run(t, newSource(), "tableau", "--", "--help")
		if code != ExitUsage || out != "" || !strings.Contains(errs, "unexpected argument") {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
	t.Run("an unknown command", func(t *testing.T) {
		for _, argv := range [][]string{{"help", "tableu"}, {"tableu", "--help"}} {
			code, out, errs := run(t, newSource(), argv...)
			if code != ExitUsage || out != "" || !strings.Contains(errs, `unknown command "tableu"`) {
				t.Errorf("%v: exit %d, stdout %q, stderr %q", argv, code, out, errs)
			}
		}
		if code, _, errs := run(t, newSource(), "help", "tableau", "gates"); code != ExitUsage || !strings.Contains(errs, "unexpected argument") {
			t.Errorf("two operands: exit %d, stderr %q", code, errs)
		}
	})
	t.Run("each command lists exactly its options", func(t *testing.T) {
		if len(ownOptions) != len(commands) {
			t.Fatalf("the test knows %d commands, the table %d", len(ownOptions), len(commands))
		}
		for _, c := range commands {
			code, out, _ := run(t, newSource(), "help", c.Name)
			if code != ExitOK {
				t.Fatalf("help %s: exit %d", c.Name, code)
			}
			want := []string{"-C <dir>"}
			if c.Group == "read" {
				want = append(append(want, commonOptions...), ownOptions[c.Name]...)
			}
			want = append(want, "-h, --help")
			var got []string
			for _, l := range strings.Split(out, "\n") {
				if m := helpLine.FindStringSubmatch(l); m != nil {
					got = append(got, m[1])
				}
			}
			if strings.Join(got, "|") != strings.Join(want, "|") {
				t.Errorf("%s lists\n  %v\nwant\n  %v", c.Name, got, want)
			}
			wantUsage := "usage: tabloio [-C <dir>] " + c.Name
			if c.Operands != "" {
				wantUsage += " " + c.Operands
			}
			if first, _, _ := strings.Cut(out, "\n"); first != wantUsage+" [<options>]" {
				t.Errorf("%s: usage line %q", c.Name, first)
			}
		}
	})
	t.Run("the usage lists the commands by group in table order", func(t *testing.T) {
		var names []string
		for _, l := range strings.Split(usage, "\n") {
			if strings.HasPrefix(l, "  ") {
				names = append(names, strings.Fields(l)[0])
			}
		}
		var want []string
		for _, c := range commands {
			want = append(want, c.Name)
		}
		if strings.Join(names, " ") != strings.Join(want, " ") {
			t.Errorf("usage names %v, table %v", names, want)
		}
	})
}

func TestVersion(t *testing.T) {
	code, out, errs := run(t, newSource(), "version")
	if code != ExitOK || out != "tabloio 1.2.3\n" || errs != "" {
		t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
	}
	for _, argv := range [][]string{{"version", "extra"}, {"version", "--level", "glance"}, {"version", "--bogus"}} {
		code, out, errs := run(t, newSource(), argv...)
		if code != ExitUsage || out != "" || !strings.HasPrefix(errs, "tabloio version: ") {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", argv, code, out, errs)
		}
	}
}

func TestUsageErrorShape(t *testing.T) {
	_, _, errs := run(t, newSource(), "tableau", "--bogus")
	if want := "tabloio tableau: unknown option --bogus\nRun 'tabloio tableau --help' for its options.\n"; errs != want {
		t.Errorf("stderr %q, want %q", errs, want)
	}
	_, _, errs = run(t, newSource(), "tableu")
	if want := "tabloio: unknown command \"tableu\"\nRun 'tabloio help' for the commands.\n"; errs != want {
		t.Errorf("stderr %q, want %q", errs, want)
	}
}
