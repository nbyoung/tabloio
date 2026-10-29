package action

import "context"

// Source supplies the facts a plan needs and the two services tablo owns.
// Tablo implements it through tablo's library; a test supplies fixed answers.
type Source interface {
	// Project returns the gate, state and reason keys in file order, and the
	// trunk's branch name, "" when undetermined.
	Project(ctx context.Context) (Project, error)
	// Task returns the derived facts of one task that exists.
	Task(ctx context.Context, id string) (Task, error)
	// Trailer returns one task trailer line without its newline; kind is
	// "authorised", "reviewed" or "reaffirmed", and gate is "" unless reviewed.
	Trailer(ctx context.Context, kind, id, gate string) (string, error)
	// Validate returns the validator's diagnostics for a tree.
	Validate(ctx context.Context, tree string) ([]Diagnostic, error)
}

// Project is what a plan reads of gates.yaml and the trunk.
type Project struct {
	Gates, States, Reasons []string
	Trunk                  string
}

// Task is the derived facts of one task.
type Task struct {
	ID, Title, Assignee string
	Children            []Child    // in display order; empty for a leaf
	Authorities         []string   // nearest first; the root lists its own assignee, the owner
	Junctions           []Junction // one per gate after the first, in gate order
	Status              *Status    // the recorded status; nil when no file exists
	Reviews             []Review   // the effective reviews in the branch's history
}

// Child is one child of a task and its place.
type Child struct {
	ID    string
	Order int // 0 when the child states none
}

// The kinds of a junction.
const (
	plain         = "plain"
	recursive     = "recursive"
	notApplicable = "not-applicable"
)

// Junction is one junction of a task, resolved through its ancestors.
type Junction struct {
	Gate                         string
	Kind                         string // "plain", "recursive" or "not-applicable"
	Contributor, Model, Reviewer string // resolved; plain only
	URL                          string // recursive only: the subproject, a path from the repository's root for a submodule
	Submodule                    bool   // recursive only: URL is a submodule path
}

// Status is a status file as tablo reads it; "" for a field the file lacks.
type Status struct{ Gate, State, Reason, Note string }

// Review is a Reviewed: commit that takes effect: the gate it accepts and
// the reviewer's email on it.
type Review struct{ Gate, By string }

// Diagnostic is one finding of tablo's validator.
type Diagnostic struct{ Severity, Code, Path, Message string }

// junction returns the task's junction at a gate, or nil when the task has
// none there: at the first gate, or at a gate the project lacks.
func (t *Task) junction(gate string) *Junction {
	for i := range t.Junctions {
		if t.Junctions[i].Gate == gate {
			return &t.Junctions[i]
		}
	}
	return nil
}

// reviewer returns a plain junction's reviewer as tablo's task view resolves
// it: the one stated, the default, or "" for none, which a waiver
// (reviewer: false) gives. It supplies no default of its own.
func (t *Task) reviewer(j *Junction) string {
	if j == nil || j.Kind != plain {
		return ""
	}
	return j.Reviewer
}
