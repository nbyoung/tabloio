package render

import (
	"strconv"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { register(buildAssignment) }

// assignmentSeeAlso lists the views related to the task assignment, in the
// order of the See also line.
var assignmentSeeAlso = []string{"queue", "authority", "task", "tableau", "context", "blockage", "history", "audit"}

// buildAssignment builds the task assignment view: one row per email at
// glance; at detail a section per email with its tasks, its junctions by
// gate, its counts and its authority; at provenance the file and the
// ancestor behind each position, and the commits.
func buildAssignment(v *view.Assignment, o Options) doc.Doc {
	p := newPage(&v.Head, o)
	titles := taskTitles(v)

	b := p.peopleBlocks(v)
	if o.Level >= view.Detail {
		for _, s := range v.Sections {
			b = append(b, p.personSection(s)...)
		}
	}
	if o.Level >= view.Provenance {
		for _, s := range v.Sections {
			b = append(b, p.personProvenance(s, titles)...)
		}
		b = append(b, p.commitsSection(v)...)
	}
	return p.frame(frame{
		Name:     "assignment",
		Title:    one(txt("Task assignment")),
		Question: "What does each person carry?",
		Legend:   true,
		SeeAlso:  assignmentSeeAlso,
		Body:     b,
	})
}

// peopleBlocks is the glance: the table of the emails and, when a task hands
// its next junction to a subproject, the line that names them.
func (p *page) peopleBlocks(v *view.Assignment) []doc.Block {
	if len(v.People) == 0 {
		return []doc.Block{emptyForm("No email is in view.")}
	}
	var rows [][]doc.Cell
	for _, r := range v.People {
		var models []doc.Inline
		for i, m := range r.Models {
			if i > 0 {
				models = append(models, txt(", "))
			}
			models = append(models, code(m))
		}
		rows = append(rows, []doc.Cell{
			cellText(r.Email),
			cellText(strconv.Itoa(r.Assigned)),
			cellText(strconv.Itoa(r.ContributesNext)),
			cellText(strconv.Itoa(r.ReviewsNext)),
			cell(models),
		})
	}
	b := []doc.Block{doc.Table{
		Align: []doc.Align{doc.Left, doc.Right, doc.Right, doc.Right, doc.Left},
		Head:  heads("Email", "Assigned", "Contributes next", "Reviews next", "Models"),
		Rows:  rows,
	}}
	if len(v.Recursive) > 0 {
		var lead []doc.Inline
		if m, ok := p.marks["subproject"]; ok {
			lead = one(symbol(m.Symbol, m.Key), txt(" "))
		}
		b = append(b, para(spans(lead,
			one(txt(strconv.Itoa(len(v.Recursive))+" tasks hand their next junction to a subproject: ")),
			ids(v.Recursive), one(txt(".")))...))
	}
	return b
}

// personSection builds the detail of one email.
func (p *page) personSection(s view.PersonSection) []doc.Block {
	b := []doc.Block{heading(2, txt(s.Email))}
	b = append(b, p.assigned(s)...)
	b = append(b, p.noContributor(s)...)
	b = append(b, p.junctionSection("Contributes", s.Contributes, false)...)
	b = append(b, p.junctionSection("Reviews", s.Reviews, true)...)
	b = append(b, p.modelsByGate(s)...)
	b = append(b, p.countsByGate(s)...)
	b = append(b, p.authoritySection(s)...)
	return b
}

// taskCell prints the task cell of a row about that task, linked to its file.
func (p *page) taskCell(t view.TaskRef) []doc.Inline { return p.fileLink(t, task(t)) }

// assigned is `### Tasks assigned: n`: the task, its parent, its status and
// its next junction, with the key line of the marks the table shows.
func (p *page) assigned(s view.PersonSection) []doc.Block {
	b := []doc.Block{countHeading(3, "Tasks assigned", len(s.Assigned))}
	if len(s.Assigned) == 0 {
		return b
	}
	shown := map[string]bool{}
	var rows []taskRow
	for _, r := range s.Assigned {
		var next []doc.Inline
		if r.Next != nil {
			for _, m := range r.Next.Marks {
				shown[m] = true
			}
			next = p.gate(r.Gate)
			if len(r.Next.Marks) > 0 {
				next = spans(next, one(txt(" ")), p.markSymbols(r.Next.Marks))
			}
		}
		var under []doc.Inline
		if r.Under != "" {
			under = one(code(r.Under))
		}
		rows = append(rows, taskRow{ID: r.ID, Task: p.taskCell(r.TaskRef), Rest: []doc.Cell{
			cell(under), cell(p.status(r.Status)), cell(next),
		}})
	}
	b = append(b, doc.Table{Head: heads("Task", "Under", "Status", "Next junction"), Rows: p.foldRows(rows)})
	if k := p.markKey(shown); k != nil {
		b = append(b, *k)
	}
	return b
}

// noContributor is `### No contributor in this project: n`, the tasks whose
// next junction a subproject carries; a section with no task prints nothing.
func (p *page) noContributor(s view.PersonSection) []doc.Block {
	if len(s.Recursive) == 0 {
		return nil
	}
	var mark []doc.Inline
	if m, ok := p.marks["subproject"]; ok {
		mark = one(txt(" "), symbol(m.Symbol, m.Key))
	}
	var rows []taskRow
	for _, r := range s.Recursive {
		rows = append(rows, taskRow{ID: r.Task.ID, Task: p.taskCell(r.Task), Rest: []doc.Cell{
			cell(spans(p.gate(r.Gate), mark)),
			cell(one(code(r.Subproject.URL))),
			cell(spans(one(code(r.Subproject.Task.ID), txt(" ")), p.phrase(r.Status))),
		}})
	}
	return []doc.Block{
		countHeading(3, "No contributor in this project", len(s.Recursive)),
		doc.Table{Head: heads("Task", "Next gate", "Subproject", "Its root task stands at"), Rows: p.foldRows(rows)},
	}
}

// nextCount counts the junctions that are next.
func nextCount(rows []view.JunctionRow) int {
	n := 0
	for _, r := range rows {
		if r.When == "next" {
			n++
		}
	}
	return n
}

// junctionSection is `### Contributes: n junctions, m next` or `### Reviews:
// ...`, then a group per gate in the legend's order (R10 folds each table).
// Under Reviews the table names the contributor in place of the reviewer.
func (p *page) junctionSection(name string, rows []view.JunctionRow, reviews bool) []doc.Block {
	if len(rows) == 0 {
		return []doc.Block{countHeading(3, name, 0)}
	}
	b := []doc.Block{heading(3, txt(name+": "+strconv.Itoa(len(rows))+" junctions, "+strconv.Itoa(nextCount(rows))+" next"))}
	for _, g := range p.byGate(rows) {
		head := heads("Task", "Model", "Reviewer", "When")
		if reviews {
			head = heads("Task", "Contributor", "Model", "When")
		}
		var trs []taskRow
		for _, r := range g.rows {
			var model []doc.Inline
			if r.Model != "" {
				model = one(code(r.Model))
			}
			rest := []doc.Cell{cell(model), cell(p.person(r.Reviewer)), cellText(r.When)}
			if reviews {
				rest = []doc.Cell{cell(p.person(r.Contributor)), cell(model), cellText(r.When)}
			}
			trs = append(trs, taskRow{ID: r.Task.ID, Task: p.taskCell(r.Task), Rest: rest})
		}
		b = append(b,
			para(spans(one(doc.Strong(p.gate(g.gate))),
				one(txt(" · "+strconv.Itoa(len(g.rows))+" junctions · "+strconv.Itoa(nextCount(g.rows))+" next")))...),
			doc.Table{Head: head, Rows: p.foldRows(trs)})
	}
	return b
}

// gateGroup is the junctions of one gate.
type gateGroup struct {
	gate string
	rows []view.JunctionRow
}

// byGate groups rows by gate in the legend's order, and prints a group only
// for a gate that holds a row. A gate the legend lacks follows, in the order
// it first appears. Rows keep their order inside a group.
func (p *page) byGate(rows []view.JunctionRow) []gateGroup {
	at := map[string]int{}
	var groups []gateGroup
	add := func(key string) int {
		if i, ok := at[key]; ok {
			return i
		}
		at[key] = len(groups)
		groups = append(groups, gateGroup{gate: key})
		return at[key]
	}
	present := map[string]bool{}
	for _, r := range rows {
		present[r.Gate] = true
	}
	for _, g := range p.h.Legend.Gates {
		if present[g.Key] {
			add(g.Key)
		}
	}
	for _, r := range rows {
		i := add(r.Gate)
		groups[i].rows = append(groups[i].rows, r)
	}
	return groups
}

// modelsByGate is `### Models by gate`, for an agent: the models it runs, by
// gate, with the task that states each.
func (p *page) modelsByGate(s view.PersonSection) []doc.Block {
	if len(s.Models) == 0 {
		return nil
	}
	var rows [][]doc.Cell
	for _, m := range s.Models {
		rows = append(rows, []doc.Cell{
			cell(p.gate(m.Gate)),
			cell(one(code(m.Model))),
			cell(one(code(m.StatedBy))),
			cellText(strconv.Itoa(m.Junctions)),
			cellText(strconv.Itoa(m.Next)),
		})
	}
	return []doc.Block{
		heading(3, txt("Models by gate")),
		doc.Table{
			Align: []doc.Align{doc.Left, doc.Left, doc.Left, doc.Right, doc.Right},
			Head:  heads("Gate", "Model", "Stated by", "Junctions", "Next"),
			Rows:  rows,
		},
	}
}

// countsByGate is `### Counts by gate`: every gate of the legend, then the
// total in bold.
func (p *page) countsByGate(s view.PersonSection) []doc.Block {
	by := map[string]view.GateCount{}
	for _, c := range s.Counts {
		by[c.Gate] = c
	}
	var keys []string
	listed := map[string]bool{}
	for _, g := range p.h.Legend.Gates {
		keys = append(keys, g.Key)
		listed[g.Key] = true
	}
	for _, c := range s.Counts {
		if !listed[c.Gate] {
			keys = append(keys, c.Gate)
			listed[c.Gate] = true
		}
	}
	var total view.GateCount
	var rows [][]doc.Cell
	numbers := func(c view.GateCount) []int {
		return []int{c.Standing, c.Contributes, c.ContributesNext, c.Reviews, c.ReviewsNext}
	}
	for _, k := range keys {
		c := by[k]
		total.Standing += c.Standing
		total.Contributes += c.Contributes
		total.ContributesNext += c.ContributesNext
		total.Reviews += c.Reviews
		total.ReviewsNext += c.ReviewsNext
		row := []doc.Cell{cell(p.gate(k))}
		for _, n := range numbers(c) {
			row = append(row, cellText(strconv.Itoa(n)))
		}
		rows = append(rows, row)
	}
	last := []doc.Cell{cell(one(doc.Strong{txt("Total")}))}
	for _, n := range numbers(total) {
		last = append(last, cell(one(doc.Strong{txt(strconv.Itoa(n))})))
	}
	rows = append(rows, last)
	return []doc.Block{
		heading(3, txt("Counts by gate")),
		doc.Table{
			Align: []doc.Align{doc.Left, doc.Right, doc.Right, doc.Right, doc.Right, doc.Right},
			Head:  heads("Gate", "Tasks standing here", "Contributes", "of which next", "Reviews", "of which next"),
			Rows:  rows,
		},
	}
}

// authoritySection is `### Authority`: a sentence per subtree, or None.
func (p *page) authoritySection(s view.PersonSection) []doc.Block {
	b := []doc.Block{heading(3, txt("Authority"))}
	if len(s.Authority) == 0 {
		return append(b, para(txt("None.")))
	}
	for _, t := range s.Authority {
		b = append(b, para(spans(one(txt("Over ")), task(t.Task),
			one(txt(" and its "+strconv.Itoa(t.Descendants)+" descendants.")))...))
	}
	return b
}

// taskTitles collects the titles the view names, by id, so that a Stated,
// which holds an id alone, prints as a task (R4).
func taskTitles(v *view.Assignment) map[string]string {
	titles := map[string]string{}
	add := func(t view.TaskRef) {
		if t.Title != "" {
			titles[t.ID] = t.Title
		}
	}
	add(v.Project)
	for _, s := range v.Sections {
		for _, r := range s.Assigned {
			add(r.TaskRef)
		}
		for _, r := range s.Recursive {
			add(r.Task)
		}
		for _, r := range s.Contributes {
			add(r.Task)
		}
		for _, r := range s.Reviews {
			add(r.Task)
		}
		for _, r := range s.Authority {
			add(r.Task)
		}
	}
	return titles
}

// stated words where a field stands: the task that states it (R4) or the
// plain default, the key in a code span with its file, and the commit (R6),
// joined by commas.
func stated(s view.Stated, titles map[string]string) []doc.Inline {
	var parts [][]doc.Inline
	switch s.By {
	case "":
	case "default":
		parts = append(parts, one(txt("the plain default")))
	default:
		parts = append(parts, task(view.TaskRef{ID: s.By, Title: titles[s.By]}))
	}
	if s.Key != "" {
		key := one(code(s.Key))
		if s.File != "" {
			key = append(key, txt(" in "), code(s.File))
		}
		parts = append(parts, key)
	}
	if s.Commit != "" {
		parts = append(parts, one(txt("commit "), code(short(s.Commit))))
	}
	return joined(", ", parts...)
}

// tasksCell prints the count of a position's tasks, a colon and every id.
func tasksCell(tasks []string) []doc.Inline {
	return spans(one(txt(strconv.Itoa(len(tasks))+": ")), ids(tasks))
}

// personProvenance builds `## <email>, provenance`: where each position
// stands, grouped by the tasks that share a source.
func (p *page) personProvenance(s view.PersonSection, titles map[string]string) []doc.Block {
	of := func(kinds ...string) []view.Position {
		var out []view.Position
		for _, pos := range s.Positions {
			for _, k := range kinds {
				if pos.Kind == k {
					out = append(out, pos)
				}
			}
		}
		return out
	}
	kindWord := map[string]string{"assignee": "Assignee", "authority": "Authority"}

	b := []doc.Block{heading(2, txt(s.Email+", provenance"))}
	var rows [][]doc.Cell
	for _, pos := range of("assignee", "authority") {
		rows = append(rows, []doc.Cell{cellText(kindWord[pos.Kind]), cell(tasksCell(pos.Tasks)), cell(stated(pos.From, titles))})
	}
	if len(rows) > 0 {
		b = append(b, doc.Table{Head: heads("Position", "Tasks", "Stated in"), Rows: rows})
	}
	rows = nil
	for _, pos := range of("contributes") {
		rows = append(rows, []doc.Cell{
			cell(p.gate(pos.Gate)), cell(tasksCell(pos.Tasks)), cell(stated(pos.From, titles)),
			cell(p.person(pos.Reviewer)), cell(stated(pos.RevFrom, titles)),
		})
	}
	if len(rows) > 0 {
		b = append(b, doc.Table{
			Head: heads("Contributes at", "Tasks", "Contributor and model from", "Reviewer", "Reviewer from"),
			Rows: rows,
		})
	}
	rows = nil
	for _, pos := range of("reviews") {
		rows = append(rows, []doc.Cell{
			cell(p.gate(pos.Gate)), cell(tasksCell(pos.Tasks)),
			cell(stated(pos.RevFrom, titles)), cell(stated(pos.From, titles)),
		})
	}
	if len(rows) > 0 {
		b = append(b, doc.Table{
			Head: heads("Reviews at", "Tasks", "Reviewer from", "Contributor and model from"),
			Rows: rows,
		})
	}
	rows = nil
	for _, pos := range of("recursive") {
		var pin []doc.Inline
		if pos.Pin != "" {
			pin = one(code(short(pos.Pin)))
		}
		rows = append(rows, []doc.Cell{cell(ids(pos.Tasks)), cell(stated(pos.From, titles)), cell(pin)})
	}
	if len(rows) > 0 {
		b = append(b, doc.Table{Head: heads("Task", "Stated in", "Pinned at"), Rows: rows})
	}
	return b
}

// commitsSection is `## Commits: n`, the commits the positions name, and the
// commands that reproduce them.
func (p *page) commitsSection(v *view.Assignment) []doc.Block {
	b := []doc.Block{countHeading(2, "Commits", len(v.Commits))}
	if len(v.Commits) > 0 {
		var rows [][]doc.Cell
		for _, c := range v.Commits {
			rows = append(rows, []doc.Cell{
				cell(commitSubject(c)), cellText(c.Date), cell(p.actor(c.Author)), cell(p.actor(c.Committer)),
			})
		}
		b = append(b, doc.Table{Head: heads("Commit", "Date", "Author", "Committer"), Rows: rows})
	}
	if pre := commandLines(v.Commands); pre != nil {
		b = append(b, *pre)
	}
	return b
}
