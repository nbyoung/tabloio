package render

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { register(buildDelegation) }

// delegationSeeAlso lists the views related to the authority delegation, in
// the order of the See also line.
var delegationSeeAlso = []string{"task", "assignment", "history", "audit", "queue"}

// collapseFrom is the number of children from which the glance tree folds a
// parent's children to a count (design c48a, decision 1).
const collapseFrom = 5

// buildDelegation builds the authority delegation view: the tree and the
// count line at glance; the whole tree, the chains and the defaults at
// detail; and the deciding commits at provenance.
func buildDelegation(v *view.Delegation, o Options) doc.Doc {
	p := newPage(&v.Head, o)
	detail := o.Level >= view.Detail

	b := p.treeBlocks(v)
	b = append(b, p.countLine(v))
	if detail {
		b = append(b, p.chains(v)...)
	}
	if o.Level >= view.Provenance {
		b = append(b, p.deciding(v)...)
	}
	return p.frame(frame{
		Name:     "authority",
		Title:    one(txt("Authority delegation")),
		Question: "Who may accept what?",
		Legend:   true,
		SeeAlso:  delegationSeeAlso,
		Body:     b,
	})
}

// treeBlocks is the tree as a preformatted block, or the bold sentence that
// stands in its place when the view holds no row (R9).
func (p *page) treeBlocks(v *view.Delegation) []doc.Block {
	if len(v.Rows) == 0 {
		if p.h.Params.Proposed {
			return []doc.Block{emptyForm("No task is proposed.")}
		}
		return []doc.Block{emptyForm("No task is in view.")}
	}
	return []doc.Block{doc.Pre{Lines: drawTree(v.Rows, p.collapsed(v.Rows))}}
}

// collapsed marks the rows the tree hides and counts the children of the
// parents that carry the count (R11). At glance a parent hides its children
// when it has five or more, each a leaf with the parent's assignee and none
// proposed. Detail, provenance and Options.Unfold draw every row.
func (p *page) collapsed(rows []view.DelegationRow) map[int]int {
	if p.o.Level >= view.Detail || p.o.Unfold {
		return nil
	}
	folds := map[int]int{} // the index of a parent to its number of children
	for i, r := range rows {
		if !r.Parent {
			continue
		}
		n, ok := 0, true
		for j := i + 1; j < len(rows) && rows[j].Depth > r.Depth; j++ {
			c := rows[j]
			if c.Depth != r.Depth+1 || c.Parent || c.Assignee != r.Assignee || c.Proposed {
				ok = false
				break
			}
			n++
		}
		if ok && n >= collapseFrom {
			folds[i] = n
		}
	}
	return folds
}

// oneLine puts a text on one line: a line break prints as a space.
var oneLine = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ")

// drawTree draws the tree: a first line that reads Task and Assignee, then a
// row per task. A row prints two spaces per step of depth, the id, a space
// and the title; then spaces up to the assignee column; then two spaces per
// step of depth again and the assignee. The column starts two characters
// past the longest task text among the rows drawn, counted in code points.
// After the assignee, two spaces and a word follow where one applies.
func drawTree(rows []view.DelegationRow, folds map[int]int) []string {
	type line struct {
		task, assignee, word string
		depth                int
	}
	var lines []line
	width := 0
	for i := 0; i < len(rows); i++ {
		r := rows[i]
		l := line{
			task:     strings.Repeat("  ", r.Depth) + r.ID + " " + oneLine.Replace(r.Title),
			assignee: oneLine.Replace(r.Assignee),
			depth:    r.Depth,
		}
		switch n, folded := folds[i]; {
		case folded:
			l.word = strconv.Itoa(n) + " children"
		case r.Proposed && r.Differs:
			l.word = "proposed (differs from trunk)"
		case r.Proposed:
			l.word = "proposed"
		}
		lines = append(lines, l)
		width = max(width, utf8.RuneCountInString(l.task))
		if _, folded := folds[i]; folded {
			i += folds[i] // the children are hidden
		}
	}
	column := width + 2
	pad := func(s string, to int) string {
		return s + strings.Repeat(" ", max(to-utf8.RuneCountInString(s), 0))
	}
	out := []string{pad("Task", column) + "Assignee"}
	for _, l := range lines {
		s := pad(l.task, column) + strings.Repeat("  ", l.depth) + l.assignee
		if l.word != "" {
			s += "  " + l.word
		}
		out = append(out, s)
	}
	return out
}

// countLine is `<n> tasks, <n> delegations, <n> authorised, <n> proposed.`;
// it counts every task in view, the hidden ones included.
func (p *page) countLine(v *view.Delegation) doc.Block {
	var delegations, proposed int
	for _, r := range v.Rows {
		if r.Delegated {
			delegations++
		}
		if r.Proposed {
			proposed++
		}
	}
	n := strconv.Itoa
	return para(txt(n(len(v.Rows)) + " tasks, " + n(delegations) + " delegations, " +
		n(len(v.Rows)-proposed) + " authorised, " + n(proposed) + " proposed."))
}

// chains builds the section of the parents: each one's authorities, its
// children and the junction defaults it states.
func (p *page) chains(v *view.Delegation) []doc.Block {
	b := []doc.Block{countHeading(2, "Chains and defaults", len(v.Parents))}
	for _, c := range v.Parents {
		b = append(b, heading(3, spans(p.fileLink(c.TaskRef, task(c.TaskRef)), one(txt(" · ")), p.person(c.Assignee))...))
		b = append(b, para(spans(
			one(txt("Authorities of its "+strconv.Itoa(len(c.Children))+" children, nearest first: ")),
			p.authorities(c.Authorities), one(txt(".")))...))
		b = append(b, para(spans(one(txt("Children: ")), ids(c.Children), one(txt(".")))...))
		b = append(b, p.defaults(c)...)
	}
	return b
}

// defaults is the table of the junction entries a parent states, with its
// key line (R2), or the sentence that says it states none.
func (p *page) defaults(c view.Chain) []doc.Block {
	if len(c.Defaults) == 0 {
		return []doc.Block{para(code(c.ID), txt(" states no junction defaults."))}
	}
	shown := map[string]bool{}
	var rows [][]doc.Cell
	for _, d := range c.Defaults {
		for _, m := range d.Marks {
			shown[m] = true
		}
		var model []doc.Inline
		if d.Model != "" {
			model = one(code(d.Model))
		}
		rows = append(rows, []doc.Cell{
			cell(p.gate(d.Gate)),
			cell(p.markSymbols(d.Marks)),
			cell(fromAncestor(p.person(d.Contributor), d.ContributorFrom)),
			cell(fromAncestor(model, d.ModelFrom)),
			cell(fromAncestor(p.person(d.Reviewer), d.ReviewerFrom)),
		})
	}
	b := []doc.Block{doc.Table{
		Align: []doc.Align{doc.Left, doc.Centre, doc.Left, doc.Left, doc.Left},
		Head:  heads("Gate", "Marks", "Contributor", "Model", "Reviewer"),
		Rows:  rows,
	}}
	if k := p.markKey(shown); k != nil {
		b = append(b, *k)
	}
	return b
}

// fromAncestor adds `, from <id>` to a field that an ancestor supplies.
func fromAncestor(value []doc.Inline, from string) []doc.Inline {
	if len(value) == 0 || from == "" {
		return value
	}
	return spans(value, one(txt(", from "), code(from)))
}

// deciding builds the table of the commits that decide the tasks, newest
// first, and the commands that reproduce a row.
func (p *page) deciding(v *view.Delegation) []doc.Block {
	b := []doc.Block{countHeading(2, "Deciding commits", len(v.Deciding))}
	if len(v.Deciding) > 0 {
		var rows [][]doc.Cell
		for _, d := range v.Deciding {
			rows = append(rows, []doc.Cell{
				cell(commit(d.Commit.Hash)),
				cellText(d.Commit.Date),
				cell(p.actor(d.Commit.Author)),
				cell(p.actor(d.Commit.Committer)),
				cell(p.acceptsBy(d)),
				cellText(authority(d.By)),
				cell(spans(one(txt(strconv.Itoa(len(d.Tasks))+": ")), ids(d.Tasks))),
			})
		}
		b = append(b, doc.Table{
			Head: heads("Commit", "Date", "Author", "Committer", "Accepts by", "The authority is", "Tasks"),
			Rows: rows,
		})
	}
	if pre := commandLines(v.Commands); pre != nil {
		b = append(b, *pre)
	}
	return b
}

// acceptsBy words the way a commit decides: a change on the trunk, a merge
// of the commit it brings in, or an Authorised: trailer.
func (p *page) acceptsBy(d view.Deciding) []doc.Inline {
	out := acceptsBy(d.Way)
	if d.Way == "merge" && d.Merged != "" {
		out = append(out, txt(" of "), code(short(d.Merged)))
	}
	return out
}
