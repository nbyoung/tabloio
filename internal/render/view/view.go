// Package view holds the data of each view as tabloio's renderers read it.
//
// The types are tabloio's own. Each mirrors the data object that tablo emits
// for a view as JSON, and the JSON tags are the contract. Decoding ignores
// unknown fields and leaves absent ones zero, so tablo may add a field within
// a minor series.
package view

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Level orders the three levels of disclosure; each includes the one before.
type Level int

// The levels, from the least to the most.
const (
	Glance Level = iota
	Detail
	Provenance
)

var levelNames = [...]string{"glance", "detail", "provenance"}

// ParseLevel reads "glance", "detail" or "provenance".
func ParseLevel(s string) (Level, error) {
	for i, n := range levelNames {
		if s == n {
			return Level(i), nil
		}
	}
	return 0, fmt.Errorf("view: unknown level %q", s)
}

// String returns the level's word, or "level(n)" for a value out of range.
func (l Level) String() string {
	if l < 0 || int(l) >= len(levelNames) {
		return fmt.Sprintf("level(%d)", int(l))
	}
	return levelNames[l]
}

// MarshalText writes the level's word.
func (l Level) MarshalText() ([]byte, error) {
	if l < 0 || int(l) >= len(levelNames) {
		return nil, fmt.Errorf("view: unknown level %d", int(l))
	}
	return []byte(levelNames[l]), nil
}

// UnmarshalText reads the level's word.
func (l *Level) UnmarshalText(b []byte) error {
	v, err := ParseLevel(string(b))
	if err != nil {
		return err
	}
	*l = v
	return nil
}

// View is the data of one view. Every view struct embeds Head, which
// supplies the method.
//
// The method is Header, not Head: a struct that embeds a field named Head
// does not promote a method of the same name.
type View interface{ Header() *Head }

// ErrUnknown is the error Decode returns for a view name it does not know.
var ErrUnknown = errors.New("view: unknown view")

// decoders maps a command name to a constructor of its view struct. Each
// view's file registers it in init, so a new view changes no shared file.
var decoders = map[string]func() View{}

// register adds a view to Decode. It panics on a duplicate name.
func register(name string, newView func() View) {
	if _, ok := decoders[name]; ok {
		panic("view: duplicate view " + name)
	}
	decoders[name] = newView
}

// Decode reads the data object of one view. The names are tabloio's command
// names: gates, task, authority, assignment, queue, blockage, tableau,
// context, history, audit. It returns ErrUnknown for any other name.
func Decode(name string, data []byte) (View, error) {
	newView, ok := decoders[name]
	if !ok {
		return nil, fmt.Errorf("%w %q", ErrUnknown, name)
	}
	v := newView()
	if err := json.Unmarshal(data, v); err != nil {
		return nil, fmt.Errorf("view %s: %w", name, err)
	}
	return v, nil
}

// Head is what every view's data carries; each view struct embeds it.
type Head struct {
	Level   Level   `json:"level"`   // the deepest level the data holds
	Project TaskRef `json:"project"` // the root task
	Ref     Ref     `json:"ref"`
	Params  Params  `json:"params"` // the parameters in force, as tablo resolves them
	Legend  Legend  `json:"legend"`
}

// Header returns the head itself.
func (h *Head) Header() *Head { return h }

// Ref is the commit in view.
type Ref struct {
	Name    string `json:"name"`   // as given, "HEAD" by default
	From    string `json:"from"`   // the start of a range; history only
	Commit  string `json:"commit"` // the full hash
	Date    string `json:"date"`   // its author date, YYYY-MM-DD
	OnTrunk bool   `json:"on_trunk"`
}

// Params echoes the focusing parameters; a zero field is not in force.
type Params struct {
	Task       string   `json:"task"` // empty when the view covers the root
	Person     string   `json:"person"`
	Window     *int     `json:"window"`
	Columns    []string `json:"columns"`
	Historical bool     `json:"historical"`
	Proposed   bool     `json:"proposed"`
	Stale      int      `json:"stale"`
}

// Legend is gates.yaml and the method's marks, in file order.
type Legend struct {
	Gates   []Gate   `json:"gates"`
	States  []State  `json:"states"`
	Reasons []Reason `json:"reasons"`
	Marks   []Mark   `json:"marks"`
}

// Gate is one gate of gates.yaml.
type Gate struct {
	Key      string `json:"key"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Criteria string `json:"criteria"`
}

// State is one state of gates.yaml.
type State struct {
	Key      string `json:"key"`
	Symbol   string `json:"symbol"`
	Synopsis string `json:"synopsis"`
	Severity int    `json:"severity"`
}

// Reason is one reason of gates.yaml; the reserved one is the method's.
type Reason struct {
	Key      string `json:"key"`
	Symbol   string `json:"symbol"`
	Synopsis string `json:"synopsis"`
	Reserved bool   `json:"reserved"` // true for review
}

// Mark is a junction mark of the method. Its key is person, agent, reviewer,
// subproject or exempt.
type Mark struct {
	Key      string `json:"key"`
	Symbol   string `json:"symbol"`
	Meaning  string `json:"meaning"`
	Junction string `json:"junction"`
}

// TaskRef names a task.
type TaskRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Person is an actor; Name is empty where only the email is known.
type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Status is a status as a view shows it: stated, rolled up, or from a snapshot.
type Status struct {
	Gate     string  `json:"gate"`
	State    string  `json:"state"`
	Reason   string  `json:"reason"`
	Note     string  `json:"note"`
	Date     string  `json:"date"`
	Recorder string  `json:"recorder"`
	From     *Origin `json:"from"` // nil for a status file
}

// Origin says where a derived status comes from.
type Origin struct {
	Kind string  `json:"kind"` // "rollup" or "snapshot"
	Task TaskRef `json:"task"` // the child, or the subproject's task
	URL  string  `json:"url"`  // snapshot: the subproject
	Pin  string  `json:"pin"`  // snapshot: the commit read
	Gate string  `json:"gate"` // snapshot: where the subproject's task stands
}

// Commit is a Git fact.
type Commit struct {
	Hash          string   `json:"hash"`
	Subject       string   `json:"subject"`
	Date          string   `json:"date"` // author date
	Author        Person   `json:"author"`
	Committer     Person   `json:"committer"`
	AuthorTime    string   `json:"author_time"` // RFC 3339 with offset
	CommitterTime string   `json:"committer_time"`
	Trailers      []string `json:"trailers"`
	Files         []File   `json:"files"`
}

// File is a path a commit changes; Change is added, changed, removed or pin.
type File struct {
	Path   string `json:"path"`
	Change string `json:"change"`
}

// Command is a line a reader runs to reproduce or resolve a fact.
type Command struct {
	Text    string `json:"text"`
	Comment string `json:"comment"`
}

// Marked pairs a junction's marks, in the legend's order, with whether the
// person in view acts there.
type Marked struct {
	Marks []string `json:"marks"` // keys of Legend.Marks
	Acts  bool     `json:"acts"`
}
