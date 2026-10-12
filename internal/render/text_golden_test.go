package render

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"golang.org/x/text/width"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/text"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// textCase is one text golden file: a fixture at a level and a width.
type textCase struct {
	view, name string
	level      view.Level
	width      int    // 0 for the zero options
	suffix     string // before ".txt", for example ".w100"
}

func (c textCase) file() string {
	return filepath.Join("testdata", c.view, c.name+"."+c.level.String()+c.suffix+".txt")
}

// fixture finds <view>/<name>.json or <view>/in/<name>.json, as the temporal
// views keep theirs.
func (c textCase) fixture(t testing.TB) view.View {
	t.Helper()
	p := filepath.Join("testdata", c.view, c.name+".json")
	if _, err := os.Stat(p); err != nil {
		p = filepath.Join("testdata", c.view, "in", c.name+".json")
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	v, err := view.Decode(c.view, data)
	if err != nil {
		t.Fatalf("%s: %v", p, err)
	}
	return v
}

func (c textCase) options() Options {
	return Options{Format: Text, Level: c.level, Text: text.Options{Width: c.width}}
}

// textVariants are the three golden files beyond the Markdown ones (T9).
var textVariants = []textCase{
	{"tableau", "weather", view.Detail, 100, ".w100"},
	{"tableau", "weather", view.Detail, 48, ".w48"},
	{"queue", "tooling-agent", view.Glance, 132, ".w132"},
}

// textCases lists a text case beside each Markdown golden file, then the
// variants (T8, T9).
func textCases(t testing.TB) []textCase {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", "*", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	var out []textCase
	for _, f := range files {
		base := strings.TrimSuffix(filepath.Base(f), ".md")
		i := strings.LastIndex(base, ".")
		level, err := view.ParseLevel(base[i+1:])
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		out = append(out, textCase{filepath.Base(filepath.Dir(f)), base[:i], level, 0, ""})
	}
	return append(out, textVariants...)
}

func renderText(t testing.TB, c textCase) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Render(&buf, c.fixture(t), c.options()); err != nil {
		t.Fatalf("%s: %v", c.file(), err)
	}
	return buf.String()
}

// T8 and T9: every text golden file against its rendering, byte for byte.
func TestTextGolden(t *testing.T) {
	cases := textCases(t)
	if len(cases) != 126 {
		t.Errorf("%d text cases, want 126", len(cases))
	}
	for _, c := range cases {
		t.Run(c.view+"/"+c.name+"."+c.level.String()+c.suffix, func(t *testing.T) {
			got := renderText(t, c)
			if *update {
				if err := os.WriteFile(c.file(), []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(c.file())
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Errorf("%s differs from its golden file\n--- got ---\n%s\n--- want ---\n%s", c.file(), got, want)
			}
		})
	}
}

// T6 and T9: the golden files of the global tableau equal the drafts, but
// for the drafts that decision 8 of design e0f7 outdates. The closing
// command prints on one line past the width (8c), so the draft
// tableau.weather.detail.w48.txt, which wraps it at 48 cells, left the
// comparison.
func TestTextDrafts(t *testing.T) {
	drafts := map[string]string{
		"tableau/weather.detail.txt":           "tableau.weather.detail.txt",
		"tableau/weather.detail.w100.txt":      "tableau.weather.detail.w100.txt",
		"blockage/illustration.provenance.txt": "blockage.illustration.provenance.txt",
		"queue/weather-ada.detail.txt":         "queue.weather-ada.detail.txt",
	}
	for golden, draft := range drafts {
		want, err := os.ReadFile(filepath.Join("..", "..", "design", "e0f7", draft))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join("testdata", filepath.FromSlash(golden)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s differs from the draft %s", golden, draft)
		}
	}
}

// cellsOf is the measure of T10, stated apart from the writer's.
func cellsOf(s string) int {
	n := 0
	for _, r := range s {
		switch {
		case r < 0x20, r >= 0x7F && r < 0xA0, unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf):
		case width.LookupRune(r).Kind() == width.EastAsianWide, width.LookupRune(r).Kind() == width.EastAsianFullwidth:
			n += 2
		default:
			n++
		}
	}
	return n
}

// preLines collects the lines of every Pre of the document, as the writer
// prints them without their indent.
func preLines(blocks []doc.Block, into map[string]bool) {
	for _, b := range blocks {
		switch b := b.(type) {
		case doc.Pre:
			for _, ln := range b.Lines {
				ln = strings.Map(func(r rune) rune {
					switch {
					case r == '\t':
						return ' '
					case unicode.IsControl(r):
						return unicode.ReplacementChar
					}
					return r
				}, ln)
				into[strings.TrimSpace(ln)] = true
			}
		case doc.List:
			for _, it := range b.Items {
				preLines(it.Blocks, into)
			}
		}
	}
}

// T10: the invariants of every text golden.
func TestTextInvariants(t *testing.T) {
	for _, c := range textCases(t) {
		t.Run(c.view+"/"+c.name+"."+c.level.String()+c.suffix, func(t *testing.T) {
			out := renderText(t, c)
			limit := c.width
			if limit == 0 {
				limit = text.DefaultWidth
			}
			pre := map[string]bool{}
			preLines(buildFor(t, c.fixture(t), c.options()).Blocks, pre)
			if !strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\n\n") {
				t.Error("the output does not end in exactly one newline")
			}
			for _, r := range out {
				if r != '\n' && unicode.IsControl(r) {
					t.Errorf("control character %U", r)
				}
			}
			var box []string
			lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
			for i, ln := range lines {
				// The closing command prints on one line, past the limit if need be (decision 8c).
				isPre := pre[strings.TrimSpace(ln)] && strings.HasPrefix(ln, "    ") ||
					i == len(lines)-1 && strings.HasPrefix(ln, "Command: ")
				if strings.HasSuffix(ln, " ") {
					t.Errorf("line %d ends in a space: %q", i+1, ln)
				}
				if cellsOf(ln) > limit && !isPre {
					t.Errorf("line %d is %d cells, over %d: %q", i+1, cellsOf(ln), limit, ln)
				}
				if !isPre {
					prev := rune(0)
					for _, r := range ln {
						if r == '️' && prev != 0 && cellsOf(string(prev)) == 1 {
							t.Errorf("line %d holds a selector after %U", i+1, prev)
						}
						prev = r
					}
				}
				switch {
				case strings.HasPrefix(ln, "┌"):
					box = []string{ln}
				case box != nil:
					box = append(box, ln)
					if strings.HasPrefix(ln, "└") {
						for _, b := range box {
							if cellsOf(b) != cellsOf(box[0]) {
								t.Errorf("line %d: a box line of %d cells under a rule of %d: %q", i+1, cellsOf(b), cellsOf(box[0]), b)
							}
						}
						box = nil
					}
				}
			}
		})
	}
}

// T11: the writer invents no symbol.
func TestTextInventsNoSymbol(t *testing.T) {
	const allowed = "·…→›—×═─│┌┬┐├┼┤└┴┘┼�"
	for _, c := range textCases(t) {
		t.Run(c.view+"/"+c.name+"."+c.level.String()+c.suffix, func(t *testing.T) {
			p := filepath.Join("testdata", c.view, c.name+".json")
			if _, err := os.Stat(p); err != nil {
				p = filepath.Join("testdata", c.view, "in", c.name+".json")
			}
			fx, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range renderText(t, c) {
				if r > 0x7F && !strings.ContainsRune(allowed, r) && !bytes.ContainsRune(fx, r) {
					t.Errorf("the output holds %q (%U), which the fixture does not", r, r)
				}
			}
		})
	}
}

// run joins the Text, Code and Symbol spans of a run, with the content of
// its Strong spans and the text of its Links, as the writer prints them
// before the marks.
func run(xs []doc.Inline) string {
	var b strings.Builder
	for _, x := range xs {
		switch x := x.(type) {
		case doc.Text:
			b.WriteString(string(x))
		case doc.Code:
			b.WriteString(string(x))
		case doc.Symbol:
			b.WriteString(string(x))
		case doc.Strong:
			b.WriteString(run(x))
		case doc.Link:
			b.WriteString(run(x.Text))
		}
	}
	return b.String()
}

// words collects the words of the runs of the blocks.
func words(blocks []doc.Block, into map[string]int) {
	for _, b := range blocks {
		switch b := b.(type) {
		case doc.Heading:
			addWords(run(b.Text), into)
		case doc.Para:
			addWords(run(b.Text), into)
		case doc.Table:
			for _, c := range b.Head {
				addWords(run(c.Text), into)
			}
			for _, r := range b.Rows {
				for _, c := range r {
					addWords(run(c.Text), into)
				}
			}
		case doc.List:
			for _, it := range b.Items {
				addWords(run(it.Text), into)
				words(it.Blocks, into)
			}
		}
	}
}

// tokenWord is a word without the marks the writer adds around it.
func tokenWord(w string) string {
	return strings.Trim(strings.ReplaceAll(w, "️", ""), "*<>:,.;()")
}

func addWords(s string, into map[string]int) {
	for _, w := range strings.Fields(s) {
		if w = tokenWord(w); w != "" && cellsOf(w) <= 16 {
			into[w]++
		}
	}
}

// T12: nothing is lost against the document.
func TestTextLosesNothing(t *testing.T) {
	for _, c := range textCases(t) {
		if c.width != 0 {
			continue
		}
		t.Run(c.view+"/"+c.name+"."+c.level.String(), func(t *testing.T) {
			want := map[string]int{}
			words(buildFor(t, c.fixture(t), c.options()).Blocks, want)
			out := strings.Map(func(r rune) rune {
				if strings.ContainsRune("│┌┬┐├┼┤└┴┘─═", r) {
					return ' '
				}
				return r
			}, renderText(t, c))
			got := map[string]int{}
			for _, w := range strings.Fields(out) {
				if w = tokenWord(w); w != "" {
					got[w]++
				}
			}
			for w, n := range want {
				if got[w] < n {
					t.Errorf("the word %q occurs %d times in the document and %d in the output", w, n, got[w])
				}
			}
		})
	}
}

// T13: through Render, a width under 20 is ErrWidth and writes nothing.
func TestTextWidth(t *testing.T) {
	v := readFixture(t, "task", "e9c6")
	for _, w := range []int{-1, 1, 19} {
		var buf bytes.Buffer
		err := Render(&buf, v, Options{Format: Text, Level: view.Glance, Text: text.Options{Width: w}})
		if !errors.Is(err, text.ErrWidth) || buf.Len() != 0 {
			t.Errorf("width %d: err %v, wrote %d bytes", w, err, buf.Len())
		}
	}
	for _, w := range []int{0, 20, 80} {
		if err := Render(io.Discard, v, Options{Format: Text, Level: view.Glance, Text: text.Options{Width: w}}); err != nil {
			t.Errorf("width %d: %v", w, err)
		}
	}
	// Markdown ignores the width.
	if err := Render(io.Discard, v, Options{Level: view.Glance, Text: text.Options{Width: 1}}); err != nil {
		t.Errorf("Markdown with a width: %v", err)
	}
}

// closing returns the closing command of a text rendering as written: it
// stands alone on one line, the last of the output.
func closing(t *testing.T, out string) string {
	t.Helper()
	for _, para := range strings.Split(out, "\n\n") {
		if strings.HasPrefix(para, "Command: ") {
			para = strings.TrimSuffix(para, "\n")
			if strings.Contains(para, "\n") || !strings.HasSuffix(out, "\n"+para+"\n") {
				t.Fatalf("the closing command does not stand on one last line: %q", para)
			}
			return para
		}
	}
	t.Fatalf("no closing command in\n%s", out)
	return ""
}

// T14: Render and the command.
func TestTextCommand(t *testing.T) {
	v := readFixture(t, "task", "e9c6")
	if err := Render(io.Discard, v, Options{Format: Format(9), Level: view.Glance}); !errors.Is(err, ErrFormat) {
		t.Errorf("Format(9): %v", err)
	}
	for _, c := range []struct {
		width int
		want  string
	}{
		{0, "--level glance --text"},
		{80, "--level glance --text"},
		{100, "--level glance --text --width 100"},
		{20, "--level glance --text --width 20"},
	} {
		got := closing(t, renderString(t, v, Options{Format: Text, Level: view.Glance, Text: text.Options{Width: c.width}}))
		if !strings.HasSuffix(got, c.want) {
			t.Errorf("width %d: %q does not end in %q", c.width, got, c.want)
		}
	}
	if got := closing(t, renderString(t, v, Options{Level: view.Glance})); !strings.HasSuffix(got, "--level glance`") {
		t.Errorf("Markdown command changed: %q", got)
	}
}

// T15: links in text.
func TestTextLinks(t *testing.T) {
	v := readFixture(t, "task", "e9c6")
	for _, c := range []struct{ project, want string }{
		{"", "Proposed views <PLAN.md#views>"},
		{".", "Proposed views <PLAN.md#views>"},
		{"../..", "Proposed views <../../PLAN.md#views>"},
	} {
		out := renderString(t, v, Options{Format: Text, Level: view.Detail, Text: text.Options{Width: 200}, Links: Links{Project: c.project}})
		if !strings.Contains(out, c.want) {
			t.Errorf("Project %q: no %q in\n%s", c.project, c.want, out)
		}
		if strings.Contains(out, ".tableaux/tasks/") {
			t.Errorf("Project %q: the text holds a task-file link", c.project)
		}
	}
	for _, c := range textCases(t) {
		o := c.options()
		o.Links.Project = "."
		var buf bytes.Buffer
		if err := Render(&buf, c.fixture(t), o); err != nil {
			t.Fatal(err)
		}
		if buf.String() != renderText(t, c) {
			t.Errorf("%s differs under Project \".\"", c.file())
		}
	}
}
