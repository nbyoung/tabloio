package view

import (
	"encoding/json"
	"fmt"
)

func init() {
	register("tableau", func() View { return new(Tableau) })
	register("context", func() View { return &Tableau{Contextual: true} })
}

// Tableau is the global and the contextual tableau; Decode("tableau") and
// Decode("context") both return it, with Contextual set for the second.
type Tableau struct {
	Head
	Contextual bool          `json:"contextual"`
	Subject    *TaskRef      `json:"subject"`  // the contextual tableau's task; nil for a person
	Columns    []Column      `json:"columns"`  // every gate once, in order, alone or in a folded run
	Rows       []GridRow     `json:"rows"`     // the rows of the level, in display order
	Statuses   []StatusCell  `json:"statuses"` // provenance from here
	Rollups    []Rollup      `json:"rollups"`
	MarkGroups []MarkGroup   `json:"mark_groups"`
	Snapshots  []SnapshotRow `json:"snapshots"`
	Commands   []Command     `json:"commands"`
}

// UnmarshalJSON reads the data of a tableau and checks that every row holds
// one cell per column, so that a malformed grid fails at the door with an
// error that names the row.
func (t *Tableau) UnmarshalJSON(data []byte) error {
	type plain Tableau         // no methods: avoids the recursion
	contextual := t.Contextual // set by Decode("context")
	if err := json.Unmarshal(data, (*plain)(t)); err != nil {
		return err
	}
	t.Contextual = t.Contextual || contextual
	return t.Check()
}

// Check reports the first row that does not hold one cell per column.
func (t *Tableau) Check() error {
	for _, r := range t.Rows {
		if len(r.Cells) != len(t.Columns) {
			return fmt.Errorf("view: tableau row %q holds %d cells for %d columns", r.ID, len(r.Cells), len(t.Columns))
		}
	}
	return nil
}

// Column is one column of the grid.
type Column struct {
	Gates  []string `json:"gates"` // one gate, or the run a folded column stands for
	Folded bool     `json:"folded"`
	Count  int      `json:"count"`   // folded: the tasks in view whose current gate lies in the run
	ByGate []int    `json:"by_gate"` // folded: the same count per gate of the run
}

// GridRow is one task.
type GridRow struct {
	TaskRef
	Depth  int        `json:"depth"`
	Parent bool       `json:"parent"`
	Hidden int        `json:"hidden"` // descendants the level does not draw
	Role   string     `json:"role"`   // contextual: "", spine or sibling
	Status Status     `json:"status"` // with its origin, for the notes
	Cells  []GridCell `json:"cells"`  // one per column
}

// GridCell is one junction as the tableau shows it.
type GridCell struct {
	Kind   string `json:"kind"`   // empty, status, marks, exempt
	State  string `json:"state"`  // status
	Reason string `json:"reason"` // status
	Marked        // marks; Acts encloses any kind of cell
}

// StatusCell is the commit that records a status at one gate, for the rows
// that share it.
type StatusCell struct {
	Tasks  []string `json:"tasks"` // the rows one commit records at one gate
	Gate   string   `json:"gate"`
	Commit Commit   `json:"commit"`
}

// Rollup is a parent's status and the child it takes it from.
type Rollup struct {
	Parent TaskRef `json:"parent"`
	Status Status  `json:"status"`
	Child  TaskRef `json:"child"` // the child it takes its status from
}

// MarkGroup is the junction that stands behind the marks of some rows at one
// gate, with the task that supplies each field.
type MarkGroup struct {
	Gate            string   `json:"gate"`
	Marks           []string `json:"marks"`
	Tasks           []string `json:"tasks"`
	Contributor     string   `json:"contributor"`
	Model           string   `json:"model"`
	Reviewer        string   `json:"reviewer"`
	ContributorFrom string   `json:"contributor_from"` // a task id, or "default"
	ReviewerFrom    string   `json:"reviewer_from"`
	ExemptFrom      string   `json:"exempt_from"`
}

// SnapshotRow is a row that a subproject's snapshot supplies.
//
// The design embeds SubRef here, but SubRef carries a field tagged "task"
// and so does the row, and encoding/json drops the inner one. The root task
// of the subproject therefore has a key of its own, "root".
type SnapshotRow struct {
	Task   string  `json:"task"` // the row
	URL    string  `json:"url"`  // the subproject
	Root   TaskRef `json:"root"` // the subproject's root task at the pin
	Pin    string  `json:"pin"`
	SetBy  Commit  `json:"set_by"`
	Status Status  `json:"status"` // of the subproject's task, with its origin
}
