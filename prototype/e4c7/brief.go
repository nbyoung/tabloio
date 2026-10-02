package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// The types read the view data of the tablo prototypes: the work queue
// (886d) and the task and gate views (493e). Unknown fields are ignored.

type queueView struct {
	Ref    string      `json:"ref"`
	Person string      `json:"person"`
	Items  []queueItem `json:"items"`
}

type queueItem struct {
	Kind       string `json:"kind"`
	Task       string `json:"task"`
	Title      string `json:"title"`
	Gate       string `json:"gate"`
	Cause      string `json:"cause"`
	StatusDate string `json:"status_date"`
	Dependents int    `json:"dependents"`
}

type reference struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type junction struct {
	Contributor string      `json:"contributor"`
	Gate        string      `json:"gate"`
	Kind        string      `json:"kind"`
	Model       string      `json:"model"`
	References  []reference `json:"references"`
	Reviewer    string      `json:"reviewer"`
	Symbol      string      `json:"symbol"`
}

type requirement struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	From      string `json:"from"`
	To        string `json:"to"`
	Text      string `json:"text"`
	Condition string `json:"condition"`
	// Set on a cross-project entry.
	URL    string `json:"url"`
	Commit string `json:"commit"`
}

type status struct {
	Gate     string `json:"gate"`
	State    string `json:"state"`
	Reason   string `json:"reason"`
	Note     string `json:"note"`
	Date     string `json:"date"`
	Recorder string `json:"recorder"`
}

type taskView struct {
	Detail struct {
		Description string        `json:"description"`
		References  []reference   `json:"references"`
		Junctions   []junction    `json:"junctions"`
		Requires    []requirement `json:"requires"`
		Dependents  []requirement `json:"dependents"`
	} `json:"detail"`
	Glance struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Assignee string `json:"assignee"`
		Parent   struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"parent"`
		Status status `json:"status"`
	} `json:"glance"`
	Provenance struct {
		Sources []struct {
			Gate    string            `json:"gate"`
			Sources map[string]string `json:"sources"`
		} `json:"junction_sources"`
	} `json:"provenance"`
}

type gateView struct {
	Detail struct {
		Gates []struct {
			Key      string `json:"key"`
			Criteria string `json:"criteria"`
		} `json:"gates"`
	} `json:"detail"`
}

func load(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// selectItem returns the queue item for a task and gate. An item of a kind
// with no gate, an authorisation owed, matches on the task alone when gate is
// empty.
func selectItem(q queueView, task, gate string) (queueItem, error) {
	for _, it := range q.Items {
		if it.Task == task && (gate == "" || it.Gate == gate) {
			return it, nil
		}
	}
	return queueItem{}, fmt.Errorf("no item for %s %s in the queue of %s", task, gate, q.Person)
}

func (t taskView) junction(gate string) (junction, bool) {
	for _, j := range t.Detail.Junctions {
		if j.Gate == gate {
			return j, true
		}
	}
	return junction{}, false
}

// source names the task whose entry supplies a junction field, "default" for
// the plain default, or "" when the data does not say.
func (t taskView) source(gate, field string) string {
	for _, s := range t.Provenance.Sources {
		if s.Gate == gate {
			return s.Sources[field]
		}
	}
	return ""
}

// inputs holds the view data a brief draws on.
type inputs struct {
	Queue queueView
	Task  taskView
	Gate  gateView
}

func (in inputs) criteria(gate string) string {
	for _, g := range in.Gate.Detail.Gates {
		if g.Key == gate {
			return g.Criteria
		}
	}
	return ""
}

// render writes the brief for one queue item.
func render(w io.Writer, in inputs, it queueItem) {
	t := in.Task
	p := func(format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }
	who := in.Queue.Person
	j, hasJ := t.junction(it.Gate)
	gateLine := it.Gate
	if hasJ && j.Symbol != "" {
		gateLine = j.Symbol + " " + it.Gate
	}

	p("# Brief: %s %s at %s\n\n", it.Task, it.Title, it.Gate)
	p("This brief stands alone. It covers one item of the work queue of %s on `%s`: %s.\n\n", who, in.Queue.Ref, it.Kind)

	// 1. The item.
	p("## The item\n\n")
	if hasJ {
		contributor := orNone(j.Contributor, t.Glance.Assignee)
		p("- Contributor: %s", contributor)
		if j.Model != "" {
			p(", model `%s`", j.Model)
		}
		p("%s\n", from(t.source(it.Gate, "contributor")))
		if rv := reviewerOf(j, t); rv != "" {
			p("- Reviewer: %s%s\n", rv, from(t.source(it.Gate, "reviewer")))
		} else {
			p("- Reviewer: none\n")
		}
	}
	p("- Gate: %s", gateLine)
	if c := in.criteria(it.Gate); c != "" {
		p(": %s", c)
	}
	p("\n")
	if hasJ && len(j.References) > 0 {
		p("- Junction references, which expand the criteria for this task:\n")
		for _, r := range j.References {
			p("  - %s\n", refLine(r))
		}
	}
	if it.Cause != "" {
		p("- Cause: %s\n", it.Cause)
	}
	p("\n")
	if j.Model != "" {
		p("The `model` is the plan's statement (F20). Run on a model that matches it, as an identifier or by prefix, and record the model you ran in the `Model:` trailer.\n\n")
	}

	// 2. The task.
	p("## The task\n\n")
	if t.Detail.Description != "" {
		p("%s\n\n", t.Detail.Description)
	}
	if len(t.Detail.References) > 0 {
		p("References:\n\n")
		for _, r := range t.Detail.References {
			p("- %s\n", refLine(r))
		}
		p("\n")
	}
	if t.Glance.Parent.ID != "" {
		p("Under: %s (`%s`)\n\n", t.Glance.Parent.Title, t.Glance.Parent.ID)
	}
	if src := t.source(it.Gate, "contributor"); src != "" && src != "default" && src != t.Glance.ID {
		p("Your junction comes from the entry of task `%s`.\n\n", src)
	}

	// 3. The requirements.
	p("## Requirements\n\n")
	if len(t.Detail.Requires) == 0 {
		p("Requires: none.\n")
	}
	for _, r := range t.Detail.Requires {
		p("Requires: `%s` %s from %s to %s, %s: %s", r.ID, r.Title, r.From, r.To, r.Text, r.Condition)
		if r.URL != "" {
			p(" (project %s, read at commit %s)", r.URL, r.Commit)
		}
		p("\n")
	}
	p("\n")
	unblocks, others := splitDependents(t.Detail.Dependents, it.Gate)
	for _, d := range unblocks {
		p("Unblocks: `%s` %s at %s, %s\n", d.ID, d.Title, d.To, d.Text)
	}
	for _, d := range others {
		p("Also needed by: `%s` %s from %s to %s, %s\n", d.ID, d.Title, d.From, d.To, d.Text)
	}
	if len(unblocks)+len(others) == 0 {
		p("Unblocks: nothing.\n")
	}
	p("\n")

	// 4. The status.
	p("## Status\n\n")
	s := t.Glance.Status
	p("%s %s", s.Gate, s.State)
	if s.Reason != "" {
		p(" (%s)", s.Reason)
	}
	p(", %s, %s", s.Date, s.Recorder)
	if s.Note != "" {
		p(": %s", s.Note)
	}
	p("\n\n")

	// 5. The commit.
	p("## When done\n\n")
	renderCommit(w, in, it, j, hasJ)

	// 6. Reproduce.
	p("## Reproduce\n\n```\n")
	if it.Gate != "" {
		p("tabloio queue --person %s --brief %s %s\n", who, it.Task, it.Gate)
	} else {
		p("tabloio queue --person %s --brief %s\n", who, it.Task)
	}
	p("tabloio task %s\n```\n", it.Task)
}

func reviewerOf(j junction, t taskView) string {
	if j.Reviewer != "" {
		return j.Reviewer
	}
	if j.Model != "" {
		return t.Glance.Assignee
	}
	return ""
}

func orNone(s, fallback string) string {
	if s != "" {
		return s
	}
	return fallback
}

func from(src string) string {
	switch src {
	case "":
		return ""
	case "default", "assignee":
		return "  (" + src + ")"
	}
	return "  (from `" + src + "`)"
}

func refLine(r reference) string {
	if r.Text == "" {
		return r.URL
	}
	return r.Text + ", " + r.URL
}

func splitDependents(ds []requirement, gate string) (unblocks, others []requirement) {
	for _, d := range ds {
		if d.From == gate {
			unblocks = append(unblocks, d)
		} else {
			others = append(others, d)
		}
	}
	return
}

// renderCommit writes the commit the agent makes, which depends on the item's
// kind and, for work, on the reviewer the junction resolves.
func renderCommit(w io.Writer, in inputs, it queueItem, j junction, hasJ bool) {
	p := func(format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }
	t := in.Task
	id := it.Task
	agent := orNone(j.Contributor, in.Queue.Person)
	model := "<the identifier the harness reports>"
	if j.Model != "" {
		model = "<the identifier the harness reports, matching " + j.Model + ">"
	}
	// A work item's junction says whether an agent contributes; the other
	// kinds are for an agent, as the brief is.
	isAgent := j.Model != "" || it.Kind != "work ready"
	trailers := func(lines ...string) {
		if len(lines) == 0 && !isAgent {
			p("The contributor is a person, so the commit needs no trailer.\n\n")
			return
		}
		p("Trailers, last, as one final paragraph with no blank line between the lines:\n\n```\n")
		for _, l := range lines {
			p("%s\n", l)
		}
		if isAgent {
			p("Model: %s\nCo-Authored-By: <your model's display name> <noreply@anthropic.com>\n", model)
		}
		p("```\n\n")
	}
	role := "the contributor"
	if isAgent {
		role = "the agent"
	}
	identity := fmt.Sprintf("Author and committer are %s, `%s` (D3). Commit on a branch off `%s`; do not merge.\n\n", role, agent, in.Queue.Ref)

	switch it.Kind {
	case "work ready":
		rv := reviewerOf(j, t)
		p("Commit the work and `.tableaux/status/%s.yaml`. %s", id, identity)
		switch {
		case rv != "" && rv != agent:
			p("The reviewer `%s` accepts this gate, so record the hand-off. The status stays at the gate it stands at, with the reason `review`:\n\n```\ngate: %s\nstate: nominal\nreason: review\nnote: <one line: what is on the branch>\n```\n\n", rv, t.Glance.Status.Gate)
			trailers()
			p("The reviewer merges the branch and passes the gate with a commit carrying `Reviewed: %s %s`.\n\n", id, it.Gate)
		case rv == agent && it.Gate != "defined":
			p("You review yourself, so the commit that records the status passes the gate:\n\n```\ngate: %s\nstate: nominal\nnote: <one line: what is on the branch>\n```\n\n", it.Gate)
			trailers(fmt.Sprintf("Reviewed: %s %s", id, it.Gate))
			p("The owner merges the branch.\n\n")
		default:
			p("No one reviews this junction:\n\n```\ngate: %s\nstate: nominal\nnote: <one line: what is on the branch>\n```\n\n", it.Gate)
			if it.Gate == "defined" {
				p("Authorisation stands as the review of `defined`, so the commit needs no `Reviewed:` trailer.\n\n")
			}
			trailers()
		}
	case "work waiting":
		p("Do not start. The work waits: %s. Nothing is committed.\n\n", it.Cause)
	case "reaffirmation":
		p("Confirm the status as it stands, with an empty commit. %s", identity)
		trailers("Reaffirmed: " + id)
	case "review owed":
		p("Accept the work with an empty commit, as the reviewer. %s", identity)
		trailers(fmt.Sprintf("Reviewed: %s %s", id, it.Gate))
	case "authorisation owed":
		p("Accept the task as it stands, with an empty commit on the trunk. %s", identity)
		trailers("Authorised: " + id)
	default:
		p("The kind `%s` has no commit in this prototype.\n\n", it.Kind)
	}
}
