package render

import (
	"strconv"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { registerStatus(buildTableau) }

// The related views of the two tableaux, in the order of the See also line.
var (
	tableauSeeAlso = []string{"gates", "context", "task", "blockage", "queue", "history"}
	contextSeeAlso = []string{"gates", "tableau", "task", "blockage", "queue", "history"}
)

// buildTableau builds the global and the contextual tableau: one grid, its
// key line and the notes at glance; the folded columns at detail; and the
// Git facts behind each cell at provenance. The grid derives nothing: tablo
// decides every cell, the window and the roll-ups, and the builder draws
// what arrives. It returns an error that names a row with the wrong number
// of cells.
func buildTableau(v *view.Tableau, o Options) (doc.Doc, error) {
	if err := v.Check(); err != nil {
		return doc.Doc{}, err
	}
	p := newPage(&v.Head, o)

	var b []doc.Block
	if len(v.Rows) == 0 {
		b = append(b, emptyForm("No task stands in this corner."))
	} else {
		b = append(b, p.gridTable(v))
		if k := p.gridKey(v); k != nil {
			b = append(b, *k)
		}
		// The global tableau's glance holds the notes; the contextual
		// tableau's holds the grid and its key line alone.
		if !v.Contextual || o.Level >= view.Detail {
			b = append(b, p.gridNotes(v)...)
		}
		if o.Level >= view.Detail {
			b = append(b, p.gridFolded(v)...)
		}
	}
	if o.Level >= view.Provenance {
		b = append(b, p.gridProvenance(v)...)
	}

	f := frame{
		Name:     "tableau",
		Title:    one(txt("Global tableau")),
		Question: "How does the whole project stand?",
		Legend:   true,
		SeeAlso:  tableauSeeAlso,
		Body:     b,
	}
	if v.Contextual {
		f.Name = "context"
		f.Title = one(txt("Contextual tableau"))
		f.Question = "How does my corner stand?"
		f.SeeAlso = contextSeeAlso
		switch {
		case v.Subject != nil:
			f.Title = spans(f.Title, one(txt(": ")), task(*v.Subject))
		case v.Params.Person != "":
			f.Title = spans(f.Title, one(txt(": "+v.Params.Person)))
		}
	}
	return p.frame(f), nil
}

// gateSymbol prints a gate's bare symbol, or its key where the legend has none.
func (p *page) gateSymbol(key string) doc.Inline { return symbol(p.gates[key], key) }

// gridHead prints a column header: the gate's symbol; for a run, the first
// symbol, an ellipsis and the last; for a folded column, ` ×` and its count.
func (p *page) gridHead(c view.Column) []doc.Inline {
	if len(c.Gates) == 0 {
		return nil
	}
	out := one(p.gateSymbol(c.Gates[0]))
	if len(c.Gates) > 1 {
		out = append(out, txt("…"), p.gateSymbol(c.Gates[len(c.Gates)-1]))
	}
	if c.Folded {
		out = append(out, txt(" ×"+strconv.Itoa(c.Count)))
	}
	return out
}

// gridTable is the grid: Id, Task, then one centred column per Column.
func (p *page) gridTable(v *view.Tableau) doc.Table {
	head := heads("Id", "Task")
	align := []doc.Align{doc.Left, doc.Left}
	for _, c := range v.Columns {
		head = append(head, cell(p.gridHead(c)))
		align = append(align, doc.Centre)
	}
	var rows [][]doc.Cell
	for _, r := range v.Rows {
		row := []doc.Cell{cell(p.fileLink(r.TaskRef, one(code(r.ID)))), p.gridTask(r)}
		for _, c := range r.Cells {
			row = append(row, p.gridCell(c))
		}
		rows = append(rows, row)
	}
	return doc.Table{Align: align, Head: head, Rows: rows}
}

// gridTask is the Task cell: the title indented by depth, in bold for a
// parent, then the count of hidden rows, then the role.
func (p *page) gridTask(r view.GridRow) doc.Cell {
	out := one(txt(r.Title))
	if r.Parent {
		out = one(doc.Strong{txt(r.Title)})
	}
	if r.Hidden > 0 {
		out = append(out, txt(" ("+strconv.Itoa(r.Hidden)+")"))
	}
	if r.Role != "" {
		out = append(out, txt(" · "+r.Role))
	}
	return doc.Cell{Indent: r.Depth, Text: out}
}

// gridCell prints a cell by its kind: nothing for empty; the state's symbol
// and the reason's for a status; the marks' symbols side by side; and the
// exempt mark. Brackets enclose a cell where the person in view acts (R5).
func (p *page) gridCell(c view.GridCell) doc.Cell {
	var xs []doc.Inline
	switch c.Kind {
	case "status":
		if c.State != "" {
			xs = append(xs, symbol(p.states[c.State], c.State))
		}
		if c.Reason != "" {
			xs = append(xs, symbol(p.reasons[c.Reason], c.Reason))
		}
	case "marks":
		xs = p.markSymbols(c.Marks)
	case "exempt":
		sym := "—"
		if m, ok := p.marks["exempt"]; ok && m.Symbol != "" {
			sym = m.Symbol
		}
		xs = one(doc.Symbol(sym))
	}
	if c.Acts && len(xs) > 0 {
		xs = spans(one(txt("[")), xs, one(txt("]")))
	}
	return cell(xs)
}

// gridKey is the key line (R2): the gates the headers show, then the states,
// the reasons and the marks the cells show, in legend order.
func (p *page) gridKey(v *view.Tableau) *doc.Para {
	gates, states, reasons, marks := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, c := range v.Columns {
		if n := len(c.Gates); n > 0 {
			gates[c.Gates[0]] = true
			gates[c.Gates[n-1]] = true
		}
	}
	for _, r := range v.Rows {
		for _, c := range r.Cells {
			switch c.Kind {
			case "status":
				states[c.State] = true
				reasons[c.Reason] = true
			case "marks":
				for _, m := range c.Marks {
					marks[m] = true
				}
			case "exempt":
				marks["exempt"] = true
			}
		}
	}
	var runs [][]doc.Inline
	for _, g := range p.h.Legend.Gates {
		if gates[g.Key] {
			runs = append(runs, p.gate(g.Key))
		}
	}
	for _, s := range p.h.Legend.States {
		if states[s.Key] {
			runs = append(runs, p.state(s.Key))
		}
	}
	for _, r := range p.h.Legend.Reasons {
		if reasons[r.Key] {
			runs = append(runs, p.reason(r.Key))
		}
	}
	for _, m := range p.h.Legend.Marks {
		if marks[m.Key] {
			runs = append(runs, p.markMeaning(m))
		}
	}
	if len(runs) == 0 {
		return nil
	}
	pr := para(append(joined(" · ", runs...), txt("."))...)
	return &pr
}

// markMeaning prints a mark as its symbol and its meaning with a lower-case
// initial, for a key line.
func (p *page) markMeaning(m view.Mark) []doc.Inline {
	meaning := lowerInitial(m.Meaning)
	if meaning == "" {
		meaning = m.Key
	}
	return spans(one(symbol(m.Symbol, m.Key), txt(" ")), worded(meaning))
}

// originKey identifies an origin, so that rows with one origin share a line.
func originKey(o *view.Origin) string {
	if o == nil {
		return ""
	}
	return o.Kind + "\x00" + o.Task.ID + "\x00" + o.URL + "\x00" + o.Pin + "\x00" + o.Gate
}

// gridOrigin prints where a note comes from: nothing for a status file;
// “rolled up from <task>“ for a roll-up; “<task> in `<url>` at `<pin>`,
// at <gate>“ for a snapshot, the task and the gate being the subproject's.
func (p *page) gridOrigin(o *view.Origin) []doc.Inline {
	if o == nil {
		return nil
	}
	switch o.Kind {
	case "rollup":
		return spans(one(txt("rolled up from ")), task(o.Task))
	case "snapshot":
		out := spans(task(o.Task), one(txt(" in "), code(o.URL), txt(" at "), code(o.Pin)))
		if o.Gate != "" {
			out = spans(out, one(txt(", at ")), p.gate(o.Gate))
		}
		return out
	}
	return nil
}

// gridNotes is the notes table beneath the grid. Rows that share a date, a
// note and an origin share one line, placed where the first of them stands
// in display order; a row with no date and no note gives no line. The
// section does not print when no row has either.
func (p *page) gridNotes(v *view.Tableau) []doc.Block {
	type group struct {
		ids    []string
		status view.Status
	}
	var groups []*group
	index := map[string]*group{}
	for _, r := range v.Rows {
		st := r.Status
		if st.Date == "" && st.Note == "" {
			continue
		}
		key := st.Date + "\x01" + st.Note + "\x01" + originKey(st.From)
		g, ok := index[key]
		if !ok {
			g = &group{status: st}
			index[key] = g
			groups = append(groups, g)
		}
		g.ids = append(g.ids, r.ID)
	}
	if len(groups) == 0 {
		return nil
	}
	var rows [][]doc.Cell
	for _, g := range groups {
		rows = append(rows, []doc.Cell{
			cell(ids(g.ids)), cellText(g.status.Date), cellText(g.status.Note), cell(p.gridOrigin(g.status.From)),
		})
	}
	return []doc.Block{heading(2, txt("Notes")), doc.Table{Head: heads("Rows", "Date", "Note", "From"), Rows: rows}}
}

// gridFolded is the folded columns table: one line per gate of each folded
// column, the first cell repeating the column's header without its count.
func (p *page) gridFolded(v *view.Tableau) []doc.Block {
	var rows [][]doc.Cell
	for _, c := range v.Columns {
		if !c.Folded {
			continue
		}
		plain := c
		plain.Folded = false
		for i, g := range c.Gates {
			var n []doc.Inline
			if i < len(c.ByGate) {
				n = one(txt(strconv.Itoa(c.ByGate[i])))
			}
			rows = append(rows, []doc.Cell{cell(p.gridHead(plain)), cell(p.gate(g)), cell(n)})
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return []doc.Block{
		heading(2, txt("Folded columns")),
		doc.Table{Align: []doc.Align{doc.Centre, doc.Left, doc.Right}, Head: heads("Folded column", "Gate", "Tasks"), Rows: rows},
	}
}

// gridProvenance builds the blocks that only provenance adds: the commit
// behind each status cell, each roll-up, each group of marks and each
// snapshot, then the commands.
func (p *page) gridProvenance(v *view.Tableau) []doc.Block {
	b := []doc.Block{countHeading(2, "Status cells", len(v.Statuses))}
	if len(v.Statuses) > 0 {
		var rows [][]doc.Cell
		for _, s := range v.Statuses {
			rows = append(rows, []doc.Cell{
				cell(ids(s.Tasks)), cell(p.gate(s.Gate)), cellText(s.Commit.Date), cell(p.actor(s.Commit.Author)),
				cell(commitSubject(s.Commit)), cell(p.actor(s.Commit.Committer)), cell(trailers(s.Commit)),
			})
		}
		b = append(b, doc.Table{
			Head: heads("Rows", "Cell", "Date", "Recorder", "Deciding commit", "Committer", "Trailers"), Rows: rows,
		})
	}

	b = append(b, countHeading(2, "Roll-ups", len(v.Rollups)))
	if len(v.Rollups) > 0 {
		var rows [][]doc.Cell
		for _, r := range v.Rollups {
			rows = append(rows, []doc.Cell{
				cell(task(r.Parent)), cell(p.phrase(r.Status)), cell(task(r.Child)), cellText(r.Status.Date),
			})
		}
		b = append(b, doc.Table{Head: heads("Parent", "Shows", "Rolls up from", "Date"), Rows: rows})
	}

	b = append(b, countHeading(2, "Marks", len(v.MarkGroups)))
	if len(v.MarkGroups) > 0 {
		shown := map[string]bool{}
		var rows [][]doc.Cell
		for _, g := range v.MarkGroups {
			for _, m := range g.Marks {
				shown[m] = true
			}
			rows = append(rows, []doc.Cell{
				cell(p.gate(g.Gate)), cell(p.markSymbols(g.Marks)), cell(ids(g.Tasks)),
				cell(p.markContributor(g)), cell(p.markReviewer(g)),
			})
		}
		b = append(b, doc.Table{
			Align: []doc.Align{doc.Left, doc.Centre, doc.Left, doc.Left, doc.Left},
			Head:  heads("Gate", "Mark", "Rows", "Contributor and model", "Reviewer"), Rows: rows,
		})
		if k := p.markKey(shown); k != nil {
			b = append(b, *k)
		}
	}

	b = append(b, countHeading(2, "Snapshots", len(v.Snapshots)))
	if len(v.Snapshots) > 0 {
		var rows [][]doc.Cell
		for _, s := range v.Snapshots {
			setBy := spans(commit(s.SetBy.Hash), prefixed(", ", textRun(s.SetBy.Date)), prefixed(", ", p.person(s.SetBy.Author.Email)))
			root := spans(task(s.Root), prefixed(": ", p.phrase(s.Status)))
			var from []doc.Inline
			if s.Status.From != nil {
				from = task(s.Status.From.Task)
			}
			rows = append(rows, []doc.Cell{
				cell(one(code(s.Task))), cell(one(code(s.URL))), cell(one(code(s.Pin))), cell(setBy), cell(root), cell(from),
			})
		}
		b = append(b, doc.Table{
			Head: heads("Row", "Subproject", "Pin", "Pin set by", "Root task at the pin", "Rolls up from"), Rows: rows,
		})
	}
	if pre := commandLines(v.Commands); pre != nil {
		b = append(b, *pre)
	}
	return b
}

// trailers prints a commit's trailers as code spans joined by commas.
func trailers(c view.Commit) []doc.Inline {
	var runs [][]doc.Inline
	for _, t := range c.Trailers {
		runs = append(runs, one(code(t)))
	}
	return joined(", ", runs...)
}

// markFrom prints where a field comes from: `, from <id>` after the value,
// `, by default` for the plain default.
func markFrom(src string) []doc.Inline {
	switch src {
	case "":
		return nil
	case "default":
		return one(txt(", by default"))
	}
	return one(txt(", from "), code(src))
}

// markContributor prints the fourth cell of a marks line: the contributor,
// the model and the source; for an exempt group, `Exempt, from <id>`.
func (p *page) markContributor(g view.MarkGroup) []doc.Inline {
	if g.ExemptFrom != "" {
		return spans(one(txt("Exempt")), markFrom(g.ExemptFrom))
	}
	if g.Contributor == "" && g.Model == "" {
		return nil
	}
	out := p.person(g.Contributor)
	if g.Model != "" {
		if len(out) > 0 {
			out = append(out, txt(", "))
		}
		out = append(out, code(g.Model))
	}
	return spans(out, markFrom(g.ContributorFrom))
}

// markReviewer prints the reviewer and the source, or nothing.
func (p *page) markReviewer(g view.MarkGroup) []doc.Inline {
	if g.Reviewer == "" {
		return nil
	}
	return spans(p.person(g.Reviewer), markFrom(g.ReviewerFrom))
}
