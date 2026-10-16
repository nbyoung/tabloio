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

// historyEffects words the keys of a line's effect (8ed1 decision 4). A key
// outside the table prints as it is.
var historyEffects = map[string]string{
	"authorised_commit":  "an authority commits the task file on the trunk, so the task is authorised",
	"authorised_merge":   "an authority merges the change, so the task is authorised",
	"authorised_trailer": "an authority's trailer authorises the task",
	"proposed":           "no authority accepts the change, so the task stands proposed",
	"handoff":            "the contributor hands the work to the reviewer",
	"accepts":            "the reviewer accepts the work at the gate",
	"stands":             "none; the authorisation stands as the review",
	"none":               "none",
}

// first is the first event of a line, which gives its date, its author and
// its commit; a line with no event gives the zero event.
func first(l view.Line) view.Event {
	if len(l.Events) == 0 {
		return view.Event{}
	}
	return l.Events[0]
}

// taken is the subproject a line comes from, or nil for the project's own.
func taken(l view.Line) *view.SubEvent { return first(l).Subproject }

// ownLines are the project's own lines, which the counts cover.
func ownLines(v *view.History) []view.Line {
	var out []view.Line
	for _, l := range v.Lines {
		if taken(l) == nil {
			out = append(out, l)
		}
	}
	return out
}

// buildHistory builds the history view: the count line and the lines of the
// range at glance, the lines by day with their facts at detail, and the Git
// facts of each line at provenance.
func buildHistory(v *view.History, o Options) doc.Doc {
	p := newPage(&v.Head, o)
	title := one(txt("History"))
	if v.Subject.ID != "" {
		title = spans(title, one(txt(": ")), task(v.Subject))
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
// to <last day>: ` and the six kinds, over the project's own lines. One day
// prints once; no line prints no day.
func (p *page) historyCount(v *view.History) doc.Para {
	s := count(v.Events, "event", "events") + " in " + count(v.Commits, "commit", "commits")
	if own := ownLines(v); len(own) > 0 {
		from, to := first(own[0]).Date, first(own[len(own)-1]).Date
		s += ", " + from
		if to != from {
			s += " to " + to
		}
	}
	return para(txt(s + ": " + historyKindCounts(v.Kinds) + "."))
}

// historyFolds reports whether the glance folds its lines into the days.
func (p *page) historyFolds(v *view.History) bool {
	return p.o.Level < view.Detail && len(ownLines(v)) > historyFoldAbove && !p.o.Unfold
}

// historyGlance is the day table, when the lines span three days or fold,
// and the line table or, in its place, the sentence of the fold. The glance
// draws the project's own lines alone, as tablo gives them there (8ed1
// decision 2).
func (p *page) historyGlance(v *view.History) []doc.Block {
	var b []doc.Block
	fold := p.historyFolds(v)
	if len(v.Days) >= historyDayTable || fold {
		b = append(b, p.historyDayTable(v))
	}
	if fold {
		b = append(b, emptyForm(count(len(ownLines(v)), "line", "lines")+" fold into the days above."))
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
	for _, l := range ownLines(v) {
		e := first(l)
		rows = append(rows, []doc.Cell{
			cellText(e.Date),
			cell(p.person(e.By)),
			cell(historyEvent(l)),
			cell(historyTask(l)),
			cell(p.historyLast(l)),
		})
	}
	return doc.Table{
		Head: heads("Date", "By", "Event", "Task", "Gate and state, or pin"),
		Rows: rows,
	}
}

// historyKindsOf is the kinds of a line's events, in their order.
func historyKindsOf(l view.Line) []string {
	var out []string
	for _, e := range l.Events {
		out = append(out, e.Kind)
	}
	return out
}

// historyEvent is the kinds of a line joined by `, `, with ` (proposal)` after
// them when the branch adds the line.
func historyEvent(l view.Line) []doc.Inline {
	s := strings.Join(historyKindsOf(l), ", ")
	if l.Proposal {
		s += " (proposal)"
	}
	return one(txt(s))
}

// historyTask is the task of a line, and for a line taken from a subproject,
// the subproject and its task.
func historyTask(l view.Line) []doc.Inline {
	out := task(l.Task)
	if s := taken(l); s != nil {
		out = append(out, txt(", taken from "), code(s.URL), txt(" "), code(s.Task))
	}
	return out
}

// historyRecorded is the first status a line's events record, or nil.
func historyRecorded(l view.Line) *view.EventStatus {
	for _, e := range l.Events {
		if e.Kind == "status" && e.Recorded != nil {
			return e.Recorded
		}
	}
	return nil
}

// historyLast is the last cell of a line: the first of the status phrase, the
// gate of a review, the move of a pin, and `the task file appears`.
func (p *page) historyLast(l view.Line) []doc.Inline {
	if r := historyRecorded(l); r != nil {
		return p.phrase(view.Status{Gate: r.Gate, State: r.State, Reason: r.Reason})
	}
	for _, e := range l.Events {
		if e.Kind == "reviewed" && e.Gate != "" {
			return p.gate(e.Gate)
		}
	}
	for _, e := range l.Events {
		if e.Kind == "pin" && e.Pin != nil {
			return historyPin(*e.Pin)
		}
	}
	if historyHas(historyKindsOf(l), "task") {
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

// historyNest finds, for each line taken from a subproject, the line of the
// project's own whose pin of that subproject follows it next; the renderer
// draws the taken line beneath that pin (8ed1 decision 2). A taken line with
// no such pin after it stands in its own place.
func historyNest(lines []view.Line) map[int][]int {
	under := map[int][]int{}
	for i, l := range lines {
		s := taken(l)
		if s == nil {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if taken(lines[j]) != nil {
				continue
			}
			pinned := false
			for _, e := range lines[j].Events {
				pinned = pinned || e.Kind == "pin" && e.Pin != nil && e.Pin.URL == s.URL
			}
			if pinned {
				under[j] = append(under[j], i)
				break
			}
		}
	}
	return under
}

// historyDays is the detail: per day a heading and a list of its lines.
func (p *page) historyDays(v *view.History) []doc.Block {
	days := map[string]view.Day{}
	for _, d := range v.Days {
		days[d.Date] = d
	}
	under := historyNest(v.Lines)
	nested := map[int]bool{}
	for _, ks := range under {
		for _, k := range ks {
			nested[k] = true
		}
	}
	var top []int
	for i := range v.Lines {
		if !nested[i] {
			top = append(top, i)
		}
	}
	var b []doc.Block
	for i := 0; i < len(top); {
		date := first(v.Lines[top[i]]).Date
		j := i
		for j < len(top) && first(v.Lines[top[j]]).Date == date {
			j++
		}
		head := date
		if d, ok := days[date]; ok {
			head += " · " + count(d.Events, "event", "events") + " in " + count(d.Commits, "commit", "commits") +
				": " + historyKindsOccurring(d.Kinds)
		}
		var items []doc.Item
		for _, k := range top[i:j] {
			var sub []view.Line
			for _, n := range under[k] {
				sub = append(sub, v.Lines[n])
			}
			items = append(items, p.historyItem(v, v.Lines[k], sub))
		}
		b = append(b, heading(2, txt(head)), doc.List{Items: items})
		i = j
	}
	return b
}

// historyRun is a line's leading cells, joined by ` · `, without the empty ones.
func (p *page) historyRun(l view.Line) []doc.Inline {
	e := first(l)
	runs := [][]doc.Inline{
		one(doc.Strong{txt(e.Date)}),
		p.person(e.By),
		historyEvent(l),
		historyTask(l),
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

// historyItem is one line at detail: the line and its facts beneath it, and
// beneath a pin the lines taken from its subproject.
func (p *page) historyItem(v *view.History, l view.Line, sub []view.Line) doc.Item {
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
	if r := historyRecorded(l); r != nil && r.Note != "" {
		fact("Note:", one(txt(r.Note)))
	}
	if l.Effect != "" {
		words, ok := historyEffects[l.Effect]
		if !ok {
			words = l.Effect
		}
		fact("Effect:", one(txt(words)))
	}
	if l.SubjectAfter != nil {
		facts = append(facts, doc.Item{Text: spans(task(v.Subject), one(txt(" after: ")), p.status(*l.SubjectAfter))})
	}
	if m := p.historyModel(l); len(m) > 0 {
		fact("Model:", m)
	}
	if l.Committer != "" {
		fact("Committer:", p.person(l.Committer))
	}
	if len(sub) > 0 {
		facts = append(facts, p.historySub(sub))
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

// historyModel reads a line's Model: trailer by its check's reading. With no
// check and a trailer it is the model alone; with neither it is empty.
func (p *page) historyModel(l view.Line) []doc.Inline {
	c := l.Check
	if c == nil {
		if l.Model != "" {
			return one(code(l.Model))
		}
		return nil
	}
	junction := func() []doc.Inline {
		switch {
		case c.Stated == "":
			return nil
		case c.Gate == "":
			return one(txt("; the junction states "), code(c.Stated))
		}
		return spans(one(txt("; the ")), p.gate(c.Gate), one(txt(" junction states "), code(c.Stated)))
	}
	switch c.Reading {
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
	case "unstated":
		if l.Model != "" {
			return spans(one(code(l.Model)), one(txt("; the junction states no model")))
		}
	}
	if l.Model != "" {
		return one(code(l.Model))
	}
	return nil
}

// historySub is the lines taken from a subproject beneath the pin they
// follow: their events, counted here, and one item per line.
func (p *page) historySub(lines []view.Line) doc.Item {
	events, commits := 0, map[string]bool{}
	for _, l := range lines {
		events += len(l.Events)
		for _, e := range l.Events {
			if e.Commit != "" {
				commits[e.Commit] = true
			}
		}
	}
	head := "Subproject events: " + count(events, "event", "events")
	if len(commits) > 0 {
		head += " in " + count(len(commits), "commit", "commits")
	}
	var items []doc.Item
	for _, l := range lines {
		e := first(l)
		runs := [][]doc.Inline{
			one(txt(e.Date)),
			p.person(e.By),
			historyEvent(l),
			one(code(e.Subproject.Task)),
			p.historyLast(l),
		}
		if l.Model != "" {
			runs = append(runs, one(code(l.Model)))
		}
		items = append(items, doc.Item{Text: historyJoin(runs)})
	}
	return doc.Item{
		Text:   one(txt(head+" of "), code(first(lines[0]).Subproject.URL)),
		Blocks: []doc.Block{doc.List{Items: items}},
	}
}

// historyProvenance holds the Git facts of a line, after its other facts:
// Commit, Subject, Author, Committer, Trailers and Files.
func (p *page) historyProvenance(l view.Line) []doc.Item {
	var out []doc.Item
	add := func(label string, value []doc.Inline) {
		out = append(out, doc.Item{Text: spans(one(txt(label+" ")), value)})
	}
	c := l.Commit
	if c == nil {
		return nil
	}
	if c.Hash != "" {
		add("Commit:", one(code(c.Hash)))
	}
	if c.Subject != "" {
		add("Subject:", one(txt(c.Subject)))
	}
	if a := p.historyActor(c.Author, c.AuthorTime); len(a) > 0 {
		add("Author:", a)
	}
	if a := p.historyActor(c.Committer, ""); len(a) > 0 {
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
