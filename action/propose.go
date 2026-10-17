package action

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Proposal is a new task.
type Proposal struct {
	Title, Description string
	Parent             string      // required: propose never creates a root
	Assignee           string      // "" means the author's email
	Order              int         // 0 means one above the largest order among the parent's children
	References         []Reference // in the order given
	ID                 string      // "" draws a fresh id
}

// Reference is one entry of a task's references.
type Reference struct{ URL, Text string }

// freshDraws bounds the draws of a fresh id.
const freshDraws = 1000

// descriptionWidth is the width the folded description wraps at, before its
// indent.
const descriptionWidth = 78

// Propose plans the commit that adds tasks/<id>.yaml. It writes no junction
// and no requirement, since a new task inherits its junctions, and never an
// Authorised: trailer: the commit is the task's deciding commit (D4).
func (r *Repo) Propose(ctx context.Context, p Proposal, o Options) (*Plan, error) {
	s, err := r.begin(ctx)
	if err != nil {
		return nil, err
	}
	var refusals Refusals
	bad := func(text string) { refusals = append(refusals, Refusal{Code: "bad-field", Text: text}) }
	broken := func(v string) bool { return strings.ContainsAny(v, "\r\n") }
	switch {
	case strings.TrimSpace(p.Title) == "":
		bad("the title is empty")
	case broken(p.Title):
		bad("the title holds a line break")
	}
	if strings.TrimSpace(p.Description) == "" {
		bad("the description is empty")
	}
	if a := p.Assignee; a != "" && (!strings.Contains(a, "@") || broken(a)) {
		bad("the assignee " + strconv.Quote(a) + " is no email address")
	}
	for _, ref := range p.References {
		if strings.TrimSpace(ref.URL) == "" {
			bad("a reference has no URL")
		}
		if broken(ref.URL) || broken(ref.Text) {
			bad("a reference holds a line break")
		}
	}
	if p.Order < 0 {
		bad("the order " + strconv.Itoa(p.Order) + " is negative")
	}
	if p.Parent == "" {
		bad("a proposal needs a parent: propose never creates a root")
	} else {
		refusals = append(refusals, s.checkIDs(p.Parent)...)
	}
	taken, err := s.taken(ctx)
	if err != nil {
		return nil, err
	}
	switch {
	case p.ID != "" && !idPattern.MatchString(p.ID):
		bad("the id " + strconv.Quote(p.ID) + " is not four lowercase hexadecimal digits")
	case p.ID != "" && taken[p.ID]:
		refusals = append(refusals, Refusal{Code: "id-taken", Task: p.ID, Text: "a task or status file of that name stands at HEAD or in the working tree"})
	}
	if len(refusals) > 0 {
		return nil, refusals
	}

	order := p.Order
	if order == 0 {
		parent, err := s.src.Task(ctx, p.Parent)
		if err != nil {
			return nil, err
		}
		for _, c := range parent.Children {
			order = max(order, c.Order)
		}
		order++
	}
	id := p.ID
	if id == "" {
		if id, err = r.fresh(taken); err != nil {
			return nil, err
		}
	}
	assignee := p.Assignee
	if assignee == "" {
		assignee = s.author.Email
	}
	plan := &Plan{
		Command: "propose", Tasks: []string{id}, NewID: id,
		Subject: "Propose " + lowerTitle(p.Title) + " (" + id + ")",
		Files:   []FileChange{{Path: s.taskPath(id), New: []byte(taskFile(p, assignee, order))}},
	}
	return s.finish(ctx, plan, o, nil)
}

// taken returns the ids that name a file under tasks/ or status/ at HEAD or
// in the working tree.
func (s *session) taken(ctx context.Context) (map[string]bool, error) {
	ids := map[string]bool{}
	for id := range s.ids {
		ids[id] = true
	}
	out, err := s.git(ctx, "ls-tree", "--name-only", "HEAD", "--", s.project+"/status/")
	if err != nil {
		return nil, err
	}
	names := strings.Fields(out)
	for _, dir := range []string{"tasks", "status"} {
		entries, err := os.ReadDir(filepath.Join(s.top, filepath.FromSlash(s.project), dir))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		for _, e := range entries {
			names = append(names, e.Name())
		}
	}
	for _, name := range names {
		if id, ok := strings.CutSuffix(path.Base(name), ".yaml"); ok {
			ids[id] = true
		}
	}
	return ids, nil
}

// fresh draws two bytes from Rand, in hexadecimal, until they name no file
// that stands.
func (r *Repo) fresh(taken map[string]bool) (string, error) {
	src := r.Rand
	if src == nil {
		src = rand.Reader
	}
	b := make([]byte, 2)
	for range freshDraws {
		if _, err := io.ReadFull(src, b); err != nil {
			return "", err
		}
		if id := hex.EncodeToString(b); !taken[id] {
			return id, nil
		}
	}
	return "", &ExecError{Tool: "rand", Exit: -1, Stderr: "no free task id after " + strconv.Itoa(freshDraws) + " draws"}
}

// taskFile writes the task file of design/b618/task.yaml: the title, the
// description folded at 78 columns, the assignee, the references as flow
// mappings, and the parent with its id in double quotes.
func taskFile(p Proposal, assignee string, order int) string {
	var b strings.Builder
	b.WriteString("title: " + scalar(p.Title) + "\ndescription: >\n")
	for _, l := range wrap(strings.Fields(p.Description), descriptionWidth) {
		b.WriteString("  " + l + "\n")
	}
	b.WriteString("assignee: " + scalar(assignee) + "\n")
	if len(p.References) > 0 {
		b.WriteString("references:\n")
		for _, ref := range p.References {
			b.WriteString("  - { url: " + flowScalar(ref.URL))
			if ref.Text != "" {
				b.WriteString(", text: " + flowScalar(ref.Text))
			}
			b.WriteString(" }\n")
		}
	}
	b.WriteString(`parent: { id: "` + p.Parent + `", order: ` + strconv.Itoa(order) + " }\n")
	return b.String()
}

// wrap joins words into lines of at most width characters, a longer word
// alone on its line.
func wrap(words []string, width int) []string {
	var lines []string
	line := ""
	for _, w := range words {
		switch {
		case line == "":
			line = w
		case utf8.RuneCountInString(line)+1+utf8.RuneCountInString(w) <= width:
			line += " " + w
		default:
			lines = append(lines, line)
			line = w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// taskPath returns a task's file from the repository's root.
func (s *session) taskPath(id string) string { return s.project + "/tasks/" + id + ".yaml" }
