package view

func init() { register("assignment", func() View { return new(Assignment) }) }

// Assignment is the task assignment view.
type Assignment struct {
	Head
	People    []PersonRow     `json:"people"`    // one per email in view, in order of first appearance in the tree
	Recursive []string        `json:"recursive"` // tasks whose next junction is recursive
	Sections  []PersonSection `json:"sections"`  // detail: one per row of People
	Commits   []Commit        `json:"commits"`   // provenance: every commit a Stated names, newest first
	Commands  []Command       `json:"commands"`
}

// PersonRow is the glance of one email.
type PersonRow struct {
	Email           string   `json:"email"`
	Assigned        int      `json:"assigned"`
	ContributesNext int      `json:"contributes_next"`
	ReviewsNext     int      `json:"reviews_next"`
	Models          []string `json:"models"`
}

// PersonSection is what one email carries.
type PersonSection struct {
	Email       string         `json:"email"`
	Assigned    []AssignedRow  `json:"assigned"`
	Recursive   []RecursiveRow `json:"recursive"`
	Contributes []JunctionRow  `json:"contributes"` // gate order, then display order
	Reviews     []JunctionRow  `json:"reviews"`
	Models      []ModelRow     `json:"models"` // an agent: its models by gate
	Counts      []GateCount    `json:"counts"` // every gate, in order
	Authority   []Subtree      `json:"authority"`
	Positions   []Position     `json:"positions"` // provenance
}

// AssignedRow is a task assigned to an email.
type AssignedRow struct {
	TaskRef
	Under  string  `json:"under"` // the parent's id
	Status Status  `json:"status"`
	Next   *Marked `json:"next"` // nil for a parent
	Gate   string  `json:"next_gate"`
}

// RecursiveRow is a task whose next junction a subproject carries.
type RecursiveRow struct {
	Task       TaskRef `json:"task"`
	Gate       string  `json:"gate"`
	Subproject SubRef  `json:"subproject"`
	Status     Status  `json:"status"` // of the subproject's task
}

// JunctionRow is one junction where an email contributes or reviews.
type JunctionRow struct {
	Task        TaskRef `json:"task"`
	Gate        string  `json:"gate"`
	Contributor string  `json:"contributor"`
	Model       string  `json:"model"`
	Reviewer    string  `json:"reviewer"`
	When        string  `json:"when"` // passed, next, later
}

// ModelRow is a model an agent runs at a gate.
type ModelRow struct {
	Gate      string `json:"gate"`
	Model     string `json:"model"`
	StatedBy  string `json:"stated_by"` // the id of the task whose entry states it
	Junctions int    `json:"junctions"`
	Next      int    `json:"next"`
}

// GateCount is what an email holds at one gate.
type GateCount struct {
	Gate            string `json:"gate"`
	Standing        int    `json:"standing"` // tasks assigned whose status stands at the gate
	Contributes     int    `json:"contributes"`
	ContributesNext int    `json:"contributes_next"`
	Reviews         int    `json:"reviews"`
	ReviewsNext     int    `json:"reviews_next"`
}

// Subtree is a task an email has authority over, with its descendants.
type Subtree struct {
	Task        TaskRef `json:"task"`
	Descendants int     `json:"descendants"`
}

// Position groups the tasks at which one source gives the email a position.
type Position struct {
	Kind     string   `json:"kind"` // assignee, authority, contributes, reviews, recursive
	Gate     string   `json:"gate"`
	Tasks    []string `json:"tasks"`
	Reviewer string   `json:"reviewer"`
	From     Stated   `json:"from"`          // the assignee, the authority, or the contributor and model
	RevFrom  Stated   `json:"reviewer_from"` // the reviewer
	URL      string   `json:"url"`           // recursive: the subproject
	Pin      string   `json:"pin"`           // recursive: the commit it is read at
}

// Stated says where a field stands.
type Stated struct {
	By     string `json:"by"`     // a task id, or "default"
	Key    string `json:"key"`    // for example junctions.release
	File   string `json:"file"`   // the file that holds it
	Commit string `json:"commit"` // the commit that last changed the line
}
