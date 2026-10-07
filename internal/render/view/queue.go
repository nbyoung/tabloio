package view

import "encoding/json"

func init() { register("queue", func() View { return new(Queue) }) }

// Queue is the contributor work queue of one person.
type Queue struct {
	Head
	Items []QueueItem     `json:"items"` // kind order, then display order; reaffirmations oldest first
	Brief json.RawMessage `json:"brief"` // set with the brief parameter; task e4c7 decodes it
}

// QueueItem is one thing the person does.
type QueueItem struct {
	Kind  string  `json:"kind"` // review, authorisation, ready, reaffirmation, waiting
	Task  TaskRef `json:"task"`
	Gate  string  `json:"gate"` // the next gate; a reaffirmation: the gate the status names
	Model string  `json:"model"`
	Since string  `json:"since"` // the date of the status
	Age   int     `json:"age"`   // reaffirmation: days from the status to the ref's commit
	Cause string  `json:"cause"` // waiting, authorisation: worded text (R17)

	Junction       *Junction     `json:"junction"` // detail from here
	Sources        []FieldSource `json:"sources"`  // who supplies contributor, model, reviewer
	TaskReferences []Reference   `json:"task_references"`
	Requires       []Requirement `json:"requires"`
	Unblocks       []Requirement `json:"unblocks"`         // dependents this junction's gate meets
	AlsoRequiredBy []Requirement `json:"also_required_by"` // the other dependents
	ParentUnblocks []ParentEdge  `json:"parent_unblocks"`
	Status         *Status       `json:"status"`
	StatusCommit   *Commit       `json:"status_commit"`
}

// ParentEdge is a requirement on a parent that a child's work would let pass.
type ParentEdge struct {
	Parent   TaskRef     `json:"parent"`
	Gate     string      `json:"gate"`
	Requires Requirement `json:"requires"` // Task is the dependent
}
