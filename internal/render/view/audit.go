package view

func init() { register("audit", func() View { return new(Audit) }) }

// Audit is the audit view.
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

// Finding is the tasks one rule hits at one gate for one reason.
type Finding struct {
	Rule     string    `json:"rule"`     // the code of corpus/RULES.md; empty for a proposed task and a stale status
	Severity string    `json:"severity"` // error, warning, information
	Kind     string    `json:"kind"`     // the rule in a few words: worded text (R17)
	Tasks    []TaskRef `json:"tasks"`
	Gate     string    `json:"gate"`
	Files    []string  `json:"files"`
	Message  string    `json:"message"` // worded text
	Action   string    `json:"action"`  // worded text
	Resolver string    `json:"resolver"`

	Sentence string    `json:"sentence"` // provenance from here: the rule's sentence in the method
	Source   Reference `json:"source"`   // where the sentence stands
	Facts    []Fact    `json:"facts"`    // worded pairs: the junction, the reading
	Commits  []Commit  `json:"commits"`  // the commits involved, oldest first
	Commands []Command `json:"commands"` // what shows the finding, then what resolves it
}
