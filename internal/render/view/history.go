package view

func init() { register("history", func() View { return new(History) }) }

// History is the history view: what each commit does to each task in view,
// oldest first. The counts cover the project's own lines, at every level.
type History struct {
	Head
	Subject  TaskRef     `json:"subject"`  // the task whose subtree is in view: the task parameter, the root by default
	Events   int         `json:"events"`   // the events the project's own lines hold
	Commits  int         `json:"commits"`  // the distinct commits of those lines
	Kinds    []KindCount `json:"kinds"`    // task, authorised, status, reaffirmed, reviewed, pin
	Days     []Day       `json:"days"`     // one per author date of those lines
	Lines    []Line      `json:"lines"`    // oldest first; a subproject's lines from detail on
	Commands []Command   `json:"commands"` // provenance: the git log commands that give the list
}

// Day counts one day's events of the project's own lines.
type Day struct {
	Date    string      `json:"date"`
	Events  int         `json:"events"`
	Commits int         `json:"commits"`
	Kinds   []KindCount `json:"kinds"` // the kinds that occur, in the fixed order
}

// Line is what one commit does to one task: its events, each the shared
// Event (8ed1 decision 1). The line's kinds are its events', and its date,
// author and commit the first event's. A line taken from a subproject marks
// each event with the subproject (8ed1 decision 2).
type Line struct {
	Task     TaskRef `json:"task"`     // the task of the project in view
	Proposal bool    `json:"proposal"` // off the trunk: the trunk's tip does not reach the commit
	Events   []Event `json:"events"`   // never empty; in the order of the kinds

	After        *Status     `json:"after"`         // detail from here: the task's status at the commit
	Unchanged    bool        `json:"unchanged"`     // a parent's files state the same status
	SubjectAfter *Status     `json:"subject_after"` // the subject's status, on a line of another task
	Effect       string      `json:"effect"`        // one of the keys the renderer words, or empty
	Model        string      `json:"model"`         // the commit's first Model: trailer
	Check        *ModelCheck `json:"check"`         // the model check of the line's status or reaffirmed event
	Committer    string      `json:"committer"`     // only where it differs from the author

	Commit *Commit `json:"commit"` // provenance
}
