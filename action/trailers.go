package action

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Review plans an empty commit with one "Reviewed: <id> <gate>" per task.
func (r *Repo) Review(ctx context.Context, ids []string, gate string, o Options) (*Plan, error) {
	s, project, tasks, err := r.empty(ctx, ids)
	if err != nil {
		return nil, err
	}
	var refusals Refusals
	if !slices.Contains(project.Gates, gate) {
		return nil, Refusals{{Code: "unknown-gate", Gate: gate, Rule: "H1", Text: "names no gate of gates.yaml"}}
	}
	for _, t := range tasks {
		refuse := func(code, rule, text string) {
			refusals = append(refusals, Refusal{Code: code, Task: t.ID, Gate: gate, Rule: rule, Text: text})
		}
		j := t.junction(gate)
		reviewer := t.reviewer(j)
		switch {
		case gate == project.Gates[0] || j == nil || j.Kind == notApplicable:
			refuse("gate-not-applicable", "H1", "the gate does not apply to the task, so the method cannot read the trailer")
		case len(t.Children) > 0:
			refuse("not-leaf", "", "a parent's junctions are defaults; it has no work to accept")
		case gate == definedGate:
			refuse("defined-review", "S11", "authorisation stands as the review of defined; authorise the task instead")
		case j.Kind == recursive:
			refuse("recursive", "", "the junction is recursive; the subproject reviews its own work")
		case reviewer == "":
			refuse("no-reviewer", "", "the junction states no reviewer, so the trailer has no effect")
		case !s.holds(reviewer):
			refuse("wrong-hand", "H2", fmt.Sprintf("the reviewer is %s; a review from another hand has no effect", reviewer))
		}
	}
	if len(refusals) > 0 {
		return nil, refusals
	}
	p := &Plan{Command: "review", Tasks: ids, Subject: "Accept " + about(tasks) + " at " + gate}
	var requests []request
	for _, id := range ids {
		requests = append(requests, request{"reviewed", id, gate})
	}
	return s.finish(ctx, p, o, requests)
}

// Reaffirm plans an empty commit with one "Reaffirmed: <id>" per task.
func (r *Repo) Reaffirm(ctx context.Context, ids []string, o Options) (*Plan, error) {
	s, project, tasks, err := r.empty(ctx, ids)
	if err != nil {
		return nil, err
	}
	var refusals Refusals
	var warnings []Warning
	for _, t := range tasks {
		refuse := func(code, rule, text string) {
			refusals = append(refusals, Refusal{Code: code, Task: t.ID, Rule: rule, Text: text})
		}
		switch {
		case len(t.Children) > 0:
			refuse("not-leaf", "S2", "a parent has no status file to confirm")
		case t.Status == nil:
			refuse("no-status", "", "the task has no status file to confirm")
		default:
			next := nextJunction(&t, project.Gates, t.Status.Gate)
			if next != nil && next.Kind == recursive {
				refuse("recursive", "S9", fmt.Sprintf("the next junction, at %s, is recursive; the status date comes from the subproject", next.Gate))
				continue
			}
			if w := s.notContributor(&t, next); w != nil {
				warnings = append(warnings, *w)
			}
		}
	}
	if len(refusals) > 0 {
		return nil, refusals
	}
	p := &Plan{Command: "reaffirm", Tasks: ids, Subject: "Reaffirm " + about(tasks), Warnings: warnings}
	var requests []request
	for _, id := range ids {
		requests = append(requests, request{"reaffirmed", id, ""})
	}
	return s.finish(ctx, p, o, requests)
}

// Authorise plans an empty commit with one "Authorised: <id>" per task.
func (r *Repo) Authorise(ctx context.Context, ids []string, o Options) (*Plan, error) {
	s, project, tasks, err := r.empty(ctx, ids)
	if err != nil {
		return nil, err
	}
	if project.Trunk == "" || s.branch != project.Trunk {
		text := "the trunk is undetermined; authorisation reads the trunk's first-parent history alone"
		if project.Trunk != "" {
			text = fmt.Sprintf("the checkout is not on the trunk %s; authorisation reads the trunk's first-parent history alone", project.Trunk)
		}
		return nil, Refusals{{Code: "off-trunk", Text: text}}
	}
	var refusals Refusals
	for _, t := range tasks {
		if !slices.ContainsFunc(t.Authorities, s.holds) {
			refusals = append(refusals, Refusal{Code: "wrong-hand", Task: t.ID,
				Text: "the deciding commit counts only from an authority: " + strings.Join(t.Authorities, ", ")})
		}
	}
	if len(refusals) > 0 {
		return nil, refusals
	}
	p := &Plan{Command: "authorise", Tasks: ids, Subject: "Authorise " + about(tasks)}
	var requests []request
	for _, id := range ids {
		requests = append(requests, request{"authorised", id, ""})
	}
	return s.finish(ctx, p, o, requests)
}

// definedGate is the gate whose junction the authorisation reviews.
const definedGate = "defined"

// empty begins a plan of one of the three empty commits: it refuses no id,
// a repeated id and an id that names no task, then reads the project and
// each task.
func (r *Repo) empty(ctx context.Context, ids []string) (*session, Project, []Task, error) {
	s, err := r.begin(ctx)
	if err != nil {
		return nil, Project{}, nil, err
	}
	var refusals Refusals
	if len(ids) == 0 {
		refusals = append(refusals, Refusal{Code: "bad-field", Text: "name at least one task"})
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			refusals = append(refusals, Refusal{Code: "bad-field", Task: id, Text: "the task is named twice"})
		}
		seen[id] = true
	}
	refusals = append(refusals, s.checkIDs(ids...)...)
	if len(refusals) > 0 {
		return nil, Project{}, nil, refusals
	}
	project, err := s.src.Project(ctx)
	if err != nil {
		return nil, Project{}, nil, err
	}
	if len(project.Gates) == 0 {
		return nil, Project{}, nil, &ExecError{Tool: "tablo", Exit: -1, Stderr: "the project states no gate"}
	}
	tasks := make([]Task, len(ids))
	for i, id := range ids {
		if tasks[i], err = s.src.Task(ctx, id); err != nil {
			return nil, Project{}, nil, err
		}
		tasks[i].ID = id
	}
	return s, project, tasks, nil
}

// about names the tasks of a subject: "<title> (<id>)" for one, "tasks <id>,
// <id>" for two to four, "<n> tasks" for more.
func about(tasks []Task) string {
	switch n := len(tasks); {
	case n == 1:
		return lowerTitle(tasks[0].Title) + " (" + tasks[0].ID + ")"
	case n <= 4:
		ids := make([]string, n)
		for i, t := range tasks {
			ids[i] = t.ID
		}
		return "tasks " + strings.Join(ids, ", ")
	default:
		return strconv.Itoa(n) + " tasks"
	}
}

// nextJunction returns the task's first junction after a gate that applies,
// or nil at the last applicable gate.
func nextJunction(t *Task, gates []string, gate string) *Junction {
	i := slices.Index(gates, gate)
	if i < 0 {
		return nil
	}
	for _, g := range gates[i+1:] {
		if j := t.junction(g); j != nil && j.Kind != notApplicable {
			return j
		}
	}
	return nil
}

// notContributor warns when the hand holds neither the contributor of the
// junction a commit is at nor the task's assignee.
func (s *session) notContributor(t *Task, at *Junction) *Warning {
	contributor := ""
	if at != nil && at.Kind == plain {
		contributor = at.Contributor
	}
	if s.holds(contributor) || s.holds(t.Assignee) {
		return nil
	}
	text := "the hand that commits is not the assignee " + t.Assignee
	if contributor != "" && contributor != t.Assignee {
		text = "the hand that commits is neither the contributor " + contributor + " nor the assignee " + t.Assignee
	}
	return &Warning{Code: "not-contributor", Task: t.ID, Text: text}
}
