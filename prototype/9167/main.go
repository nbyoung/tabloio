// Command 9167 renders the history and audit views as Markdown from the JSON
// that tablo's temporal prototype (8ed1) emits. It reads view data only.
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

// Event is one history record, in the shape of schemas/history.schema.yaml.
type Event struct {
	Date   string `json:"date"`
	Commit string `json:"commit"`
	By     string `json:"by"`
	Task   string `json:"task"`
	Event  string `json:"event"`
	Gate   string `json:"gate"`
	State  string `json:"state"`
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

// Finding is one audit record.
type Finding struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Task     string `json:"task"`
	Message  string `json:"message"`
	Since    string `json:"since"`
	Range    string `json:"range"`
}

// Level is how much a rendering discloses; Markdown fixes one per rendering.
type Level int

// The three levels of VIEWS.md.
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
	return 0, fmt.Errorf("unknown level %q (glance, detail, provenance)", s)
}

// Options are the focusing parameters the prototype honours.
type Options struct {
	Level  Level
	Title  string // the range, for the heading and the commands
	Repo   string
	Stale  int
	Person string
	Task   string
}

func cell(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	return strings.ReplaceAll(s, "\n", " ")
}

func table(w io.Writer, head []string, rows [][]string) {
	pf(w, "| %s |\n", strings.Join(head, " | "))
	sep := make([]string, len(head))
	for i := range sep {
		sep[i] = "---"
	}
	pf(w, "| %s |\n", strings.Join(sep, " | "))
	for _, r := range rows {
		c := make([]string, len(r))
		for i, v := range r {
			c[i] = cell(v)
		}
		pf(w, "| %s |\n", strings.Join(c, " | "))
	}
}

// row is one history line: the events one commit made to one task.
type row struct {
	Date, Commit, By, Task string
	Events                 []Event
}

// group merges the events of one commit on one task, as VIEWS.md's example
// does ("task, authorised"). Order stays as given.
func group(evs []Event) []row {
	var rows []row
	for _, e := range evs {
		if n := len(rows); n > 0 {
			l := &rows[n-1]
			if l.Commit == e.Commit && l.Task == e.Task && l.By == e.By {
				l.Events = append(l.Events, e)
				continue
			}
		}
		rows = append(rows, row{e.Date, e.Commit, e.By, e.Task, []Event{e}})
	}
	return rows
}

func names(evs []Event) string {
	var n []string
	for _, e := range evs {
		n = append(n, e.Event)
	}
	return strings.Join(n, ", ")
}

// status reads "gate, state" for a status event; the state is absent when a
// gate has no state of its own.
func status(e Event) string {
	s := e.Gate
	if e.State != "" {
		s += " " + e.State
	} else if e.Gate != "" {
		s += ", no state"
	}
	if e.Reason != "" {
		s += " (" + e.Reason + ")"
	}
	return s
}

func glanceDetail(evs []Event) string {
	var p []string
	for _, e := range evs {
		switch e.Event {
		case "status":
			p = append(p, status(e))
		case "reviewed":
			p = append(p, "at "+e.Gate)
		case "pin":
			p = append(p, e.Note)
		}
	}
	return strings.Join(p, "; ")
}

func detailText(evs []Event) string {
	s := glanceDetail(evs)
	var notes []string
	for _, e := range evs {
		if e.Event != "pin" && e.Note != "" {
			notes = append(notes, e.Note)
		}
	}
	if len(notes) > 0 {
		s += ": " + strings.Join(notes, "; ")
	}
	return s
}

// History renders the history view.
func History(w io.Writer, evs []Event, o Options) {
	var kept []Event
	for _, e := range evs {
		if o.Person != "" && e.By != o.Person {
			continue
		}
		if o.Task != "" && e.Task != o.Task {
			continue
		}
		kept = append(kept, e)
	}
	pf(w, "# History\n\n")
	pf(w, "%s", scope(o))
	if len(kept) == 0 {
		pf(w, "No events.\n")
		return
	}
	rows := group(kept)
	pf(w, "%d events in %d commits, %s to %s.\n\n", len(kept), commits(rows), rows[0].Date, rows[len(rows)-1].Date)
	head := []string{"Date", "By", "Task", "Event", "Detail"}
	if o.Level == Provenance {
		head = []string{"Date", "By", "Task", "Event", "Commit", "Detail"}
	}
	var out [][]string
	for _, r := range rows {
		d := glanceDetail(r.Events)
		if o.Level >= Detail {
			d = detailText(r.Events)
		}
		line := []string{r.Date, r.By, "`" + r.Task + "`", names(r.Events), d}
		if o.Level == Provenance {
			line = []string{r.Date, r.By, "`" + r.Task + "`", names(r.Events), "`" + r.Commit + "`", d}
		}
		out = append(out, line)
	}
	table(w, head, out)
	if o.Level == Provenance {
		pf(w, "\nReproduce:\n\n```\n%s\n```\n", historyCmd(o))
	}
}

func commits(rows []row) int {
	seen := map[string]bool{}
	for _, r := range rows {
		seen[r.Commit] = true
	}
	return len(seen)
}

func scope(o Options) string {
	var p []string
	if o.Title != "" {
		p = append(p, "range `"+o.Title+"`")
	}
	if o.Task != "" {
		p = append(p, "task `"+o.Task+"`")
	}
	if o.Person != "" {
		p = append(p, "person "+o.Person)
	}
	if len(p) == 0 {
		return ""
	}
	return "For " + strings.Join(p, ", ") + ".\n\n"
}

func historyCmd(o Options) string {
	c := "git log --root --reverse --raw --no-abbrev"
	if o.Title != "" {
		c += " " + o.Title
	}
	return c
}

var severityOrder = map[string]int{"error": 0, "warning": 1, "information": 2}

// Audit renders the audit view: findings, errors first.
func Audit(w io.Writer, fs []Finding, o Options) {
	var kept []Finding
	for _, f := range fs {
		if o.Task != "" && f.Task != o.Task {
			continue
		}
		kept = append(kept, f)
	}
	sort.SliceStable(kept, func(i, j int) bool {
		a, b := severityOrder[kept[i].Severity], severityOrder[kept[j].Severity]
		if a != b {
			return a < b
		}
		if kept[i].Rule != kept[j].Rule {
			return kept[i].Rule < kept[j].Rule
		}
		return kept[i].Task < kept[j].Task
	})
	pf(w, "# Audit\n\n%s", scope(o))
	if len(kept) == 0 {
		pf(w, "No findings.\n")
		return
	}
	pf(w, "%s.\n\n", counts(kept))
	switch o.Level {
	case Glance:
		kinds := map[string][]string{}
		var order []string
		for _, f := range kept {
			k := f.Rule + "/" + f.Severity
			if _, ok := kinds[k]; !ok {
				order = append(order, k)
			}
			kinds[k] = append(kinds[k], "`"+f.Task+"`")
		}
		var rows [][]string
		for _, k := range order {
			rs := strings.SplitN(k, "/", 2)
			rows = append(rows, []string{rs[0], rs[1], fmt.Sprint(len(kinds[k])), strings.Join(kinds[k], ", ")})
		}
		table(w, []string{"Rule", "Severity", "Count", "Tasks"}, rows)
	default:
		head := []string{"Rule", "Severity", "Task", "Range", "Message"}
		if o.Level == Provenance {
			head = append(head, "Since")
		}
		var rows [][]string
		for _, f := range kept {
			r := []string{f.Rule, f.Severity, "`" + f.Task + "`", f.Range, f.Message}
			if o.Level == Provenance {
				r = append(r, "`"+f.Since+"`")
			}
			rows = append(rows, r)
		}
		table(w, head, rows)
	}
	if o.Level == Provenance {
		pf(w, "\nEach `Since` commit first shows the finding. Reproduce:\n\n```\ntablo audit --ref <to>")
		if o.Stale > 0 {
			pf(w, " --stale %d", o.Stale)
		}
		pf(w, "\ngit show <since>\n```\n")
	}
}

func counts(fs []Finding) string {
	n := map[string]int{}
	for _, f := range fs {
		n[f.Severity]++
	}
	var p []string
	for _, s := range []string{"error", "warning", "information"} {
		if n[s] > 0 {
			p = append(p, fmt.Sprintf("%d %s", n[s], plural(s, n[s])))
		}
	}
	r := map[string]int{}
	for _, f := range fs {
		r[f.Range]++
	}
	var q []string
	for _, s := range []string{"introduced", "standing", "resolved"} {
		if r[s] > 0 {
			q = append(q, fmt.Sprintf("%d %s", r[s], s))
		}
	}
	return strings.Join(p, ", ") + " (" + strings.Join(q, ", ") + ")"
}

func plural(s string, n int) string {
	if n == 1 {
		return s
	}
	if s == "information" {
		return "items of information"
	}
	return s + "s"
}

func load(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func main() {
	view := flag.String("view", "history", "history or audit")
	in := flag.String("in", "", "view data JSON file")
	level := flag.String("level", "glance", "glance, detail or provenance")
	rng := flag.String("range", "", "the range the data covers, for the heading")
	task := flag.String("task", "", "keep one task")
	person := flag.String("person", "", "keep one actor (history)")
	stale := flag.Int("stale", 0, "the stale age the data used (audit)")
	flag.Parse()
	l, err := ParseLevel(*level)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	o := Options{Level: l, Title: *rng, Stale: *stale, Person: *person, Task: *task}
	switch *view {
	case "history":
		var evs []Event
		if err := load(*in, &evs); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		History(os.Stdout, evs, o)
	case "audit":
		var fs []Finding
		if err := load(*in, &fs); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		Audit(os.Stdout, fs, o)
	default:
		fmt.Fprintln(os.Stderr, "unknown view", *view)
		os.Exit(2)
	}
}

// pf writes formatted text; a failed write to stdout has no recovery here.
func pf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}
