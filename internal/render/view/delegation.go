package view

func init() { register("authority", func() View { return new(Delegation) }) }

// Delegation is the authority delegation view.
type Delegation struct {
	Head
	Rows     []DelegationRow `json:"rows"`     // every task in view, in display order
	Parents  []Chain         `json:"parents"`  // detail: every parent in view, in display order
	Deciding []Deciding      `json:"deciding"` // provenance: newest first
	Commands []Command       `json:"commands"` // provenance
}

// DelegationRow is one task of the tree.
type DelegationRow struct {
	TaskRef
	Depth     int    `json:"depth"`
	Assignee  string `json:"assignee"`
	Parent    bool   `json:"parent"`             // the task has children
	Delegated bool   `json:"delegated"`          // its assignee differs from its parent's
	Proposed  bool   `json:"proposed"`           // its deciding commit leaves it unauthorised
	Differs   bool   `json:"differs_from_trunk"` // off the trunk: the branch changes the file
}

// Chain is what one parent gives its subtree.
type Chain struct {
	TaskRef
	Assignee    string      `json:"assignee"`
	Authorities []Authority `json:"authorities"` // of its children, nearest first (design e3ed)
	Children    []string    `json:"children"`
	Defaults    []Default   `json:"defaults"` // the gates at which this parent states an entry
}

// Default is a junction entry a parent states, with each field resolved.
type Default struct {
	Gate string `json:"gate"`
	Marked
	Contributor     string `json:"contributor"`
	Model           string `json:"model"`
	Reviewer        string `json:"reviewer"`
	ContributorFrom string `json:"contributor_from"` // an ancestor's id when it supplies the field
	ModelFrom       string `json:"model_from"`
	ReviewerFrom    string `json:"reviewer_from"`
}

// Deciding groups the tasks one commit decides.
type Deciding struct {
	Commit Commit   `json:"commit"`
	Way    string   `json:"way"`    // change, merge, trailer
	Merged string   `json:"merged"` // a merge: the commit it brings in
	By     string   `json:"by"`     // author, committer, both, or empty
	Tasks  []string `json:"tasks"`
}
