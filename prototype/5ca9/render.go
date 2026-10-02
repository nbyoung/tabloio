package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// Render draws the envelope's data in one format. It is a generic stand-in
// for the per-view renderers of task 1093: a heading, the scalar fields, then
// each list of objects as a table and each object as a section. Go maps lose
// key order, so keys sort, with the usual identifying keys first.
func Render(v *view, e *Envelope, format string) string {
	var b strings.Builder
	title := strings.ToUpper(v.name[:1]) + v.name[1:]
	heading(&b, format, 1, title)
	ref, _ := e.Ref["name"].(string)
	fmt.Fprintf(&b, "%s at %s, level %v\n\n", v.tablo, code(format, ref), e.Data["level"])
	section(&b, format, 2, e.Data)
	return b.String()
}

func code(format, s string) string {
	if format == "markdown" {
		return "`" + s + "`"
	}
	return s
}

func heading(b *strings.Builder, format string, depth int, s string) {
	if format == "markdown" {
		pf(b, "%s %s\n\n", strings.Repeat("#", depth), s)
		return
	}
	rule := "═"
	if depth > 1 {
		rule = "─"
	}
	pf(b, "%s\n%s\n\n", s, strings.Repeat(rule, utf8.RuneCountInString(s)))
}

var first = []string{"id", "task", "key", "gate", "kind", "rule", "date", "title"}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	rank := func(k string) int {
		for i, f := range first {
			if f == k {
				return i
			}
		}
		return len(first)
	}
	sort.Slice(keys, func(i, j int) bool {
		ri, rj := rank(keys[i]), rank(keys[j])
		if ri != rj {
			return ri < rj
		}
		return keys[i] < keys[j]
	})
	return keys
}

func isScalar(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return false
	}
	return true
}

func section(b *strings.Builder, format string, depth int, m map[string]any) {
	skip := map[string]bool{"view": true, "level": true, "ref": true}
	keys := sortedKeys(m)
	for _, k := range keys {
		if v := m[k]; isScalar(v) && !skip[k] {
			pf(b, "- %s: %s\n", k, cell(v))
		}
	}
	b.WriteString("\n")
	for _, k := range keys {
		switch v := m[k].(type) {
		case []any:
			if len(v) == 0 {
				continue
			}
			heading(b, format, depth, k)
			list(b, format, depth, v)
		case map[string]any:
			heading(b, format, depth, k)
			section(b, format, depth+1, v)
		}
	}
}

func list(b *strings.Builder, format string, depth int, l []any) {
	var rows []map[string]any
	for _, x := range l {
		m, ok := x.(map[string]any)
		if !ok {
			parts := make([]string, len(l))
			for i, y := range l {
				parts[i] = cell(y)
			}
			b.WriteString(strings.Join(parts, ", ") + "\n\n")
			return
		}
		rows = append(rows, m)
	}
	cols := map[string]bool{}
	for _, r := range rows {
		for k := range r {
			cols[k] = true
		}
	}
	all := map[string]any{}
	for k := range cols {
		all[k] = nil
	}
	header := sortedKeys(all)
	body := make([][]string, len(rows))
	for i, r := range rows {
		for _, k := range header {
			body[i] = append(body[i], cell(r[k]))
		}
	}
	table(b, format, header, body)
}

// cell shows a scalar as itself and anything nested as compact JSON.
func cell(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return strings.ReplaceAll(x, "|", "/")
	case float64:
		return fmt.Sprintf("%g", x)
	}
	j, _ := json.Marshal(v)
	return strings.ReplaceAll(string(j), "|", "/")
}

func table(b *strings.Builder, format string, header []string, body [][]string) {
	w := make([]int, len(header))
	for i, h := range header {
		w[i] = utf8.RuneCountInString(h)
	}
	for _, r := range body {
		for i, c := range r {
			if n := utf8.RuneCountInString(c); n > w[i] {
				w[i] = n
			}
		}
	}
	pad := func(s string, n int) string {
		return s + strings.Repeat(" ", n-utf8.RuneCountInString(s))
	}
	if format == "markdown" {
		line := func(r []string) {
			b.WriteString("|")
			for i, c := range r {
				b.WriteString(" " + pad(c, w[i]) + " |")
			}
			b.WriteString("\n")
		}
		line(header)
		sep := make([]string, len(header))
		for i := range sep {
			sep[i] = strings.Repeat("-", w[i])
		}
		line(sep)
		for _, r := range body {
			line(r)
		}
		b.WriteString("\n")
		return
	}
	rule := func(l, m, r string) {
		b.WriteString(l)
		for i, n := range w {
			b.WriteString(strings.Repeat("─", n+2))
			if i < len(w)-1 {
				b.WriteString(m)
			}
		}
		b.WriteString(r + "\n")
	}
	line := func(r []string) {
		b.WriteString("│")
		for i, c := range r {
			b.WriteString(" " + pad(c, w[i]) + " │")
		}
		b.WriteString("\n")
	}
	rule("┌", "┬", "┐")
	line(header)
	rule("├", "┼", "┤")
	for _, r := range body {
		line(r)
	}
	rule("└", "┴", "┘")
	b.WriteString("\n")
}
