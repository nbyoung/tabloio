package render

import (
	"strconv"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { register(buildGates) }

// buildGates builds the gate definition view: the four tables of the legend,
// with the criteria, severities and synopses at detail and the sources at
// provenance.
func buildGates(v *view.Gates, o Options) doc.Doc {
	p := newPage(&v.Head, o)
	f := frame{
		Name:     "gates",
		Title:    one(txt("Gate definition")),
		Question: "What do the columns and symbols mean?",
	}
	if v.Task != nil {
		f.Title = spans(f.Title, one(txt(" for ")), task(*v.Task))
	}
	var b []doc.Block
	b = append(b, p.gatesTable(v)...)
	b = append(b, p.statesTable(v)...)
	b = append(b, p.reasonsTable(v)...)
	b = append(b, p.marksTable(v)...)
	if o.Level >= view.Provenance {
		b = append(b, p.gatesProvenance(v)...)
	}
	f.Body = b
	return p.frame(f)
}

// source is the line under a table at provenance: `Source: ...`.
func source(xs ...doc.Inline) doc.Para {
	return para(spans(one(txt("Source: ")), xs, one(txt(".")))...)
}

// inFile reads `<key>` in `.tableaux/gates.yaml`.
func inFile(key string) []doc.Inline {
	return one(code(key), txt(" in "), code(".tableaux/gates.yaml"))
}

func (p *page) gatesTable(v *view.Gates) []doc.Block {
	gates := p.h.Legend.Gates
	b := []doc.Block{heading(2, txt("Gates"))}
	if len(gates) == 0 {
		return []doc.Block{countHeading(2, "Gates", 0)}
	}
	detail := p.o.Level >= view.Detail
	head := heads("#", "Symbol", "Gate", "Key")
	align := []doc.Align{doc.Right, doc.Centre, doc.Left, doc.Left}
	if detail {
		head = append(head, cellText("Criteria"))
		align = append(align, doc.Left)
	}
	forTask := map[string]view.GateFor{}
	withTask := detail && v.Task != nil
	if withTask {
		head = append(head, cell(one(txt("For "), code(v.Task.ID))))
		align = append(align, doc.Left)
		for _, g := range v.ForTask {
			forTask[g.Gate] = g
		}
	}
	var rows [][]doc.Cell
	for i, g := range gates {
		row := []doc.Cell{
			cellText(strconv.Itoa(i + 1)),
			cell(one(symbol(g.Symbol, g.Key))),
			cellText(g.Name),
			cell(one(code(g.Key))),
		}
		if detail {
			row = append(row, cellText(g.Criteria))
		}
		if withTask {
			row = append(row, cell(p.forTask(forTask[g.Key])))
		}
		rows = append(rows, row)
	}
	b = append(b, doc.Table{Align: align, Head: head, Rows: rows})
	if p.o.Level >= view.Provenance {
		b = append(b, source(inFile("gates")...))
	}
	return b
}

// forTask prints how one task's junction expands a gate: `Applies`,
// `— The gate does not apply`, or `Applies; reference: <text>, <url>,
// stated by <id>`, one reference after another.
func (p *page) forTask(g view.GateFor) []doc.Inline {
	if !g.Applies {
		var out []doc.Inline
		meaning := "The gate does not apply"
		if m, ok := p.marks["exempt"]; ok {
			out = append(out, symbol(m.Symbol, m.Key), txt(" "))
			if m.Meaning != "" {
				meaning = m.Meaning
			}
		}
		return append(out, worded(meaning)...)
	}
	out := one(txt("Applies"))
	for _, r := range g.References {
		out = append(out, txt("; reference: "))
		if r.Text != "" {
			out = append(out, txt(r.Text+", "))
		}
		out = append(out, code(r.URL))
		if g.StatedBy != "" {
			out = append(out, txt(", stated by "), code(g.StatedBy))
		}
	}
	return out
}

func (p *page) statesTable(*view.Gates) []doc.Block {
	states := p.h.Legend.States
	if len(states) == 0 {
		return []doc.Block{countHeading(2, "States", 0)}
	}
	detail := p.o.Level >= view.Detail
	head := heads("Symbol", "State")
	align := []doc.Align{doc.Centre, doc.Left}
	if detail {
		head = append(head, heads("Severity", "Synopsis")...)
		align = append(align, doc.Right, doc.Left)
	}
	var rows [][]doc.Cell
	for _, s := range states {
		row := []doc.Cell{cell(one(symbol(s.Symbol, s.Key))), cell(one(code(s.Key)))}
		if detail {
			row = append(row, cellText(strconv.Itoa(s.Severity)), cellText(s.Synopsis))
		}
		rows = append(rows, row)
	}
	b := []doc.Block{heading(2, txt("States")), doc.Table{Align: align, Head: head, Rows: rows}}
	if p.o.Level >= view.Provenance {
		b = append(b, source(inFile("states")...))
	}
	return b
}

func (p *page) reasonsTable(*view.Gates) []doc.Block {
	reasons := p.h.Legend.Reasons
	if len(reasons) == 0 {
		return []doc.Block{countHeading(2, "Reasons", 0)}
	}
	detail := p.o.Level >= view.Detail
	head := heads("Symbol", "Reason")
	align := []doc.Align{doc.Centre, doc.Left}
	if detail {
		head = append(head, cellText("Synopsis"))
		align = append(align, doc.Left)
	}
	var rows [][]doc.Cell
	var reserved *view.Reason
	for i, r := range reasons {
		row := []doc.Cell{cell(one(symbol(r.Symbol, r.Key))), cell(one(code(r.Key)))}
		if detail {
			row = append(row, cellText(r.Synopsis))
		}
		rows = append(rows, row)
		if r.Reserved && reserved == nil {
			reserved = &reasons[i]
		}
	}
	b := []doc.Block{heading(2, txt("Reasons")), doc.Table{Align: align, Head: head, Rows: rows}}
	if detail && reserved != nil {
		b = append(b, p.reservedSentence(*reserved))
	}
	if p.o.Level >= view.Provenance {
		xs := inFile("reasons")
		if reserved != nil {
			xs = append(xs, txt("; the reserved meaning of "), code(reserved.Key), txt(" in "), code("README.md"), txt(", Gates"))
		}
		b = append(b, source(xs...))
	}
	return b
}

// reservedSentence is the one sentence the gate definition keeps (design
// e3ed, G3): what the method reserves a reason for, with the reason's symbol
// and the reviewer mark's taken from the legend.
func (p *page) reservedSentence(r view.Reason) doc.Para {
	rm, ok := p.marks["reviewer"]
	if !ok {
		rm = view.Mark{Key: "reviewer"}
	}
	return para(
		txt("The method reserves "), code(r.Key),
		txt(": the contributor has handed the next junction's work to its reviewer, and the work queue lists the review as owed. "),
		symbol(r.Symbol, r.Key),
		txt(" beside a state says the work waits for its reviewer; "),
		symbol(rm.Symbol, rm.Key),
		txt(" in a cell says a reviewer accepts the work there."),
	)
}

func (p *page) marksTable(v *view.Gates) []doc.Block {
	marks := p.h.Legend.Marks
	if len(marks) == 0 {
		return []doc.Block{countHeading(2, "Junction marks", 0)}
	}
	detail := p.o.Level >= view.Detail
	head := heads("Mark", "Meaning")
	align := []doc.Align{doc.Centre, doc.Left}
	if detail {
		head = append(head, cellText("Junction"))
		align = append(align, doc.Left)
	}
	var rows [][]doc.Cell
	for _, m := range marks {
		row := []doc.Cell{cell(one(symbol(m.Symbol, m.Key))), cell(worded(m.Meaning))}
		if detail {
			row = append(row, cell(worded(m.Junction)))
		}
		rows = append(rows, row)
	}
	b := []doc.Block{heading(2, txt("Junction marks")), doc.Table{Align: align, Head: head, Rows: rows}}
	if p.o.Level >= view.Provenance {
		xs := one(code("README.md"), txt(", Junctions"))
		if v.Language != "" {
			xs = append(xs, txt(", at language "+v.Language))
		}
		b = append(b, source(xs...))
	}
	return b
}

// gatesProvenance builds the Provenance section: the facts, the files with
// the commits that last changed them, and the commands.
func (p *page) gatesProvenance(v *view.Gates) []doc.Block {
	r := p.h.Ref
	refValue := one(code(refName(r)))
	if r.Commit != "" {
		refValue = append(refValue, txt(" at "), code(short(r.Commit)))
	}
	if r.Date != "" {
		refValue = append(refValue, txt(", "+r.Date))
	}
	if r.OnTrunk {
		refValue = append(refValue, txt(", on the trunk"))
	} else {
		refValue = append(refValue, txt(", off the trunk"))
	}
	versionFile := func(key string) []doc.Inline {
		return one(code(key), txt(" in "), code(".tableaux/version.yaml"))
	}
	facts := [][]doc.Cell{
		{cellText("Language version"), cellText(v.Language), cell(versionFile("tableaux"))},
		{cellText("Trunk"), cell(codeOrNone(v.Trunk)), cell(versionFile("trunk"))},
		{cellText("Ref"), cell(refValue), cell(one(txt("The "), code(flagRef), txt(" parameter")))},
	}
	b := []doc.Block{
		heading(2, txt("Provenance")),
		doc.Table{Head: heads("Fact", "Value", "Source"), Rows: facts},
	}
	if len(v.Files) > 0 {
		var rows [][]doc.Cell
		for _, f := range v.Files {
			c := f.Commit
			rows = append(rows, []doc.Cell{
				cell(one(code(f.Path))),
				cell(commit(c.Hash)),
				cellText(c.Date),
				cell(p.actor(c.Author)),
				cell(p.actor(c.Committer)),
				cellText(c.Subject),
			})
		}
		b = append(b, doc.Table{
			Head: heads("File", "Last changed by", "Date", "Author", "Committer", "Subject"),
			Rows: rows,
		})
	}
	if pre := commandLines(v.Commands); pre != nil {
		b = append(b, *pre)
	}
	return b
}

// codeOrNone prints a value in a code span, or nothing when it is empty.
func codeOrNone(s string) []doc.Inline {
	if s == "" {
		return nil
	}
	return one(code(s))
}
