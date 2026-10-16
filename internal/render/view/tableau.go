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
	Contextual bool      `json:"-"`
	Subject    *TaskRef  `json:"subject"`  // the contextual tableau's task; nil for a person and for the global tableau
	Columns    []Column  `json:"columns"`  // every gate once, in order, alone or in a folded run
	Rows       []GridRow `json:"rows"`     // the rows of the level, in display order
	Commands   []Command `json:"commands"` // provenance
}

// UnmarshalJSON reads the data of a tableau and checks that every row holds
// one cell per gate of the legend, so that a malformed grid fails at the
// door with an error that names the row.
func (t *Tableau) UnmarshalJSON(data []byte) error {
	type plain Tableau         // no methods: avoids the recursion
	contextual := t.Contextual // set by Decode("context")
	if err := json.Unmarshal(data, (*plain)(t)); err != nil {
		return err
	}
	t.Contextual = contextual
	return t.Check()
}

// Check reports the first row that does not hold one cell per gate of the
// legend.
func (t *Tableau) Check() error {
	for _, r := range t.Rows {
		if len(r.Cells) != len(t.Legend.Gates) {
			return fmt.Errorf("view: tableau row %q holds %d cells for %d gates", r.ID, len(r.Cells), len(t.Legend.Gates))
		}
	}
	return nil
}

// Column is one column of the grid: a gate, or the run of gates a fold
// stands for.
type Column struct {
	Gates  []string `json:"gates"` // one gate, or the run a folded column stands for
	Folded bool     `json:"folded"`
	Count  int      `json:"count"`   // the rows of the detail level whose status names a gate of the column
	ByGate []int    `json:"by_gate"` // the same count per gate of the column
}

// GridRow is one task.
type GridRow struct {
	TaskRef
	Depth  int        `json:"depth"`
	Parent bool       `json:"parent"`
	Hidden int        `json:"hidden"` // the tasks beneath it with no row, whose nearest ancestor with a row it is
	Label  string     `json:"label"`  // contextual: "", spine or sibling
	Status Status     `json:"status"` // with its origin, for the notes
	Cells  []GridCell `json:"cells"`  // one per gate of the legend, whatever the columns

	// Provenance: the fields of the task definition, for this row's task.
	StatusCommit *Commit    `json:"status_commit"` // the deciding commit of the last status of the chain
	OwnCommit    *Commit    `json:"own_commit"`    // a snapshot: the deciding commit of the task's own file
	Chain        []Step     `json:"chain"`         // the statuses the status comes through
	Junctions    []Junction `json:"junctions"`     // every gate after undefined, in order
	Sources      []Supplier `json:"sources"`       // of those junctions, by gate, then field
	Linkages     []Linkage  `json:"linkages"`      // one per subproject a recursive junction reads
}

// GridCell is one task at one gate.
type GridCell struct {
	Kind   string `json:"kind"`   // empty, status, marks, exempt
	State  string `json:"state"`  // status
	Reason string `json:"reason"` // status
	Marked        // marks; Acts encloses any kind of cell
}
