package text_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/text"
)

func cell(s string) doc.Cell { return doc.Cell{Text: []doc.Inline{doc.Text(s)}} }

func cells(ss ...string) []doc.Cell {
	var out []doc.Cell
	for _, s := range ss {
		out = append(out, cell(s))
	}
	return out
}

func draw(t *testing.T, d doc.Doc, w int) string {
	t.Helper()
	var b bytes.Buffer
	if err := text.Write(&b, d, text.Options{Width: w}); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// T6, at the edge: a width that fits exactly, and one cell less.
func TestTableEdge(t *testing.T) {
	d := doc.Doc{Blocks: []doc.Block{doc.Table{
		Head: cells("Id", "Title"),
		Rows: [][]doc.Cell{cells("a1", "two words")},
	}}}
	// natural: (2+2) + (9+2) + 3 bars = 18
	padded := draw(t, d, 20)
	if !strings.Contains(padded, "│ a1 │ two words │") {
		t.Errorf("padded form missing:\n%s", padded)
	}
	// Padded needs 18 cells, but the minimum width is 20; use a wider table.
	d2 := doc.Doc{Blocks: []doc.Block{doc.Table{
		Head: cells("Id", "Title"),
		Rows: [][]doc.Cell{cells("a1", "abcdefgh ijklmnopqrstu")},
	}}}
	// padded: (2+2) + (22+2) + 3 = 31
	for _, c := range []struct {
		w      int
		padded bool
	}{{31, true}, {30, false}} {
		out := draw(t, d2, c.w)
		if got := strings.Contains(out, "│ a1 │ abcdefgh ijklmnopqrstu │"); got != c.padded {
			t.Errorf("width %d: padded = %v\n%s", c.w, got, out)
		}
		for _, ln := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
			if n := len([]rune(ln)); n > c.w {
				t.Errorf("width %d: line of %d cells: %s", c.w, n, ln)
			}
		}
	}
}

// T7: the row rule and the alignments.
func TestRowRule(t *testing.T) {
	const w = 28
	wide := "alpha beta gamma delta epsilon"
	rules := func(rows ...[]doc.Cell) int {
		out := draw(t, doc.Doc{Blocks: []doc.Block{doc.Table{Head: cells("Id", "Note"), Rows: rows}}}, w)
		return strings.Count(out, "├")
	}
	if n := rules(cells("a", wide), cells("b", wide)); n != 1 {
		t.Errorf("distinct first cells: %d rules, want only the header's", n)
	}
	if n := rules(cells("a", wide), cells("", wide)); n != 2 {
		t.Errorf("empty first cell: %d rules, want 2", n)
	}
	if n := rules(cells("a b c d e f g h i j k l m", "x"), cells("b", wide)); n != 2 {
		t.Errorf("wrapping first cell: %d rules, want 2", n)
	}
	d := doc.Doc{Blocks: []doc.Block{doc.Table{
		Align: []doc.Align{doc.Left, doc.Centre, doc.Right},
		Head:  cells("left", "centre", "right"),
		Rows:  [][]doc.Cell{cells("a", "b", "c")},
	}}}
	want := strings.Join([]string{
		"┌──────┬────────┬───────┐",
		"│ left │ centre │ right │",
		"├──────┼────────┼───────┤",
		"│ a    │   b    │     c │",
		"└──────┴────────┴───────┘",
		"",
	}, "\n")
	if got := draw(t, d, 80); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	// An odd gap: the smaller half goes left.
	d.Blocks[0] = doc.Table{Align: []doc.Align{doc.Centre}, Head: cells("abcd"), Rows: [][]doc.Cell{cells("x")}}
	if got := draw(t, d, 80); !strings.Contains(got, "│  x   │") {
		t.Errorf("odd gap:\n%s", got)
	}
}
