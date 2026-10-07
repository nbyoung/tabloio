package render

import (
	"strconv"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { registerStatus(buildQueue) }

// queueSeeAlso lists the views related to the work queue, in the order of
// the See also line.
var queueSeeAlso = []string{"task", "blockage", "assignment", "context", "tableau", "authority"}

// queueKinds lists the five kinds of item in VIEWS.md's order, with the
// words of each: the Kind cell, the section heading, and the count line.
var queueKinds = []struct{ key, label, section, one, many string }{
	{"review", "Review owed", "Reviews owed", "review owed", "reviews owed"},
	{"authorisation", "Authorisation owed", "Authorisations owed", "authorisation owed", "authorisations owed"},
	{"ready", "Work ready", "Work ready", "work ready", "work ready"},
	{"reaffirmation", "Reaffirmation", "Reaffirmations", "reaffirmation", "reaffirmations"},
	{"waiting", "Work waiting", "Work waiting", "work waiting", "work waiting"},
}

// buildBrief is the hook of A6, a stub. The Agent briefs task (e4c7) replaces
// it with internal/render/brief.go, which decodes Queue.Brief and writes the
// brief; until then a queue with a brief at provenance is a format this
// package has not built.
func buildBrief(*view.Queue, Options) (doc.Doc, error) {
	return doc.Doc{}, ErrFormat
}

// buildQueue builds the contributor work queue: the count line and one line
// per item at glance; each item expanded in five sections at detail; and at
// provenance the brief where the data holds one, else the detail, since the
// queue's provenance is the brief and nothing else (decision 4). It returns
// the error of the brief hook.
//
// R1 asks a deeper level to keep the table of the level above, so detail
// draws the table of glance before the sections.
func buildQueue(v *view.Queue, o Options) (doc.Doc, error) {
	if o.Level >= view.Provenance && len(v.Brief) > 0 && string(v.Brief) != "null" {
		return buildBrief(v, o)
	}
	p := newPage(&v.Head, o)

	b := []doc.Block{p.queueCounts(v)}
	if len(v.Items) == 0 {
		b = append(b, emptyForm("The queue is empty."))
	} else {
		b = append(b, p.queueTable(v))
	}
	if o.Level >= view.Detail {
		for i, k := range queueKinds {
			b = append(b, p.queueSection(v, i+1, k.key, k.section)...)
		}
	}
	return p.frame(frame{
		Name:     "queue",
		Title:    one(txt("Contributor work queue")),
		Question: "What do I do next?",
		Legend:   true,
		SeeAlso:  queueSeeAlso,
		Body:     b,
	}), nil
}

// queueCounts is the count line: the items, then each kind, zeros included.
func (p *page) queueCounts(v *view.Queue) doc.Para {
	n := map[string]int{}
	for _, it := range v.Items {
		n[it.Kind]++
	}
	var kinds []string
	for _, k := range queueKinds {
		kinds = append(kinds, count(n[k.key], k.one, k.many))
	}
	return para(txt(count(len(v.Items), "item", "items") + ": " + joinWords(kinds, ", ") + "."))
}

func queueLabel(kind string) string {
	for _, k := range queueKinds {
		if k.key == kind {
			return k.label
		}
	}
	return kind
}

// queueSince prints the Since cell: the date, and for a reaffirmation the
// age after it.
func queueSince(it view.QueueItem) string {
	if it.Kind != "reaffirmation" {
		return it.Since
	}
	age := count(it.Age, "day", "days")
	if it.Since == "" {
		return age
	}
	return it.Since + ", " + age
}

// queueAlike reports whether two items agree in everything but the task, so
// that a run of them folds (R10). An item with a cause never folds: the
// cause belongs to its task.
func queueAlike(a, b view.QueueItem) bool {
	return a.Cause == "" && b.Cause == "" && a.Kind == b.Kind && a.Gate == b.Gate &&
		a.Model == b.Model && a.Since == b.Since && a.Age == b.Age
}

// queueRuns splits the items into runs of consecutive alike items, as
// half-open ranges.
func queueRuns(items []view.QueueItem) [][2]int {
	var runs [][2]int
	for i := 0; i < len(items); {
		j := i + 1
		for j < len(items) && queueAlike(items[i], items[j]) {
			j++
		}
		runs = append(runs, [2]int{i, j})
		i = j
	}
	return runs
}

// queueTable is the glance table: Kind, Task, Gate, Model, Since. The run
// fold (R10) works on each run of one kind; foldRows puts the task cell
// first, so the Kind cell goes in front of its rows here.
func (p *page) queueTable(v *view.Queue) doc.Table {
	var rows [][]doc.Cell
	for _, run := range queueRuns(v.Items) {
		var tr []taskRow
		for _, it := range v.Items[run[0]:run[1]] {
			t := p.fileLink(it.Task, task(it.Task))
			if it.Cause != "" {
				t = spans(t, one(txt("; ")), worded(it.Cause))
			}
			var model []doc.Inline
			if it.Model != "" {
				model = one(code(it.Model))
			}
			tr = append(tr, taskRow{ID: it.Task.ID, Task: t, Rest: []doc.Cell{
				cell(p.gate(it.Gate)), cell(model), cellText(queueSince(it)),
			}})
		}
		kind := cellText(queueLabel(v.Items[run[0]].Kind))
		for _, r := range p.foldRows(tr) {
			rows = append(rows, append([]doc.Cell{kind}, r...))
		}
	}
	return doc.Table{Head: heads("Kind", "Task", "Gate", "Model", "Since"), Rows: rows}
}

// queueSection is one of the five sections of detail: its heading with the
// count, then each item as a bold paragraph and a list, a long run folded.
func (p *page) queueSection(v *view.Queue, n int, kind, name string) []doc.Block {
	var items []view.QueueItem
	for _, it := range v.Items {
		if it.Kind == kind {
			items = append(items, it)
		}
	}
	b := []doc.Block{countHeading(2, strconv.Itoa(n)+". "+name, len(items))}
	for _, run := range queueRuns(items) {
		rest := items[run[0]:run[1]]
		shown := rest
		if len(rest) > foldAbove && !p.o.Unfold {
			shown = rest[:foldShown]
		}
		for _, it := range shown {
			b = append(b, p.queueItem(it)...)
		}
		if len(shown) < len(rest) {
			var more []string
			for _, it := range rest[foldShown:] {
				more = append(more, it.Task.ID)
			}
			head := spans(one(doc.Strong{txt("… and " + strconv.Itoa(len(more)) + " more")}), p.queueTail(rest[0]))
			b = append(b, para(spans(head, one(txt(": ")), ids(more), one(txt(".")))...))
		}
	}
	return b
}

// queueTail prints what follows an item's bold task: ` at <gate>, model
// `<model>“, or for a reaffirmation `, <gate>, <date>, <age> days`.
func (p *page) queueTail(it view.QueueItem) []doc.Inline {
	if it.Kind == "reaffirmation" {
		out := spans(one(txt(", ")), p.gate(it.Gate))
		if s := queueSince(it); s != "" {
			out = append(out, txt(", "+s))
		}
		return out
	}
	out := spans(one(txt(" at ")), p.gate(it.Gate))
	if it.Model != "" {
		out = append(out, txt(", model "), code(it.Model))
	}
	return out
}

// queueLine makes a list item `<label> <text>`, with nested blocks.
func queueLine(label string, text []doc.Inline, blocks ...doc.Block) doc.Item {
	return doc.Item{Text: spans(one(txt(label+" ")), text), Blocks: blocks}
}

// queueItem is one item at detail: a bold paragraph and a list of the lines
// the data holds.
func (p *page) queueItem(it view.QueueItem) []doc.Block {
	head := spans(one(doc.Strong(p.fileLink(it.Task, task(it.Task)))), p.queueTail(it))
	var l doc.List
	add := func(x doc.Item) { l.Items = append(l.Items, x) }

	if g, ok := p.legendGate(it.Gate); ok {
		text := spans(p.gate(it.Gate), one(txt(", "+g.Name)))
		if g.Criteria != "" {
			text = append(text, txt(": "+g.Criteria))
		}
		add(queueLine("Gate:", text))
	}
	if j := it.Junction; j != nil {
		p.queueJunction(it, j, add)
	}
	var refs [][]doc.Inline
	if it.Junction != nil {
		for _, r := range it.Junction.References {
			refs = append(refs, p.reference(r))
		}
	}
	for _, r := range it.TaskReferences {
		refs = append(refs, p.reference(r))
	}
	if len(refs) > 0 {
		add(queueLine("References:", joined(", ", refs...)))
	}
	if it.Kind == "authorisation" && it.Cause != "" {
		add(queueLine("Cause:", worded(it.Cause)))
	}
	p.queueEntries("Requires:", it.Requires, add)
	if it.Kind == "waiting" && it.Cause != "" {
		add(queueLine("Waits for:", worded(it.Cause)))
	}
	p.queueEntries("Unblocks:", it.Unblocks, add)
	p.queueEntries("Also required by:", it.AlsoRequiredBy, add)
	p.queueParents(it.ParentUnblocks, add)
	if st := p.queueStatus(it); len(st) > 0 {
		add(queueLine("Status:", st))
	}
	switch it.Kind {
	case "ready":
		add(queueLine("Brief:", one(code(p.queueBriefCommand(it)))))
	case "review":
		add(queueLine("To accept:", one(code("tabloio review "+it.Task.ID+" "+it.Gate))))
	case "authorisation":
		add(queueLine("To authorise:", one(code("tabloio authorise "+it.Task.ID))))
	case "reaffirmation":
		add(queueLine("To reaffirm:", one(code("tabloio reaffirm "+it.Task.ID))))
	case "waiting":
		add(queueLine("Do not start:", one(txt("the cause stands."))))
	}
	b := []doc.Block{para(head...)}
	if len(l.Items) > 0 {
		b = append(b, l)
	}
	return b
}

// legendGate finds a gate of the legend by its key.
func (p *page) legendGate(key string) (view.Gate, bool) {
	for _, g := range p.h.Legend.Gates {
		if g.Key == key {
			return g, true
		}
	}
	return view.Gate{}, false
}

// queueSource finds the task that supplies one field of the item's junction.
func queueSource(it view.QueueItem, field string) string {
	for _, s := range it.Sources {
		if s.Gate == it.Gate && s.Field == field {
			return s.By
		}
	}
	return ""
}

// queueJunction adds the Contributor and Reviewer lines.
func (p *page) queueJunction(it view.QueueItem, j *view.Junction, add func(doc.Item)) {
	if j.Contributor != "" || j.Model != "" {
		var out []doc.Inline
		for _, m := range j.Marks {
			if m == "person" || m == "agent" {
				out = append(out, symbol(p.marks[m].Symbol, m), txt(" "))
				break
			}
		}
		out = append(out, p.person(j.Contributor)...)
		if j.Model != "" {
			if j.Contributor != "" {
				out = append(out, txt(", model "))
			} else {
				out = append(out, txt("model "))
			}
			out = append(out, code(j.Model))
		}
		out = append(out, markFrom(queueSource(it, "contributor"))...)
		add(queueLine("Contributor:", out))
	}
	if j.Reviewer != "" {
		var out []doc.Inline
		if m, ok := p.marks["reviewer"]; ok {
			out = append(out, symbol(m.Symbol, m.Key), txt(" "))
		}
		out = append(out, p.person(j.Reviewer)...)
		if j.ByDefault {
			out = append(out, txt(", the assignee"))
		} else {
			out = append(out, markFrom(queueSource(it, "reviewer"))...)
		}
		add(queueLine("Reviewer:", out))
	}
}

// queueEntry prints a requirement as `<task>, <from> → <to>, <text>:
// <condition>`.
func (p *page) queueEntry(r view.Requirement) []doc.Inline {
	out := task(r.Task)
	if r.Subproject != "" {
		out = append(out, txt(" in "), code(r.Subproject))
	}
	return append(out, txt(", "+edge(r.From, r.To)+", "+r.Text+": "+condition(r.Met, r.Due)))
}

// queueEntries adds a line of requirements: on the line itself when there
// is one, nested when there are several. A run of more than eight entries
// that agree in everything but the task folds (R10).
func (p *page) queueEntries(label string, rs []view.Requirement, add func(doc.Item)) {
	switch len(rs) {
	case 0:
		return
	case 1:
		add(queueLine(label, p.queueEntry(rs[0])))
		return
	}
	var inner doc.List
	for i := 0; i < len(rs); {
		j := i + 1
		for j < len(rs) && queueAlikeEntry(rs[i], rs[j]) {
			j++
		}
		if j-i > foldAbove && !p.o.Unfold {
			for _, r := range rs[i : i+foldShown] {
				inner.Items = append(inner.Items, doc.Item{Text: p.queueEntry(r)})
			}
			var more []string
			for _, r := range rs[i+foldShown : j] {
				more = append(more, r.Task.ID)
			}
			text := spans(one(txt("… and "+strconv.Itoa(len(more))+" more: ")), ids(more),
				one(txt(", "+edge(rs[i].From, rs[i].To)+", "+rs[i].Text+": "+condition(rs[i].Met, rs[i].Due))))
			inner.Items = append(inner.Items, doc.Item{Text: text})
		} else {
			for _, r := range rs[i:j] {
				inner.Items = append(inner.Items, doc.Item{Text: p.queueEntry(r)})
			}
		}
		i = j
	}
	add(doc.Item{Text: one(txt(label)), Blocks: []doc.Block{inner}})
}

func queueAlikeEntry(a, b view.Requirement) bool {
	return a.From == b.From && a.To == b.To && a.Text == b.Text && a.Met == b.Met &&
		a.Due == b.Due && a.Subproject == b.Subproject
}

// queueParents adds the line of what the item's parent unblocks, each entry
// as “<parent> at <gate> is required by <entry>“.
func (p *page) queueParents(es []view.ParentEdge, add func(doc.Item)) {
	entry := func(e view.ParentEdge) []doc.Inline {
		return spans(task(e.Parent), one(txt(" at ")), p.gate(e.Gate), one(txt(" is required by ")), p.queueEntry(e.Requires))
	}
	switch len(es) {
	case 0:
	case 1:
		add(queueLine("Its parent unblocks:", entry(es[0])))
	default:
		var inner doc.List
		for _, e := range es {
			inner.Items = append(inner.Items, doc.Item{Text: entry(e)})
		}
		add(doc.Item{Text: one(txt("Its parent unblocks:")), Blocks: []doc.Block{inner}})
	}
}

// queueStatus prints the status line: the phrase, the date, the recorder,
// the commit and, after a colon, the note.
func (p *page) queueStatus(it view.QueueItem) []doc.Inline {
	if it.Status == nil {
		return nil
	}
	parts := [][]doc.Inline{p.phrase(*it.Status)}
	if it.Status.Date != "" {
		parts = append(parts, one(txt(it.Status.Date)))
	}
	if r := p.person(it.Status.Recorder); len(r) > 0 {
		parts = append(parts, r)
	}
	if it.StatusCommit != nil {
		if c := commit(it.StatusCommit.Hash); len(c) > 0 {
			parts = append(parts, c)
		}
	}
	out := joined(", ", parts...)
	if it.Status.Note != "" {
		out = append(out, txt(": "+it.Status.Note))
	}
	return out
}

// queueBriefCommand is the command that writes the brief of an item (A7):
// `tabloio queue --person <email> --ref <ref> --brief <task> <gate>`.
func (p *page) queueBriefCommand(it view.QueueItem) string {
	s := "tabloio queue"
	if par := p.h.Params; par.Person != "" {
		s += " " + flagPerson + " " + par.Person
	}
	if ref := refName(p.h.Ref); ref != "HEAD" {
		s += " " + flagRef + " " + ref
	}
	return s + " --brief " + it.Task.ID + " " + it.Gate
}
