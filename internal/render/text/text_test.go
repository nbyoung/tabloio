package text

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/doc"
)

func tx(s string) []doc.Inline { return []doc.Inline{doc.Text(s)} }

func write(t *testing.T, d doc.Doc, width int) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Write(&buf, d, Options{Width: width}); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func para(in ...doc.Inline) doc.Doc { return doc.Doc{Blocks: []doc.Block{doc.Para{Text: in}}} }

// T2: cleaning.
func TestClean(t *testing.T) {
	const gear = "⚙️"
	const bolt = "⚡️" // two cells wide: keeps its selector
	for _, c := range []struct {
		name string
		in   doc.Doc
		want string
	}{
		{"text", para(doc.Text("a " + gear)), "a ⚙\n"},
		{"code", para(doc.Code(gear)), "⚙\n"},
		{"symbol", para(doc.Symbol(gear)), "⚙\n"},
		{"strong", para(doc.Strong{doc.Text(gear)}), "*⚙*\n"},
		{"link", para(doc.Link{Text: tx(gear), URL: "u" + gear}), "⚙ <u⚙>\n"},
		{"wide keeps it", para(doc.Symbol(bolt)), bolt + "\n"},
		{"pre keeps it", doc.Doc{Blocks: []doc.Block{doc.Pre{Lines: []string{gear}}}}, "    " + gear + "\n"},
		{"escape", para(doc.Text("x\x1b[31my\x07")), "x�[31my�\n"},
		{"white space", para(doc.Text("a \t\n  b")), "a b\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := write(t, c.in, 0); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// T3: wrapping.
func TestWrap(t *testing.T) {
	for _, c := range []struct {
		in   string
		w    int
		want []string
	}{
		{"abcdefghij kl 🟢🟢🟢🟢 x", 5, []string{"abcde", "fghij", "kl", "🟢🟢", "🟢🟢", "x"}},
		{"ééé", 2, []string{"éé", "é"}},
		{"👩‍💻👩‍💻", 4, []string{"👩‍💻", "👩‍💻"}},
		{"", 5, nil},
		{"a b c", 3, []string{"a b", "c"}},
	} {
		got := wrap(c.in, c.w)
		if strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("wrap(%q, %d) = %q, want %q", c.in, c.w, got, c.want)
		}
	}
}

// T4: each span.
func TestSpans(t *testing.T) {
	for _, c := range []struct {
		name string
		in   []doc.Inline
		want string
	}{
		{"text", tx("a  b"), "a b\n"},
		{"code", []doc.Inline{doc.Code("x_y")}, "x_y\n"},
		{"symbol", []doc.Inline{doc.Symbol("🟢")}, "🟢\n"},
		{"strong", []doc.Inline{doc.Strong{doc.Text("Sensor node")}}, "*Sensor node*\n"},
		{"empty strong", []doc.Inline{doc.Text("a"), doc.Strong{}, doc.Text("b")}, "ab\n"},
		{"link", []doc.Inline{doc.Link{Text: tx("Proposed views"), URL: "PLAN.md#views"}}, "Proposed views <PLAN.md#views>\n"},
		{"link text is the url", []doc.Inline{doc.Link{Text: tx("a.md"), URL: "a.md"}}, "a.md\n"},
		{"link without url", []doc.Inline{doc.Link{Text: tx("x")}}, "x\n"},
		{"link without text", []doc.Inline{doc.Link{URL: "u"}}, "u\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := write(t, para(c.in...), 0); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// T5: each block.
func TestBlocks(t *testing.T) {
	d := doc.Doc{Blocks: []doc.Block{
		doc.Heading{Level: 1, Text: tx("Title")},
		doc.Heading{Level: 2, Text: tx("Section")},
		doc.Heading{Level: 3, Text: tx("Sub")},
		doc.Para{Text: tx("one two three four five six")},
		doc.List{Ordered: true, Items: []doc.Item{
			{Text: tx("first item that wraps around"), Blocks: []doc.Block{
				doc.List{Items: []doc.Item{{Text: tx("nested")}, {}}},
				doc.Pre{Lines: []string{"a\tb   ", "", "tail"}},
			}},
			{Text: tx("second")},
		}},
		doc.Table{},
		doc.List{},
		doc.Pre{},
		doc.Heading{},
	}}
	want := strings.Join([]string{
		"Title",
		"═════",
		"",
		"Section",
		"───────",
		"",
		"Sub",
		"···",
		"",
		"one two three four",
		"five six",
		"",
		"1. first item that",
		"   wraps around",
		"   - nested",
		"   -",
		"       a b",
		"",
		"       tail",
		"2. second",
		"",
	}, "\n")
	if got := write(t, d, 20); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if got := write(t, doc.Doc{}, 0); got != "" {
		t.Errorf("empty document wrote %q", got)
	}
}

type failAfter struct{ n, calls int }

func (f *failAfter) Write(p []byte) (int, error) {
	f.calls++
	return min(f.n, len(p)), errBoom
}

var errBoom = errors.New("boom")

// T13: determinism, the write and the width.
func TestWrite(t *testing.T) {
	d := para(doc.Text("same bytes every time"))
	if a, b := write(t, d, 0), write(t, d, 0); a != b {
		t.Error("two renderings differ")
	}
	if write(t, d, 0) != write(t, d, 80) {
		t.Error("width 0 and 80 differ")
	}
	if write(t, d, 20) == "" {
		t.Error("width 20 renders nothing")
	}
	var buf countWriter
	if err := Write(&buf, d, Options{}); err != nil || buf.calls != 1 {
		t.Errorf("err %v after %d calls, want one call", err, buf.calls)
	}
	f := &failAfter{n: 3}
	if err := Write(f, d, Options{}); err != errBoom || f.calls != 1 {
		t.Errorf("err %v after %d calls, want boom after one", err, f.calls)
	}
	for _, w := range []int{-1, 1, 19} {
		var b bytes.Buffer
		if err := Write(&b, d, Options{Width: w}); !errors.Is(err, ErrWidth) || b.Len() != 0 {
			t.Errorf("width %d: err %v, wrote %d bytes", w, err, b.Len())
		}
	}
}

type countWriter struct{ calls int }

func (c *countWriter) Write(p []byte) (int, error) { c.calls++; return io.Discard.Write(p) }

// A paragraph marked OneLine prints on one line past the width; one that
// is not marked wraps.
func TestOneLine(t *testing.T) {
	spans := []doc.Inline{doc.Text("Command: "), doc.Code("tabloio queue --person someone@example.com --level detail --text")}
	got := write(t, doc.Doc{Blocks: []doc.Block{doc.Para{Text: spans, OneLine: true}}}, 20)
	if want := "Command: tabloio queue --person someone@example.com --level detail --text\n"; got != want {
		t.Errorf("OneLine: got %q, want %q", got, want)
	}
	got = write(t, doc.Doc{Blocks: []doc.Block{doc.Para{Text: spans}}}, 20)
	if strings.Count(got, "\n") < 2 {
		t.Errorf("a plain paragraph does not wrap: %q", got)
	}
}
