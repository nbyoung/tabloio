package render

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/markdown"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// fixtureName names one fixture file: testdata/<view>/<name>.json.
type fixtureName struct{ view, name string }

func (f fixtureName) String() string { return f.view + "/" + f.name }

// fixtures lists every fixture under testdata.
func fixtures(t testing.TB) []fixtureName {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("testdata", "*", "*.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no fixture: %v", err)
	}
	var out []fixtureName
	for _, p := range paths {
		out = append(out, fixtureName{
			view: filepath.Base(filepath.Dir(p)),
			name: strings.TrimSuffix(filepath.Base(p), ".json"),
		})
	}
	return out
}

// levelsOf lists the levels from glance up to the one the data holds.
func levelsOf(v view.View) []view.Level {
	var out []view.Level
	for l := view.Glance; l <= v.Header().Level; l++ {
		out = append(out, l)
	}
	return out
}

func buildFor(t testing.TB, v view.View, o Options) doc.Doc {
	t.Helper()
	d, err := buildDoc(v, o)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// A plain-text writer for the document model, used by T6 and T15. It prints
// every block and inline kind as plain lines and records the kinds it meets.

func plainInline(xs []doc.Inline, seen map[string]bool) string {
	var b strings.Builder
	for _, x := range xs {
		switch x := x.(type) {
		case doc.Text:
			seen["Text"] = true
			b.WriteString(string(x))
		case doc.Code:
			seen["Code"] = true
			b.WriteString(string(x))
		case doc.Symbol:
			seen["Symbol"] = true
			b.WriteString(string(x))
		case doc.Strong:
			seen["Strong"] = true
			b.WriteString(plainInline(x, seen))
		case doc.Link:
			seen["Link"] = true
			b.WriteString(plainInline(x.Text, seen) + " <" + x.URL + ">")
		default:
			panic(fmt.Sprintf("unknown inline %T", x))
		}
	}
	return b.String()
}

func plainBlock(b doc.Block, seen map[string]bool) []string {
	var lines []string
	switch b := b.(type) {
	case doc.Heading:
		seen["Heading"] = true
		lines = append(lines, strings.Repeat("=", b.Level)+" "+plainInline(b.Text, seen))
	case doc.Para:
		seen["Para"] = true
		lines = append(lines, plainInline(b.Text, seen))
	case doc.Table:
		seen["Table"] = true
		row := func(cells []doc.Cell) string {
			var out []string
			for _, c := range cells {
				out = append(out, strings.Repeat("  ", c.Indent)+plainInline(c.Text, seen))
			}
			return strings.Join(out, "\t")
		}
		lines = append(lines, row(b.Head))
		for _, r := range b.Rows {
			lines = append(lines, row(r))
		}
	case doc.List:
		seen["List"] = true
		for i, it := range b.Items {
			marker := "* "
			if b.Ordered {
				marker = strconv.Itoa(i+1) + ". "
			}
			lines = append(lines, marker+plainInline(it.Text, seen))
			for _, nb := range it.Blocks {
				for _, ln := range plainBlock(nb, seen) {
					lines = append(lines, "  "+ln)
				}
			}
		}
	case doc.Pre:
		seen["Pre"] = true
		lines = append(lines, b.Lines...)
	default:
		panic(fmt.Sprintf("unknown block %T", b))
	}
	return lines
}

func plainDoc(d doc.Doc, seen map[string]bool) []string {
	var lines []string
	for _, b := range d.Blocks {
		lines = append(lines, plainBlock(b, seen)...)
		lines = append(lines, "")
	}
	return lines
}

// T6: levels nest. For every fixture, each heading and each table header
// cell at one level appears at the next, in order.
func TestLevelsNest(t *testing.T) {
	for _, f := range fixtures(t) {
		t.Run(f.String(), func(t *testing.T) {
			v := readFixture(t, f.view, f.name)
			var prev []string
			for _, l := range levelsOf(v) {
				d := buildFor(t, v, Options{Level: l})
				var items []string
				for _, b := range d.Blocks {
					switch b := b.(type) {
					case doc.Heading:
						items = append(items, "heading "+plainInline(b.Text, map[string]bool{}))
					case doc.Table:
						for _, c := range b.Head {
							items = append(items, "column "+plainInline(c.Text, map[string]bool{}))
						}
					}
				}
				if l > view.Glance && !isSubsequence(prev, items) {
					t.Errorf("%s does not hold %s:\n%v\n%v", l, l-1, items, prev)
				}
				prev = items
			}
		})
	}
}

func isSubsequence(sub, of []string) bool {
	i := 0
	for _, x := range of {
		if i < len(sub) && sub[i] == x {
			i++
		}
	}
	return i == len(sub)
}

// T7: the renderer invents no symbol. Each rune outside ASCII in the output
// occurs in the fixture, or is one of · … → › — ×.
func TestNoInventedSymbol(t *testing.T) {
	const own = "·…→›—×"
	for _, f := range fixtures(t) {
		t.Run(f.String(), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", f.view, f.name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			v := readFixture(t, f.view, f.name)
			for _, l := range levelsOf(v) {
				out := renderString(t, v, Options{Level: l, Links: Links{Project: "."}})
				for _, r := range out {
					if r < 0x80 || strings.ContainsRune(own, r) || strings.ContainsRune(string(raw), r) {
						continue
					}
					t.Errorf("%s: the rune %q (U+%04X) is in the output and not in the fixture", l, r, r)
				}
			}
		})
	}
}

// T10: determinism. The same data and options give the same bytes, with
// Links.Views built in another order.
func TestDeterminism(t *testing.T) {
	names := []string{"gates", "history", "blockage", "authority", "assignment", "queue", "context", "tableau", "audit"}
	forward, backward := map[string]string{}, map[string]string{}
	for i, n := range names {
		forward[n] = strings.ToUpper(n) + ".md"
		backward[names[len(names)-1-i]] = strings.ToUpper(names[len(names)-1-i]) + ".md"
	}
	for _, f := range fixtures(t) {
		t.Run(f.String(), func(t *testing.T) {
			v := readFixture(t, f.view, f.name)
			for _, l := range levelsOf(v) {
				a := renderString(t, v, Options{Level: l, Links: Links{Project: ".", Views: forward}})
				b := renderString(t, v, Options{Level: l, Links: Links{Project: ".", Views: backward}})
				if a != b {
					t.Errorf("%s: two renderings differ", l)
				}
				if c := renderString(t, v, Options{Level: l, Links: Links{Project: ".", Views: forward}}); a != c {
					t.Errorf("%s: a repeated rendering differs", l)
				}
			}
		})
	}
}

// T11: the ref and the commit. One fixture four ways.
func TestRefAndCommit(t *testing.T) {
	v := readFixture(t, "task", "e9c6").(*view.Task)
	const person = "--person nbyoung@nbyoung.com"
	cases := []struct {
		name   string
		mutate func(*view.Task)
		o      Options
		frame  string // in the parameter line
		cmd    string
		absent string // in neither
	}{
		{"glance", nil, Options{Level: view.Glance},
			"Tableaux tooling · ref `main` · task", "`tabloio task e9c6 " + person + " --ref main --level glance`", "3cdae52"},
		{"glance with a stamp", nil, Options{Level: view.Glance, Stamp: true},
			"ref `main` at `3cdae52`, 2026-10-05 · task", "`tabloio task e9c6 " + person + " --ref main --level glance`", ""},
		{"provenance", nil, Options{Level: view.Provenance},
			"ref `main` at `3cdae52`, 2026-10-05 · task", "`tabloio task e9c6 " + person + " --ref 3cdae52 --level provenance`", ""},
		{"HEAD at glance", func(v *view.Task) { v.Ref.Name = "HEAD" }, Options{Level: view.Glance},
			"ref `HEAD` · task", "`tabloio task e9c6 " + person + " --level glance`", "--ref"},
		{"HEAD at provenance", func(v *view.Task) { v.Ref.Name = "HEAD" }, Options{Level: view.Provenance},
			"ref `HEAD` at `3cdae52`, 2026-10-05 · task", "`tabloio task e9c6 " + person + " --ref 3cdae52 --level provenance`", ""},
		{"off the trunk", func(v *view.Task) { v.Ref.OnTrunk = false }, Options{Level: view.Glance},
			"ref `main`, off the trunk · task", "`tabloio task e9c6 " + person + " --ref main --level glance`", ""},
		{"a range", func(v *view.Task) { v.Ref.From = "v1" }, Options{Level: view.Glance},
			"ref `v1..main` · task", "`tabloio task e9c6 " + person + " --ref v1..main --level glance`", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cp := *v
			if c.mutate != nil {
				c.mutate(&cp)
			}
			out := renderString(t, &cp, c.o)
			if !strings.Contains(out, c.frame) {
				t.Errorf("the frame lacks %q:\n%s", c.frame, strings.SplitN(out, "\n", 4)[2])
			}
			if !strings.Contains(out, "Command: "+c.cmd+"\n") {
				t.Errorf("the command lacks %s:\n%s", c.cmd, out[strings.LastIndex(out, "Command:"):])
			}
			if c.absent != "" && strings.Contains(out, c.absent) {
				t.Errorf("the output holds %q", c.absent)
			}
		})
	}
}

// TestCommand covers the flags no fixture of the two views sets.
func TestCommand(t *testing.T) {
	two := 2
	h := &view.Head{Ref: view.Ref{Name: "main", Commit: "0123456789abcdef"}}
	h.Params = view.Params{Task: "e9c6", Person: "a@b.c", Window: &two, Historical: true, Proposed: true, Stale: 7}
	got := command("context", "", h, Options{Level: view.Detail})
	want := "tabloio context --task e9c6 --person a@b.c --ref main --window 2 --historical --proposed --stale 7 --level detail"
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	h.Params = view.Params{Columns: []string{"design", "unit"}}
	if got, want := command("tableau", "", h, Options{Level: view.Provenance}),
		"tabloio tableau --ref 0123456 --columns design,unit --level provenance"; got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	h.Project = view.TaskRef{Title: "Plan"}
	h.Params = view.Params{Task: "e9c6", Person: "a@b.c", Window: &two, Historical: true, Proposed: true, Stale: 7}
	p := newPage(h, Options{Level: view.Glance})
	got = plainInline(p.paramLine("Q?").Text, map[string]bool{})
	want = "Q? Plan · ref main, off the trunk · task e9c6 · person a@b.c · window 2 · historical junctions on · proposed only · stale 7 days · level glance"
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

// T12: links.
func TestLinks(t *testing.T) {
	v := readFixture(t, "task", "e9c6").(*view.Task)
	cp := *v
	cp.References = append([]view.Reference{
		{Text: "Absolute", URL: "https://example.org/a b"},
		{URL: "docs/x.md"},
	}, v.References...)
	cases := []struct {
		name  string
		links Links
		want  []string
		not   []string
	}{
		{"no project", Links{}, []string{
			"# Task definition: `e9c6` Abstract views\n",
			"| `c2ad` Roles |",
			"- [Proposed views](PLAN.md#views)",
			"- [Absolute](https://example.org/a%20b)",
			"- [docs/x.md](docs/x.md)",
			"Legend: `tabloio gates`.\n",
		}, []string{".tableaux/tasks", "See also"}},
		{"this directory", Links{Project: "."}, []string{
			"# Task definition: [`e9c6` Abstract views](.tableaux/tasks/e9c6.yaml)\n",
			"| [`c2ad` Roles](.tableaux/tasks/c2ad.yaml) |",
			"| [`99f0` Gate definition view in Markdown](.tableaux/tasks/99f0.yaml) |",
			"- [Proposed views](PLAN.md#views)",
			"- [Absolute](https://example.org/a%20b)",
		}, []string{"| [`e9c6`", "(…"}},
		{"two levels up", Links{Project: "../.."}, []string{
			"# Task definition: [`e9c6` Abstract views](../../.tableaux/tasks/e9c6.yaml)\n",
			"| [`c2ad` Roles](../../.tableaux/tasks/c2ad.yaml) |",
			"- [Proposed views](../../PLAN.md#views)",
			"- [Absolute](https://example.org/a%20b)",
			"- [docs/x.md](../../docs/x.md)",
		}, nil},
		{"a view named", Links{Views: map[string]string{"gates": "GATES.md"}}, []string{
			"Legend: [gates](GATES.md).\n",
		}, []string{"See also"}},
		{"related views named", Links{Views: map[string]string{"gates": "GATES.md", "history": "docs/history.md", "audit": "AUDIT.md"}}, []string{
			"Legend: [gates](GATES.md).\n",
			"See also: [history](docs/history.md), `tabloio blockage`, `tabloio authority`, `tabloio assignment`, `tabloio queue`, `tabloio context`, `tabloio tableau`, [audit](AUDIT.md).\n",
		}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := renderString(t, &cp, Options{Level: view.Detail, Links: c.links})
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("the output lacks %q", w)
				}
			}
			for _, w := range c.not {
				if strings.Contains(out, w) {
					t.Errorf("the output holds %q", w)
				}
			}
		})
	}
	// At glance no See also line stands, whatever the links.
	out := renderString(t, v, Options{Level: view.Glance, Links: Links{Views: map[string]string{"history": "H.md"}}})
	if strings.Contains(out, "See also") {
		t.Error("a See also line stands at glance")
	}
}

// alike makes n dependents that agree in every cell but the task.
func alike(n int, from int) []view.Requirement {
	var out []view.Requirement
	for i := 0; i < n; i++ {
		out = append(out, view.Requirement{
			Task: view.TaskRef{ID: fmt.Sprintf("d%03d", from+i), Title: "Dependent"},
			From: "design", To: "mockup", Text: "The same text", Met: true, Due: true,
		})
	}
	return out
}

// tableRows returns the lines of the first table of a heading's section.
func section(out, heading string) []string {
	var lines []string
	in := false
	for _, ln := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(ln, "## "):
			in = strings.HasPrefix(ln, heading)
		case in && strings.HasPrefix(ln, "|"):
			lines = append(lines, ln)
		}
	}
	return lines
}

// T13: the empty forms and the fold.
func TestEmptyAndFold(t *testing.T) {
	v := readFixture(t, "task", "e9c6").(*view.Task)

	empty := *v
	empty.Requires, empty.Dependents = nil, nil
	out := renderString(t, &empty, Options{Level: view.Detail})
	for _, w := range []string{"## Requires: none\n\n## Dependents: none\n\n## Junctions\n"} {
		if !strings.Contains(out, w) {
			t.Errorf("an empty task lacks %q:\n%s", w, out)
		}
	}
	if got := section(out, "## Requires"); len(got) != 0 {
		t.Errorf("an empty section holds a table: %v", got)
	}

	differing := alike(9, 0)
	differing[4].Text = "A different text"

	cases := []struct {
		name   string
		deps   []view.Requirement
		unfold bool
		rows   int      // table lines, header and delimiter included
		fold   string   // the fold row, if any
		shown  []string // ids shown as the first cell
	}{
		{"eight alike", alike(8, 0), false, 2 + 8, "", nil},
		{"nine alike", alike(9, 0), false, 2 + 4,
			"| … and 6 more: `d003` `d004` `d005` `d006` `d007` `d008` | design → mockup | The same text | met |", []string{"d000", "d001", "d002"}},
		{"nine, one differs in the middle", differing, false, 2 + 9, "", nil},
		{"nine unfolded", alike(9, 0), true, 2 + 9, "", nil},
		{"twelve, then one that differs", append(alike(12, 0), view.Requirement{
			Task: view.TaskRef{ID: "x1", Title: "Other"}, From: "design", To: "design", Text: "Other", Met: false, Due: false}),
			false, 2 + 4 + 1,
			"| … and 9 more: `d003` `d004` `d005` `d006` `d007` `d008` `d009` `d010` `d011` | design → mockup | The same text | met |", []string{"d000", "d001", "d002", "x1"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cp := *v
			cp.Dependents = c.deps
			out := renderString(t, &cp, Options{Level: view.Detail, Unfold: c.unfold})
			rows := section(out, "## Dependents")
			if len(rows) != c.rows {
				t.Errorf("%d table lines, want %d:\n%s", len(rows), c.rows, strings.Join(rows, "\n"))
			}
			if c.fold == "" && strings.Contains(out, "more:") {
				t.Errorf("a fold stands:\n%s", strings.Join(rows, "\n"))
			}
			if c.fold != "" && !slices.Contains(rows, c.fold) {
				t.Errorf("the fold row is missing:\n%s", strings.Join(rows, "\n"))
			}
			for _, id := range c.shown {
				if !strings.Contains(strings.Join(rows, "\n"), "`"+id+"`") {
					t.Errorf("%s is not shown", id)
				}
			}
		})
	}
	// The fold counts "more", and no id of the folded rows is lost.
	cp := *v
	cp.Dependents = alike(30, 0)
	out = renderString(t, &cp, Options{Level: view.Detail})
	for i := 0; i < 30; i++ {
		if !strings.Contains(out, fmt.Sprintf("`d%03d`", i)) {
			t.Errorf("d%03d is lost from the fold", i)
		}
	}
}

type foreign struct{ view.Head }

// A writer that fails after n bytes.
type failAfter struct {
	n   int
	err error
}

func (f *failAfter) Write(p []byte) (int, error) {
	if len(p) <= f.n {
		f.n -= len(p)
		return len(p), nil
	}
	n := f.n
	f.n = 0
	return n, f.err
}

// T14: the errors.
func TestErrors(t *testing.T) {
	gates := readFixture(t, "gates", "draft") // holds detail
	task := readFixture(t, "task", "weather-3c5d")

	render := func(v view.View, o Options) error { return Render(&bytes.Buffer{}, v, o) }
	if err := render(gates, Options{Level: view.Provenance}); !errors.Is(err, ErrLevel) {
		t.Errorf("a detail fixture asked for provenance: %v", err)
	}
	glance := *gates.(*view.Gates)
	glance.Level = view.Glance
	if err := render(&glance, Options{Level: view.Detail}); !errors.Is(err, ErrLevel) {
		t.Errorf("a glance fixture asked for detail: %v", err)
	}
	if err := render(gates, Options{Level: view.Level(-1)}); !errors.Is(err, ErrLevel) {
		t.Errorf("a negative level: %v", err)
	}
	if err := render(task, Options{Level: view.Detail}); err != nil {
		t.Errorf("a detail fixture asked for detail: %v", err)
	}
	if err := render(gates, Options{Format: Text, Level: view.Glance}); !errors.Is(err, ErrFormat) {
		t.Errorf("Format Text: %v", err)
	}
	if err := render(gates, Options{Format: Format(9)}); !errors.Is(err, ErrFormat) {
		t.Errorf("an unknown format: %v", err)
	}
	if err := render(&foreign{}, Options{}); !errors.Is(err, ErrView) {
		t.Errorf("a foreign view: %v", err)
	}
	if err := render(nil, Options{}); !errors.Is(err, ErrView) {
		t.Errorf("a nil view: %v", err)
	}
	if err := render((*view.Gates)(nil), Options{}); !errors.Is(err, ErrView) {
		t.Errorf("a nil Gates: %v", err)
	}

	boom := errors.New("boom")
	var full bytes.Buffer
	if err := Render(&full, gates, Options{Level: view.Detail}); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{0, 1, full.Len() - 1} {
		w := &failAfter{n: n, err: boom}
		if err := Render(w, gates, Options{Level: view.Detail}); err != boom {
			t.Errorf("a writer that fails after %d bytes: got %v, want its own error", n, err)
		}
	}
	if err := Render(&failAfter{n: full.Len(), err: boom}, gates, Options{Level: view.Detail}); err != nil {
		t.Errorf("a writer with room fails: %v", err)
	}
}

// T15: the document model serves a second writer. A plain-text writer prints
// every block and inline kind the builders emit.
func TestSecondWriter(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range fixtures(t) {
		v := readFixture(t, f.view, f.name)
		for _, l := range levelsOf(v) {
			d := buildFor(t, v, Options{Level: l, Links: Links{Project: ".", Views: map[string]string{"gates": "GATES.md"}}})
			lines := plainDoc(d, seen)
			if len(lines) == 0 {
				t.Errorf("%s at %s: no line", f, l)
			}
			for _, ln := range lines {
				if strings.HasSuffix(ln, "\n") {
					t.Errorf("a line holds a line end: %q", ln)
				}
			}
		}
	}
	for _, kind := range []string{"Heading", "Para", "Table", "List", "Pre", "Text", "Code", "Symbol", "Strong", "Link"} {
		if !seen[kind] {
			t.Errorf("no builder emits a %s", kind)
		}
	}
}

func TestWorded(t *testing.T) {
	cases := []struct {
		in   string
		want []doc.Inline
	}{
		{"", nil},
		{"plain", []doc.Inline{doc.Text("plain")}},
		{"a `b` c", []doc.Inline{doc.Text("a "), doc.Code("b"), doc.Text(" c")}},
		{"`b`", []doc.Inline{doc.Code("b")}},
		{"a `b` and `c`", []doc.Inline{doc.Text("a "), doc.Code("b"), doc.Text(" and "), doc.Code("c")}},
		{"a ` b", []doc.Inline{doc.Text("a ` b")}},
		{"`a` `", []doc.Inline{doc.Code("a"), doc.Text(" `")}},
	}
	for _, c := range cases {
		got := worded(c.in)
		if len(got) != len(c.want) {
			t.Errorf("%q: %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q: %v, want %v", c.in, got, c.want)
			}
		}
	}
}

func TestStandsAndConditions(t *testing.T) {
	cases := []struct {
		j    view.Junction
		want string
	}{
		{view.Junction{Stands: "passed"}, "passed"},
		{view.Junction{Stands: "passed", Reviewed: true}, "passed, reviewed"},
		{view.Junction{Stands: "here"}, "passed; the status stands here"},
		{view.Junction{Stands: "here", Reviewed: true}, "passed, reviewed; the status stands here"},
		{view.Junction{Stands: "next"}, "next"},
		{view.Junction{Stands: "later", Reviewed: true}, "later"},
		{view.Junction{Stands: "exempt"}, "does not apply"},
	}
	for _, c := range cases {
		if got := stands(c.j); got != c.want {
			t.Errorf("%+v: %q, want %q", c.j, got, c.want)
		}
	}
	for _, c := range []struct {
		met, due bool
		want     string
	}{{true, true, "met"}, {false, true, "unmet"}, {false, false, "unmet, not yet due"}, {true, false, "met, not yet due"}} {
		if got := condition(c.met, c.due); got != c.want {
			t.Errorf("%v %v: %q, want %q", c.met, c.due, got, c.want)
		}
	}
}

// TestLegendKeys covers R2 and A2: a key with no symbol prints as the key.
func TestLegendKeys(t *testing.T) {
	v := *readFixture(t, "task", "e9c6").(*view.Task)
	v.Legend.States = []view.State{{Key: "nominal"}} // no symbol
	v.Legend.Gates = nil
	out := renderString(t, &v, Options{Level: view.Glance})
	if !strings.Contains(out, "| Status | design nominal |") {
		t.Errorf("a key with no symbol:\n%s", out)
	}
}

// TestEmptyForm covers R9's bold sentence, which the views of the sibling
// tasks use in place of a first table.
func TestEmptyForm(t *testing.T) {
	var buf bytes.Buffer
	d := doc.Doc{Blocks: []doc.Block{emptyForm("No cause holds any task."), countHeading(2, "Requires", 0), countHeading(2, "Requires", 3)}}
	if err := markdown.Write(&buf, d); err != nil {
		t.Fatal(err)
	}
	if want := "**No cause holds any task.**\n\n## Requires: none\n\n## Requires: 3\n"; buf.String() != want {
		t.Errorf("got %q want %q", buf.String(), want)
	}
}
