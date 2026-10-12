// Package text writes a doc.Doc as plain Unicode text at a width in cells.
//
// Blocks print with one blank line between them, lines end in \n, no line
// ends in a space, and the output ends in one newline. The output holds no
// control character but \n and no escape sequence. No line is wider than
// the width, but for the lines of a Pre, which print as given, and for a
// paragraph marked OneLine, such as the closing command.
//
// The writer offers no option for the variation selector U+FE0F: outside a
// Pre it writes a code point one cell wide without a selector after it, and
// counts the cells of what it writes.
package text

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/nbyoung/tabloio/internal/render/doc"
)

// The widths, in cells.
const (
	DefaultWidth = 80 // what Options.Width 0 means
	MinWidth     = 20 // the least width; Write refuses a smaller one
)

// ErrWidth is what Write returns for a width under MinWidth.
var ErrWidth = errors.New("text: the width is under 20 cells")

const (
	minText   = 16     // the least width a nested block or an indented line wraps at
	wordCap   = 24     // a word up to this width never breaks inside a box table
	preIndent = "    " // before each line of a Pre
)

// Options holds the width. The zero value is 80 cells.
type Options struct {
	Width int // cells per line; 0 is DefaultWidth, and any other value under MinWidth is an error
}

// Write prints d to w in one call to w.Write, so a failing writer sees the
// whole text at once and its error passes through unwrapped. An empty
// document writes nothing. A width of 0 is DefaultWidth; for any other
// width under MinWidth, a negative one included, Write returns ErrWidth and
// writes nothing: it adjusts no width.
func Write(w io.Writer, d doc.Doc, o Options) error {
	width := o.Width
	if width == 0 {
		width = DefaultWidth
	}
	if width < MinWidth {
		return ErrWidth
	}
	var parts []string
	for _, b := range d.Blocks {
		if s := strings.Join(block(b, width), "\n"); s != "" {
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
func block(b doc.Block, w int) []string {
	switch b := b.(type) {
	case doc.Heading:
		return heading(b, w)
	case doc.Para:
		if b.OneLine {
			return oneLine(b)
		}
		return wrap(flat(b.Text), w)
	case doc.Table:
		return table(b, w)
	case doc.List:
		return list(b, w)
	case doc.Pre:
		return pre(b)
	}
	return nil
}

// oneLine prints a paragraph on one line, past the width if need be.
func oneLine(p doc.Para) []string {
	if s := flat(p.Text); s != "" {
		return []string{s}
	}
	return nil
}

// heading prints the text and, under it, a rule as wide as its widest line:
// ═ for level 1, ─ for level 2, · below.
func heading(h doc.Heading, w int) []string {
	lines := wrap(flat(h.Text), w)
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
		n = max(n, cells(ln))
	}
	return append(lines, strings.Repeat(rule, n))
}

// list prints each item after "- " or "<n>. ", its further lines and its
// nested blocks indented by the marker's width.
func list(l doc.List, w int) []string {
	var lines []string
	for i, it := range l.Items {
		marker := "- "
		if l.Ordered {
			marker = strconv.Itoa(i+1) + ". "
		}
		pad := strings.Repeat(" ", len(marker))
		inner := max(w-len(marker), minText)
		text := wrap(flat(it.Text), inner)
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
			for _, ln := range block(b, inner) {
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
func pre(p doc.Pre) []string {
	var lines []string
	for _, ln := range p.Lines {
		ln = strings.TrimRight(cleanPre(ln), " ")
		if ln != "" {
			ln = preIndent + ln
		}
		lines = append(lines, ln)
	}
	return lines
}

// cleanPre keeps a line as given but for its control characters: a tab
// prints as a space and any other as U+FFFD.
func cleanPre(s string) string {
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
func flat(xs []doc.Inline) string {
	return strings.TrimSpace(clean(spans(xs)))
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
