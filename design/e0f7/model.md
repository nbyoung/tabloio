# Model e0f7: the text writer

This draft holds the Go of `internal/render/text` in full, as the trial of [design e0f7](../e0f7.md) compiles it: three files, 621 lines with their comments. The design's prose states the rules; this draft is their exact form, so that the implementation takes no decision. The golden drafts beside it are this code's output. The implementation lands these files as they stand and adds the tests of the design.

## `text.go`

The options, `Write`, the blocks and the spans.

```go
// Package text writes a doc.Doc as plain Unicode text at a width in cells.
//
// Blocks print with one blank line between them, lines end in \n, no line
// ends in a space, and the output ends in one newline. The output holds no
// control character but \n and no escape sequence. No line is wider than
// the width, but for the lines of a Pre, which print as given.
package text

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/nbyoung/tabloio/internal/render/doc"
)

// Selector is the policy for U+FE0F, the variation selector that asks for
// the emoji form of a symbol, where it follows a code point one cell wide.
type Selector int

// The policies.
const (
	Strip Selector = iota // remove the selector: the symbol takes one cell
	Keep                  // keep the selector and count the pair as two cells
)

// ParseSelector reads "strip" or "keep".
func ParseSelector(s string) (Selector, error) {
	switch s {
	case "strip":
		return Strip, nil
	case "keep":
		return Keep, nil
	}
	return Strip, fmt.Errorf("unknown selector policy %q", s)
}

func (s Selector) String() string {
	if s == Keep {
		return "keep"
	}
	return "strip"
}

// The widths, in cells.
const (
	DefaultWidth = 80 // what Options.Width 0 means
	MinWidth     = 20 // a smaller Options.Width counts as this
)

const (
	minText   = 16     // the least width a nested block or an indented line wraps at
	wordCap   = 24     // a word up to this width never breaks inside a box table
	preIndent = "    " // before each line of a Pre
)

// Options holds the width and the selector policy. The zero value is 80
// cells and Strip.
type Options struct {
	Width    int // cells per line; 0 is DefaultWidth, and less than MinWidth is MinWidth
	Selector Selector
}

// writer draws blocks under one measure.
type writer struct{ measure }

// Write prints d to w in one call to w.Write, so a failing writer sees the
// whole text at once and its error passes through unwrapped. An empty
// document writes nothing.
func Write(w io.Writer, d doc.Doc, o Options) error {
	width := o.Width
	if width == 0 {
		width = DefaultWidth
	}
	width = max(width, MinWidth)
	x := writer{measure{o.Selector}}
	var parts []string
	for _, b := range d.Blocks {
		if s := strings.Join(x.block(b, width), "\n"); s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	out := strings.Join(parts, "\n\n") + "\n"
	n, err := io.WriteString(w, out)
	if err == nil && n < len(out) {
		err = io.ErrShortWrite
	}
	return err
}

// block returns the lines of one block at w cells, without line ends.
func (x writer) block(b doc.Block, w int) []string {
	switch b := b.(type) {
	case doc.Heading:
		return x.heading(b, w)
	case doc.Para:
		return x.wrap(x.flat(b.Text), w)
	case doc.Table:
		return x.table(b, w)
	case doc.List:
		return x.list(b, w)
	case doc.Pre:
		return x.pre(b)
	}
	return nil
}

// heading prints the text and, under it, a rule as wide as its widest line:
// ═ for level 1, ─ for level 2, · below.
func (x writer) heading(h doc.Heading, w int) []string {
	lines := x.wrap(x.flat(h.Text), w)
	if len(lines) == 0 {
		return nil
	}
	rule := "·"
	switch {
	case h.Level <= 1:
		rule = "═"
	case h.Level == 2:
		rule = "─"
	}
	n := 0
	for _, ln := range lines {
		n = max(n, x.width(ln))
	}
	return append(lines, strings.Repeat(rule, n))
}

// list prints each item after "- " or "<n>. ", its further lines and its
// nested blocks indented by the marker's width.
func (x writer) list(l doc.List, w int) []string {
	var lines []string
	for i, it := range l.Items {
		marker := "- "
		if l.Ordered {
			marker = strconv.Itoa(i+1) + ". "
		}
		pad := strings.Repeat(" ", len(marker))
		inner := max(w-len(marker), minText)
		text := x.wrap(x.flat(it.Text), inner)
		if len(text) == 0 {
			text = []string{""}
		}
		for j, ln := range text {
			if j == 0 {
				lines = append(lines, strings.TrimRight(marker+ln, " "))
			} else {
				lines = append(lines, pad+ln)
			}
		}
		for _, b := range it.Blocks {
			for _, ln := range x.block(b, inner) {
				if ln != "" {
					ln = pad + ln
				}
				lines = append(lines, ln)
			}
		}
	}
	return lines
}

// pre prints each line as given after four spaces, without its trailing
// spaces; an empty line stays empty. No line wraps.
func (x writer) pre(p doc.Pre) []string {
	var lines []string
	for _, ln := range p.Lines {
		ln = strings.TrimRight(x.cleanPre(ln), " ")
		if ln != "" {
			ln = preIndent + ln
		}
		lines = append(lines, ln)
	}
	return lines
}

// cleanPre keeps a line as given but for its control characters: a tab
// prints as a space and any other as U+FFFD.
func (x writer) cleanPre(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\t':
			b.WriteByte(' ')
		case unicode.IsControl(r):
			b.WriteRune(unicode.ReplacementChar)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// flat prints a run of spans as one line of plain text, cleaned and with
// no space at either end.
func (x writer) flat(xs []doc.Inline) string {
	return strings.TrimSpace(x.clean(spans(xs)))
}

// spans prints a run of spans: Text, Code and Symbol as given, Strong
// between asterisks, and a Link as its text and then its URL in angle
// brackets, unless the text is the URL.
func spans(xs []doc.Inline) string {
	var b strings.Builder
	for _, in := range xs {
		switch in := in.(type) {
		case doc.Text:
			b.WriteString(spaces(string(in)))
		case doc.Code:
			b.WriteString(spaces(string(in)))
		case doc.Symbol:
			b.WriteString(spaces(string(in)))
		case doc.Strong:
			if s := strings.TrimSpace(spans(in)); s != "" {
				b.WriteString("*" + s + "*")
			}
		case doc.Link:
			s := spans(in.Text)
			u := spaces(in.URL)
			switch {
			case strings.TrimSpace(s) == "":
				s = u
			case u != "" && u != strings.TrimSpace(s):
				s += " <" + u + ">"
			}
			b.WriteString(s)
		}
	}
	return b.String()
}

// spaces turns each run of white space into one space.
func spaces(s string) string {
	var b strings.Builder
	space := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			space = true
			continue
		}
		if space {
			b.WriteByte(' ')
			space = false
		}
		b.WriteRune(r)
	}
	if space {
		b.WriteByte(' ')
	}
	return b.String()
}
```

## `width.go`

The measure: cells, the selector policy, units and wrapping.

```go
package text

import (
	"strings"
	"unicode"

	"golang.org/x/text/width"
)

const selector = '️'

// measure counts terminal cells under one selector policy.
type measure struct{ sel Selector }

// runeWidth is the number of cells one code point takes: 0, 1 or 2.
func runeWidth(r rune) int {
	switch {
	case r < 0x20, r >= 0x7F && r < 0xA0:
		return 0
	case unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf):
		return 0
	}
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	}
	return 1
}

// width is the number of cells s takes.
func (m measure) width(s string) int {
	n, prev := 0, 0
	for _, r := range s {
		w := runeWidth(r)
		if r == selector && m.sel == Keep && prev == 1 {
			w = 1
		}
		n += w
		if r != selector {
			prev = runeWidth(r)
		} else {
			prev = 0
		}
	}
	return n
}

// clean replaces each control character with U+FFFD and, under Strip,
// removes each U+FE0F that follows a code point one cell wide.
func (m measure) clean(s string) string {
	var b strings.Builder
	prev := 0
	for _, r := range s {
		if unicode.IsControl(r) {
			r = unicode.ReplacementChar
		}
		if r == selector && m.sel == Strip && prev == 1 {
			prev = 0
			continue
		}
		prev = runeWidth(r)
		b.WriteRune(r)
	}
	return b.String()
}

// units splits a word into the pieces a hard break may fall between: a code
// point with a width, with the zero-width code points after it, and with
// whatever a zero-width joiner ties to it.
func units(s string) []string {
	var out []string
	joined := false
	for _, r := range s {
		if len(out) > 0 && (runeWidth(r) == 0 || joined) {
			out[len(out)-1] += string(r)
		} else {
			out = append(out, string(r))
		}
		joined = r == '‍'
	}
	return out
}

// wrap fills lines of at most w cells with the words of s, which single
// spaces separate. A word wider than w breaks between units.
func (m measure) wrap(s string, w int) []string {
	if s == "" {
		return nil
	}
	w = max(w, 1)
	var lines []string
	line, used := "", 0
	flush := func() {
		lines = append(lines, line)
		line, used = "", 0
	}
	for _, word := range strings.Split(s, " ") {
		ww := m.width(word)
		if used > 0 && used+1+ww <= w {
			line += " " + word
			used += 1 + ww
			continue
		}
		if used > 0 {
			flush()
		}
		if ww <= w {
			line, used = word, ww
			continue
		}
		for _, u := range units(word) {
			uw := m.width(u)
			if used > 0 && used+uw > w {
				flush()
			}
			line += u
			used += uw
		}
	}
	if used > 0 || line != "" {
		flush()
	}
	return lines
}

// widestWord is the width of the widest word of s.
func (m measure) widestWord(s string) int {
	n := 0
	for _, word := range strings.Split(s, " ") {
		n = max(n, m.width(word))
	}
	return n
}
```

## `table.go`

The table: the three forms.

```go
package text

import (
	"strings"

	"github.com/nbyoung/tabloio/internal/render/doc"
)

// tcell is a cell as the table draws it: its indent in cells and its text.
type tcell struct {
	indent int
	text   string
}

// table draws t in the first form that fits w: a box with every column at
// its natural width; a tight box, where centred columns lose their padding
// and the widest columns narrow to their floors; or the stacked form.
func (x writer) table(t doc.Table, w int) []string {
	n := len(t.Head)
	for _, r := range t.Rows {
		n = max(n, len(r))
	}
	if n == 0 {
		return nil
	}
	conv := func(cells []doc.Cell) []tcell {
		out := make([]tcell, n)
		for i, c := range cells {
			out[i] = tcell{indent: 2 * max(c.Indent, 0), text: x.flat(c.Text)}
		}
		return out
	}
	head := conv(t.Head)
	rows := make([][]tcell, len(t.Rows))
	for i, r := range t.Rows {
		rows[i] = conv(r)
	}
	align := make([]doc.Align, n)
	copy(align, t.Align)

	nat, floor := make([]int, n), make([]int, n)
	for i := range n {
		nat[i], floor[i] = 1, 1
	}
	for _, r := range append([][]tcell{head}, rows...) {
		for i, c := range r {
			nat[i] = max(nat[i], c.indent+x.width(c.text))
			floor[i] = max(floor[i], c.indent+min(x.widestWord(c.text), wordCap))
		}
	}
	// overhead is the cells the bars and the padding take.
	overhead := func(tight bool) int {
		over := n + 1
		for i := range n {
			if !tight || align[i] != doc.Centre {
				over += 2
			}
		}
		return over
	}
	if sum(nat)+overhead(false) <= w {
		return x.boxed(head, rows, align, nat, false)
	}
	if over := overhead(true); sum(floor)+over <= w {
		widths := append([]int(nil), nat...)
		for sum(widths)+over > w {
			best := -1
			for i := range n {
				if widths[i] > floor[i] && (best < 0 || widths[i] > widths[best]) {
					best = i
				}
			}
			widths[best]--
		}
		return x.boxed(head, rows, align, widths, true)
	}
	return x.stacked(head, rows, align, w)
}

// sum adds the widths.
func sum(xs []int) int {
	n := 0
	for _, v := range xs {
		n += v
	}
	return n
}

// boxed draws the table as a box. A cell wraps inside its column, and a
// row is as tall as its tallest cell.
func (x writer) boxed(head []tcell, rows [][]tcell, align []doc.Align, widths []int, tight bool) []string {
	pad := make([]string, len(widths))
	for i := range widths {
		if !tight || align[i] != doc.Centre {
			pad[i] = " "
		}
	}
	rule := func(l, mid, r string) string {
		parts := make([]string, len(widths))
		for i, w := range widths {
			parts[i] = strings.Repeat("─", w+2*len(pad[i]))
		}
		return l + strings.Join(parts, mid) + r
	}
	draw := func(r []tcell) []string {
		cols := make([][]string, len(widths))
		height := 1
		for i, c := range r {
			for _, ln := range x.wrap(c.text, widths[i]-c.indent) {
				cols[i] = append(cols[i], strings.Repeat(" ", c.indent)+ln)
			}
			height = max(height, len(cols[i]))
		}
		out := make([]string, height)
		for k := range height {
			var b strings.Builder
			b.WriteString("│")
			for i, w := range widths {
				s := ""
				if k < len(cols[i]) {
					s = cols[i][k]
				}
				gap := w - x.width(s)
				left := 0
				switch align[i] {
				case doc.Centre:
					left = gap / 2
				case doc.Right:
					left = gap
				}
				b.WriteString(pad[i] + strings.Repeat(" ", left) + s + strings.Repeat(" ", gap-left) + pad[i] + "│")
			}
			out[k] = b.String()
		}
		return out
	}
	// A rule parts the body rows when a row holds more than one line and
	// the first cells do not tell the rows apart: one is empty or wraps.
	body := make([][]string, len(rows))
	tall, told := false, true
	for i, r := range rows {
		body[i] = draw(r)
		tall = tall || len(body[i]) > 1
		if r[0].text == "" || len(x.wrap(r[0].text, widths[0]-r[0].indent)) > 1 {
			told = false
		}
	}
	ruled := tall && !told
	lines := []string{rule("┌", "┬", "┐")}
	lines = append(lines, draw(head)...)
	for i, b := range body {
		if i == 0 || ruled {
			lines = append(lines, rule("├", "┼", "┤"))
		}
		lines = append(lines, b...)
	}
	return append(lines, rule("└", "┴", "┘"))
}

// stacked draws the table as records between rules. The first record
// names the columns, joined by " · ". Each row follows: its first cell,
// then one line per further cell that is not empty, as "<head>: <cell>".
// A run of centred columns shares one line.
func (x writer) stacked(head []tcell, rows [][]tcell, align []doc.Align, w int) []string {
	rule := strings.Repeat("─", w)
	lines := []string{rule}
	var names []string
	for _, h := range head {
		if h.text != "" {
			names = append(names, h.text)
		}
	}
	if len(names) > 0 {
		lines = append(append(lines, x.fill(names, w)...), rule)
	}
	pair := func(i int, c tcell) string {
		if head[i].text == "" {
			return c.text
		}
		return head[i].text + ": " + c.text
	}
	for _, r := range rows {
		first := r[0]
		for _, ln := range x.wrap(first.text, max(w-first.indent, minText)) {
			lines = append(lines, strings.Repeat(" ", first.indent)+ln)
		}
		for i := 1; i < len(r); i++ {
			if r[i].text == "" {
				continue
			}
			pairs := []string{pair(i, r[i])}
			for align[i] == doc.Centre && i+1 < len(r) && align[i+1] == doc.Centre {
				i++
				if r[i].text != "" {
					pairs = append(pairs, pair(i, r[i]))
				}
			}
			for k, ln := range x.fill(pairs, w-4) {
				if k == 0 {
					lines = append(lines, "  "+ln)
				} else {
					lines = append(lines, "    "+ln)
				}
			}
		}
		lines = append(lines, rule)
	}
	return lines
}

// fill joins pairs by " · " into lines of at most w cells and breaks
// between pairs; a pair wider than w wraps by its words.
func (x writer) fill(pairs []string, w int) []string {
	if len(pairs) == 0 {
		return nil
	}
	const sep = " · "
	var lines []string
	line := ""
	for _, p := range pairs {
		switch {
		case line == "":
			line = p
		case x.width(line)+x.width(sep)+x.width(p) <= w:
			line += sep + p
		default:
			lines = append(lines, x.wrap(line, w)...)
			line = p
		}
	}
	return append(lines, x.wrap(line, w)...)
}
```
