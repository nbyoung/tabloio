package render

import (
	"strconv"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { register(buildBlockage) }

// blockageSeeAlso lists the views related to the work-blockage tree, in the
// order of the See also line.
var blockageSeeAlso = []string{"queue", "audit", "task", "tableau", "history"}

// kindLabels words the five kinds of cause for the count line, in the
// singular and the plural.
var kindLabels = map[string][2]string{
	"requirement":   {"unmet requirement", "unmet requirements"},
	"status":        {"status off nominal", "statuses off nominal"},
	"review":        {"review outstanding", "reviews outstanding"},
	"authorisation": {"authorisation outstanding", "authorisations outstanding"},
	"snapshot":      {"snapshot not advanced", "snapshots not advanced"},
}

// buildBlockage builds the work-blockage tree: the count line and one line
// per cause at glance; the causes with their trees, then the requirements
// not yet due, at detail; and the facts and the commands behind each cause
// at provenance.
//
// R1 asks a deeper level to keep the table of the level above, so detail
// draws the table of glance before the trees.
func buildBlockage(v *view.Blockage, o Options) doc.Doc {
	p := newPage(&v.Head, o)
	detail := o.Level >= view.Detail

	b := []doc.Block{p.blockageCounts(v)}
	if len(v.Causes) == 0 {
		b = append(b, emptyForm("No cause holds any task."))
	} else {
		b = append(b, p.blockageTable(v))
		if detail {
			b = append(b, p.blockageTrees(v))
		}
	}
	if !detail && v.NotDueN > 0 {
		b = append(b, para(txt("Next: "+count(v.NotDueN, "requirement is", "requirements are")+
			" not yet due, "+strconv.Itoa(v.NotDueUnmet)+" of them unmet.")))
	}
	if detail {
		b = append(b, p.blockageNotDue(v)...)
	}
	if o.Level >= view.Provenance {
		if pre := commandLines(v.Commands); pre != nil {
			b = append(b, *pre)
		}
	}
	return p.frame(frame{
		Name:     "blockage",
		Title:    one(txt("Work-blockage tree")),
		Question: "What waits on what?",
		Legend:   true,
		SeeAlso:  blockageSeeAlso,
		Body:     b,
	})
}

// blockageCounts is the count line: the causes, then each kind, zeros included.
func (p *page) blockageCounts(v *view.Blockage) doc.Para {
	s := count(len(v.Causes), "cause", "causes")
	if len(v.Kinds) == 0 {
		return para(txt(s + "."))
	}
	var kinds []string
	for _, k := range v.Kinds {
		l, ok := kindLabels[k.Kind]
		if !ok {
			l = [2]string{k.Kind, k.Kind}
		}
		kinds = append(kinds, count(k.Count, l[0], l[1]))
	}
	return para(txt(s + ": " + joinWords(kinds, ", ") + "."))
}

// blockageCause words a cause by its kind.
func (p *page) blockageCause(c view.Cause) []doc.Inline {
	t := task(c.Task)
	switch c.Kind {
	case "requirement":
		if c.URL != "" {
			t = append(t, txt(" in "), code(c.URL))
		}
		return spans(t, one(txt(" has not passed ")), p.gate(c.Gate))
	case "status":
		out := spans(t, one(txt(" stands at ")), p.phrase(c.Status))
		if c.Status.Note != "" {
			out = append(out, txt(": "+c.Status.Note))
		}
		return out
	case "review":
		return spans(t, one(txt(" awaits review at ")), p.gate(c.Gate))
	case "authorisation":
		return spans(t, one(txt(" is proposed")))
	case "snapshot":
		return spans(t, one(txt(" waits on "), code(c.URL), txt(" at "), code(short(c.Pin))))
	}
	return t
}

// blockageWho words who acts: the mark's symbol, the email and the act.
func (p *page) blockageWho(c view.Cause) []doc.Inline {
	var out []doc.Inline
	if c.Mark != "" {
		out = append(out, symbol(p.marks[c.Mark].Symbol, c.Mark))
	}
	if who := p.person(c.Resolver); len(who) > 0 {
		if len(out) > 0 {
			out = append(out, txt(" "))
		}
		out = append(out, who...)
	}
	if c.Act != "" {
		if len(out) > 0 {
			out = append(out, txt(" "))
		}
		out = append(out, txt(c.Act))
	}
	return out
}

// blockageTable is the glance table: Cause, Who acts, Holds.
func (p *page) blockageTable(v *view.Blockage) doc.Table {
	var rows [][]doc.Cell
	for _, c := range v.Causes {
		rows = append(rows, []doc.Cell{
			cell(p.blockageCause(c)), cell(p.blockageWho(c)), cellText(strconv.Itoa(c.Holds)),
		})
	}
	return doc.Table{
		Align: []doc.Align{doc.Left, doc.Left, doc.Right},
		Head:  heads("Cause", "Who acts", "Holds"),
		Rows:  rows,
	}
}

// blockageTrees is the ordered list of causes with their trees; at
// provenance each cause also lists its facts and its commands.
func (p *page) blockageTrees(v *view.Blockage) doc.List {
	list := doc.List{Ordered: true}
	for _, c := range v.Causes {
		text := spans(one(doc.Strong(p.blockageCause(c))), one(txt(" · ")), p.blockageWho(c),
			one(txt(" · holds "+strconv.Itoa(c.Holds))))
		item := doc.Item{Text: text}
		var inner doc.List
		if action := p.blockageAction(c); len(action) > 0 {
			inner.Items = append(inner.Items, doc.Item{Text: spans(one(txt("Action: ")), action)})
		}
		for _, h := range c.Held {
			inner.Items = append(inner.Items, p.blockageHeld(h))
		}
		if len(inner.Items) > 0 {
			item.Blocks = append(item.Blocks, inner)
		}
		if p.o.Level >= view.Provenance {
			if facts := p.blockageFacts(c); len(facts.Items) > 0 {
				item.Blocks = append(item.Blocks, facts)
			}
			if pre := commandLines(c.Resolve); pre != nil {
				item.Blocks = append(item.Blocks, *pre)
			}
		}
		list.Items = append(list.Items, item)
	}
	return list
}

// blockageAction words the one action that resolves a cause, by its act
// (886d decision 5, a3cc A4): one phrase per act, and for a status the
// reason it clears, for a snapshot what advances. A cause whose files name
// no resolver has no action line (R9).
func (p *page) blockageAction(c view.Cause) []doc.Inline {
	who := p.person(c.Resolver)
	if len(who) == 0 {
		return nil
	}
	id := code(c.Task.ID)
	var rest []doc.Inline
	switch c.Act {
	case "contributes":
		rest = one(txt(" takes "), id, txt(" through "+c.Gate+"."))
	case "records":
		rest = one(txt(" records a new status."))
		if c.Status.Reason != "" {
			rest = one(txt(" clears the "), code(c.Status.Reason), txt(" reason, or records a new status."))
		}
	case "reviews":
		rest = one(txt(" accepts the "+c.Gate+" with a commit that carries "), code("Reviewed: "+c.Task.ID+" "+c.Gate), txt("."))
	case "authorises":
		rest = one(txt(" commits the trailer "), code("Authorised: "+c.Task.ID), txt("."))
	case "advances":
		switch c.Advance {
		case "pin":
			rest = one(txt(" advances the pin of "), code(c.URL), txt(" to the tip of its trunk"))
			if c.Tip != "" {
				rest = append(rest, txt(", "), code(short(c.Tip)))
			}
			rest = append(rest, txt("."))
		case "gate":
			rest = one(txt(" records "), id, txt(" past "+c.Gate+", where "), code(c.URL), txt(" already stands."))
		}
	}
	if rest == nil {
		return nil
	}
	return spans(who, rest)
}

// blockageFacts words the typed facts behind a cause at provenance, each as
// a bold name and its value: the status, the next junction, the
// authorisation and the linkage, each where the cause holds it (R9).
func (p *page) blockageFacts(c view.Cause) doc.List {
	var facts doc.List
	add := func(name string, value []doc.Inline) {
		if len(value) > 0 {
			facts.Items = append(facts.Items, doc.Item{Text: spans(one(doc.Strong{txt(name + ".")}, txt(" ")), value, one(txt(".")))})
		}
	}
	if c.Status.Gate != "" {
		st := p.status(c.Status)
		if c.Status.Note != "" {
			st = append(st, txt(": "+c.Status.Note))
		}
		st = spans(st, prefixed(", recorded ", p.recorded(c.Status)))
		if c.StatusCommit != nil {
			st = spans(st, one(txt(", by the commit ")), p.commitLine(*c.StatusCommit))
		}
		add("The status", st)
	}
	if j := c.Junction; j != nil {
		add("The junction", p.junctionLine(c.Task.ID, *j))
	}
	if a := c.Authorisation; a != nil {
		add("The authorisation", p.authorisationLine(*a))
	}
	if l := c.Linkage; l != nil {
		add("The linkage", p.linkageLine(*l))
	}
	return facts
}

// junctionLine words a junction on one line: `<id>` at <gate>, then its
// contributor, model and reviewer, and the tasks whose entries supply it.
func (p *page) junctionLine(id string, j view.Junction) []doc.Inline {
	out := spans(one(code(id), txt(" at ")), p.gate(j.Gate))
	var fields [][]doc.Inline
	if who := p.person(j.Contributor); len(who) > 0 {
		fields = append(fields, spans(one(txt("contributor ")), who))
	}
	if j.Model != "" {
		fields = append(fields, one(txt("model "), code(j.Model)))
	}
	if who := p.person(j.Reviewer); len(who) > 0 {
		fields = append(fields, spans(one(txt("reviewer ")), who))
	}
	if j.Subproject != nil {
		fields = append(fields, spans(one(txt("the subproject "), code(j.Subproject.URL), txt(" ")), task(j.Subproject.Task)))
	}
	if len(fields) > 0 {
		out = spans(out, one(txt(": ")), joined(", ", fields...))
	}
	if len(j.From) > 0 {
		out = spans(out, one(txt(", from ")), ids(j.From))
	}
	return out
}

// authorisationWhy words why no commit decides a proposed task.
var authorisationWhy = map[string]string{
	"no_history": "no commit decides it: the source has no history",
	"no_trunk":   "no commit decides it: the trunk is undetermined",
	"off_trunk":  "no commit decides it: the ref stands off the trunk",
	"no_commit":  "no commit on the trunk decides it",
}

// authorisationLine words an authorisation on one line: its deciding
// commit and how it accepts, or why none decides.
func (p *page) authorisationLine(a view.Authorisation) []doc.Inline {
	if a.Commit == nil {
		if w, ok := authorisationWhy[a.Why]; ok {
			return one(txt(w))
		}
		return nil
	}
	out := p.commitLine(*a.Commit)
	if a.Way != "" {
		out = spans(out, one(txt(", accepting by ")), acceptsBy(a.Way))
	}
	if a.By != "" {
		out = append(out, txt(", the authority being "+authority(a.By)))
	}
	return out
}

// linkageLine words a linkage on one line: the task read, the link and the
// commit it reads, and its trunk.
func (p *page) linkageLine(l view.Linkage) []doc.Inline {
	out := spans(task(l.Task), one(txt(" in "), code(l.URL)))
	if l.Commit != "" {
		out = append(out, txt(" at "), code(short(l.Commit)))
	}
	if l.Trunk != "" {
		out = append(out, txt(", on the trunk "), code(l.Trunk))
		if l.Tip != "" {
			out = append(out, txt(" whose tip is "), code(short(l.Tip)))
		}
		if b := linkBehind[l.Behind]; b != "" {
			out = append(out, txt(", "+b))
		}
	}
	return out
}

// blockageHeld words a held task and, nested beneath it, what it holds in
// turn: “<task> at <gate>: “ and the words of its Via.
func (p *page) blockageHeld(h view.Held) doc.Item {
	t := one(code(h.Task.ID))
	switch {
	case h.Task.Title == "":
	case h.Parent:
		t = append(t, txt(" "), doc.Strong{txt(h.Task.Title)})
	default:
		t = append(t, txt(" "+h.Task.Title))
	}
	text := spans(t, one(txt(" at ")), p.gate(h.Gate), one(txt(": ")), p.blockageVia(h))
	for _, n := range h.Also {
		text = append(text, txt(" Also under cause "+strconv.Itoa(n)+"."))
	}
	item := doc.Item{Text: text}
	if len(h.Held) > 0 {
		var inner doc.List
		for _, c := range h.Held {
			inner.Items = append(inner.Items, p.blockageHeld(c))
		}
		item.Blocks = append(item.Blocks, inner)
	}
	return item
}

// blockageVia words why a task is held: the cause holds the task itself; the
// parent of the child that holds it; or the requirement that holds it.
func (p *page) blockageVia(h view.Held) []doc.Inline {
	switch h.Via {
	case "self":
		return one(txt("the cause holds the task itself."))
	case "parent":
		return one(txt("the parent of "), code(h.Child), txt("."))
	case "requirement":
		if h.Requires == nil {
			return one(txt("requires."))
		}
		r := h.Requires
		out := one(txt("requires "), code(r.Task.ID), txt(" at "))
		out = append(out, p.gate(r.From)...)
		return append(out, txt(", "+r.Text+"."))
	}
	return one(txt(h.Via))
}

// blockageNotDue is the section of the requirements not yet due, by task.
func (p *page) blockageNotDue(v *view.Blockage) []doc.Block {
	if v.NotDueN == 0 && len(v.NotDue) == 0 {
		return []doc.Block{countHeading(2, "Next", 0)}
	}
	b := []doc.Block{heading(2, txt("Next: "+count(v.NotDueN, "requirement", "requirements")+" not yet due"))}
	var list doc.List
	for _, w := range v.NotDue {
		t := one(code(w.Task.ID))
		if w.Task.Title != "" {
			t = append(t, txt(" "), doc.Strong{txt(w.Task.Title)})
		}
		text := spans(t, one(txt(" stands at ")), p.phrase(w.Status))
		text = spans(text, one(txt("; next ")), p.gate(w.Next), one(txt(".")))
		item := doc.Item{Text: text}
		var inner doc.List
		for _, r := range w.Requires {
			req := spans(p.gate(r.To), one(txt(": requires ")), task(r.Task))
			if r.Subproject != "" {
				req = spans(req, one(txt(" in "), code(r.Subproject)))
			}
			req = spans(req, one(txt(" at ")), p.gate(r.From), one(txt(", "+r.Text+": "+condition(r.Met, r.Due)+".")))
			inner.Items = append(inner.Items, doc.Item{Text: req})
		}
		if len(inner.Items) > 0 {
			item.Blocks = append(item.Blocks, inner)
		}
		list.Items = append(list.Items, item)
	}
	if len(list.Items) > 0 {
		b = append(b, list)
	}
	return b
}
