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
func table(t doc.Table, w int) []string {
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
			out[i] = tcell{indent: 2 * max(c.Indent, 0), text: flat(c.Text)}
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
			nat[i] = max(nat[i], c.indent+cells(c.text))
			floor[i] = max(floor[i], c.indent+min(widestWord(c.text), wordCap))
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
		return boxed(head, rows, align, nat, false)
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
		return boxed(head, rows, align, widths, true)
	}
	return stacked(head, rows, align, w)
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
func boxed(head []tcell, rows [][]tcell, align []doc.Align, widths []int, tight bool) []string {
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
			for _, ln := range wrap(c.text, widths[i]-c.indent) {
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
				gap := w - cells(s)
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
		if r[0].text == "" || len(wrap(r[0].text, widths[0]-r[0].indent)) > 1 {
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
func stacked(head []tcell, rows [][]tcell, align []doc.Align, w int) []string {
	rule := strings.Repeat("─", w)
	lines := []string{rule}
	var names []string
	for _, h := range head {
		if h.text != "" {
			names = append(names, h.text)
		}
	}
	if len(names) > 0 {
		lines = append(append(lines, fill(names, w)...), rule)
	}
	pair := func(i int, c tcell) string {
		if head[i].text == "" {
			return c.text
		}
		return head[i].text + ": " + c.text
	}
	for _, r := range rows {
		first := r[0]
		for _, ln := range wrap(first.text, max(w-first.indent, minText)) {
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
			for k, ln := range fill(pairs, w-4) {
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
func fill(pairs []string, w int) []string {
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
		case cells(line)+cells(sep)+cells(p) <= w:
			line += sep + p
		default:
			lines = append(lines, wrap(line, w)...)
			line = p
		}
	}
	return append(lines, wrap(line, w)...)
}
