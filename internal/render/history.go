package render

import (
	"strconv"
	"strings"
	"time"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { register(buildHistory) }

// The history's own folds (design 9167, decision 1): the day table stands
// from historyDayTable days, and the glance folds its lines above
// historyFoldAbove lines. Both are apart from the run fold of alike rows.
const (
	historyDayTable  = 3
	historyFoldAbove = 50
)

// historySeeAlso lists the views related to the history, in the order of the
// See also line.
var historySeeAlso = []string{"gates", "task", "queue", "audit", "tableau"}

// historyKinds are the six kinds of event, in the fixed order of the count line.
var historyKinds = []string{"task", "authorised", "status", "reaffirmed", "reviewed", "pin"}

// historyEffects words the keys of a line's effect. A key outside the table
// prints as it is.
var historyEffects = map[string]string{
	"authorised-commit":  "an authority commits the task file on the trunk, so the task is authorised",
	"authorised-merge":   "an authority merges the change, so the task is authorised",
	"authorised-trailer": "an authority's trailer authorises the task",
	"proposed":           "no authority accepts the change, so the task stands proposed",
	"handoff":            "the contributor hands the work to the reviewer",
	"accepts":            "the reviewer accepts the work at the gate",
	"stands":             "none; the authorisation stands as the review",
	"none":               "none",
}

// buildHistory builds the history view: the count line and the lines of the
// range at glance, the lines by day with their facts at detail, and the Git
// facts of each line at provenance.
func buildHistory(v *view.History, o Options) doc.Doc {
	p := newPage(&v.Head, o)
	title := one(txt("History"))
	if v.Subject != nil {
		title = spans(title, one(txt(": ")), task(*v.Subject))
	}
	b := []doc.Block{p.historyCount(v)}
	switch {
	case len(v.Lines) == 0:
		b = append(b, emptyForm("No event in view."))
	case o.Level >= view.Detail:
		b = append(b, p.historyDays(v)...)
	default:
		b = append(b, p.historyGlance(v)...)
	}
	if o.Level >= view.Provenance {
		if pre := commandLines(v.Commands); pre != nil {
			b = append(b, *pre)
		}
	}
	return p.frame(frame{
		Name:     "history",
		Title:    title,
		Question: "What happened, when, and who did it?",
		Legend:   true,
		SeeAlso:  historySeeAlso,
		Body:     b,
	})
}

// historyKindCounts prints the six kinds, each as its count and its key, zeros
// included, in the fixed order. A kind the data does not hold counts zero.
func historyKindCounts(kinds []view.KindCount) string {
	n := map[string]int{}
	for _, k := range kinds {
		n[k.Kind] = k.Count
	}
	var parts []string
	for _, k := range historyKinds {
		parts = append(parts, historyKindWord(n[k], k))
	}
	return strings.Join(parts, ", ")
}

// historyKindsOccurring prints the kinds of a day that occur, as the data
// gives them.
func historyKindsOccurring(kinds []view.KindCount) string {
	var parts []string
	for _, k := range kinds {
		if k.Count > 0 {
			parts = append(parts, historyKindWord(k.Count, k.Kind))
		}
	}
	return strings.Join(parts, ", ")
}

// historyKindWord prints a count and the key it counts: `4 status`.
func historyKindWord(n int, key string) string { return count(n, key, key) }

// historyCount is the count line: `<n> event(s) in <m> commit(s), <first day>
// to <last day>: ` and the six kinds. One day prints once; no line prints no day.
func (p *page) historyCount(v *view.History) doc.Para {
	s := count(v.Events, "event", "events") + " in " + count(v.Commits, "commit", "commits")
	if n := len(v.Lines); n > 0 {
		first, last := v.Lines[0].Date, v.Lines[n-1].Date
		s += ", " + first
		if last != first {
			s += " to " + last
		}
	}
	return para(txt(s + ": " + historyKindCounts(v.Kinds) + "."))
}

// historyFolds reports whether the glance folds its lines into the days.
func (p *page) historyFolds(v *view.History) bool {
	return p.o.Level < view.Detail && len(v.Lines) > historyFoldAbove && !p.o.Unfold
}

// historyGlance is the day table, when the lines span three days or fold,
// and the line table or, in its place, the sentence of the fold.
func (p *page) historyGlance(v *view.History) []doc.Block {
	var b []doc.Block
	fold := p.historyFolds(v)
	if len(v.Days) >= historyDayTable || fold {
		b = append(b, p.historyDayTable(v))
	}
	if fold {
		b = append(b, emptyForm(count(len(v.Lines), "line", "lines")+" fold into the days above."))
	} else {
		b = append(b, p.historyLineTable(v))
	}
	return b
}

func (p *page) historyDayTable(v *view.History) doc.Table {
	var rows [][]doc.Cell
	for _, d := range v.Days {
		rows = append(rows, []doc.Cell{
			cellText(d.Date),
			cellText(strconv.Itoa(d.Events)),
			cellText(strconv.Itoa(d.Commits)),
			cellText(historyKindsOccurring(d.Kinds)),
		})
	}
	return doc.Table{
		Align: []doc.Align{doc.Left, doc.Right, doc.Right, doc.Left},
		Head:  heads("Day", "Events", "Commits", "Kinds"),
		Rows:  rows,
	}
}

func (p *page) historyLineTable(v *view.History) doc.Table {
	var rows [][]doc.Cell
	for _, l := range v.Lines {
		rows = append(rows, []doc.Cell{
			cellText(l.Date),
			cell(p.person(l.By)),
			cell(historyEvent(l)),
			cell(task(l.Task)),
			cell(p.historyLast(l)),
		})
	}
	return doc.Table{
		Head: heads("Date", "By", "Event", "Task", "Gate and state, or pin"),
		Rows: rows,
	}
}

// historyEvent is the kinds of a line joined by `, `, with ` (proposal)` after
// them when the branch adds the line.
func historyEvent(l view.Line) []doc.Inline {
	s := strings.Join(l.Kinds, ", ")
	if l.Proposal {
		s += " (proposal)"
	}
	return one(txt(s))
}

// historyLast is the last cell of a line: the first of the status phrase, the
// gate of a review, the move of a pin, and `the task file appears`.
func (p *page) historyLast(l view.Line) []doc.Inline {
	switch {
	case l.Status != nil:
		return p.status(*l.Status)
	case l.Gate != "":
		return p.gate(l.Gate)
	case l.Pin != nil:
		return historyPin(*l.Pin)
	case historyHas(l.Kinds, "task"):
		return one(txt("the task file appears"))
	}
	return nil
}

// historyPin prints a pin as its url, the old commit, an arrow and the new
// commit, or as the url and the new commit alone for a first pin. Hashes are
// cut to seven.
func historyPin(m view.PinMove) []doc.Inline {
	out := one(code(m.URL), txt(" "))
	if m.Old == "" {
		return append(out, txt("new "), code(short(m.New)))
	}
	return append(out, code(short(m.Old)), txt(" → "), code(short(m.New)))
}

func historyHas(kinds []string, kind string) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

// historyDays is the detail: per day a heading and a list of its lines.
func (p *page) historyDays(v *view.History) []doc.Block {
	days := map[string]view.Day{}
	for _, d := range v.Days {
		days[d.Date] = d
	}
	var b []doc.Block
	for i := 0; i < len(v.Lines); {
		j := i
		for j < len(v.Lines) && v.Lines[j].Date == v.Lines[i].Date {
			j++
		}
		date := v.Lines[i].Date
		head := date
		if d, ok := days[date]; ok {
			head += " · " + count(d.Events, "event", "events") + " in " + count(d.Commits, "commit", "commits") +
				": " + historyKindsOccurring(d.Kinds)
		}
		var items []doc.Item
		for _, l := range v.Lines[i:j] {
			items = append(items, p.historyItem(v, l))
		}
		b = append(b, heading(2, txt(head)), doc.List{Items: items})
		i = j
	}
	return b
}

// historyRun is a line's leading cells, joined by ` · `, without the empty ones.
func (p *page) historyRun(l view.Line) []doc.Inline {
	runs := [][]doc.Inline{
		one(doc.Strong{txt(l.Date)}),
		p.person(l.By),
		historyEvent(l),
		task(l.Task),
		p.historyLast(l),
	}
	return historyJoin(runs)
}

// historyJoin joins the runs by ` · ` and drops the empty ones.
func historyJoin(runs [][]doc.Inline) []doc.Inline {
	var kept [][]doc.Inline
	for _, r := range runs {
		if len(r) > 0 {
			kept = append(kept, r)
		}
	}
	return joined(" · ", kept...)
}

// historyItem is one line at detail: the line and its facts beneath it.
func (p *page) historyItem(v *view.History, l view.Line) doc.Item {
	var facts []doc.Item
	fact := func(label string, value []doc.Inline, blocks ...doc.Block) {
		facts = append(facts, doc.Item{Text: spans(one(txt(label+" ")), value), Blocks: blocks})
	}
	if l.After != nil {
		value := p.status(*l.After)
		if l.Unchanged {
			value = append(value, txt(", unchanged"))
		}
		fact("Status after:", value)
	}
	if l.Status != nil && l.Status.Note != "" {
		fact("Note:", one(txt(l.Status.Note)))
	}
	if l.Effect != "" {
		words, ok := historyEffects[l.Effect]
		if !ok {
			words = l.Effect
		}
		fact("Effect:", one(txt(words)))
	}
	if l.SubjectAfter != nil {
		subject := v.Project
		if v.Subject != nil {
			subject = *v.Subject
		}
		facts = append(facts, doc.Item{Text: spans(task(subject), one(txt(" after: ")), p.status(*l.SubjectAfter))})
	}
	if m := p.historyModel(l); len(m) > 0 {
		fact("Model:", m)
	}
	if l.Committer != "" {
		fact("Committer:", p.person(l.Committer))
	}
	if l.Sub != nil {
		facts = append(facts, p.historySub(l.Sub))
	}
	if p.o.Level >= view.Provenance {
		facts = append(facts, p.historyProvenance(l)...)
	}
	it := doc.Item{Text: p.historyRun(l)}
	if len(facts) > 0 {
		it.Blocks = []doc.Block{doc.List{Items: facts}}
	}
	return it
}

// historyModel reads a line's Model: trailer by its reading. With no reading
// and a trailer it is the model alone; with neither it is empty.
func (p *page) historyModel(l view.Line) []doc.Inline {
	junction := func() []doc.Inline {
		switch {
		case l.Stated == "":
			return nil
		case l.StatedGate == "":
			return one(txt("; the junction states "), code(l.Stated))
		}
		return spans(one(txt("; the ")), p.gate(l.StatedGate), one(txt(" junction states "), code(l.Stated)))
	}
	switch l.Reading {
	case "agrees":
		if l.Model != "" {
			return spans(one(code(l.Model)), junction())
		}
	case "differs":
		if l.Model != "" {
			return spans(one(code(l.Model)), junction(), one(txt(", which differs")))
		}
	case "missing":
		return spans(one(txt("no trailer")), junction())
	case "exempt":
		return one(txt("none recorded; exempt"))
	}
	if l.Model != "" {
		return one(code(l.Model))
	}
	return nil
}

// historySub is the subproject's own events between two pins: the count and,
// nested beneath it, one item per line.
func (p *page) historySub(s *view.SubEvents) doc.Item {
	it := doc.Item{Text: spans(
		one(txt("Subproject events: "+count(s.Events, "event", "events")+" in "+count(s.Commits, "commit", "commits")+" of "), code(s.URL)),
	)}
	if len(s.Lines) == 0 {
		return it
	}
	var items []doc.Item
	for _, l := range s.Lines {
		runs := [][]doc.Inline{
			one(txt(l.Date)),
			p.person(l.By),
			historyEvent(l),
			task(l.Task),
			p.historyLast(l),
		}
		if l.Model != "" {
			runs = append(runs, one(code(l.Model)))
		}
		items = append(items, doc.Item{Text: historyJoin(runs)})
	}
	it.Blocks = []doc.Block{doc.List{Items: items}}
	return it
}

// historyProvenance holds the Git facts of a line, after its other facts:
// Commit, Subject, Author, Committer, Trailers, Files and Reproduce.
func (p *page) historyProvenance(l view.Line) []doc.Item {
	var out []doc.Item
	add := func(label string, value []doc.Inline, blocks ...doc.Block) {
		out = append(out, doc.Item{Text: spans(one(txt(label+" ")), value), Blocks: blocks})
	}
	if c := l.Commit; c != nil {
		if c.Hash != "" {
			add("Commit:", one(code(c.Hash)))
		}
		if c.Subject != "" {
			add("Subject:", one(txt(c.Subject)))
		}
		if a := p.historyActor(c.Author, c.AuthorTime); len(a) > 0 {
			add("Author:", a)
		}
		if a := p.historyActor(c.Committer, c.CommitterTime); len(a) > 0 {
			add("Committer:", a)
		}
		if len(c.Trailers) > 0 {
			var runs [][]doc.Inline
			for _, t := range c.Trailers {
				runs = append(runs, one(code(t)))
			}
			add("Trailers:", joined(", ", runs...))
		}
		if len(c.Files) > 0 {
			var runs [][]doc.Inline
			for _, f := range c.Files {
				run := one(code(f.Path))
				if f.Change != "" {
					run = append(run, txt(" ("+f.Change+")"))
				}
				runs = append(runs, run)
			}
			add("Files:", joined(", ", runs...))
		}
	}
	if pre := commandLines(l.Commands); pre != nil {
		add("Reproduce:", nil, *pre)
	}
	return out
}

// historyActor prints a person of a commit as Git writes it: `Name <email>,
// <time>`.
func (p *page) historyActor(a view.Person, when string) []doc.Inline {
	var s string
	switch {
	case a.Name != "" && a.Email != "":
		s = a.Name + " <" + a.Email + ">"
	case a.Email != "":
		s = "<" + a.Email + ">"
	default:
		s = a.Name
	}
	if s == "" {
		return nil
	}
	if when != "" {
		s += ", " + gitTime(when)
	}
	return one(txt(s))
}

// gitTime prints an RFC 3339 time as Git prints it with --date=iso:
// `2026-09-29 16:38:16 -0400`. A time that does not parse stays as given.
func gitTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02 15:04:05 -0700")
}
