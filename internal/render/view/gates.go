package view

func init() { register("gates", func() View { return new(Gates) }) }

// Gates is the gate definition view.
type Gates struct {
	Head
	Task     *TaskRef     `json:"task"`     // set with the task parameter
	ForTask  []GateFor    `json:"for_task"` // detail, with a task: one per gate, in order
	Language string       `json:"language"` // provenance: the language version
	Trunk    string       `json:"trunk"`    // provenance
	Files    []FileCommit `json:"files"`    // provenance: gates.yaml and version.yaml
	Commands []Command    `json:"commands"` // provenance
}

// GateFor says how one task's junction expands a gate.
type GateFor struct {
	Gate       string      `json:"gate"`
	Applies    bool        `json:"applies"`
	References []Reference `json:"references"`
	StatedBy   string      `json:"stated_by"` // the task whose entry states them
}

// Reference is a reference of a task or a junction.
type Reference struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// FileCommit is a file with the commit that last changed it.
type FileCommit struct {
	Path   string `json:"path"`
	Commit Commit `json:"commit"`
}
