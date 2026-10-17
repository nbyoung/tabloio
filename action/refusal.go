package action

import (
	"strconv"
	"strings"
)

// Refusal is a rule that stops a plan. Code is stable; Rule is the RULES.md
// id the commit would break, when one exists.
type Refusal struct{ Code, Task, Gate, Rule, Text string }

// Error returns "<task> <gate>: <text> (<code>, <rule>)", leaving out the
// parts the refusal lacks.
func (e Refusal) Error() string {
	var b strings.Builder
	if where := strings.TrimSpace(e.Task + " " + e.Gate); where != "" {
		b.WriteString(where + ": ")
	}
	b.WriteString(e.Text + " (" + e.Code)
	if e.Rule != "" {
		b.WriteString(", " + e.Rule)
	}
	b.WriteString(")")
	return b.String()
}

// Refusals is every refusal of one plan; a plan function returns it as its error.
type Refusals []Refusal

// Error returns each refusal's text, one per line.
func (e Refusals) Error() string {
	lines := make([]string, len(e))
	for i, r := range e {
		lines[i] = r.Error()
	}
	return strings.Join(lines, "\n")
}

// ExecError is a git or tablo run that failed for a reason outside the method.
type ExecError struct {
	Tool   string
	Args   []string
	Exit   int // the exit code; -1 when the tool did not run or ran in process
	Stderr string
}

// Error returns the tool, its arguments, the exit code and what it printed
// on standard error.
func (e *ExecError) Error() string {
	head := strings.TrimSpace(e.Tool + " " + strings.Join(e.Args, " "))
	if e.Exit >= 0 {
		head += ": exit " + strconv.Itoa(e.Exit)
	}
	if msg := strings.TrimSpace(e.Stderr); msg != "" {
		head += ": " + msg
	}
	return head
}
