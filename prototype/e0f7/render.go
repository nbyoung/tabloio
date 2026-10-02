package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Tableau is the part of tablo's global-tableau view data the renderer reads.
type Tableau struct {
	Ref     string `json:"ref"`
	Level   string `json:"level"`
	Columns []struct {
		Gate   string `json:"gate"`
		Symbol string `json:"symbol"`
	} `json:"columns"`
	Folded []struct {
		Side  string `json:"side"`
		From  string `json:"from"`
		To    string `json:"to"`
		Count int    `json:"count"`
	} `json:"folded"`
	Rows []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Depth int    `json:"depth"`
		Cells []struct {
			Symbols string `json:"symbols"`
		} `json:"cells"`
	} `json:"rows"`
}

// ParseTableau reads the JSON tablo view emits for the global tableau.
func ParseTableau(data []byte) (*Tableau, error) {
	var t Tableau
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

const minTask = 8

// RenderTableau draws t as a box table no wider than width cells, when the
// columns allow it. It shrinks the task column first, with an ellipsis.
func RenderTableau(t *Tableau, width int, m Model) string {
	head := []string{"Id", "Task"}
	for _, c := range t.Columns {
		head = append(head, Prepare(c.Symbol, m))
	}
	var body [][]string
	for _, r := range t.Rows {
		row := []string{r.ID, strings.Repeat("  ", r.Depth) + r.Title}
		for _, c := range r.Cells {
			row = append(row, Prepare(c.Symbols, m))
		}
		body = append(body, row)
	}
	n := len(head)
	w := make([]int, n)
	for _, row := range append([][]string{head}, body...) {
		for i, s := range row {
			if x := Width(s, m); x > w[i] {
				w[i] = x
			}
		}
	}
	total := func() int { // borders and one space of padding each side
		s := 1
		for _, x := range w {
			s += x + 3
		}
		return s
	}
	if over := total() - width; over > 0 {
		w[1] = max(minTask, w[1]-over)
	}
	line := func(l, mid, r string) string {
		parts := make([]string, n)
		for i, x := range w {
			parts[i] = strings.Repeat("─", x+2)
		}
		return l + strings.Join(parts, mid) + r + "\n"
	}
	rowOut := func(row []string) string {
		parts := make([]string, n)
		for i, s := range row {
			align := byte('c')
			if i < 2 {
				align = 'l'
			}
			parts[i] = " " + Pad(Truncate(s, w[i], m), w[i], align, m) + " "
		}
		return "│" + strings.Join(parts, "│") + "│\n"
	}
	var b strings.Builder
	b.WriteString(line("┌", "┬", "┐"))
	b.WriteString(rowOut(head))
	b.WriteString(line("├", "┼", "┤"))
	for _, r := range body {
		b.WriteString(rowOut(r))
	}
	b.WriteString(line("└", "┴", "┘"))
	for _, f := range t.Folded {
		fmt.Fprintf(&b, "folded %s: %s to %s, %d tasks\n", f.Side, f.From, f.To, f.Count)
	}
	return b.String()
}

// Probe draws each symbol of t between bars, so that a terminal shows at a
// glance whether it agrees with model m: the right bars line up when it does.
func Probe(t *Tableau, m Model) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-16s %s\n", "gate", "symbol (bars align when the terminal agrees)")
	seen := map[string]bool{}
	add := func(label, s string) {
		if seen[s] || s == "" {
			return
		}
		seen[s] = true
		s = Prepare(s, m)
		fmt.Fprintf(&b, "%-16s │%s│ width %d\n", label, Pad(s, 6, 'l', m), Width(s, m))
	}
	for _, c := range t.Columns {
		add(c.Gate, c.Symbol)
	}
	for _, r := range t.Rows {
		for _, c := range r.Cells {
			add("cell", c.Symbols)
		}
	}
	return b.String()
}
