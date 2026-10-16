package view

func init() { register("blockage", func() View { return new(Blockage) }) }

// Blockage is the work-blockage tree.
type Blockage struct {
	Head
	Causes      []Cause     `json:"causes"`        // largest first, then display order
	Kinds       []KindCount `json:"kinds"`         // the five kinds, in VIEWS.md's order
	NotDue      []Waiting   `json:"not_due"`       // detail: by task, in display order
	NotDueN     int         `json:"not_due_count"` // the requirements not yet due
	NotDueUnmet int         `json:"not_due_unmet"` // those of them not met
	Commands    []Command   `json:"commands"`      // provenance: what reproduces the facts
}

// KindCount is a count of one kind of thing: a cause of the tree, or an
// event of the history.
type KindCount struct {
	Kind  string `json:"kind"` // requirement, status, review, authorisation, snapshot
	Count int    `json:"count"`
}

// Cause is one root of the tree.
type Cause struct {
	Kind     string  `json:"kind"`
	Task     TaskRef `json:"task"`     // the task the cause stands at
	URL      string  `json:"url"`      // a cross-project requirement: the project of Task; a snapshot: the link
	Gate     string  `json:"gate"`     // requirement: the gate not passed; review: the gate under review; else the gate the status names
	Status   Status  `json:"status"`   // of Task
	Pin      string  `json:"pin"`      // snapshot: the commit the link reads
	Tip      string  `json:"tip"`      // snapshot: the tip of the subproject's trunk
	Resolver string  `json:"resolver"` // the person who acts; "" where the files name none
	Mark     string  `json:"mark"`     // the resolver's mark key, or empty
	Act      string  `json:"act"`      // contributes, records, reviews, authorises, advances
	Advance  string  `json:"advance"`  // snapshot: what advances, pin or gate
	Holds    int     `json:"holds"`    // distinct tasks beneath, at every depth
	Held     []Held  `json:"held"`     // detail

	StatusCommit  *Commit        `json:"status_commit"` // provenance from here: the deciding commit of Status
	Junction      *Junction      `json:"junction"`      // the next junction of Task
	Authorisation *Authorisation `json:"authorisation"` // authorisation: the task's
	Linkage       *Linkage       `json:"linkage"`       // snapshot: the link
	Resolve       []Command      `json:"resolve"`       // review, authorisation: the commit that resolves
}

// Held is a task a cause holds, and what it holds in turn.
type Held struct {
	Task     TaskRef      `json:"task"`
	Gate     string       `json:"gate"`
	Parent   bool         `json:"parent"`
	Via      string       `json:"via"`      // self, parent, requirement
	Child    string       `json:"child"`    // parent: the child that holds it
	Requires *Requirement `json:"requires"` // requirement: the entry
	Also     []int        `json:"also"`     // the other causes it stands under, counted from 1
	Held     []Held       `json:"held"`
}

// Waiting is a task with requirements not yet due.
type Waiting struct {
	Task     TaskRef       `json:"task"`
	Status   Status        `json:"status"`
	Next     string        `json:"next"`
	Requires []Requirement `json:"requires"`
}
