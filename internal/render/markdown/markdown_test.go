package markdown

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/doc"
)

func render(t *testing.T, blocks ...doc.Block) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Write(&buf, doc.Doc{Blocks: blocks}); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func txt(s string) []doc.Inline { return []doc.Inline{doc.Text(s)} }

// T9: escaping, in a paragraph, a cell, a list item and a heading.
func TestEscaping(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"backslash", `a\b`, `a\\b`},
		{"backtick", "a`b", "a\\`b"},
		{"star", "a*b", `a\*b`},
		{"underscore", "a_b", `a\_b`},
		{"open bracket", "a[b", `a\[b`},
		{"close bracket", "a]b", `a\]b`},
		{"angle", "a<b", `a\<b`},
		{"bar", "a|b", `a\|b`},
		{"ampersand", "a&b", "a&amp;b"},
		{"white space run", "a \t b", "a b"},
		{"line break", "a\nb\r\n\r\nc", "a b c"},
		{"plain", "a#b>c+d-e.f", "a#b>c+d-e.f"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := render(t,
				doc.Para{Text: txt(c.in)},
				doc.Table{Head: []doc.Cell{{Text: txt("H")}}, Rows: [][]doc.Cell{{{Text: txt(c.in)}}}},
				doc.List{Items: []doc.Item{{Text: txt(c.in)}}},
				doc.Heading{Level: 2, Text: txt(c.in)},
			)
			want := c.want + "\n\n| H |\n|---|\n| " + c.want + " |\n\n- " + c.want + "\n\n## " + c.want + "\n"
			if got != want {
				t.Errorf("got\n%q\nwant\n%q", got, want)
			}
		})
	}
}

func TestEscapingAtStart(t *testing.T) {
	cases := []struct{ in, want string }{
		{"# title", `\# title`},
		{"> quote", `\> quote`},
		{"+ item", `\+ item`},
		{"- item", `\- item`},
		{"12. step", `12\. step`},
		{"3) step", `3\) step`},
		{"  - spaced", `\- spaced`},
		{"1 a", "1 a"},
		{"a - b", "a - b"},
	}
	for _, c := range cases {
		got := render(t, doc.Para{Text: txt(c.in)}, doc.List{Items: []doc.Item{{Text: txt(c.in)}}})
		want := c.want + "\n\n- " + c.want + "\n"
		if got != want {
			t.Errorf("%q: got %q want %q", c.in, got, want)
		}
	}
	// Only the start of a block is special; a cell and a heading are not.
	got := render(t, doc.Heading{Level: 2, Text: txt("# a")})
	if got != "## # a\n" {
		t.Errorf("heading: %q", got)
	}
}

func TestInlines(t *testing.T) {
	cases := []struct {
		name string
		in   []doc.Inline
		want string
	}{
		{"code", []doc.Inline{doc.Code("e9c6")}, "`e9c6`"},
		{"code with a backtick", []doc.Inline{doc.Code("a`b")}, "`` a`b ``"},
		{"code with a run of two", []doc.Inline{doc.Code("a``b")}, "``` a``b ```"},
		{"code with a bar", []doc.Inline{doc.Code("a|b")}, `` + "`a\\|b`"},
		{"code with a line break", []doc.Inline{doc.Code("a\nb")}, "`a b`"},
		{"empty code", []doc.Inline{doc.Text("a"), doc.Code(""), doc.Text("b")}, "ab"},
		{"symbol", []doc.Inline{doc.Symbol("⚙️"), doc.Text(" "), doc.Text("function")}, "⚙️ function"},
		{"strong", []doc.Inline{doc.Strong{doc.Text("a_b")}}, `**a\_b**`},
		{"empty strong", []doc.Inline{doc.Strong{}}, ""},
		{"link", []doc.Inline{doc.Link{Text: txt("a"), URL: "x/y#z"}}, "[a](x/y#z)"},
		{"link with a space and parentheses", []doc.Inline{doc.Link{Text: []doc.Inline{doc.Code("c")}, URL: "a b/(c).md"}}, "[`c`](a%20b/%28c%29.md)"},
		{"spaces between spans", []doc.Inline{doc.Text("a "), doc.Code("b"), doc.Text(" c")}, "a `b` c"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := render(t, doc.Para{Text: c.in})
			want := c.want + "\n"
			if c.want == "" {
				want = ""
			}
			if got != want {
				t.Errorf("got %q want %q", got, want)
			}
		})
	}
}

func TestPre(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  string
	}{
		{"plain", []string{"git log", "git show # c"}, "```\ngit log\ngit show # c\n```\n"},
		{"fence inside", []string{"a ``` b"}, "````\na ``` b\n````\n"},
		{"longer fence inside", []string{"`", "`````"}, "``````\n`\n`````\n``````\n"},
		{"empty", nil, "```\n```\n"},
	}
	for _, c := range cases {
		if got := render(t, doc.Pre{Lines: c.lines}); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestTable(t *testing.T) {
	got := render(t, doc.Table{
		Align: []doc.Align{doc.Right, doc.Centre, doc.Left},
		Head:  []doc.Cell{{Text: txt("#")}, {Text: txt("Symbol")}, {Text: txt("Name")}},
		Rows: [][]doc.Cell{
			{{Text: txt("1")}, {Text: []doc.Inline{doc.Symbol("❔")}}, {}},
			{{Text: txt("2")}, {}, {Indent: 2, Text: txt("x")}},
			{{}, {}, {}},
			{{Text: txt("short")}},
		},
	})
	want := "| # | Symbol | Name |\n" +
		"|--:|:-:|---|\n" +
		"| 1 | ❔ |  |\n" +
		"| 2 |  | &nbsp;&nbsp;&nbsp;&nbsp;x |\n" +
		"|  |  |  |\n" +
		"| short |  |  |\n"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestList(t *testing.T) {
	got := render(t, doc.List{
		Items: []doc.Item{
			{Text: txt("one"), Blocks: []doc.Block{
				doc.List{Items: []doc.Item{{Text: txt("nested")}}},
				doc.Pre{Lines: []string{"cmd"}},
			}},
			{Text: txt("two")},
			{},
		},
	}, doc.List{Ordered: true, Items: []doc.Item{{Text: txt("a")}, {Text: txt("b")}}})
	want := "- one\n  - nested\n  ```\n  cmd\n  ```\n- two\n-\n\n1. a\n2. b\n"
	if got != want {
		t.Errorf("got\n%q\nwant\n%q", got, want)
	}
}

func TestBlocks(t *testing.T) {
	if got := render(t); got != "" {
		t.Errorf("an empty document prints %q", got)
	}
	got := render(t,
		doc.Heading{Level: 1, Text: txt("T")},
		doc.Para{Text: txt("p")},
		doc.Para{},
		doc.Heading{Level: 9, Text: txt("deep")},
		doc.Heading{Level: 0},
	)
	if want := "# T\n\np\n\n###### deep\n\n#\n"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

type failing struct{ left int }

var errFull = errors.New("full")

func (f *failing) Write(p []byte) (int, error) {
	if len(p) > f.left {
		n := f.left
		f.left = 0
		return n, errFull
	}
	f.left -= len(p)
	return len(p), nil
}

func TestWriteError(t *testing.T) {
	d := doc.Doc{Blocks: []doc.Block{doc.Para{Text: txt("some text")}}}
	if err := Write(&failing{left: 4}, d); err != errFull {
		t.Errorf("got %v, want the writer's own error", err)
	}
	if err := Write(&failing{left: 100}, d); err != nil {
		t.Error(err)
	}
}

type short struct{}

func (short) Write(p []byte) (int, error) { return len(p) - 1, nil }

func TestShortWrite(t *testing.T) {
	d := doc.Doc{Blocks: []doc.Block{doc.Para{Text: txt("x")}}}
	if err := Write(short{}, d); err == nil {
		t.Error("a short write passes")
	}
}

// T8: the table layout, over every golden file under testdata.
func TestGoldenLayout(t *testing.T) {
	root := filepath.Join("..", "testdata")
	n := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		n++
		t.Run(strings.TrimPrefix(path, root+string(filepath.Separator)), func(t *testing.T) {
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			checkLayout(t, string(b))
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("no golden file")
	}
}

// checkLayout checks one rendering: it ends in one newline, no line ends in
// a space, and each row of a table holds as many cells as the header.
func checkLayout(t *testing.T, s string) {
	t.Helper()
	if !strings.HasSuffix(s, "\n") || strings.HasSuffix(s, "\n\n") {
		t.Error("the file does not end in exactly one newline")
	}
	if strings.Contains(s, "\r") {
		t.Error("the file holds a carriage return")
	}
	fenced, width := false, 0
	for i, ln := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		if strings.HasSuffix(ln, " ") || strings.HasSuffix(ln, "\t") {
			t.Errorf("line %d ends in white space: %q", i+1, ln)
		}
		if strings.HasPrefix(ln, "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		if !strings.HasPrefix(ln, "|") {
			width = 0
			continue
		}
		n := cellCount(ln)
		if width == 0 {
			width = n // the header row
		} else if n != width {
			t.Errorf("line %d has %d cells, the header has %d: %q", i+1, n, width, ln)
		}
	}
}

// cellCount counts the cells of a table row: the bars not escaped and not
// inside a code span, less one.
func cellCount(row string) int {
	bars, inCode, run := 0, false, 0
	fence := 0
	r := []rune(row)
	for i := 0; i < len(r); i++ {
		switch {
		case r[i] == '\\' && i+1 < len(r):
			i++
		case r[i] == '`':
			run = 1
			for i+1 < len(r) && r[i+1] == '`' {
				run++
				i++
			}
			if !inCode {
				inCode, fence = true, run
			} else if run == fence {
				inCode = false
			}
		case r[i] == '|' && !inCode:
			bars++
		}
	}
	return bars - 1
}

func TestCellCount(t *testing.T) {
	for row, want := range map[string]int{
		"| a | b |":      2,
		"|  |  | c |":    3,
		`| a\|b | c |`:   2,
		"| `a|b` | c |":  2,
		"| `` a`|b `` |": 1,
		"|---|:-:|--:|":  3,
		"| [x](a%20b) |": 1,
	} {
		if got := cellCount(row); got != want {
			t.Errorf("%q: %d, want %d", row, got, want)
		}
	}
}
