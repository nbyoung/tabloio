package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

func init() { register(buildAudit) }

// auditSeeAlso lists the views related to the audit, in the order of the See
// also line.
var auditSeeAlso = []string{"history", "task", "queue", "tableau"}

// buildAudit builds the audit view: the count line and the findings merged by
// severity and rule at glance, one row per finding at detail, and one section
// per finding at provenance. The kind, the message and the sentence are
// tablo's words (R17); the builder words the action from the act and the
// typed facts as every other view words them (dada decision 8).
func buildAudit(v *view.Audit, o Options) doc.Doc {
	p := newPage(&v.Head, o)
	b := []doc.Block{p.auditCount(v)}
	switch {
	case len(v.Findings) == 0:
		b = append(b, emptyForm("No finding."))
	case o.Level >= view.Detail:
		b = append(b, p.auditDetailTable(v))
	default:
		b = append(b, p.auditGlanceTable(v))
	}
	if v.Most != nil && len(v.Findings) > 0 {
		b = append(b, p.auditMost(v))
	}
	if o.Level >= view.Provenance {
		for _, f := range v.Findings {
			b = append(b, p.auditSection(f)...)
		}
	}
	return p.frame(frame{
		Name:     "audit",
		Title:    one(txt("Audit")),
		Question: "Where do files and history disagree?",
		Legend:   true,
		SeeAlso:  auditSeeAlso,
		Body:     b,
	})
}

func auditTotal(v *view.Audit) int { return v.Errors + v.Warnings + v.Information }

// auditCount is the count line: every severity named, zeros included, in
// bold, then the total; at detail it also counts the rows of the table.
func (p *page) auditCount(v *view.Audit) doc.Para {
	counts := count(v.Errors, "error", "errors") + ", " +
		count(v.Warnings, "warning", "warnings") + ", " +
		count(v.Information, "item", "items") + " of information"
	tail := ": " + count(auditTotal(v), "finding", "findings")
	if p.o.Level >= view.Detail && len(v.Findings) > 0 {
		tail += " in " + count(len(v.Findings), "row", "rows")
	}
	return para(doc.Strong{txt(counts)}, txt(tail+"."))
}

// auditMost is the line naming the person with the most to resolve.
func (p *page) auditMost(v *view.Audit) doc.Para {
	return para(spans(
		one(doc.Strong{txt("Most to resolve:")}, txt(" ")),
		p.person(v.Most.Email),
		one(txt(", "+strconv.Itoa(v.Most.Count)+" of "+strconv.Itoa(auditTotal(v))+".")),
	)...)
}

func auditIDs(tasks []view.TaskRef) []string {
	list := make([]string, len(tasks))
	for i, t := range tasks {
		list[i] = t.ID
	}
	return list
}

// auditTasks prints the tasks of a cell: one task by R4, several as ids.
func auditTasks(tasks []view.TaskRef) []doc.Inline {
	if len(tasks) == 1 {
		return task(tasks[0])
	}
	return ids(auditIDs(tasks))
}

// auditRow is the findings of one severity and rule merged into one row of
// the glance table.
type auditRow struct {
	severity, rule, kind string
	findings             int
	tasks                []view.TaskRef
	resolvers            []string
}

// auditRows merges the findings of one severity and rule, in the order of the
// first of each. A finding with no rule (a proposed task, a stale status)
// merges only with those of its own kind. Findings is the sum of the task
// counts; the tasks and the resolvers are listed once each, in the data's order.
func auditRows(findings []view.Finding) []*auditRow {
	var rows []*auditRow
	byKey := map[string]*auditRow{}
	seenTask := map[*auditRow]map[string]bool{}
	seenResolver := map[*auditRow]map[string]bool{}
	for _, f := range findings {
		key := f.Severity + "\x00" + f.Rule
		switch {
		case f.Key != "":
			key = f.Severity + "\x00" + f.Key
		case f.Rule == "":
			key += "\x00" + f.Kind
		}
		r := byKey[key]
		if r == nil {
			r = &auditRow{severity: f.Severity, rule: f.Rule, kind: f.Kind}
			byKey[key] = r
			seenTask[r], seenResolver[r] = map[string]bool{}, map[string]bool{}
			rows = append(rows, r)
		}
		r.findings += len(f.Tasks)
		for _, t := range f.Tasks {
			if !seenTask[r][t.ID] {
				seenTask[r][t.ID] = true
				r.tasks = append(r.tasks, t)
			}
		}
		if f.Resolver != "" && !seenResolver[r][f.Resolver] {
			seenResolver[r][f.Resolver] = true
			r.resolvers = append(r.resolvers, f.Resolver)
		}
	}
	return rows
}

func (p *page) auditGlanceTable(v *view.Audit) doc.Table {
	var rows [][]doc.Cell
	for _, r := range auditRows(v.Findings) {
		var resolvers [][]doc.Inline
		for _, e := range r.resolvers {
			resolvers = append(resolvers, p.person(e))
		}
		rows = append(rows, []doc.Cell{
			cellText(r.severity),
			cellText(r.rule),
			cell(worded(r.kind)),
			cellText(strconv.Itoa(r.findings)),
			cell(auditTasks(r.tasks)),
			cell(joined(", ", resolvers...)),
		})
	}
	return doc.Table{
		Align: []doc.Align{doc.Left, doc.Left, doc.Left, doc.Right, doc.Left, doc.Left},
		Head:  heads("Severity", "Rule", "Kind", "Findings", "Tasks", "Resolver"),
		Rows:  rows,
	}
}

// auditFiles prints files as code spans joined by `, `.
func auditFiles(files []string) []doc.Inline {
	var runs [][]doc.Inline
	for _, f := range files {
		runs = append(runs, one(code(f)))
	}
	return joined(", ", runs...)
}

func (p *page) auditDetailTable(v *view.Audit) doc.Table {
	var rows [][]doc.Cell
	for _, f := range v.Findings {
		rows = append(rows, []doc.Cell{
			cellText(f.Rule),
			cellText(f.Severity),
			cell(auditTasks(f.Tasks)),
			cell(p.gate(f.Gate)),
			cell(auditFiles(f.Files)),
			cell(worded(f.Message)),
			cell(p.auditAction(f)),
			cell(p.person(f.Resolver)),
		})
	}
	return doc.Table{
		Head: heads("Rule", "Severity", "Task", "Gate", "File", "Message", "Action", "Resolver"),
		Rows: rows,
	}
}

// auditSection is the provenance of one finding: its rule and the sentence
// in the method, its facts, the commits involved and the commands.
func (p *page) auditSection(f view.Finding) []doc.Block {
	lead := one(txt(f.Rule))
	if f.Rule == "" {
		lead = worded(f.Kind)
	}
	title := spans(lead, one(txt(", "+f.Severity+": ")), ids(auditIDs(f.Tasks)))
	if f.Gate != "" {
		title = spans(title, one(txt(" at ")), p.gate(f.Gate))
	}
	b := []doc.Block{heading(2, title...)}
	if f.Sentence != "" || f.Source.Text != "" || f.Source.URL != "" {
		rule := one(doc.Strong{txt("Rule.")})
		if f.Sentence != "" {
			rule = spans(rule, one(txt(` "`)), worded(f.Sentence), one(txt(`"`)))
		}
		if f.Source.Text != "" || f.Source.URL != "" {
			rule = spans(rule, one(txt(" ")), p.reference(f.Source))
		}
		b = append(b, para(rule...))
	}
	b = append(b, p.auditFacts(f)...)
	if len(f.Commits) > 0 {
		b = append(b, p.auditCommits(f.Commits))
	}
	if pre := auditCommands(f.Commands); pre != nil {
		b = append(b, *pre)
	}
	return b
}

// auditAction words the action that resolves a finding by its act, one
// phrase per act in the words of VIEWS.md#audit, with the task's id where
// the finding names one task. The act none has no action (R9).
func (p *page) auditAction(f view.Finding) []doc.Inline {
	id := "<id>"
	if len(f.Tasks) == 1 {
		id = f.Tasks[0].ID
	}
	switch f.Act {
	case "authorise":
		return one(txt("Authorise: commit "), code("Authorised: "+id))
	case "review":
		return one(txt("Review: commit "), code("Reviewed: "+id+" "+f.Gate))
	case "reaffirm":
		return one(txt("Reaffirm: commit "), code("Reaffirmed: "+id))
	case "record_handoff":
		return spans(one(txt("Record the hand-off: the reason ")), p.reason("review"), one(txt(" when the work is done; or work on")))
	case "clear_handoff":
		return one(txt("Clear the hand-off: record the gate, or the next hand-off"))
	case "revise":
		return one(txt("Revise the file"))
	case "move_pin":
		return one(txt("Move the pin or the "), code("commit"))
	case "checkout":
		return one(txt("Check out the submodule"))
	case "state_commit":
		return one(txt("State the "), code("commit"), txt(" the requirement lacks"))
	case "raise_language":
		return one(txt("Raise the subproject's language"))
	case "await":
		for _, x := range f.Facts {
			if r := x.Requirement; r != nil {
				return spans(one(txt("Await "), code(r.Task.ID), txt(" at ")), p.gate(r.From))
			}
		}
		return one(txt("Await the task the requirement names"))
	}
	return nil
}

// auditFacts words the typed facts of a finding, one paragraph per kind of
// fact and per value: the tasks whose fact reads alike share it, in the
// order of Tasks.
func (p *page) auditFacts(f view.Finding) []doc.Block {
	kinds := []struct {
		name, sep string
		value     func(view.FindingFacts) []doc.Inline
	}{
		{"The junction", " at ", func(x view.FindingFacts) []doc.Inline {
			if x.Junction == nil {
				return nil
			}
			return p.junctionFields(*x.Junction)
		}},
		{"The status", ": ", func(x view.FindingFacts) []doc.Inline {
			if x.Status == nil {
				return nil
			}
			return p.statusLine(*x.Status)
		}},
		{"The requirement", ": ", func(x view.FindingFacts) []doc.Inline {
			if x.Requirement == nil {
				return nil
			}
			return p.queueEntry(*x.Requirement)
		}},
		{"The authorisation", ": ", func(x view.FindingFacts) []doc.Inline {
			if x.Authorisation == nil {
				return nil
			}
			return p.authorisationLine(*x.Authorisation)
		}},
		{"The linkage", ": ", func(x view.FindingFacts) []doc.Inline {
			if x.Linkage == nil {
				return nil
			}
			return p.linkageLine(*x.Linkage)
		}},
		{"The model check", ": ", func(x view.FindingFacts) []doc.Inline {
			if x.Model == nil {
				return nil
			}
			return p.modelLine(*x.Model)
		}},
	}
	var b []doc.Block
	for _, k := range kinds {
		type group struct {
			ids   []string
			value []doc.Inline
		}
		var groups []*group
		index := map[string]*group{}
		for _, x := range f.Facts {
			v := k.value(x)
			if len(v) == 0 {
				continue
			}
			key := fmt.Sprintf("%#v", v)
			g, ok := index[key]
			if !ok {
				g = &group{value: v}
				index[key] = g
				groups = append(groups, g)
			}
			g.ids = append(g.ids, x.Task)
		}
		for _, g := range groups {
			b = append(b, para(spans(one(doc.Strong{txt(k.name + ".")}, txt(" ")), ids(g.ids), one(txt(k.sep)), g.value, one(txt(".")))...))
		}
	}
	return b
}

// modelLine words a model check on one line: the commit and the gate, the
// model the junction states, the commit's Model: trailer, the reading, and
// the agent the commit names in Co-Authored-By: where it names one.
func (p *page) modelLine(m view.ModelCheck) []doc.Inline {
	out := spans(commit(m.Commit), one(txt(" at ")), p.gate(m.Gate))
	if m.Stated != "" {
		out = append(out, txt(", the junction states "), code(m.Stated))
	}
	if m.Trailer != "" {
		out = append(out, txt(", the trailer "), code("Model: "+m.Trailer))
	}
	return spans(append(out, txt(": "+reading(m.Reading))), p.coAuthor(m.CoAuthor))
}

// auditCommits is the table of the commits a finding involves, oldest first;
// the committer is empty where it equals the author.
func (p *page) auditCommits(commits []view.Commit) doc.Table {
	var rows [][]doc.Cell
	for _, c := range commits {
		var committer []doc.Inline
		if c.Committer != c.Author {
			committer = p.actor(c.Committer)
		}
		var trailers [][]doc.Inline
		for _, t := range c.Trailers {
			trailers = append(trailers, one(code(t)))
		}
		rows = append(rows, []doc.Cell{
			cell(commit(c.Hash)),
			cellText(c.Date),
			cell(p.actor(c.Author)),
			cell(committer),
			cellText(c.Subject),
			cell(joined(", ", trailers...)),
		})
	}
	return doc.Table{Head: heads("Commit", "Date", "Author", "Committer", "Subject", "Trailers"), Rows: rows}
}

// auditCommands prints the commands of a finding, each under its comment,
// with a blank line between them.
func auditCommands(cmds []view.Command) *doc.Pre {
	if len(cmds) == 0 {
		return nil
	}
	pre := &doc.Pre{}
	for i, c := range cmds {
		if i > 0 {
			pre.Lines = append(pre.Lines, "")
		}
		if c.Comment != "" {
			pre.Lines = append(pre.Lines, "# "+c.Comment)
		}
		pre.Lines = append(pre.Lines, strings.Split(c.Text, "\n")...)
	}
	return pre
}
