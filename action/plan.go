package action

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Agent names what ran. The zero value adds no trailer.
type Agent struct {
	Model    string // the value of Model:
	CoAuthor string // the value of Co-Authored-By:, "Name <email>"; refused without Model
}

// Options are common to the five commands.
type Options struct {
	Agent   Agent
	Subject string // replaces the default subject; one line
	Body    string // paragraphs between the subject and the trailers; "" for none
}

// Plan is a commit that does not exist yet: the dry-run form.
type Plan struct {
	Command           string       // "propose", "record", "review", "reaffirm" or "authorise"
	Tasks             []string     // the ids the commit concerns, in the order given
	NewID             string       // propose: the id allocated
	Branch            string       // the checked-out branch; "" on a detached HEAD
	Base              string       // the full hash of HEAD when planned
	Author, Committer Identity     // as Git will write them
	Subject, Body     string       // the message without its trailers
	Trailers          []Trailer    // task trailers in the order given, then Model, then Co-Authored-By
	Files             []FileChange // sorted by path
	Pins              []PinChange  // sorted by path
	Tree              string       // the tree the commit will have
	Diff              string       // git's diff from the tree of Base to Tree; "" for an empty commit
	Warnings          []Warning
}

// Identity is a name and an email as Git writes them.
type Identity struct{ Name, Email string }

// Trailer is one line of the message's final paragraph, "Key: Value".
type Trailer struct{ Key, Value string }

// FileChange is a file the commit writes.
type FileChange struct {
	Path     string // relative to the repository root, with forward slashes
	Old, New []byte // Old is nil for a file the commit creates
}

// PinChange is a submodule path and two full hashes: the gitlink at Base and
// the commit the submodule's checkout stands at.
type PinChange struct{ Path, Old, New string }

// Warning is a finding that does not stop a plan.
type Warning struct{ Code, Task, Text string }

// Result is a commit that exists.
type Result struct{ Commit string } // the full hash of the new commit

// Message returns the commit message: the subject; a blank line and the body
// when there is one; a blank line and the trailers, one per line, as one
// final paragraph. It ends with one newline.
func (p *Plan) Message() string {
	var b strings.Builder
	b.WriteString(p.Subject + "\n")
	if p.Body != "" {
		b.WriteString("\n" + p.Body + "\n")
	}
	if len(p.Trailers) > 0 {
		b.WriteString("\n")
		for _, t := range p.Trailers {
			b.WriteString(t.Key + ": " + t.Value + "\n")
		}
	}
	return b.String()
}

// String returns the shown form (design/b618/plan.txt): the line "<Command>
// <ids> on <branch> at <base>", the author and committer, the message
// indented by four spaces between blank lines, the diff, then one line per
// warning.
func (p *Plan) String() string {
	var b strings.Builder
	on := "on " + p.Branch
	if p.Branch == "" {
		on = "on a detached HEAD"
	}
	b.WriteString(capital(p.Command) + " " + strings.Join(p.Tasks, " ") + " " + on + " at " + short(p.Base) + "\n")
	b.WriteString("author    " + p.Author.String() + "\n")
	b.WriteString("committer " + p.Committer.String() + "\n\n")
	for _, l := range strings.Split(strings.TrimSuffix(p.Message(), "\n"), "\n") {
		if l != "" {
			l = "    " + l
		}
		b.WriteString(l + "\n")
	}
	if p.Diff != "" {
		b.WriteString("\n" + p.Diff)
	} else if len(p.Warnings) > 0 {
		b.WriteString("\n")
	}
	for _, w := range p.Warnings {
		b.WriteString("warning: " + w.Code + ": " + w.Text + "\n")
	}
	return b.String()
}

// String returns "Name <email>".
func (i Identity) String() string { return i.Name + " <" + i.Email + ">" }

// short returns the first seven digits of a hash.
func short(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

// capital returns s with its first letter in upper case.
func capital(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// lowerTitle returns a title as a subject names it: its first letter in
// lower case when its second is, so that "Write commands" reads "write
// commands" and "API docs" stays.
func lowerTitle(title string) string {
	r1, n1 := utf8.DecodeRuneInString(title)
	r2, _ := utf8.DecodeRuneInString(title[n1:])
	if unicode.IsUpper(r1) && unicode.IsLower(r2) {
		return string(unicode.ToLower(r1)) + title[n1:]
	}
	return title
}
