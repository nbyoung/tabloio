// Command a3cc renders the status views from tablo's view data as Markdown.
//
// It reads one JSON file (or standard input) and dispatches on its "view"
// field: global-tableau, contextual-tableau, work-blockage-tree or
// contributor-work-queue. It never reads a .tableaux directory.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Column is one gate column of a tableau.
type Column struct {
	Gate   string `json:"gate"`
	Symbol string `json:"symbol"`
}

// Cell is one cell of a tableau row.
type Cell struct {
	Gate    string `json:"gate"`
	Kind    string `json:"kind"`
	Symbols string `json:"symbols"`
}

// Status is a row's current status.
type Status struct {
	Gate         string `json:"gate"`
	State        string `json:"state"`
	Reason       string `json:"reason"`
	Note         string `json:"note"`
	Date         string `json:"date"`
	RolledUpFrom string `json:"rolled_up_from"`
	Recorder     string `json:"recorder"`
	Commit       string `json:"commit"`
	Subproject   bool   `json:"subproject_snapshot"`
}

// Row is one task of a tableau.
type Row struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Depth      int    `json:"depth"`
	Parent     bool   `json:"parent"`
	Role       string `json:"role"`
	Collapsed  int    `json:"collapsed_children"`
	Status     Status `json:"status"`
	NextGate   string `json:"next_gate"`
	Cells      []Cell `json:"cells"`
	Authorised bool   `json:"authorised"`
}

// Fold is a run of columns folded into one count.
type Fold struct {
	Side  string `json:"side"`
	From  string `json:"from"`
	To    string `json:"to"`
	Count int    `json:"count"`
}

// Tableau is the global or contextual tableau.
type Tableau struct {
	View    string   `json:"view"`
	Ref     string   `json:"ref"`
	Task    string   `json:"task"`
	Person  string   `json:"person"`
	Columns []Column `json:"columns"`
	Folded  []Fold   `json:"folded"`
	Rows    []Row    `json:"rows"`
	Next    []string `json:"next_gates_in_view"`
}

// Held is a task a blockage cause holds.
type Held struct {
	Task  string `json:"task"`
	Title string `json:"title"`
	Gate  string `json:"gate"`
}

// Cause is one root of the work-blockage tree.
type Cause struct {
	Kind     string `json:"kind"`
	Cause    string `json:"cause"`
	Resolver string `json:"resolver"`
	Action   string `json:"action"`
	Holds    int    `json:"holds"`
	Tree     []Held `json:"tree"`
}

// NotDue is a requirement that is not yet due.
type NotDue struct {
	Task     string `json:"task"`
	Requires string `json:"requires"`
	From     string `json:"from"`
	To       string `json:"to"`
	Text     string `json:"text"`
}

// Blockage is the work-blockage tree.
type Blockage struct {
	Ref        string   `json:"ref"`
	Causes     []Cause  `json:"causes"`
	NotYetDue  []NotDue `json:"not_yet_due"`
	NotDerived []string `json:"not_derived"`
}

// Item is one queue item.
type Item struct {
	Kind       string `json:"kind"`
	Task       string `json:"task"`
	Title      string `json:"title"`
	Gate       string `json:"gate"`
	Cause      string `json:"cause"`
	StatusDate string `json:"status_date"`
	Dependents int    `json:"dependents"`
}

// Queue is a contributor's work queue.
type Queue struct {
	Ref    string `json:"ref"`
	Person string `json:"person"`
	Items  []Item `json:"items"`
}

// gateSymbols names the symbol of each gate. The view data lists symbols only
// for the columns in the window, so a folded column needs this table.
var gateSymbols = map[string]string{
	"undefined": "❔", "defined": "📝", "mockup": "📌", "function": "⚙️",
	"performance": "⚡", "reliability": "⚓", "design": "📐",
	"implementation": "🛠️", "unit": "🧩", "integrate": "🖼️",
	"validate": "🌍", "release": "🚀",
}

func sym(g string) string {
	if s, ok := gateSymbols[g]; ok {
		return s
	}
	return g
}

func esc(s string) string { return strings.ReplaceAll(s, "|", `\|`) }

// foldHeader names a folded run by its end symbols and its count.
func foldHeader(f Fold) string {
	s := sym(f.From)
	if f.To != f.From {
		s += "…" + sym(f.To)
	}
	return fmt.Sprintf("%s ×%d", s, f.Count)
}

// Render writes the view in data as Markdown. level is "glance" or "detail".
func Render(data []byte, level string) (string, error) {
	var head struct {
		View string `json:"view"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return "", err
	}
	switch head.View {
	case "global-tableau", "contextual-tableau":
		var t Tableau
		if err := json.Unmarshal(data, &t); err != nil {
			return "", err
		}
		return renderTableau(t, level), nil
	case "work-blockage-tree":
		var b Blockage
		if err := json.Unmarshal(data, &b); err != nil {
			return "", err
		}
		return renderBlockage(b, level), nil
	case "contributor-work-queue":
		var q Queue
		if err := json.Unmarshal(data, &q); err != nil {
			return "", err
		}
		return renderQueue(q, level), nil
	}
	return "", fmt.Errorf("unknown view %q", head.View)
}

func rowNote(r Row) string {
	var p []string
	if r.Status.Note != "" {
		p = append(p, r.Status.Note)
	}
	if r.Status.RolledUpFrom != "" {
		p = append(p, "rolls up from `"+r.Status.RolledUpFrom+"`")
	}
	if r.Status.Subproject {
		p = append(p, "subproject at `"+r.Status.Commit+"`")
	}
	if r.Status.Date != "" {
		p = append(p, r.Status.Date)
	}
	return strings.Join(p, "; ")
}

// hidden counts the rows below row i that glance leaves out.
func hidden(rows []Row, i int) int {
	n := 0
	for _, r := range rows[i+1:] {
		if r.Depth <= rows[i].Depth {
			break
		}
		n++
	}
	return n
}

func renderTableau(t Tableau, level string) string {
	var b strings.Builder
	switch {
	case t.Task != "":
		fmt.Fprintf(&b, "## Contextual tableau: task `%s` on `%s`\n\n", t.Task, t.Ref)
	case t.Person != "":
		fmt.Fprintf(&b, "## Contextual tableau: %s on `%s`\n\n", t.Person, t.Ref)
	default:
		fmt.Fprintf(&b, "## Global tableau on `%s`\n\n", t.Ref)
	}
	// Columns: the before folds, the window, the after folds.
	hdr := []string{"Id", "Task"}
	var before, after []Fold
	for _, f := range t.Folded {
		if f.Side == "before" {
			before = append(before, f)
		} else {
			after = append(after, f)
		}
	}
	for _, f := range before {
		hdr = append(hdr, foldHeader(f))
	}
	for _, c := range t.Columns {
		hdr = append(hdr, c.Symbol)
	}
	for _, f := range after {
		hdr = append(hdr, foldHeader(f))
	}
	hdr = append(hdr, "Status")
	b.WriteString("| " + strings.Join(hdr, " | ") + " |\n")
	align := []string{"---", "---"}
	for i := 2; i < len(hdr)-1; i++ {
		align = append(align, ":-:")
	}
	align = append(align, "---")
	b.WriteString("|" + strings.Join(align, "|") + "|\n")
	for i, r := range t.Rows {
		if level == "glance" && r.Depth > 1 {
			continue
		}
		name := esc(r.Title)
		if r.Parent {
			name = "**" + name + "**"
		}
		if r.Role == "corner" {
			name += " ◀"
		}
		if level == "glance" && r.Depth == 1 {
			if n := hidden(t.Rows, i); n > 0 {
				name += fmt.Sprintf(" (%d)", n)
			}
		}
		if r.Collapsed > 0 {
			name += fmt.Sprintf(" (%d)", r.Collapsed)
		}
		name = strings.Repeat("&nbsp;&nbsp;", r.Depth) + name
		cols := []string{"`" + r.ID + "`", name}
		for range before {
			cols = append(cols, "")
		}
		for _, c := range r.Cells {
			cols = append(cols, c.Symbols)
		}
		for range after {
			cols = append(cols, "")
		}
		if level == "detail" {
			cols = append(cols, esc(rowNote(r)))
		} else {
			cols = append(cols, r.Status.Date)
		}
		b.WriteString("| " + strings.Join(cols, " | ") + " |\n")
	}
	if len(t.Folded) > 0 {
		b.WriteString("\nFolded columns show the count of leaf tasks whose current gate lies in them.\n")
	}
	if t.Person != "" {
		b.WriteString("\nRoles: ◀ corner; the other rows are the spine and the siblings.\n")
	}
	return b.String()
}

func renderBlockage(bl Blockage, level string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Work-blockage tree on `%s`\n\n", bl.Ref)
	if len(bl.Causes) == 0 {
		b.WriteString("Nothing waits.\n")
		return b.String()
	}
	for _, c := range bl.Causes {
		fmt.Fprintf(&b, "- **%s** (%s, holds %d)\n", c.Cause, c.Resolver, c.Holds)
		if level == "detail" {
			fmt.Fprintf(&b, "  - action: %s\n", c.Action)
			for _, h := range c.Tree {
				fmt.Fprintf(&b, "  - `%s` %s at %s %s\n", h.Task, h.Title, sym(h.Gate), h.Gate)
			}
		}
	}
	if level == "detail" {
		if len(bl.NotYetDue) > 0 {
			b.WriteString("\nNext, not yet due:\n\n")
			for _, n := range bl.NotYetDue {
				fmt.Fprintf(&b, "- `%s` requires `%s` (%s) from %s to %s\n", n.Task, n.Requires, n.Text, n.From, n.To)
			}
		}
		if len(bl.NotDerived) > 0 {
			b.WriteString("\nNot derived: " + strings.Join(bl.NotDerived, ", ") + ".\n")
		}
	}
	return b.String()
}

func renderQueue(q Queue, level string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Work queue for %s on `%s`\n\n", q.Person, q.Ref)
	if len(q.Items) == 0 {
		b.WriteString("The queue is empty.\n")
		return b.String()
	}
	if level == "detail" {
		b.WriteString("| Kind | Task | Gate | Detail | Dependents | Status date |\n|---|---|---|---|:-:|---|\n")
	} else {
		b.WriteString("| Kind | Task | Gate | Detail |\n|---|---|---|---|\n")
	}
	for _, it := range q.Items {
		gate := ""
		if it.Gate != "" {
			gate = sym(it.Gate) + " " + it.Gate
		}
		row := []string{it.Kind, "`" + it.Task + "` " + esc(it.Title), gate, esc(it.Cause)}
		if level == "detail" {
			row = append(row, fmt.Sprint(it.Dependents), it.StatusDate)
		}
		b.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}
	return b.String()
}

func main() {
	level := flag.String("level", "detail", "glance or detail")
	flag.Parse()
	var in io.Reader = os.Stdin
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		in = f
	}
	data, err := io.ReadAll(in)
	if err == nil {
		var out string
		if out, err = Render(data, *level); err == nil {
			fmt.Print(out)
			return
		}
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
