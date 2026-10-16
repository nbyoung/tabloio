package view

func init() { register("gates", func() View { return new(Gates) }) }

// Gates is the gate definition view.
type Gates struct {
	Head
	Task     *TaskRef     `json:"task"`     // set with the task parameter
	ForTask  []GateFor    `json:"for_task"` // detail, with a task: one per gate, undefined first
	Files    []FileCommit `json:"files"`    // provenance: gates.yaml, then version.yaml
	Commands []Command    `json:"commands"` // provenance
}

// GateFor says how one task's junction expands a gate.
type GateFor struct {
	Gate       string      `json:"gate"`
	Applies    bool        `json:"applies"`
	References []Reference `json:"references"`
	StatedBy   string      `json:"stated_by"` // the task whose entry states them; "" for none
}

// Reference is a reference of a task or a junction.
type Reference struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// FileCommit is a file with the newest commit that changes it.
type FileCommit struct {
	Path   string  `json:"path"`   // relative to the project directory
	Commit *Commit `json:"commit"` // nil without history, or when no commit changes it
}
