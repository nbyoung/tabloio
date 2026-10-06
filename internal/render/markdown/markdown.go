// Package markdown writes a doc.Doc as GitHub-flavoured Markdown.
//
// Blocks print with one blank line between them, lines end in \n, no line
// ends in a space, and the output ends in one newline. Tables carry no
// padding: the display width of an emoji differs between media, and an
// unpadded table changes one line where a padded one changes every row.
package markdown

import (
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/nbyoung/tabloio/internal/render/doc"
)

// indentStep is what one step of Cell.Indent prints.
const indentStep = "&nbsp;&nbsp;"

// Write prints d to w in one call to w.Write, so a failing writer sees the
// whole text at once and its error passes through unwrapped.
func Write(w io.Writer, d doc.Doc) error {
	var parts []string
	for _, b := range d.Blocks {
		if s := strings.Join(blockLines(b), "\n"); s != "" {
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

// blockLines returns the lines of one block, without line ends.
func blockLines(b doc.Block) []string {
	switch b := b.(type) {
	case doc.Heading:
		return []string{heading(b)}
	case doc.Para:
		return []string{inlines(b.Text, true)}
	case doc.Table:
		return table(b)
	case doc.List:
		return list(b)
	case doc.Pre:
		return pre(b)
	}
	return nil
}

func heading(h doc.Heading) string {
	level := min(max(h.Level, 1), 6)
	s := strings.Repeat("#", level)
	if t := inlines(h.Text, false); t != "" {
		s += " " + t
	}
	return s
}

func table(t doc.Table) []string {
	n := len(t.Head)
	lines := make([]string, 0, len(t.Rows)+2)
	lines = append(lines, row(t.Head, n))
	delim := make([]string, n)
	for i := range delim {
		delim[i] = "---"
		if i < len(t.Align) {
			switch t.Align[i] {
			case doc.Centre:
				delim[i] = ":-:"
			case doc.Right:
				delim[i] = "--:"
			}
		}
	}
	lines = append(lines, "|"+strings.Join(delim, "|")+"|")
	for _, r := range t.Rows {
		lines = append(lines, row(r, n))
	}
	return lines
}

// row prints cells as one table row, padded with empty cells to n.
func row(cells []doc.Cell, n int) string {
	out := make([]string, max(n, len(cells)))
	for i, c := range cells {
		out[i] = strings.Repeat(indentStep, c.Indent) + inlines(c.Text, false)
	}
	return "| " + strings.Join(out, " | ") + " |"
}

func list(l doc.List) []string {
	var lines []string
	for i, it := range l.Items {
		marker := "- "
		if l.Ordered {
			marker = strconv.Itoa(i+1) + ". "
		}
		lines = append(lines, strings.TrimRight(marker+inlines(it.Text, true), " "))
		pad := strings.Repeat(" ", len(marker))
		for _, b := range it.Blocks {
			for _, ln := range blockLines(b) {
				if ln != "" {
					ln = pad + ln
				}
				lines = append(lines, ln)
			}
		}
	}
	return lines
}

func pre(p doc.Pre) []string {
	run := 0
	for _, ln := range p.Lines {
		run = max(run, longestRun(ln, '`'))
	}
	fence := strings.Repeat("`", max(3, run+1))
	lines := make([]string, 0, len(p.Lines)+2)
	lines = append(lines, fence)
	lines = append(lines, p.Lines...)
	return append(lines, fence)
}

// inlines prints a run of spans. At the start of a paragraph or an item,
// start is true and a leading Text also escapes what would open a block.
func inlines(xs []doc.Inline, start bool) string {
	return strings.TrimSpace(spans(xs, start))
}

// spans prints a run of spans as it stands, with its edge spaces.
func spans(xs []doc.Inline, start bool) string {
	var b strings.Builder
	for i, x := range xs {
		switch x := x.(type) {
		case doc.Text:
			b.WriteString(text(string(x), start && i == 0))
		case doc.Code:
			b.WriteString(code(string(x)))
		case doc.Symbol:
			b.WriteString(string(x))
		case doc.Strong:
			if s := strings.TrimSpace(spans(x, false)); s != "" {
				b.WriteString("**" + s + "**")
			}
		case doc.Link:
			b.WriteString("[" + spans(x.Text, false) + "](" + url(x.URL) + ")")
		}
	}
	return b.String()
}

// text escapes project or fixed text: a run of white space, a line break
// included, becomes one space; a backslash precedes each character that
// Markdown reads as markup; and & prints as &amp;. With start, a leading
// character that opens a block is escaped too.
func text(s string, start bool) string {
	var b strings.Builder
	space := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			space = true
			continue
		}
		if space {
			if b.Len() > 0 || !start {
				b.WriteByte(' ')
			}
			space = false
		}
		switch r {
		case '\\', '`', '*', '_', '[', ']', '<', '|':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '&':
			b.WriteString("&amp;")
		default:
			b.WriteRune(r)
		}
	}
	if space && (b.Len() > 0 || !start) {
		b.WriteByte(' ')
	}
	out := b.String()
	if start {
		out = escapeStart(out)
	}
	return out
}

// escapeStart escapes the first character of a paragraph or an item when it
// would open a heading, a quote or a list, and the dot of a leading number.
func escapeStart(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '#', '>', '+', '-':
		return `\` + s
	}
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i > 0 && i < len(s) && (s[i] == '.' || s[i] == ')') {
		return s[:i] + `\` + s[i:]
	}
	return s
}

// code prints s as a code span: with a fence one backtick longer than the
// longest run inside it, and a space each side when it holds a backtick. A
// bar prints as \|, and a line break as a space.
func code(s string) string {
	if s == "" {
		return ""
	}
	s = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "|", `\|`).Replace(s)
	run := longestRun(s, '`')
	if run == 0 {
		return "`" + s + "`"
	}
	fence := strings.Repeat("`", run+1)
	return fence + " " + s + " " + fence
}

// url percent-encodes the spaces and parentheses of a link target.
func url(s string) string {
	return strings.NewReplacer(" ", "%20", "(", "%28", ")", "%29").Replace(s)
}

func longestRun(s string, c byte) int {
	best, cur := 0, 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			cur++
			best = max(best, cur)
		} else {
			cur = 0
		}
	}
	return best
}
