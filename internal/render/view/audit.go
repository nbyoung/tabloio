package view

func init() { register("audit", func() View { return new(Audit) }) }

// Audit is the audit view. The three counts and Most cover every finding in
// view; Findings lists the rows at or above the minimum.
type Audit struct {
	Head
	Findings    []Finding  `json:"findings"` // errors, warnings, information; then rule, gate, display order
	Errors      int        `json:"errors"`   // counts of findings, one per task a rule hits
	Warnings    int        `json:"warnings"`
	Information int        `json:"information"`
	Most        *Resolving `json:"most"` // the person with the most to resolve; nil when none
}

// Resolving is the number of findings one person resolves.
type Resolving struct {
	Email string `json:"email"`
	Count int    `json:"count"`
}

// Finding is the tasks one kind hits at one gate for one reason.
type Finding struct {
	Key      string    `json:"key"`      // a rule id of corpus/RULES.md, or proposed, stale or sole_review
	Rule     string    `json:"rule"`     // the rule id; empty for the three kinds no rule states
	Severity string    `json:"severity"` // error, warning, information, as the audit lists it
	Kind     string    `json:"kind"`     // the rule in a few words: worded text (R17)
	Tasks    []TaskRef `json:"tasks"`
	Gate     string    `json:"gate"`
	Act      string    `json:"act"` // one of the ten acts; none where nothing resolves it
	Resolver string    `json:"resolver"`

	Files   []string `json:"files"`   // detail from here
	Message string   `json:"message"` // worded text

	Sentence string         `json:"sentence"` // provenance from here: the rule's sentence in the method
	Source   Reference      `json:"source"`   // where the sentence stands
	Facts    []FindingFacts `json:"facts"`    // one per task that has a typed fact, in the order of Tasks
	Commits  []Commit       `json:"commits"`  // the commits involved, oldest first
	Commands []Command      `json:"commands"` // what shows the finding, then what resolves it
}

// FindingFacts is what the Derivation knows of one task of a finding, each
// fact in the type every other view gives it; a fact the finding does not
// carry is nil.
type FindingFacts struct {
	Task          string         `json:"task"`          // the task's id
	Junction      *Junction      `json:"junction"`      // the junction at the finding's gate
	Status        *Status        `json:"status"`        // the task's status
	Requirement   *Requirement   `json:"requirement"`   // the requires entry the finding stands at
	Authorisation *Authorisation `json:"authorisation"` // a proposed task: who decides, and why nothing does
	Linkage       *Linkage       `json:"linkage"`       // the subproject the junction reads
	Model         *ModelCheck    `json:"model"`         // the commit's Model: trailer against the model stated
}
