package view

func init() { register("history", func() View { return new(History) }) }

// History is the history view.
type History struct {
	Head
	Subject  *TaskRef    `json:"subject"` // the task parameter; nil for the whole project
	Lines    []Line      `json:"lines"`   // oldest first
	Events   int         `json:"events"`
	Commits  int         `json:"commits"`
	Kinds    []KindCount `json:"kinds"`    // task, authorised, status, reaffirmed, reviewed, pin
	Days     []Day       `json:"days"`     // one per day that holds a line
	Commands []Command   `json:"commands"` // provenance: the git log commands that give the list
}

// Day counts one day's events.
type Day struct {
	Date    string      `json:"date"`
	Events  int         `json:"events"`
	Commits int         `json:"commits"`
	Kinds   []KindCount `json:"kinds"` // the kinds that occur, in the fixed order
}

// Line is what one commit does to one task.
type Line struct {
	Date     string   `json:"date"`
	By       string   `json:"by"`
	Kinds    []string `json:"kinds"`
	Task     TaskRef  `json:"task"`
	Proposal bool     `json:"proposal"` // off the trunk: the branch adds the event
	Status   *Status  `json:"status"`   // status, reaffirmed: what the event records
	Gate     string   `json:"gate"`     // reviewed
	Pin      *PinMove `json:"pin"`      // pin; Old is empty when the linkage first appears

	After        *Status    `json:"after"` // detail from here: the task's status after the event
	Unchanged    bool       `json:"unchanged"`
	SubjectAfter *Status    `json:"subject_after"` // a subtree: the subject's status after the event
	Effect       string     `json:"effect"`        // one of the keys the renderer words, or empty
	Model        string     `json:"model"`         // the Model: trailer
	Stated       string     `json:"stated"`        // the model the junction states
	StatedGate   string     `json:"stated_gate"`
	Reading      string     `json:"reading"`    // agrees, differs, missing, exempt, or empty
	Committer    string     `json:"committer"`  // only where it differs from By
	Sub          *SubEvents `json:"subproject"` // pin: the subproject's own events between the commits

	Commit   *Commit   `json:"commit"`   // provenance
	Commands []Command `json:"commands"` // provenance: what reproduces this line
}

// SubEvents is a subproject's history between two pins.
type SubEvents struct {
	URL     string `json:"url"`
	Events  int    `json:"events"`
	Commits int    `json:"commits"`
	Lines   []Line `json:"lines"` // glance fields, and Model; empty when tablo folds them to the count
}
