package action

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/nbyoung/tablo"
	"github.com/nbyoung/tablo/view"
)

// Tablo is the Source that calls tablo's library in process, as the owner
// ruled at review on 2026-10-06 (design b618, decision 2). It is the one file
// of the package that imports tablo. It loads HEAD once and answers every
// question of a plan from that snapshot, so a Tablo serves one plan: Repo
// makes a new one for each when its Source is nil. The library's git
// processes inherit the process environment; Repo.Env does not reach them.
type Tablo struct {
	Dir string // the directory -C names; "" is the working directory

	mu   sync.Mutex
	repo *tablo.Repository
	head *tablo.Snapshot
}

// Project reads the gates view: the gate, state and reason keys of the
// legend in file order, and the trunk the source names.
func (t *Tablo) Project(ctx context.Context) (Project, error) {
	var env struct {
		Source struct {
			Trunk struct{ Name *string }
		}
		Data view.Gates
	}
	if err := t.view(ctx, &env, tablo.Gates, tablo.Params{Level: tablo.LevelGlance}); err != nil {
		return Project{}, err
	}
	var p Project
	for _, g := range env.Data.Legend.Gates {
		p.Gates = append(p.Gates, g.Key)
	}
	for _, s := range env.Data.Legend.States {
		p.States = append(p.States, s.Key)
	}
	for _, r := range env.Data.Legend.Reasons {
		p.Reasons = append(p.Reasons, r.Key)
	}
	if name := env.Source.Trunk.Name; name != nil {
		p.Trunk = *name
	}
	return p, nil
}

// Task reads the task view at provenance for the definition, the children,
// the authorities, the junctions and the status, and the history view of the
// task for its reviews.
func (t *Tablo) Task(ctx context.Context, id string) (Task, error) {
	var env struct{ Data view.Task }
	if err := t.view(ctx, &env, tablo.Task, tablo.Params{Task: id, Level: tablo.LevelProvenance}); err != nil {
		return Task{}, err
	}
	v := env.Data
	task := Task{ID: id, Title: v.Task.Title, Assignee: v.Assignee}
	switch v.Status.Kind {
	case "recorded":
		task.Status = &Status{Gate: v.Status.Gate, State: v.Status.State, Reason: v.Status.Reason, Note: v.Status.Note}
	case "snapshot":
		// The state comes from the subproject; the file holds the gate.
		task.Status = &Status{Gate: v.Status.Gate}
	}
	if d := v.TaskDetail; d != nil {
		for _, c := range d.Children {
			child := Child{ID: c.ID}
			if c.Order != nil {
				child.Order = *c.Order
			}
			task.Children = append(task.Children, child)
		}
		for _, a := range d.Authorities {
			task.Authorities = append(task.Authorities, a.Email)
		}
		for _, j := range d.Junctions {
			task.Junctions = append(task.Junctions, junctionOf(j))
		}
	}
	reviews, err := t.reviews(ctx, &task)
	if err != nil {
		return Task{}, err
	}
	task.Reviews = reviews
	return task, nil
}

// junctionOf spells a junction of the view as Source does: the kind
// not_applicable as "not-applicable", and a submodule as its path.
func junctionOf(j view.Junction) Junction {
	out := Junction{Gate: j.Gate, Kind: strings.ReplaceAll(j.Kind, "_", "-")}
	switch out.Kind {
	case plain:
		out.Contributor, out.Model, out.Reviewer = j.Contributor, j.Model, j.Reviewer
	case recursive:
		if j.Subproject != nil {
			out.URL, out.Submodule = j.Subproject.URL, j.Subproject.Form == "submodule"
		}
	}
	return out
}

// reviews reads the Reviewed: events of the task's own lines in the history
// of HEAD and keeps those that take effect: the one tablo's line reads as
// accepting, and any whose author or committer is the junction's reviewer.
func (t *Tablo) reviews(ctx context.Context, task *Task) ([]Review, error) {
	var env struct{ Data view.History }
	if err := t.view(ctx, &env, tablo.History, tablo.Params{Task: task.ID, Level: tablo.LevelProvenance}); err != nil {
		return nil, err
	}
	var out []Review
	for _, l := range env.Data.Lines {
		if l.Task.ID != task.ID || len(l.Events) == 0 || l.Events[0].Subproject != nil {
			continue
		}
		first := true
		for _, e := range l.Events {
			if e.Kind != "reviewed" {
				continue
			}
			reviewer := task.reviewer(task.junction(e.Gate))
			by := ""
			switch {
			case l.LineProvenance != nil && l.Commit != nil && reviewer != "" &&
				(l.Commit.Author.Email == reviewer || l.Commit.Committer.Email == reviewer):
				by = reviewer
			case first && l.LineDetail != nil && l.Effect == "accepts":
				by = e.By
			}
			if by != "" {
				out = append(out, Review{Gate: e.Gate, By: by})
			}
			first = false
		}
	}
	return out, nil
}

// Trailer composes one task trailer through tablo.Trailer.Line.
func (t *Tablo) Trailer(_ context.Context, kind, id, gate string) (string, error) {
	line, err := tablo.Trailer{Kind: kind, Task: id, Gate: gate}.Line()
	if err != nil {
		return "", &ExecError{Tool: "tablo", Args: []string{"trailer", kind, id, gate}, Exit: -1, Stderr: err.Error()}
	}
	return line, nil
}

// Validate loads a bare tree and returns the snapshot's diagnostics.
func (t *Tablo) Validate(ctx context.Context, tree string) ([]Diagnostic, error) {
	repo, err := t.open(ctx)
	if err != nil {
		return nil, err
	}
	snap, err := repo.Load(ctx, tree)
	if err != nil {
		return nil, &ExecError{Tool: "tablo", Args: []string{"--ref", tree, "validate"}, Exit: -1, Stderr: err.Error()}
	}
	var out []Diagnostic
	for _, d := range snap.Diagnostics() {
		out = append(out, Diagnostic{Severity: d.Severity, Code: d.Code, Path: d.Path, Message: d.Message})
	}
	return out, nil
}

// open opens the repository once.
func (t *Tablo) open(ctx context.Context) (*tablo.Repository, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.repo != nil {
		return t.repo, nil
	}
	dir := t.Dir
	if dir == "" {
		dir = "."
	}
	repo, err := tablo.Open(ctx, dir, tablo.Options{})
	if err != nil {
		return nil, &ExecError{Tool: "tablo", Args: []string{"-C", dir}, Exit: -1, Stderr: err.Error()}
	}
	t.repo = repo
	return repo, nil
}

// view derives one view of HEAD and decodes its envelope into out. An
// envelope with no data, which an error diagnostic of what tablo cannot read
// leaves, is an *ExecError that quotes the diagnostics.
func (t *Tablo) view(ctx context.Context, out any, v tablo.View, p tablo.Params) error {
	args := []string{"--ref", "HEAD", string(v)}
	if p.Task != "" && v == tablo.Task {
		args = append(args, p.Task)
	} else if p.Task != "" {
		args = append(args, "--task", p.Task)
	}
	args = append(args, "--level", string(p.Level))
	fail := func(text string) error { return &ExecError{Tool: "tablo", Args: args, Exit: -1, Stderr: text} }

	repo, err := t.open(ctx)
	if err != nil {
		return err
	}
	t.mu.Lock()
	snap := t.head
	t.mu.Unlock()
	if snap == nil {
		if snap, err = repo.Load(ctx, "HEAD"); err != nil {
			return fail(err.Error())
		}
		t.mu.Lock()
		t.head = snap
		t.mu.Unlock()
	}
	env, err := snap.View(ctx, v, p)
	if err != nil {
		return fail(err.Error())
	}
	if env.Data == nil {
		var texts []string
		for _, d := range env.Diagnostics {
			if d.Severity == "error" {
				texts = append(texts, d.Code+" "+d.Path+": "+d.Message)
			}
		}
		return fail("no view data: " + strings.Join(texts, "; "))
	}
	if err := json.Unmarshal(env.JSON(), out); err != nil {
		return fail(err.Error())
	}
	return nil
}
