package render

import (
	"strconv"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { register(buildTask) }

// taskSeeAlso lists the views related to the task definition, in the order
// of the See also line.
var taskSeeAlso = []string{"history", "blockage", "authority", "assignment", "queue", "context", "tableau", "audit"}

// buildTask builds the task definition view: the field table at glance; the
// description, the place in the tree, the requirements, the junctions and
// the subproject snapshot at detail; and the Git facts behind them at
// provenance.
func buildTask(v *view.Task, o Options) doc.Doc {
	p := newPage(&v.Head, o)
	detail := o.Level >= view.Detail
	prov := o.Level >= view.Provenance

	b := []doc.Block{p.fieldTable(v)}
	if prov && v.StatusCommit != nil {
		b = append(b, heading(2, txt("Status commit")), p.commitTable(*v.StatusCommit, nil))
	}
	if detail {
		b = append(b, p.description(v)...)
		b = append(b, heading(2, txt("Place in the tree")), p.placeTable(v))
	}
	if prov && v.Authorisation != nil && v.Authorisation.Commit != nil {
		b = append(b, heading(2, txt("Authorisation commit")), p.commitTable(*v.Authorisation.Commit, v.Authorisation))
	}
	if detail {
		b = append(b, p.requires(v)...)
		b = append(b, p.dependents(v)...)
		b = append(b, p.junctions(v)...)
		if v.Snapshot != nil {
			b = append(b, heading(2, txt("Subproject snapshot")), p.snapshotTable(v.Snapshot))
		}
	}
	if prov {
		b = append(b, p.provenance(v)...)
	}

	return p.frame(frame{
		Name:       "task",
		Positional: v.Task.ID,
		Title:      spans(one(txt("Task definition: ")), p.fileLink(v.Task, task(v.Task))),
		Question:   "What is this task and where does it stand?",
		Legend:     true,
		SeeAlso:    taskSeeAlso,
		Body:       b,
	})
}

// fieldTable makes a table of a name and a value per row.
func fieldTable(name string, rows [][]doc.Cell) doc.Table {
	return doc.Table{Head: heads(name, "Value"), Rows: rows}
}

func field(name string, value []doc.Inline) []doc.Cell {
	return []doc.Cell{cellText(name), cell(value)}
}

func (p *page) fieldTable(v *view.Task) doc.Table {
	parent := one(txt("None"))
	if v.Parent != nil {
		parent = task(v.Parent.TaskRef)
		if v.Parent.Order != nil {
			parent = append(parent, txt(", order "+strconv.Itoa(*v.Parent.Order)))
		}
	}
	return fieldTable("Field", [][]doc.Cell{
		field("Task", task(v.Task)),
		field("Assignee", p.person(v.Assignee)),
		field("Parent", parent),
		field("Status", p.status(v.Status)),
		field("Note", one(txt(v.Status.Note))),
		field("Recorded", p.recorded(v.Status)),
	})
}

// recorded prints when a status was recorded: `<date> by <recorder>`; for a
// roll-up `<date>, the oldest among its children`; for a snapshot `<date>,
// from the subproject`.
func (p *page) recorded(st view.Status) []doc.Inline {
	if st.Date == "" {
		return nil
	}
	if st.From != nil {
		switch st.From.Kind {
		case "rollup":
			return one(txt(st.Date + ", the oldest among its children"))
		case "snapshot":
			return one(txt(st.Date + ", from the subproject"))
		}
	}
	if st.Recorder == "" {
		return one(txt(st.Date))
	}
	return spans(one(txt(st.Date+" by ")), p.person(st.Recorder))
}

func (p *page) description(v *view.Task) []doc.Block {
	if v.Description == "" && len(v.References) == 0 {
		return nil
	}
	b := []doc.Block{heading(2, txt("Description"))}
	if v.Description != "" {
		b = append(b, para(txt(v.Description)))
	}
	if len(v.References) > 0 {
		list := doc.List{}
		for _, r := range v.References {
			list.Items = append(list.Items, doc.Item{Text: p.reference(r)})
		}
		b = append(b, para(txt("References:")), list)
	}
	return b
}

func (p *page) placeTable(v *view.Task) doc.Table {
	var path [][]doc.Inline
	for _, t := range v.Path {
		path = append(path, task(t))
	}
	order := []doc.Inline(nil)
	if v.Parent != nil {
		pos := "no order, among " + strconv.Itoa(v.Siblings) + " under "
		if v.Parent.Order != nil {
			pos = strconv.Itoa(*v.Parent.Order) + " of " + strconv.Itoa(v.Siblings) + " under "
		}
		order = one(txt(pos), code(v.Parent.ID))
	}
	authorised := "No, proposed"
	if v.Authorised {
		authorised = "Yes"
	}
	return fieldTable("Field", [][]doc.Cell{
		field("Path", joined(" › ", path...)),
		field("Order", order),
		field("Children", p.children(v.Children)),
		field("Authorities", p.authorities(v.Authorities)),
		field("Authorised", one(txt(authorised))),
	})
}

// children prints `None`, or the children's ids with their shared status
// phrase when all agree, else one `id status` pair per child.
func (p *page) children(rows []view.ChildRow) []doc.Inline {
	if len(rows) == 0 {
		return one(txt("None"))
	}
	agree := len(rows) > 1
	for _, r := range rows[1:] {
		a, b := r.Status, rows[0].Status
		if a.Gate != b.Gate || a.State != b.State || a.Reason != b.Reason {
			agree = false
		}
	}
	if agree {
		var list []string
		for _, r := range rows {
			list = append(list, r.ID)
		}
		return spans(ids(list), one(txt(", each at ")), p.phrase(rows[0].Status))
	}
	var pairs [][]doc.Inline
	for _, r := range rows {
		pairs = append(pairs, spans(one(code(r.ID), txt(" ")), p.phrase(r.Status)))
	}
	return joined("; ", pairs...)
}

// authorities groups consecutive ancestors with one email:
// `<email> (`2034`, `bc63`)`, the groups joined by `; then `.
func (p *page) authorities(list []view.Authority) []doc.Inline {
	var groups [][]doc.Inline
	for i := 0; i < len(list); {
		j := i
		var by []doc.Inline
		for ; j < len(list) && list[j].Email == list[i].Email; j++ {
			if j > i {
				by = append(by, txt(", "))
			}
			by = append(by, code(list[j].By))
		}
		g := spans(p.person(list[i].Email), one(txt(" (")), by, one(txt(")")))
		groups = append(groups, g)
		i = j
	}
	return joined("; then ", groups...)
}

func (p *page) requires(v *view.Task) []doc.Block {
	b := []doc.Block{countHeading(2, "Requires", len(v.Requires))}
	if len(v.Requires) == 0 {
		return b
	}
	var rows []taskRow
	for _, r := range v.Requires {
		var status []doc.Inline
		if r.Status != nil {
			status = p.status(*r.Status)
		}
		rows = append(rows, taskRow{ID: r.Task.ID, Task: p.requirementTask(r), Rest: []doc.Cell{
			cellText(edge(r.From, r.To)), cellText(r.Text), cell(status), cellText(condition(r.Met, r.Due)),
		}})
	}
	return append(b, doc.Table{Head: heads("Task", "Edge", "What passes", "Its status", "Condition"), Rows: p.foldRows(rows)})
}

func (p *page) dependents(v *view.Task) []doc.Block {
	b := []doc.Block{countHeading(2, "Dependents", len(v.Dependents))}
	if len(v.Dependents) == 0 {
		return b
	}
	var rows []taskRow
	for _, r := range v.Dependents {
		rows = append(rows, taskRow{ID: r.Task.ID, Task: p.requirementTask(r), Rest: []doc.Cell{
			cellText(edge(r.From, r.To)), cellText(r.Text), cellText(condition(r.Met, r.Due)),
		}})
	}
	return append(b, doc.Table{Head: heads("Task", "Edge", "What passes", "Condition"), Rows: p.foldRows(rows)})
}

// requirementTask prints the task cell of a requirement: the task, linked to
// its file, and for a cross-project entry the subproject.
func (p *page) requirementTask(r view.Requirement) []doc.Inline {
	out := p.fileLink(r.Task, task(r.Task))
	if r.Subproject != "" {
		out = append(out, txt(" in "), code(r.Subproject))
	}
	return out
}

func (p *page) junctions(v *view.Task) []doc.Block {
	recursive := false
	for _, j := range v.Junctions {
		recursive = recursive || j.Subproject != nil
	}
	head := heads("Gate", "Marks", "Contributor", "Model", "Reviewer")
	align := []doc.Align{doc.Left, doc.Centre, doc.Left, doc.Left, doc.Left}
	if recursive {
		head = append(head, cellText("Subproject"))
		align = append(align, doc.Left)
	}
	head = append(head, cellText("Stands"))
	align = append(align, doc.Left)

	shown := map[string]bool{}
	var rows [][]doc.Cell
	for _, j := range v.Junctions {
		for _, m := range j.Marks {
			shown[m] = true
		}
		reviewer := p.person(j.Reviewer)
		if j.ByDefault {
			reviewer = spans(one(txt("the assignee")), prefixed(", ", reviewer))
		}
		var model []doc.Inline
		if j.Model != "" {
			model = one(code(j.Model))
		}
		row := []doc.Cell{
			cell(p.gate(j.Gate)),
			cell(p.markSymbols(j.Marks)),
			cell(p.person(j.Contributor)),
			cell(model),
			cell(reviewer),
		}
		if recursive {
			var sub []doc.Inline
			if j.Subproject != nil {
				sub = spans(one(txt(j.Subproject.URL+" ")), task(j.Subproject.Task))
			}
			row = append(row, cell(sub))
		}
		row = append(row, cellText(stands(j)))
		rows = append(rows, row)
	}
	b := []doc.Block{heading(2, txt("Junctions"))}
	if len(rows) == 0 {
		return []doc.Block{countHeading(2, "Junctions", 0)}
	}
	b = append(b, doc.Table{Align: align, Head: head, Rows: rows})
	if k := p.markKey(shown); k != nil {
		b = append(b, *k)
	}
	return b
}

// prefixed puts sep before a run, or returns nothing for an empty run.
func prefixed(sep string, xs []doc.Inline) []doc.Inline {
	if len(xs) == 0 {
		return nil
	}
	return append(one(txt(sep)), xs...)
}

// stands words where the status stands at a junction: `passed`, `passed;
// the status stands here`, `next`, `later` or `does not apply`, with
// `, reviewed` after `passed` when the junction is reviewed.
func stands(j view.Junction) string {
	s := j.Stands
	switch j.Stands {
	case "passed", "here":
		s = "passed"
		if j.Reviewed {
			s += ", reviewed"
		}
		if j.Stands == "here" {
			s += "; the status stands here"
		}
	case "exempt":
		s = "does not apply"
	}
	return s
}

func (p *page) snapshotTable(s *view.Snapshot) doc.Table {
	var status, note, recorded []doc.Inline
	if st := s.Status; st != nil {
		status, note, recorded = p.status(*st), textRun(st.Note), p.recorded(*st)
	}
	return fieldTable("Field", [][]doc.Cell{
		field("Subproject", one(code(s.URL))),
		field("Task", task(s.Task)),
		field("Assignee", p.person(s.Assignee)),
		field("Status", status),
		field("Note", note),
		field("Recorded", recorded),
		field("Children", p.children(s.Children)),
	})
}

// commitTable is a field table of a commit; a with the authorisation adds
// how the commit authorises and whose authority it is.
func (p *page) commitTable(c view.Commit, a *view.Authorisation) doc.Table {
	var trailers [][]doc.Inline
	for _, t := range c.Trailers {
		trailers = append(trailers, one(code(t)))
	}
	rows := [][]doc.Cell{
		field("Commit", commitSubject(c)),
		field("Date", one(txt(c.Date))),
		field("Author", p.actor(c.Author)),
		field("Committer", p.actor(c.Committer)),
	}
	if a != nil {
		rows = append(rows,
			field("Accepts by", acceptsBy(a.Way)),
			field("The authority is", one(txt(authority(a.By)))))
	} else {
		rows = append(rows, field("Trailers", joined(", ", trailers...)))
	}
	return fieldTable("Field", rows)
}

func acceptsBy(way string) []doc.Inline {
	switch way {
	case "change":
		return one(txt("a change on the trunk"))
	case "merge":
		return one(txt("a merge"))
	case "trailer":
		return one(txt("an "), code("Authorised:"), txt(" trailer"))
	}
	return one(txt(way))
}

func authority(by string) string {
	switch by {
	case "author":
		return "the author"
	case "committer":
		return "the committer"
	case "both":
		return "the author and the committer"
	}
	return by
}

// provenance builds the blocks that only provenance adds, after the
// junctions: the roll-up, the linkages, where each junction field comes
// from, the reviews, the models, the newest events and the commands.
func (p *page) provenance(v *view.Task) []doc.Block {
	var b []doc.Block
	if len(v.Considered) > 0 {
		b = append(b, heading(2, txt("Roll-up")), doc.Table{Head: heads("Name", "Value"), Rows: p.rollupRows(v)})
	}
	if len(v.Linkages) > 0 {
		var rows [][]doc.Cell
		for _, l := range v.Linkages {
			for i, f := range p.linkageFacts(l) {
				var gate []doc.Inline
				if i == 0 {
					gate = p.linkageGates(l.Gates)
				}
				rows = append(rows, []doc.Cell{cell(gate), cell(f[0]), cell(f[1])})
			}
		}
		b = append(b, heading(2, txt("Linkages")), doc.Table{Head: heads("Gate", "Name", "Value"), Rows: rows})
	}
	if len(v.Sources) > 0 {
		b = append(b, heading(2, txt("Where each junction field comes from")), p.sourcesTable(v))
	}
	b = append(b, p.reviews(v)...)
	b = append(b, p.models(v)...)
	b = append(b, p.events(v)...)
	if pre := commandLines(v.Commands); pre != nil {
		b = append(b, *pre)
	}
	return b
}

// rollupRows are the rows of a parent's roll-up: the children it
// considers, then each status of the chain, the parent's own first, to the
// leaf whose file states it.
func (p *page) rollupRows(v *view.Task) [][]doc.Cell {
	rows := [][]doc.Cell{field("Children considered", ids(v.Considered))}
	for _, st := range v.Chain {
		name := task(st.Task)
		if st.URL != "" {
			name = append(name, txt(" in "), code(st.URL))
		}
		rows = append(rows, []doc.Cell{cell(name), cell(spans(p.status(st.Status), prefixed("; ", p.recorded(st.Status))))})
	}
	return rows
}

// linkageGates prints the gates whose junctions read a linkage: the first,
// and the count of the others.
func (p *page) linkageGates(gates []string) []doc.Inline {
	if len(gates) == 0 {
		return nil
	}
	out := p.gate(gates[0])
	if n := len(gates) - 1; n > 0 {
		out = append(out, txt(" and "+count(n, "more", "more")))
	}
	return out
}

// The words of a linkage's keys (A9 of design e3ed, which tablo now gives
// as typed facts): one phrase per form, reason, trunk reading and pin
// reading.
var (
	linkForms = map[string]string{
		"submodule": "A path that is a Git submodule, so the submodule's pin fixes the commit",
		"directory": "A directory of the same repository, read at the commit in view",
		"url":       "A repository URL, so the entry's `commit` field fixes the commit",
		"none":      "No link",
	}
	linkWhy = map[string]string{
		"no_link": "Not read: no link reaches the subproject",
		"no_task": "Not read: the subproject has no such task",
	}
	linkOnTrunk = map[string]string{
		"yes":     "the pin is on the trunk",
		"no":      "the pin is off the trunk",
		"unknown": "whether the pin is on the trunk is unknown",
	}
	linkBehind = map[string]string{
		"yes": "behind its tip",
		"no":  "at its tip",
	}
)

// linkageFacts words a linkage as Name and Value pairs, in a fixed order,
// each where the linkage holds it (R9).
func (p *page) linkageFacts(l view.Linkage) [][2][]doc.Inline {
	var out [][2][]doc.Inline
	add := func(name []doc.Inline, value []doc.Inline) {
		if len(value) > 0 {
			out = append(out, [2][]doc.Inline{name, value})
		}
	}
	url := one(code(l.URL))
	if l.File != "" {
		url = append(url, txt(", stated in "), code(l.File))
		if l.Line > 0 {
			url = append(url, txt(", line "+strconv.Itoa(l.Line)))
		}
	}
	add(one(code("url")), url)
	add(one(txt("Form of the "), code("url")), worded(linkForms[l.Form]))
	if l.Commit != "" {
		add(one(txt("Pin")), one(code(short(l.Commit)), txt(", in full "), code(l.Commit)))
	}
	if l.Form == "submodule" || l.Form == "url" {
		field := one(txt("Not stated"))
		if l.StatedCommit != "" {
			field = one(code(l.StatedCommit))
		} else if l.Form == "submodule" {
			field = one(txt("Not stated; nothing restates the pin in the task file"))
		}
		add(one(code("commit"), txt(" field")), field)
	}
	if l.Moved {
		add(one(txt("Checkout")), one(txt("Stands elsewhere than the recorded pin "), code(short(l.Recorded))))
	}
	if l.Pinned != nil {
		add(one(txt("The pinned commit")), p.commitLine(*l.Pinned))
	}
	id := one(code(l.ID))
	if l.ID == "" && l.Task.ID != "" {
		id = one(txt("Not stated; the junction reads the subproject's root, "), code(l.Task.ID))
	}
	add(one(code("id")), id)
	if l.Why != "" {
		add(one(txt("Task read")), one(txt(linkWhy[l.Why])))
	}
	if l.Trunk != "" {
		trunk := one(code(l.Trunk))
		if l.Tip != "" {
			trunk = append(trunk, txt(", tip "), code(short(l.Tip)))
		}
		if w := linkOnTrunk[l.OnTrunk]; w != "" {
			trunk = append(trunk, txt("; "+w))
			if b := linkBehind[l.Behind]; b != "" {
				trunk = append(trunk, txt(", "+b))
			}
		}
		add(one(txt("The subproject's trunk")), trunk)
	}
	if l.SetBy != nil {
		add(one(txt("Pin set by")), p.commitLine(*l.SetBy))
	}
	return out
}

// commitLine words a commit on one line: the short hash, the subject, the
// date, and the author and committer, as one person when they agree.
func (p *page) commitLine(c view.Commit) []doc.Inline {
	out := spans(commitSubject(c), prefixed(", ", textRun(c.Date)))
	if c.Author == c.Committer {
		return spans(out, prefixed(", author and committer ", p.actor(c.Author)))
	}
	return spans(out, prefixed(", author ", p.actor(c.Author)), prefixed(", committer ", p.actor(c.Committer)))
}

// sourcesTable lists, for each junction field, the task that supplies it;
// the gate stands on the first row of its run.
func (p *page) sourcesTable(v *view.Task) doc.Table {
	titles := map[string]view.TaskRef{}
	for _, t := range v.Path {
		titles[t.ID] = t
	}
	var rows [][]doc.Cell
	last := ""
	for i, s := range v.Sources {
		var gate []doc.Inline
		if i == 0 || s.Gate != last {
			gate = p.gate(s.Gate)
		}
		last = s.Gate
		var value []doc.Inline
		switch s.Field {
		case "model":
			value = one(code(s.Value))
		case "contributor", "reviewer":
			value = p.person(s.Value)
		default:
			value = one(txt(s.Value))
		}
		var by []doc.Inline
		if w, ok := statedBy[s.By]; ok {
			by = one(txt(w))
		} else if s.By != "" {
			t, ok := titles[s.By]
			if !ok {
				t = view.TaskRef{ID: s.By}
			}
			by = task(t)
		}
		rows = append(rows, []doc.Cell{cell(gate), cellText(s.Field), cell(value), cell(by)})
	}
	return doc.Table{Head: heads("Gate", "Field", "Value", "Supplied by"), Rows: rows}
}

func (p *page) reviews(v *view.Task) []doc.Block {
	b := []doc.Block{countHeading(2, "Reviews", len(v.Reviews))}
	if len(v.Reviews) == 0 {
		return b
	}
	var rows [][]doc.Cell
	for _, r := range v.Reviews {
		var by []doc.Inline
		switch {
		case r.Commit != nil:
			by = spans(commitSubject(*r.Commit), prefixed(", ", textRun(r.Commit.Date)))
		case v.Authorisation != nil && v.Authorisation.Commit != nil:
			c := v.Authorisation.Commit
			by = spans(one(txt("The authorisation, ")), commit(c.Hash), prefixed(", ", textRun(c.Date)))
		}
		rows = append(rows, []doc.Cell{cell(p.gate(r.Gate)), cell(p.person(r.Reviewer)), cell(by), cellText(effect(r.Effect))})
	}
	return append(b, doc.Table{Head: heads("Gate", "Reviewer", "Accepted by", "Effect"), Rows: rows})
}

func textRun(s string) []doc.Inline {
	if s == "" {
		return nil
	}
	return one(txt(s))
}

func effect(e string) string {
	switch e {
	case "accepts":
		return "The reviewer accepts"
	case "authorisation":
		return "The authorisation stands as the review"
	case "none":
		return "None"
	}
	return e
}

func (p *page) models(v *view.Task) []doc.Block {
	b := []doc.Block{countHeading(2, "Models", len(v.Models))}
	if len(v.Models) == 0 {
		return b
	}
	var rows [][]doc.Cell
	for _, m := range v.Models {
		var stated, trailer []doc.Inline
		if m.Stated != "" {
			stated = one(code(m.Stated))
		}
		if m.Trailer != "" {
			trailer = one(code(m.Trailer))
		}
		rows = append(rows, []doc.Cell{cell(commit(m.Commit)), cell(p.gate(m.Gate)), cell(stated), cell(trailer), cellText(reading(m.Reading))})
	}
	return append(b, doc.Table{
		Head: []doc.Cell{cellText("Commit"), cellText("Junction"), cellText("The junction states"),
			cell(one(code("Model:"), txt(" trailer"))), cellText("Reading")},
		Rows: rows,
	})
}

func reading(r string) string {
	switch r {
	case "agrees":
		return "Agrees"
	case "differs":
		return "Differs"
	case "missing":
		return "No trailer"
	case "exempt":
		return "Exempt"
	case "unstated":
		return "No model stated"
	}
	return r
}

func (p *page) events(v *view.Task) []doc.Block {
	if len(v.Events) == 0 {
		return []doc.Block{countHeading(2, "Newest events", 0)}
	}
	noun := " events; the newest "
	if v.EventCount == 1 {
		noun = " event; the newest "
	}
	b := []doc.Block{
		heading(2, txt("Newest events")),
		para(txt(strconv.Itoa(v.EventCount) + noun + strconv.Itoa(len(v.Events)) + ", newest first.")),
	}
	var rows [][]doc.Cell
	for _, e := range v.Events {
		rows = append(rows, []doc.Cell{
			cellText(e.Date), cell(commit(e.Commit)), cell(p.person(e.By)),
			cellText(e.Kind), cell(p.eventRecords(e)),
		})
	}
	if more := v.EventCount - len(v.Events); more > 0 {
		rows = append(rows, []doc.Cell{cellText("… and " + strconv.Itoa(more) + " earlier"), {}, {}, {}, {}})
	}
	return append(b, doc.Table{Head: heads("Date", "Commit", "By", "Event", "What it records"), Rows: rows})
}

func joinWords(words []string, sep string) string {
	s := ""
	for i, w := range words {
		if i > 0 {
			s += sep
		}
		s += w
	}
	return s
}

// eventRecords prints what an event records: a status phrase with its note
// after a colon, a gate for a review, or `<url> `<old>` → `<new>“ for a
// pin; then, for an event taken from a subproject, the subproject and its
// task.
func (p *page) eventRecords(e view.Event) []doc.Inline {
	var x []doc.Inline
	switch {
	case e.Kind == "status" && e.Recorded != nil:
		r := e.Recorded
		x = p.phrase(view.Status{Gate: r.Gate, State: r.State, Reason: r.Reason})
		if r.Note != "" {
			x = append(x, txt(": "+r.Note))
		}
	case e.Kind == "reviewed" && e.Gate != "":
		x = p.gate(e.Gate)
	case e.Kind == "pin" && e.Pin != nil:
		x = one(txt(e.Pin.URL + " "))
		if e.Pin.Old != "" {
			x = append(x, code(short(e.Pin.Old)), txt(" "))
		}
		x = append(x, txt("→ "), code(short(e.Pin.New)))
	}
	if s := e.Subproject; s != nil {
		if len(x) > 0 {
			x = append(x, txt("; "))
		}
		x = append(x, txt("in the subproject "), code(s.URL), txt(", task "), code(s.Task))
	}
	return x
}
