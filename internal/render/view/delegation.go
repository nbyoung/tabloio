package view

func init() { register("authority", func() View { return new(Delegation) }) }

// Delegation is the authority delegation view.
type Delegation struct {
	Head
	Rows     []DelegationRow `json:"rows"`     // every task in view, in display order
	Parents  []Chain         `json:"parents"`  // detail: every row that has children, in display order
	Deciding []Deciding      `json:"deciding"` // provenance: newest first
	Commands []Command       `json:"commands"` // provenance
}

// DelegationRow is one task of the tree.
type DelegationRow struct {
	TaskRef
	Depth     int    `json:"depth"` // 0 for the task the view starts at
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
	Defaults    []Junction  `json:"defaults"` // the gates at which this parent states an entry, resolved
	Sources     []Supplier  `json:"sources"`  // of those junctions, by gate, then field
}

// Deciding groups the tasks one commit decides.
type Deciding struct {
	Commit Commit   `json:"commit"`
	Way    string   `json:"way"`    // change, merge, trailer
	Merged string   `json:"merged"` // a merge: the commit it brings in
	By     string   `json:"by"`     // author, committer, both, or empty
	Tasks  []string `json:"tasks"`
}
