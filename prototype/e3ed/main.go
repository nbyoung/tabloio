// Command e3ed renders tablo's gate definition and task definition views as
// Markdown. It reads the JSON that `tablo view` will emit and writes the
// legend or the task page at glance, detail or provenance.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Commit is a Git fact: the commit, its date and its actors.
type Commit struct {
	Hash      string `json:"hash"`
	Author    string `json:"author"`
	Committer string `json:"committer"`
	Date      string `json:"date"`
	Subject   string `json:"subject"`
}

// Symbol pairs a key with its symbol and, for marks, its meaning.
type Symbol struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
	Meaning string `json:"meaning"`
}

// GateRow is one gate in the legend's detail.
type GateRow struct {
	Key            string      `json:"key"`
	Criteria       string      `json:"criteria"`
	Applies        *bool       `json:"applies"`
	References     []Reference `json:"references"`
	ReferencesFrom string      `json:"references_from"`
}

// StateRow is one state in the legend's detail.
type StateRow struct {
	Key      string `json:"key"`
	Severity int    `json:"severity"`
	Synopsis string `json:"synopsis"`
}

// ReasonRow is one reason in the legend's detail.
type ReasonRow struct {
	Key      string `json:"key"`
	Synopsis string `json:"synopsis"`
}

// Reference is a link a task or a junction names.
type Reference struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// GateView is the gate definition view data.
type GateView struct {
	View   string `json:"view"`
	Ref    string `json:"ref"`
	Params struct {
		Task string `json:"task"`
	} `json:"params"`
	Glance struct {
		Folded  int      `json:"folded"`
		Gates   []Symbol `json:"gates"`
		States  []Symbol `json:"states"`
		Reasons []Symbol `json:"reasons"`
		Marks   []Symbol `json:"marks"`
	} `json:"glance"`
	Detail struct {
		Gates   []GateRow   `json:"gates"`
		States  []StateRow  `json:"states"`
		Reasons []ReasonRow `json:"reasons"`
	} `json:"detail"`
	Provenance struct {
		Tableaux    string            `json:"tableaux"`
		Trunk       string            `json:"trunk"`
		LastChanged map[string]Commit `json:"last_changed"`
	} `json:"provenance"`
}

// Status is a task's status line, stated or derived.
type Status struct {
	Gate     string `json:"gate"`
	State    string `json:"state"`
	Reason   string `json:"reason"`
	Note     string `json:"note"`
	Date     string `json:"date"`
	Recorder string `json:"recorder"`
	Derived  bool   `json:"derived"`
	From     string `json:"from"`
	Snapshot *struct {
		Gate string `json:"gate"`
		Pin  string `json:"pin"`
		Task string `json:"task"`
		URL  string `json:"url"`
	} `json:"snapshot"`
}

// Requirement is a requires entry or a dependent.
type Requirement struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	From      string `json:"from"`
	To        string `json:"to"`
	Text      string `json:"text"`
	Met       bool   `json:"met"`
	Due       bool   `json:"due"`
	Condition string `json:"condition"`
}

// Junction is a task's resolved junction at one gate.
type Junction struct {
	Gate        string      `json:"gate"`
	Symbol      string      `json:"symbol"`
	Kind        string      `json:"kind"`
	Marks       []string    `json:"marks"`
	Contributor string      `json:"contributor"`
	Model       string      `json:"model"`
	Reviewer    string      `json:"reviewer"`
	References  []Reference `json:"references"`
	Subproject  *struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	} `json:"subproject"`
}

// TaskView is the task definition view data.
type TaskView struct {
	View   string `json:"view"`
	Ref    string `json:"ref"`
	Glance struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Assignee string `json:"assignee"`
		Parent   *struct {
			ID    string `json:"id"`
			Order int    `json:"order"`
			Title string `json:"title"`
		} `json:"parent"`
		NextGate string `json:"next_gate"`
		Status   Status `json:"status"`
	} `json:"glance"`
	Detail struct {
		Description   string        `json:"description"`
		References    []Reference   `json:"references"`
		Requires      []Requirement `json:"requires"`
		Dependents    []Requirement `json:"dependents"`
		Children      []string      `json:"children"`
		Authorities   []string      `json:"authorities"`
		Junctions     []Junction    `json:"junctions"`
		Authorisation struct {
			State string `json:"state"`
			Way   string `json:"way"`
		} `json:"authorisation"`
	} `json:"detail"`
	Provenance struct {
		Authorisation struct {
			Commit Commit `json:"commit"`
			By     string `json:"by"`
		} `json:"authorisation"`
		StatusCommit    *Commit `json:"status_commit"`
		Command         string  `json:"command"`
		JunctionSources []struct {
			Gate    string            `json:"gate"`
			Sources map[string]string `json:"sources"`
		} `json:"junction_sources"`
		Reviews []struct {
			Gate     string `json:"gate"`
			Reviewer string `json:"reviewer"`
			Commit   string `json:"commit"`
			Date     string `json:"date"`
		} `json:"reviews"`
		Events []struct {
			Date   string `json:"date"`
			Commit string `json:"commit"`
			By     string `json:"by"`
			Event  string `json:"event"`
			Gate   string `json:"gate"`
			State  string `json:"state"`
			Reason string `json:"reason"`
		} `json:"events"`
	} `json:"provenance"`
}

// Level orders the three levels; each includes the one before.
type Level int

// The levels of every view.
const (
	Glance Level = iota
	Detail
	Provenance
)

// ParseLevel reads a level name.
func ParseLevel(s string) (Level, error) {
	switch s {
	case "glance":
		return Glance, nil
	case "detail":
		return Detail, nil
	case "provenance":
		return Provenance, nil
	}
	return 0, fmt.Errorf("unknown level %q", s)
}

func short(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}

func cell(s string) string {
	if s == "" {
		return " "
	}
	return strings.ReplaceAll(s, "|", `\|`)
}

func row(w *strings.Builder, cells ...string) {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = cell(c)
	}
	fmt.Fprintf(w, "| %s |\n", strings.Join(out, " | "))
}

func rule(w *strings.Builder, n int) {
	fmt.Fprintf(w, "|%s\n", strings.Repeat("---|", n))
}

func commitText(c Commit) string {
	return fmt.Sprintf("`%s` %s, author %s, committer %s, %q", short(c.Hash), c.Date, c.Author, c.Committer, c.Subject)
}

func pair(s Symbol, name string) string {
	return strings.TrimSpace(s.Symbol + " " + name)
}

// reservedReview states the reserved reason; the view data omits it.
const reservedReview = "The contributor has handed the next junction's work to its reviewer."

// RenderGate writes the gate definition view.
func RenderGate(w *strings.Builder, v *GateView, lv Level) {
	g := v.Glance
	title := "# Legend"
	if v.Params.Task != "" {
		title += " for task `" + v.Params.Task + "`"
	}
	fmt.Fprintf(w, "%s\n\n## Gates\n\n", title)
	var head []string
	for _, s := range g.Gates {
		head = append(head, pair(s, s.Key))
	}
	row(w, head...)
	fmt.Fprintf(w, "|%s\n", strings.Repeat(":-:|", len(head)))
	if g.Folded > 0 {
		fmt.Fprintf(w, "\n%d gates folded.\n", g.Folded)
	}
	fmt.Fprint(w, "\nStates ")
	var parts []string
	sev := map[string]int{}
	for _, s := range v.Detail.States {
		sev[s.Key] = s.Severity
	}
	for _, s := range g.States {
		parts = append(parts, pair(s, s.Key))
	}
	fmt.Fprintf(w, "%s.\n\nReasons ", strings.Join(parts, " · "))
	parts = nil
	for _, s := range g.Reasons {
		parts = append(parts, pair(s, s.Key))
	}
	fmt.Fprintf(w, "%s.\n\nMarks ", strings.Join(parts, " · "))
	parts = nil
	for _, s := range g.Marks {
		parts = append(parts, s.Symbol+" "+s.Meaning)
	}
	fmt.Fprintf(w, "%s.\n", strings.Join(parts, " · "))
	if lv < Detail {
		return
	}
	sym := map[string]Symbol{}
	for _, s := range g.Gates {
		sym[s.Key] = s
	}
	fmt.Fprint(w, "\n## Gate criteria\n\n")
	expanded := v.Params.Task != ""
	if expanded {
		row(w, "Gate", "Name", "Criteria", "Applies", "References")
		rule(w, 5)
	} else {
		row(w, "Gate", "Name", "Criteria")
		rule(w, 3)
	}
	for _, r := range v.Detail.Gates {
		s := sym[r.Key]
		if !expanded {
			row(w, pair(s, r.Key), s.Name, r.Criteria)
			continue
		}
		applies, refs := "yes", ""
		if r.Applies != nil && !*r.Applies {
			applies = "—"
		}
		var rs []string
		for _, ref := range r.References {
			rs = append(rs, link(ref))
		}
		refs = strings.Join(rs, "; ")
		if refs != "" && r.ReferencesFrom != "" {
			refs += " (from `" + r.ReferencesFrom + "`)"
		}
		row(w, pair(s, r.Key), s.Name, r.Criteria, applies, refs)
	}
	fmt.Fprint(w, "\n## States\n\n")
	row(w, "State", "Severity", "Synopsis")
	rule(w, 3)
	states := map[string]Symbol{}
	for _, s := range g.States {
		states[s.Key] = s
	}
	for _, r := range v.Detail.States {
		row(w, pair(states[r.Key], r.Key), fmt.Sprint(r.Severity), r.Synopsis)
	}
	fmt.Fprint(w, "\n## Reasons\n\n")
	row(w, "Reason", "Synopsis")
	rule(w, 2)
	reasons := map[string]Symbol{}
	for _, s := range g.Reasons {
		reasons[s.Key] = s
	}
	for _, r := range v.Detail.Reasons {
		row(w, pair(reasons[r.Key], r.Key), r.Synopsis)
	}
	row(w, "👓 review", reservedReview+" Reserved.")
	fmt.Fprint(w, "\n## Marks\n\n")
	row(w, "Mark", "Meaning")
	rule(w, 2)
	for _, m := range g.Marks {
		row(w, m.Symbol, m.Meaning)
	}
	if lv < Provenance {
		return
	}
	p := v.Provenance
	fmt.Fprintf(w, "\n## Provenance\n\nLanguage version %s, trunk `%s`, at `%s`.\n\n", p.Tableaux, p.Trunk, short(v.Ref))
	files := make([]string, 0, len(p.LastChanged))
	for f := range p.LastChanged {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		fmt.Fprintf(w, "- `%s` last changed by %s\n", f, commitText(p.LastChanged[f]))
	}
}

func link(r Reference) string {
	if r.Text == "" {
		return "<" + r.URL + ">"
	}
	return "[" + r.Text + "](" + r.URL + ")"
}

// Legend holds the symbols the task view needs and its data omits.
type Legend struct {
	gates, states, reasons map[string]string
}

// DefaultLegend holds the method's symbols, for a task view rendered without
// a gate view beside it.
func DefaultLegend() *Legend {
	return &Legend{
		states:  map[string]string{"undefined": "⚪", "nominal": "🟢", "at_risk": "🟡", "stalled": "🔴", "complete": "✅"},
		reasons: map[string]string{"overloaded": "🪫", "blocked": "⛔", "review": "👓"},
	}
}

// LegendFrom takes the symbols from a gate view, over the defaults.
func LegendFrom(g *GateView) *Legend {
	l := DefaultLegend()
	l.gates = map[string]string{}
	for _, s := range g.Glance.Gates {
		l.gates[s.Key] = s.Symbol
	}
	for _, s := range g.Glance.States {
		l.states[s.Key] = s.Symbol
	}
	for _, s := range g.Glance.Reasons {
		l.reasons[s.Key] = s.Symbol
	}
	return l
}

func statusLine(v *TaskView, l *Legend) string {
	s := v.Glance.Status
	gate := l.gates[s.Gate]
	for _, j := range v.Detail.Junctions {
		if j.Gate == s.Gate {
			gate = j.Symbol
		}
	}
	out := strings.TrimSpace(gate + " " + s.Gate)
	if s.State != "" {
		out += " " + strings.TrimSpace(l.states[s.State]+" "+s.State)
	}
	if s.Reason != "" {
		out += " " + strings.TrimSpace(l.reasons[s.Reason]+" "+s.Reason)
	}
	who := ""
	if s.Date != "" {
		who = ", " + s.Date
	}
	if s.Recorder != "" {
		who += ", " + s.Recorder
	}
	if s.Derived {
		who += ", rolled up from `" + s.From + "`"
	}
	out += who
	if s.Note != "" {
		out += ": " + s.Note
	}
	return out
}

func joinMarks(m []string) string { return strings.Join(m, "") }

// RenderTask writes the task definition view.
func RenderTask(w *strings.Builder, v *TaskView, l *Legend, lv Level) {
	g := v.Glance
	line := fmt.Sprintf("`%s` **%s** — %s", g.ID, g.Title, g.Assignee)
	if g.Parent != nil {
		line += fmt.Sprintf(" — under `%s` %s, order %d", g.Parent.ID, g.Parent.Title, g.Parent.Order)
	} else {
		line += " — the root"
	}
	fmt.Fprintf(w, "> %s — %s\n", line, statusLine(v, l))
	if sn := g.Status.Snapshot; sn != nil {
		fmt.Fprintf(w, ">\n> Snapshot of `%s` in `%s` at `%s`, at %s.\n", sn.Task, sn.URL, short(sn.Pin), sn.Gate)
	}
	if lv < Detail {
		return
	}
	d := v.Detail
	fmt.Fprintf(w, "\n## Description\n\n%s\n", d.Description)
	if len(d.References) > 0 {
		fmt.Fprint(w, "\n## References\n\n")
		for _, r := range d.References {
			fmt.Fprintf(w, "- %s\n", link(r))
		}
	}
	if len(d.Children) > 0 {
		ids := make([]string, len(d.Children))
		for i, c := range d.Children {
			ids[i] = "`" + c + "`"
		}
		fmt.Fprintf(w, "\n## Children\n\n%s, in display order.\n", strings.Join(ids, ", "))
	}
	reqs := func(head string, rs []Requirement) {
		if len(rs) == 0 {
			return
		}
		fmt.Fprintf(w, "\n## %s\n\n", head)
		row(w, "Task", "From", "To", "Text", "Condition")
		rule(w, 5)
		for _, r := range rs {
			c := r.Condition
			if r.Due {
				c += ", due"
			}
			row(w, "`"+r.ID+"` "+r.Title, r.From, r.To, r.Text, c)
		}
	}
	reqs("Requirements", d.Requires)
	reqs("Dependents", d.Dependents)
	fmt.Fprint(w, "\n## Junctions\n\n")
	row(w, "Gate", "Marks", "Contributor", "Model", "Reviewer", "Subproject")
	rule(w, 6)
	for _, j := range d.Junctions {
		sub := ""
		if j.Subproject != nil {
			sub = "`" + j.Subproject.ID + "` at `" + j.Subproject.URL + "`"
		}
		row(w, strings.TrimSpace(j.Symbol+" "+j.Gate), joinMarks(j.Marks), j.Contributor, j.Model, j.Reviewer, sub)
	}
	fmt.Fprintf(w, "\nAuthorisation: %s.\n", d.Authorisation.State)
	if len(d.Authorities) > 0 {
		fmt.Fprintf(w, "\nAuthorities, nearest first: %s.\n", strings.Join(d.Authorities, ", "))
	}
	if lv < Provenance {
		return
	}
	p := v.Provenance
	fmt.Fprint(w, "\n## Provenance\n\n")
	fmt.Fprintf(w, "- Authorisation: %s, by the %s, way: %s.\n", commitText(p.Authorisation.Commit), p.Authorisation.By, d.Authorisation.Way)
	if p.StatusCommit != nil {
		fmt.Fprintf(w, "- Status: %s.\n", commitText(*p.StatusCommit))
	}
	for _, r := range p.Reviews {
		fmt.Fprintf(w, "- Reviewed at %s: `%s` %s by %s.\n", r.Gate, short(r.Commit), r.Date, r.Reviewer)
	}
	var notable []string
	for _, s := range p.JunctionSources {
		var fs []string
		for f, from := range s.Sources {
			if from == "default" {
				continue
			}
			fs = append(fs, f+" from `"+from+"`")
		}
		sort.Strings(fs)
		if len(fs) > 0 {
			notable = append(notable, "  - "+s.Gate+": "+strings.Join(fs, ", "))
		}
	}
	if len(notable) > 0 {
		fmt.Fprint(w, "- Junction fields resolved from a task (`default` fields omitted):\n"+strings.Join(notable, "\n")+"\n")
	}
	if len(p.Events) > 0 {
		fmt.Fprint(w, "\nEvents:\n\n")
		row(w, "Date", "Event", "Gate", "Commit", "By")
		rule(w, 5)
		for _, e := range p.Events {
			row(w, e.Date, e.Event, e.Gate, "`"+short(e.Commit)+"`", e.By)
		}
	}
	if p.Command != "" {
		fmt.Fprintf(w, "\nTo list them all:\n\n```\n%s\n```\n", p.Command)
	}
}

func load(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("e3ed", flag.ContinueOnError)
	level := fs.String("level", "glance", "glance, detail or provenance")
	gates := fs.String("gates", "", "gate view JSON that supplies symbols to a task view")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: e3ed [-level l] [-gates gate.json] view.json")
	}
	lv, err := ParseLevel(*level)
	if err != nil {
		return err
	}
	var sb strings.Builder
	var head struct {
		View string `json:"view"`
	}
	if err := load(fs.Arg(0), &head); err != nil {
		return err
	}
	switch head.View {
	case "gate":
		var v GateView
		if err := load(fs.Arg(0), &v); err != nil {
			return err
		}
		RenderGate(&sb, &v, lv)
	case "task":
		var v TaskView
		if err := load(fs.Arg(0), &v); err != nil {
			return err
		}
		l := DefaultLegend()
		if *gates != "" {
			var g GateView
			if err := load(*gates, &g); err != nil {
				return err
			}
			l = LegendFrom(&g)
		}
		RenderTask(&sb, &v, l, lv)
	default:
		return fmt.Errorf("view %q is not a gate or task definition", head.View)
	}
	_, err = io.WriteString(out, sb.String())
	return err
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
