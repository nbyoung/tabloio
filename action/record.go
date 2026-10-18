package action

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

// Recording is a record request. An empty Gate, State, Reason or Note keeps
// the field as "Resolving a record" states.
type Recording struct {
	ID                        string
	Gate, State, Reason, Note string
	NoReason, NoNote          bool // remove the field
	Pin                       bool // also commit the moved pins of the task's recursive junctions
}

// Record plans the commit that writes status/<id>.yaml, moves a pin, or both.
func (r *Repo) Record(ctx context.Context, rec Recording, o Options) (*Plan, error) {
	s, err := r.begin(ctx)
	if err != nil {
		return nil, err
	}
	if refusals := s.checkIDs(rec.ID); len(refusals) > 0 {
		return nil, refusals
	}
	project, err := s.src.Project(ctx)
	if err != nil {
		return nil, err
	}
	if len(project.Gates) == 0 {
		return nil, &ExecError{Tool: "tablo", Exit: -1, Stderr: "the project states no gate"}
	}
	t, err := s.src.Task(ctx, rec.ID)
	if err != nil {
		return nil, err
	}
	t.ID = rec.ID
	res, err := s.resolve(ctx, project, &t, rec)
	if err != nil {
		return nil, err
	}

	p := &Plan{Command: "record", Tasks: []string{rec.ID}, Pins: res.pins}
	if len(res.edits) > 0 {
		path := s.statusPath(rec.ID)
		old, err := s.headFile(ctx, path)
		if err != nil {
			return nil, err
		}
		text, err := editStatus(string(old), res.edits)
		var form *statusFormError
		if errors.As(err, &form) {
			return nil, Refusals{{Code: "status-form", Task: rec.ID, Text: path + ": " + form.text}}
		}
		if text != string(old) {
			p.Files = []FileChange{{Path: path, Old: old, New: []byte(text)}}
		}
	}
	if len(p.Files) == 0 && len(res.pins) == 0 && len(res.reviews) == 0 {
		return nil, Refusals{{Code: "no-change", Task: rec.ID, Gate: res.gate, Text: "the status stands as recorded; reaffirm it to confirm it"}}
	}

	named := lowerTitle(t.Title) + " (" + rec.ID + ")"
	_, gateChanges := res.edits["gate"]
	switch {
	case gateChanges:
		p.Subject = "Record " + named + " at " + res.gate
	case res.handOff:
		p.Subject = "Hand " + named + " to review"
	case len(p.Files) == 0 && len(res.pins) > 0:
		paths := make([]string, len(res.pins))
		for i, pin := range res.pins {
			paths[i] = pin.Path
		}
		noun := "pin"
		if len(paths) > 1 {
			noun = "pins"
		}
		p.Subject = "Advance the " + noun + " of " + strings.Join(paths, ", ") + " for " + named
	default:
		p.Subject = "Update the status of " + named
	}
	p.Warnings = s.warnings(&t, res.at, o.Agent)
	return s.finish(ctx, p, o, res.reviews)
}

// resolution is a record request resolved against the task's facts.
type resolution struct {
	gate    string
	edits   map[string]*string // the fields that change: a new value, or nil to remove
	handOff bool               // the reason becomes review
	reviews []request          // the Reviewed: trailers the hand owes and holds
	pins    []PinChange
	at      *Junction // the junction the commit is at (decision 8)
}

// resolve follows "Resolving a record" of design b618, steps 1 to 8.
func (s *session) resolve(ctx context.Context, project Project, t *Task, rec Recording) (*resolution, error) {
	refuse := func(code, gate, rule, text string) Refusals {
		return Refusals{{Code: code, Task: t.ID, Gate: gate, Rule: rule, Text: text}}
	}
	gates := project.Gates
	if len(t.Children) > 0 {
		return nil, refuse("not-leaf", "", "S2", "a parent has no status file; its status derives from its children")
	}
	var refusals Refusals
	add := func(code, gate, rule, text string) { refusals = append(refusals, refuse(code, gate, rule, text)...) }
	if rec.Reason != "" && rec.NoReason {
		add("bad-field", "", "", "a reason and no reason at once")
	}
	if rec.Note != "" && rec.NoNote {
		add("bad-field", "", "", "a note and no note at once")
	}
	if rec.Note != "" && (strings.TrimSpace(rec.Note) == "" || strings.ContainsFunc(rec.Note, unicode.IsControl)) {
		add("bad-field", "", "", "the note is empty or holds a control character or a line break")
	}

	old := Status{Gate: gates[0], State: "undefined"}
	if t.Status != nil {
		old = *t.Status
	}
	gate := rec.Gate
	if gate == "" {
		gate = old.Gate
	}
	gi := slices.Index(gates, gate)
	if gi < 0 {
		return nil, append(refusals, refuse("unknown-gate", gate, "S5", "names no gate of gates.yaml")...)
	}
	j := t.junction(gate)
	if gi > 0 && (j == nil || j.Kind == notApplicable) {
		return nil, append(refusals, refuse("gate-not-applicable", gate, "S7", "the gate does not apply to the task")...)
	}
	changes := gate != old.Gate
	next := nextJunction(t, gates, gate)

	// Step 3: the state, or a file that holds the gate alone.
	var state, reason, note string
	holdsGate := false
	switch {
	case gi == 0:
		if rec.State != "" && rec.State != "undefined" {
			add("state-forbidden", gate, "S4", "the gate "+gate+" takes the state undefined alone")
		}
		state = "undefined"
	case next == nil:
		if rec.State != "" && rec.State != "complete" {
			add("last-gate-not-complete", gate, "S8", "the last applicable gate takes the state complete")
		}
		state = "complete"
	case next.Kind == recursive:
		if rec.State != "" || rec.Reason != "" || rec.Note != "" {
			add("recursive-holds-gate", gate, "S9", "the next junction, at "+next.Gate+", is recursive, so the file holds the gate alone")
		}
		holdsGate = true
	default:
		state = rec.State
		if state == "" && !changes {
			state = old.State
		}
		switch {
		case state == "":
			add("state-required", gate, "S10", "the next junction, at "+next.Gate+", is plain, so the status needs a state")
		case state == "complete":
			add("complete-early", gate, "S12", "the state complete stands at the last applicable gate alone")
		case state == "undefined":
			add("state-forbidden", gate, "S4", "the state undefined stands at the gate "+gates[0]+" alone")
		case !slices.Contains(project.States, state):
			add("unknown-state", gate, "S6", "names no state of gates.yaml: "+state)
		}
	}

	// Steps 4 and 5: the reason and the note.
	if !holdsGate {
		reason = rec.Reason
		if !rec.NoReason && reason == "" && !changes {
			reason = old.Reason
		}
		switch {
		case reason != "" && !slices.Contains(project.Reasons, reason):
			add("unknown-reason", gate, "S6", "names no reason of gates.yaml: "+reason)
		case reason == "review" && t.reviewer(next) == "":
			add("no-reviewer", gate, "", "the reason review needs a reviewer at the next junction, and none stands there")
		}
		note = rec.Note
		if !rec.NoNote && note == "" && !changes {
			note = old.Note
		}
	}

	res := &resolution{gate: gate, edits: map[string]*string{}, at: next}
	if changes {
		res.at = j
	}

	// Step 6: the reviews the gates passed owe.
	from := max(slices.Index(gates, old.Gate), 0)
	for _, g := range gates[from+1 : max(gi+1, from+1)] {
		pj := t.junction(g)
		reviewer := t.reviewer(pj)
		if g == definedGate || reviewer == "" || t.reviewed(g) {
			continue
		}
		if s.holds(reviewer) {
			res.reviews = append(res.reviews, request{"reviewed", t.ID, g})
			continue
		}
		add("unreviewed", g, "S11", fmt.Sprintf("the reviewer %s has not accepted %s; hand the work off with the reason review", reviewer, g))
	}

	// Step 7: the pins.
	if rec.Pin {
		pins, err := s.pins(ctx, t)
		if err != nil {
			return nil, err
		}
		refusals = append(refusals, pins.refusals...)
		res.pins = pins.moved
	}
	if len(refusals) > 0 {
		return nil, refusals
	}

	// Step 8: the fields that change.
	set := func(key, was, now string) {
		if was == now {
			return
		}
		if now == "" {
			res.edits[key] = nil
		} else {
			v := now
			res.edits[key] = &v
		}
	}
	set("gate", old.Gate, gate)
	set("state", old.State, state)
	set("reason", old.Reason, reason)
	set("note", old.Note, note)
	if holdsGate {
		// tablo gives a status before a recursive junction as the
		// subproject's snapshot, the gate alone, whatever else the file
		// holds; the file holds the gate alone, so the edit removes the
		// other fields wherever they stand, and Record finds whether the
		// bytes change.
		for _, key := range []string{"state", "reason", "note"} {
			if _, ok := res.edits[key]; !ok {
				res.edits[key] = nil
			}
		}
	}
	res.handOff = reason == "review" && old.Reason != "review"
	if len(res.edits) == 0 && len(res.pins) == 0 && len(res.reviews) == 0 {
		return nil, refuse("no-change", gate, "", "the status stands as recorded; reaffirm it to confirm it")
	}
	return res, nil
}

// pinResult is what step 7 finds: the moved pins and the refusals.
type pinResult struct {
	moved    []PinChange
	refusals Refusals
}

// pins compares the gitlink at HEAD of each submodule the task's recursive
// junctions read with the commit its checkout stands at.
func (s *session) pins(ctx context.Context, t *Task) (pinResult, error) {
	var paths []string
	for _, j := range t.Junctions {
		if j.Kind == recursive && j.Submodule && !slices.Contains(paths, j.URL) {
			paths = append(paths, j.URL)
		}
	}
	if len(paths) == 0 {
		return pinResult{refusals: Refusals{{Code: "no-submodule", Task: t.ID, Text: "no recursive junction of the task reads a submodule"}}}, nil
	}
	var res pinResult
	for _, p := range paths {
		listed, err := s.git(ctx, "ls-tree", "HEAD", "--", p)
		if err != nil {
			return res, err
		}
		fields := strings.Fields(listed)
		if len(fields) < 3 || fields[0] != "160000" {
			return res, &ExecError{Tool: "git", Args: []string{"ls-tree", "HEAD", "--", p}, Exit: 0, Stderr: "HEAD holds no gitlink at " + p}
		}
		out, err := s.r.git(ctx, filepath.Join(s.top, filepath.FromSlash(p)), nil, nil, "rev-parse", "HEAD")
		if err != nil {
			return res, err
		}
		if at := strings.TrimSpace(out); at != fields[2] {
			res.moved = append(res.moved, PinChange{Path: p, Old: fields[2], New: at})
		}
	}
	if len(res.moved) == 0 {
		res.refusals = Refusals{{Code: "pin-unchanged", Task: t.ID, Text: "each submodule stands at its pin: " + strings.Join(paths, ", ")}}
	}
	return res, nil
}

// warnings returns the findings of "Warnings" for a commit at a junction:
// not-contributor, model-mismatch (H3) and model-missing (H6).
func (s *session) warnings(t *Task, at *Junction, a Agent) []Warning {
	var out []Warning
	if w := s.notContributor(t, at); w != nil {
		out = append(out, *w)
	}
	if at == nil || at.Kind != plain || at.Model == "" {
		return out
	}
	switch {
	case a.Model != "" && !strings.HasPrefix(a.Model, at.Model):
		out = append(out, Warning{Code: "model-mismatch", Task: t.ID,
			Text: fmt.Sprintf("the junction at %s states the model %s, and Model: names %s", at.Gate, at.Model, a.Model)})
	case a.Model == "" && s.holds(at.Contributor):
		out = append(out, Warning{Code: "model-missing", Task: t.ID,
			Text: fmt.Sprintf("the junction at %s states the model %s, and the commit carries no Model:", at.Gate, at.Model)})
	}
	return out
}

// statusPath returns a task's status file from the repository's root.
func (s *session) statusPath(id string) string { return s.project + "/status/" + id + ".yaml" }

// headFile returns a file's bytes at HEAD, or nil when HEAD lacks it.
func (s *session) headFile(ctx context.Context, p string) ([]byte, error) {
	listed, err := s.git(ctx, "ls-tree", "--name-only", "HEAD", "--", p)
	if err != nil || listed == "" {
		return nil, err
	}
	out, err := s.r.git(ctx, s.top, nil, nil, "cat-file", "blob", "HEAD:"+p)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// reviewed reports whether a review that takes effect accepts the gate.
func (t *Task) reviewed(gate string) bool {
	for _, r := range t.Reviews {
		if r.Gate == gate {
			return true
		}
	}
	return false
}
