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
		return spans(t, one(txt(" has not passed ")), p.gate(c.Gate))
	case "status":
		out := spans(t, one(txt(" stands at ")))
		if c.Status != nil {
			out = spans(out, p.phrase(*c.Status))
			if c.Status.Note != "" {
				out = append(out, txt(": "+c.Status.Note))
			}
		}
		return out
	case "review":
		return spans(t, one(txt(" awaits review at ")), p.gate(c.Gate))
	case "authorisation":
		return spans(t, one(txt(" is proposed")))
	case "snapshot":
		return spans(t, one(txt(" waits on "), code(c.URL), txt(" at "), code(c.Pin)))
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
		if c.Action != "" {
			inner.Items = append(inner.Items, doc.Item{Text: spans(one(txt("Action: ")), worded(c.Action))})
		}
		for _, h := range c.Held {
			inner.Items = append(inner.Items, p.blockageHeld(h))
		}
		if len(inner.Items) > 0 {
			item.Blocks = append(item.Blocks, inner)
		}
		if p.o.Level >= view.Provenance {
			if len(c.Facts) > 0 {
				var facts doc.List
				for _, f := range c.Facts {
					name := append(worded(f.Name), txt("."))
					facts.Items = append(facts.Items, doc.Item{Text: spans(one(doc.Strong(name), txt(" ")), worded(f.Value))})
				}
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
