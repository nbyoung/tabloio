package render

import (
	"path"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// foldAbove is the length of a run of alike rows above which R10 folds it,
// and foldShown the rows that stay.
const (
	foldAbove = 8
	foldShown = 3
)

// page holds what every builder reads: the view's head, the options, and
// the legend's symbols by key. A builder makes one with newPage.
type page struct {
	h       *view.Head
	o       Options
	gates   map[string]string
	states  map[string]string
	reasons map[string]string
	marks   map[string]view.Mark
}

func newPage(h *view.Head, o Options) *page {
	p := &page{
		h: h, o: o,
		gates:   map[string]string{},
		states:  map[string]string{},
		reasons: map[string]string{},
		marks:   map[string]view.Mark{},
	}
	for _, g := range h.Legend.Gates {
		p.gates[g.Key] = g.Symbol
	}
	for _, s := range h.Legend.States {
		p.states[s.Key] = s.Symbol
	}
	for _, r := range h.Legend.Reasons {
		p.reasons[r.Key] = r.Symbol
	}
	for _, m := range h.Legend.Marks {
		p.marks[m.Key] = m
	}
	return p
}

// Small constructors for spans, cells and blocks.

func txt(s string) doc.Inline  { return doc.Text(s) }
func code(s string) doc.Inline { return doc.Code(s) }

// spans joins runs of spans into one run.
func spans(parts ...[]doc.Inline) []doc.Inline {
	var out []doc.Inline
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// one makes a run of the given spans.
func one(xs ...doc.Inline) []doc.Inline { return xs }

// cell makes a cell of one run of spans.
func cell(xs []doc.Inline) doc.Cell { return doc.Cell{Text: xs} }

// cellText makes a cell of fixed or project text; an empty text is an empty cell.
func cellText(s string) doc.Cell {
	if s == "" {
		return doc.Cell{}
	}
	return doc.Cell{Text: one(txt(s))}
}

// heads makes a header row of plain words.
func heads(names ...string) []doc.Cell {
	out := make([]doc.Cell, len(names))
	for i, n := range names {
		out[i] = cellText(n)
	}
	return out
}

func heading(level int, xs ...doc.Inline) doc.Heading {
	return doc.Heading{Level: level, Text: xs}
}

func para(xs ...doc.Inline) doc.Para { return doc.Para{Text: xs} }

// joined puts sep between the runs.
func joined(sep string, runs ...[]doc.Inline) []doc.Inline {
	var out []doc.Inline
	for i, r := range runs {
		if i > 0 {
			out = append(out, txt(sep))
		}
		out = append(out, r...)
	}
	return out
}

// R2 A symbol carries its name.

// keyed prints a legend entry as its symbol, a space and its key: `📐 design`.
// A key with no symbol prints as the key alone (A2).
func keyed(symbols map[string]string, key string) []doc.Inline {
	if key == "" {
		return nil
	}
	if s := symbols[key]; s != "" {
		return one(doc.Symbol(s), txt(" "+key))
	}
	return one(txt(key))
}

func (p *page) gate(key string) []doc.Inline   { return keyed(p.gates, key) }
func (p *page) state(key string) []doc.Inline  { return keyed(p.states, key) }
func (p *page) reason(key string) []doc.Inline { return keyed(p.reasons, key) }

// symbol prints a bare symbol for a grid cell or a Marks cell, or the key
// where the legend has none.
func symbol(sym, key string) doc.Inline {
	if sym == "" {
		return txt(key)
	}
	return doc.Symbol(sym)
}

// markSymbols prints the marks of a junction side by side, as their symbols.
func (p *page) markSymbols(keys []string) []doc.Inline {
	var out []doc.Inline
	for _, k := range keys {
		out = append(out, symbol(p.marks[k].Symbol, k))
	}
	return out
}

// markKey prints the key line under a table of marks (R2): the symbols the
// table shows, in legend order, each with its meaning in lower case, joined
// by " · " with a final full stop. It returns nil when the table shows none.
func (p *page) markKey(shown map[string]bool) *doc.Para {
	var runs [][]doc.Inline
	for _, m := range p.h.Legend.Marks {
		if !shown[m.Key] {
			continue
		}
		meaning := lowerInitial(m.Meaning)
		if meaning == "" {
			meaning = m.Key
		}
		runs = append(runs, spans(one(symbol(m.Symbol, m.Key), txt(" ")), worded(meaning)))
	}
	if len(runs) == 0 {
		return nil
	}
	pr := para(append(joined(" · ", runs...), txt("."))...)
	return &pr
}

func lowerInitial(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	if n == 0 {
		return s
	}
	return string(unicode.ToLower(r)) + s[n:]
}

// R3 The status phrase.

// phrase prints a status as gate, state and reason by R2, joined by spaces;
// a status with no state prints the gate alone.
func (p *page) phrase(st view.Status) []doc.Inline {
	out := p.gate(st.Gate)
	if st.State == "" {
		return out
	}
	out = append(out, txt(" "))
	out = append(out, p.state(st.State)...)
	if st.Reason != "" {
		out = append(out, txt(" "))
		out = append(out, p.reason(st.Reason)...)
	}
	return out
}

// status prints the phrase, and after a comma the origin of a derived
// status: `rolled up from `a6f7` <title>`, or `🪆 from the subproject <url>`.
func (p *page) status(st view.Status) []doc.Inline {
	out := p.phrase(st)
	if st.From == nil {
		return out
	}
	switch st.From.Kind {
	case "rollup":
		out = append(out, txt(", rolled up from "))
		out = append(out, task(st.From.Task)...)
	case "snapshot":
		out = append(out, txt(", "))
		if m, ok := p.marks["subproject"]; ok {
			out = append(out, symbol(m.Symbol, m.Key), txt(" "))
		}
		out = append(out, txt("from the subproject "), code(st.From.URL))
	}
	return out
}

// R4 A task.

// task prints a task as its id in a code span and its title.
func task(t view.TaskRef) []doc.Inline {
	if t.Title == "" {
		return one(code(t.ID))
	}
	return one(code(t.ID), txt(" "+t.Title))
}

// ids prints task ids alone, separated by spaces.
func ids(list []string) []doc.Inline {
	var out []doc.Inline
	for i, id := range list {
		if i > 0 {
			out = append(out, txt(" "))
		}
		out = append(out, code(id))
	}
	return out
}

// R5 A person.

// person prints an email, in bold where it is the person in view.
func (p *page) person(email string) []doc.Inline {
	if email == "" {
		return nil
	}
	if email == p.h.Params.Person {
		return one(doc.Strong{txt(email)})
	}
	return one(txt(email))
}

// actor prints a person of a commit fact: `Name, email` when the data holds
// the name, else the email.
func (p *page) actor(a view.Person) []doc.Inline {
	if a.Name == "" {
		return p.person(a.Email)
	}
	if a.Email == "" {
		return one(txt(a.Name))
	}
	return spans(one(txt(a.Name+", ")), p.person(a.Email))
}

// R6 A commit.

// short is the first seven characters of a hash.
func short(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

// commit prints a commit as its short hash in a code span.
func commit(hash string) []doc.Inline {
	if hash == "" {
		return nil
	}
	return one(code(short(hash)))
}

// commitSubject prints the short hash and the subject.
func commitSubject(c view.Commit) []doc.Inline {
	out := commit(c.Hash)
	if c.Subject != "" {
		out = append(out, txt(" "+c.Subject))
	}
	return out
}

// R8, R9 Counts and empty forms.

// countHeading is a heading over a list whose length depends on the project:
// the name, a colon and the count in digits, or `none` for an empty list.
func countHeading(level int, name string, n int) doc.Heading {
	c := "none"
	if n > 0 {
		c = strconv.Itoa(n)
	}
	return heading(level, txt(name+": "+c))
}

// R16 Words for a requirement.

// edge prints `<from> → <to>`.
func edge(from, to string) string { return from + " → " + to }

// condition prints `met` or `unmet`, with `, not yet due` after it when the
// requirement is not due.
func condition(met, due bool) string {
	s := "unmet"
	if met {
		s = "met"
	}
	if !due {
		s += ", not yet due"
	}
	return s
}

// R17 Two kinds of text.

// worded splits text that tablo words at its backticks into Text and Code,
// so tablo names an id, a key or a trailer in backticks and no front end
// parses more than that. A backtick with no partner stays in the text.
func worded(s string) []doc.Inline {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "`")
	if len(parts)%2 == 0 {
		n := len(parts)
		parts = append(parts[:n-2], parts[n-2]+"`"+parts[n-1])
	}
	var out []doc.Inline
	for i, part := range parts {
		switch {
		case part == "":
		case i%2 == 1:
			out = append(out, code(part))
		default:
			out = append(out, txt(part))
		}
	}
	return out
}

// R10 The run fold.

// taskRow is a row of a table of tasks: the task cell and the others.
type taskRow struct {
	ID   string
	Task []doc.Inline
	Rest []doc.Cell
}

// foldRows turns rows into table rows. A run of more than eight consecutive
// rows that agree in every cell but the task shows its first three rows,
// then one row whose task cell names the ids of all the rest and whose other
// cells repeat the shared values. A row that differs from its neighbours
// never folds; Options.Unfold turns the rule off.
func (p *page) foldRows(rows []taskRow) [][]doc.Cell {
	var out [][]doc.Cell
	line := func(t []doc.Inline, rest []doc.Cell) []doc.Cell {
		return append([]doc.Cell{cell(t)}, rest...)
	}
	for i := 0; i < len(rows); {
		j := i + 1
		for j < len(rows) && reflect.DeepEqual(rows[j].Rest, rows[i].Rest) {
			j++
		}
		if j-i <= foldAbove || p.o.Unfold {
			for _, r := range rows[i:j] {
				out = append(out, line(r.Task, r.Rest))
			}
		} else {
			for _, r := range rows[i : i+foldShown] {
				out = append(out, line(r.Task, r.Rest))
			}
			var rest []string
			for _, r := range rows[i+foldShown : j] {
				rest = append(rest, r.ID)
			}
			t := spans(one(txt("… and "+strconv.Itoa(len(rest))+" more: ")), ids(rest))
			out = append(out, line(t, rows[i].Rest))
		}
		i = j
	}
	return out
}

// R13 Links.

// fileLink links a task's phrase to its file when Links.Project is set. The
// target is <Project>/.tableaux/tasks/<id>.yaml, cleaned as a slash path.
func (p *page) fileLink(t view.TaskRef, text []doc.Inline) []doc.Inline {
	if p.o.Links.Project == "" || t.ID == "" {
		return text
	}
	return one(doc.Link{Text: text, URL: path.Join(p.o.Links.Project, ".tableaux", "tasks", t.ID+".yaml")})
}

// refURL resolves the url of a reference: an absolute URL as given, a path
// relative to the repository root as <Project>/<path>.
func (p *page) refURL(u string) string {
	if strings.Contains(u, "://") || strings.HasPrefix(u, "mailto:") ||
		strings.HasPrefix(u, "/") || strings.HasPrefix(u, "#") {
		return u
	}
	return path.Join(p.o.Links.Project, u)
}

// reference prints a reference as a link; a reference with no text shows its url.
func (p *page) reference(r view.Reference) []doc.Inline {
	text := r.Text
	if text == "" {
		text = r.URL
	}
	if r.URL == "" {
		return one(txt(text))
	}
	return one(doc.Link{Text: one(txt(text)), URL: p.refURL(r.URL)})
}

// viewRef names a view: a link when Links.Views holds it, else its command
// in a code span.
func (p *page) viewRef(name string) []doc.Inline {
	if target, ok := p.o.Links.Views[name]; ok {
		return one(doc.Link{Text: one(txt(name)), URL: target})
	}
	return one(code("tabloio " + name))
}

// commandLines prints a view's reproduce commands, one per line, its
// comment after " # ".
func commandLines(cmds []view.Command) *doc.Pre {
	if len(cmds) == 0 {
		return nil
	}
	pre := &doc.Pre{}
	for _, c := range cmds {
		line := c.Text
		if c.Comment != "" {
			line += " # " + c.Comment
		}
		pre.Lines = append(pre.Lines, line)
	}
	return pre
}
