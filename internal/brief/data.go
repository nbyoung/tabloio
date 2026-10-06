package brief

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Data is data.brief of the queue view: one work item with every fact
// the brief states. Unknown fields are ignored.
type Data struct {
	Kind       string    `json:"kind"`  // "work ready" or "work waiting"
	Cause      string    `json:"cause"` // why a waiting item waits
	Person     string    `json:"person"`
	Ref        Ref       `json:"ref"`
	Task       Task      `json:"task"`
	Gate       Gate      `json:"gate"`       // the gate the work reaches
	Junction   Junction  `json:"junction"`   // the junction, resolved
	Supplier   *Supplier `json:"supplier"`   // nil when the task's own entry or the default supplies the contributor
	Requires   []Edge    `json:"requires"`   // the task's requirements, in file order
	Dependents []Edge    `json:"dependents"` // the tasks that require this one, in display order
	Parent     *Parent   `json:"parent"`     // nil for the root
	Status     Status    `json:"status"`
}

// Ref is the commit in view. Name is the ref as the user gives it, Commit
// the full hash and Date the commit's date as YYYY-MM-DD.
type Ref struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
	Date   string `json:"date"`
}

// Task is the task of the item. Path holds its ancestors, root first.
type Task struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	References  []Reference `json:"references"`
	Path        []Link      `json:"path"`
}

// Gate is the gate with its legend. Follows is the kind of the task's next
// applicable junction after this gate: "plain", "recursive", or "" when the
// gate is the task's last applicable gate.
type Gate struct {
	Key      string `json:"key"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Criteria string `json:"criteria"`
	Follows  string `json:"follows"`
}

// Junction holds the resolved fields, each with the entry that supplies it.
// Model.Value is "" for a person, and Reviewer.Value is "" when no one
// reviews.
type Junction struct {
	Contributor Field       `json:"contributor"`
	Model       Field       `json:"model"`
	Reviewer    Field       `json:"reviewer"`
	References  []Reference `json:"references"` // each with its Source
}

// Field is one resolved value and its source.
type Field struct {
	Value  string `json:"value"`
	Source Source `json:"source"`
}

// Source names where a field resolves from. Kind is "task" (an entry of
// the task ID, Title), "assignee" (a reviewer the method gives an agent) or
// "default" (the plain default).
type Source struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Reference is a reference of a task or of a junction. Source is set on a
// junction reference.
type Reference struct {
	URL    string `json:"url"`
	Text   string `json:"text"`
	Source Source `json:"source"`
}

// Link names a task.
type Link struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Supplier is the task whose entry supplies the contributor.
type Supplier struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Edge is one requirement, seen from either end. URL and Commit are set on
// a cross-project requirement. Condition is the key VIEWS.md names.
type Edge struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	From      string `json:"from"`
	To        string `json:"to"`
	Text      string `json:"text"`
	Condition string `json:"condition"`
	URL       string `json:"url"`
	Commit    string `json:"commit"`
}

// Parent is the task's parent with the tasks that require it.
type Parent struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Dependents []Edge `json:"dependents"`
}

// Status is the task's status with its symbols, its deciding commit and
// the path of its file from the repository root.
type Status struct {
	Path         string `json:"path"`
	Gate         string `json:"gate"`
	GateSymbol   string `json:"gate_symbol"`
	State        string `json:"state"`
	StateSymbol  string `json:"state_symbol"`
	Reason       string `json:"reason"`
	ReasonSymbol string `json:"reason_symbol"`
	Note         string `json:"note"`
	Date         string `json:"date"`
	Recorder     string `json:"recorder"`
	Commit       string `json:"commit"`
}

// Request names the item a brief is asked for.
type Request struct{ Person, Task, Gate string }

// NoWorkError says the person's queue holds no work at the task and gate.
type NoWorkError struct{ Person, Task, Gate string }

// Error returns the message without the "tabloio: " prefix.
func (e *NoWorkError) Error() string {
	return fmt.Sprintf("no work for %s at %s %s; tabloio queue --person %s lists the items",
		e.Person, e.Task, e.Gate, e.Person)
}

// WaitingError says the work waits, with its cause.
type WaitingError struct{ Task, Gate, Cause string }

// Error returns the message without the "tabloio: " prefix.
func (e *WaitingError) Error() string {
	return fmt.Sprintf("%s at %s waits and no brief starts it: %s", e.Task, e.Gate, e.Cause)
}

// Decode reads the envelope's data and returns its brief. It returns a
// *NoWorkError when data holds no brief object, and an error that names the
// field when task.id, gate.key, junction.contributor.value, status.path or
// ref.commit is empty.
func Decode(data []byte, req Request) (Data, error) {
	var env struct {
		Brief json.RawMessage `json:"brief"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return Data{}, fmt.Errorf("brief: data: %w", err)
	}
	raw := bytes.TrimSpace(env.Brief)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) || bytes.Equal(raw, []byte("{}")) {
		return Data{}, &NoWorkError{Person: req.Person, Task: req.Task, Gate: req.Gate}
	}
	var d Data
	if err := json.Unmarshal(raw, &d); err != nil {
		return Data{}, fmt.Errorf("brief: data.brief: %w", err)
	}
	for _, f := range []struct{ name, value string }{
		{"task.id", d.Task.ID},
		{"gate.key", d.Gate.Key},
		{"junction.contributor.value", d.Junction.Contributor.Value},
		{"status.path", d.Status.Path},
		{"ref.commit", d.Ref.Commit},
	} {
		if f.value == "" {
			return Data{}, fmt.Errorf("brief: %s is missing in tablo's data", f.name)
		}
	}
	return d, nil
}

// Render writes the brief of d to w. It returns a *WaitingError and writes
// nothing unless d.Kind is "work ready". The same d gives the same bytes.
func Render(w io.Writer, d Data) error {
	if d.Kind != "work ready" {
		return &WaitingError{Task: d.Task.ID, Gate: d.Gate.Key, Cause: d.Cause}
	}
	var b bytes.Buffer
	for _, l := range lines(d) {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	_, err := w.Write(b.Bytes())
	return err
}
