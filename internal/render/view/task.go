package view

func init() { register("task", func() View { return new(Task) }) }

// Task is the task definition view.
type Task struct {
	Head
	Task          TaskRef        `json:"task"`
	Assignee      string         `json:"assignee"`
	Parent        *Parent        `json:"parent"` // nil for the root
	Status        Status         `json:"status"`
	Authorised    bool           `json:"authorised"`
	Description   string         `json:"description"` // detail from here
	References    []Reference    `json:"references"`
	Path          []TaskRef      `json:"path"`     // the root first, the task last
	Siblings      int            `json:"siblings"` // children of the parent, the task included
	Children      []ChildRow     `json:"children"`
	Authorities   []Authority    `json:"authorities"` // nearest first
	Requires      []Requirement  `json:"requires"`
	Dependents    []Requirement  `json:"dependents"`
	Junctions     []Junction     `json:"junctions"`     // every gate after undefined, in order
	Snapshot      *Snapshot      `json:"snapshot"`      // a snapshot status: the task read
	StatusCommit  *Commit        `json:"status_commit"` // provenance from here
	Authorisation *Authorisation `json:"authorisation"`
	Considered    []string       `json:"considered"` // a parent: the children its roll-up considers
	Chain         []Step         `json:"chain"`      // the statuses the status comes through
	Linkages      []Linkage      `json:"linkages"`   // one per subproject a recursive junction reads
	Sources       []Supplier     `json:"sources"`    // by gate, then field
	Reviews       []Review       `json:"reviews"`
	Models        []ModelCheck   `json:"models"`
	Events        []Event        `json:"events"` // newest first, at most five
	EventCount    int            `json:"event_count"`
	Commands      []Command      `json:"commands"`
}

// Parent is the parent of a task and the task's place among its siblings.
type Parent struct {
	TaskRef
	Order *int `json:"order"` // nil when the file states none
}

// ChildRow is a child with its status.
type ChildRow struct {
	TaskRef
	Status Status `json:"status"`
}

// Authority is an assignee whose authority covers the task.
type Authority struct {
	Email string `json:"email"`
	By    string `json:"by"` // the ancestor's id
}

// Requirement is an edge between two tasks, seen from either end.
type Requirement struct {
	Task       TaskRef `json:"task"` // the other task
	From       string  `json:"from"`
	To         string  `json:"to"`
	Text       string  `json:"text"`
	Met        bool    `json:"met"`
	Due        bool    `json:"due"`
	Status     *Status `json:"status"`     // of the required task; nil for a dependent
	Subproject string  `json:"subproject"` // a cross-project entry: its url
	Commit     string  `json:"commit"`     // and the commit it reads
}

// Junction is one gate's junction of the task.
type Junction struct {
	Gate string `json:"gate"`
	Marked
	Contributor string      `json:"contributor"`
	Model       string      `json:"model"`
	Reviewer    string      `json:"reviewer"`
	ByDefault   bool        `json:"reviewer_by_default"` // the assignee reviews, since no entry states one
	Subproject  *SubRef     `json:"subproject"`          // a recursive junction
	References  []Reference `json:"references"`
	Stands      string      `json:"stands"` // passed, here, next, later, exempt; "" on a parent
	Reviewed    bool        `json:"reviewed"`
}

// SubRef names the task of a subproject that a recursive junction reads.
type SubRef struct {
	URL    string  `json:"url"`
	ID     string  `json:"id"`     // the id the entry states; "" when it reads the root
	Task   TaskRef `json:"task"`   // the task read there; zero when undetermined
	Form   string  `json:"form"`   // submodule, directory, url or none
	Commit string  `json:"commit"` // the commit the link reads, in full; "" for a directory
	Why    string  `json:"why"`    // no_link or no_task; "" when the task is read
}

// Snapshot is a subproject's task as a recursive junction reads it.
type Snapshot struct {
	SubRef
	Assignee string     `json:"assignee"`
	Status   *Status    `json:"status"` // nil when the task is not read
	Children []ChildRow `json:"children"`
}

// Authorisation is the commit that authorises a task and how it does.
type Authorisation struct {
	Commit *Commit `json:"commit"` // nil when no commit decides
	Way    string  `json:"way"`    // change, merge, trailer
	By     string  `json:"by"`     // author, committer, both, or empty for a proposed task
}

// Step is one status of a chain: the task's own first, to the leaf whose
// file states it.
type Step struct {
	Task   TaskRef `json:"task"`
	URL    string  `json:"url"` // the subproject the task is of; "" in the project in view
	Status Status  `json:"status"`
}

// Linkage is one subproject a recursive junction reads, and what fixes the
// commit it reads.
type Linkage struct {
	Gates []string `json:"gates"` // the gates whose junctions read it, in order
	SubRef
	Stated               // where the url stands
	StatedCommit string  `json:"stated_commit"` // the commit field as the entry writes it; "" for none
	Recorded     string  `json:"recorded"`      // a submodule: the gitlink of the source commit
	Moved        bool    `json:"moved"`         // a submodule: the checkout stands elsewhere
	Trunk        string  `json:"trunk"`         // the subproject's trunk; "" when undetermined
	Tip          string  `json:"tip"`           // the tip of that trunk, in full
	OnTrunk      string  `json:"on_trunk"`      // unknown, no or yes
	Behind       string  `json:"behind"`        // unknown, no or yes
	Pinned       *Commit `json:"pinned"`        // the commit read
	SetBy        *Commit `json:"set_by"`        // the commit of the newest pin event of the link
}

// Fact is a name and a value in tablo's words.
type Fact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Supplier says what supplies one field of one junction.
type Supplier struct {
	Gate  string `json:"gate"`
	Field string `json:"field"` // contributor, model, reviewer, references, subproject or applies
	Value string `json:"value"`
	Stated
}

// Review is the commit that accepts a reviewed junction.
type Review struct {
	Gate     string  `json:"gate"`
	Reviewer string  `json:"reviewer"`
	Commit   *Commit `json:"commit"` // nil while nothing accepts
	Effect   string  `json:"effect"` // accepts, authorisation, none
}

// ModelCheck compares a junction's model with a commit's Model trailer.
type ModelCheck struct {
	Commit  string `json:"commit"`
	Gate    string `json:"gate"`
	Stated  string `json:"stated"`
	Trailer string `json:"trailer"`
	Reading string `json:"reading"` // agrees, differs, missing, exempt, unstated
}

// Event is one event of a task: what one commit does to it of one kind.
type Event struct {
	Date       string       `json:"date"`
	Commit     string       `json:"commit"`
	By         string       `json:"by"`         // the author's email
	Kind       string       `json:"kind"`       // task, authorised, status, reaffirmed, reviewed, pin
	Recorded   *EventStatus `json:"recorded"`   // status: what the commit records
	Gate       string       `json:"gate"`       // reviewed
	Pin        *PinMove     `json:"pin"`        // pin
	Subproject *SubEvent    `json:"subproject"` // an event taken from a subproject
}

// EventStatus is what a status event records.
type EventStatus struct {
	Gate   string `json:"gate"`
	State  string `json:"state"`
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

// PinMove is a subproject pin that moves from one commit to another.
type PinMove struct {
	URL string `json:"url"`
	Old string `json:"old"` // "" when the linkage first appears
	New string `json:"new"`
}

// SubEvent names the subproject task an event comes from.
type SubEvent struct {
	URL  string `json:"url"`
	Task string `json:"task"`
}
