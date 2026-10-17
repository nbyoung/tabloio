package view

import "encoding/json"

func init() { register("queue", func() View { return new(Queue) }) }

// Queue is the contributor work queue of one person.
type Queue struct {
	Head
	Items []QueueItem     `json:"items"` // kind order, then display order; reaffirmations oldest first
	Brief json.RawMessage `json:"brief"` // provenance, with the brief parameter; package brief decodes it from the envelope, and no renderer reads it
}

// QueueItem is one thing the person does.
type QueueItem struct {
	Kind   string  `json:"kind"` // review, authorisation, ready, reaffirmation, waiting
	Task   TaskRef `json:"task"`
	Gate   string  `json:"gate"` // the next gate; an authorisation, a reaffirmation: the gate the status names
	Model  string  `json:"model"`
	Since  string  `json:"since"`  // the date of the status
	Age    int     `json:"age"`    // reaffirmation: days from the status to the ref's commit
	Causes []Wait  `json:"causes"` // waiting: why, in the order requirement, reason, authorisation, review
	Why    string  `json:"why"`    // authorisation: the task's Authorisation.Why, "" where a commit decides but not by an authority; else ""

	Junction       *Junction      `json:"junction"` // detail from here
	Sources        []Supplier     `json:"sources"`  // who supplies contributor, model, reviewer
	TaskReferences []Reference    `json:"task_references"`
	Requires       []Requirement  `json:"requires"`
	Unblocks       []Requirement  `json:"unblocks"`         // dependents this junction's gate meets
	AlsoRequiredBy []Requirement  `json:"also_required_by"` // the other dependents
	ParentUnblocks []ParentEdge   `json:"parent_unblocks"`
	Status         *Status        `json:"status"`
	StatusCommit   *Commit        `json:"status_commit"`
	Authorisation  *Authorisation `json:"authorisation"` // an authorisation: the task's
}

// Wait is one reason a work item waits.
type Wait struct {
	Kind     string       `json:"kind"`     // requirement, reason, authorisation, review
	Requires *Requirement `json:"requires"` // requirement: the unmet entry
	Reason   string       `json:"reason"`   // reason: the status's reason, other than review
	Note     string       `json:"note"`     // reason: the status's note; else ""
	By       string       `json:"by"`       // authorisation: the nearest authority; review: the reviewer
}

// ParentEdge is a requirement on a parent that a child's work would let pass.
type ParentEdge struct {
	Parent   TaskRef     `json:"parent"`
	Requires Requirement `json:"requires"` // Task is the dependent; From is the parent's gate
}
