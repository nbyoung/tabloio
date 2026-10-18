package action

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Repo is one checkout of a Tableaux project.
type Repo struct {
	Dir    string    // the project directory, as -C names it; "" is the working directory
	Source Source    // the project's facts; nil means &Tablo{Dir: Dir}
	Env    []string  // added to the environment of every git child process
	Rand   io.Reader // the source of a fresh id; nil means crypto/rand.Reader
}

// idPattern is the form of a task id: four lowercase hexadecimal digits.
var idPattern = regexp.MustCompile(`^[0-9a-f]{4}$`)

// session is one plan in progress: what steps 1 to 3 of "What every plan
// does" read, and the Source the plan asks.
type session struct {
	r                 *Repo
	src               Source
	top               string // the repository's root, as Git prints it
	project           string // the .tableaux directory from the root, with forward slashes
	base, branch      string
	author, committer Identity
	ids               map[string]bool // the task ids at HEAD
}

// git runs git in dir with the process environment, Env, GIT_OPTIONAL_LOCKS=0
// and env, and returns its standard output. A run that fails is an
// *ExecError with Git's own text.
func (r *Repo) git(ctx context.Context, dir string, stdin []byte, env []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(append(append(os.Environ(), r.Env...), "GIT_OPTIONAL_LOCKS=0"), env...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		exit, text := -1, stderr.String()
		var x *exec.ExitError
		if errors.As(err, &x) {
			exit = x.ExitCode()
		}
		if text == "" {
			text = err.Error()
		}
		return out.String(), &ExecError{Tool: "git", Args: args, Exit: exit, Stderr: text}
	}
	return out.String(), nil
}

// exitCode returns the exit code of a git run that failed, or -1.
func exitCode(err error) int {
	var x *ExecError
	if errors.As(err, &x) {
		return x.Exit
	}
	return -1
}

// source returns the Source a plan asks: Repo.Source, or Tablo over Dir.
func (r *Repo) source() Source {
	if r.Source != nil {
		return r.Source
	}
	return &Tablo{Dir: r.Dir}
}

// begin finds the checkout, reads the identity and lists the task ids:
// steps 1 to 3 of "What every plan does".
func (r *Repo) begin(ctx context.Context) (*session, error) {
	dir := r.Dir
	if dir == "" {
		dir = "."
	}
	out, err := r.git(ctx, dir, nil, nil, "rev-parse", "--show-toplevel", "--show-prefix")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	s := &session{r: r, src: r.source(), top: lines[0], ids: map[string]bool{}}
	prefix := ""
	if len(lines) > 1 {
		prefix = lines[1]
	}
	if s.base, err = s.git(ctx, "rev-parse", "--verify", "--quiet", "HEAD^{commit}"); err != nil {
		return nil, &ExecError{Tool: "git", Args: []string{"rev-parse", "HEAD"}, Exit: exitCode(err), Stderr: "the repository has no commit yet"}
	}
	if err := s.findProject(ctx, prefix); err != nil {
		return nil, err
	}
	switch branch, err := s.git(ctx, "symbolic-ref", "--short", "--quiet", "HEAD"); {
	case err == nil:
		s.branch = branch
	case exitCode(err) != 1:
		return nil, err
	}
	if s.author, err = s.ident(ctx, "GIT_AUTHOR_IDENT"); err != nil {
		return nil, err
	}
	if s.committer, err = s.ident(ctx, "GIT_COMMITTER_IDENT"); err != nil {
		return nil, err
	}
	names, err := s.git(ctx, "ls-tree", "--name-only", "HEAD", "--", s.project+"/tasks/")
	if err != nil {
		return nil, err
	}
	for _, name := range strings.Fields(names) {
		if id, ok := strings.CutSuffix(path.Base(name), ".yaml"); ok && idPattern.MatchString(id) {
			s.ids[id] = true
		}
	}
	return s, nil
}

// git runs git at the repository's root and returns its output without the
// final newline.
func (s *session) git(ctx context.Context, args ...string) (string, error) {
	out, err := s.r.git(ctx, s.top, nil, nil, args...)
	return strings.TrimSuffix(out, "\n"), err
}

// findProject sets the project to the nearest .tableaux at or above the
// directory -C names, as tree entries of HEAD, up to the repository's root.
func (s *session) findProject(ctx context.Context, prefix string) error {
	var candidates []string
	for p := strings.TrimSuffix(prefix, "/"); ; p = path.Dir(p) {
		if p == "." || p == "" {
			candidates = append(candidates, ".tableaux")
			break
		}
		candidates = append(candidates, p+"/.tableaux")
	}
	args := append([]string{"ls-tree", "HEAD", "--"}, candidates...)
	out, err := s.git(ctx, args...)
	if err != nil {
		return err
	}
	trees := map[string]bool{}
	for _, l := range strings.Split(out, "\n") {
		if meta, p, ok := strings.Cut(l, "\t"); ok && strings.Contains(meta, " tree ") {
			trees[p] = true
		}
	}
	for _, c := range candidates {
		if trees[c] {
			s.project = c
			return nil
		}
	}
	return &ExecError{Tool: "git", Args: args, Exit: 0, Stderr: "HEAD holds no .tableaux at or above " + path.Join(".", prefix)}
}

// ident reads one identity as git commit writes it: the environment's
// GIT_AUTHOR_* or GIT_COMMITTER_* before any configuration.
func (s *session) ident(ctx context.Context, which string) (Identity, error) {
	out, err := s.git(ctx, "var", which)
	if err != nil {
		return Identity{}, err
	}
	end := strings.LastIndex(out, ">")
	start := strings.LastIndex(out[:max(end, 0)], "<")
	if start < 0 || end < 0 {
		return Identity{}, &ExecError{Tool: "git", Args: []string{"var", which}, Exit: 0, Stderr: "no identity in " + out}
	}
	return Identity{Name: strings.TrimSpace(out[:start]), Email: out[start+1 : end]}, nil
}

// holds reports whether the hand that commits holds a position: the
// author's or the committer's email equals it.
func (s *session) holds(email string) bool {
	return email != "" && (s.author.Email == email || s.committer.Email == email)
}

// checkIDs refuses an id that is no task at HEAD, before any tablo run.
func (s *session) checkIDs(ids ...string) Refusals {
	var out Refusals
	for _, id := range ids {
		if !idPattern.MatchString(id) || !s.ids[id] {
			out = append(out, Refusal{Code: "no-task", Task: id, Text: "names no task at HEAD"})
		}
	}
	return out
}

// The trailer keys a body line must not begin with, in any case.
var trailerKeys = []string{"Authorised", "Reviewed", "Reaffirmed", "Model", "Co-Authored-By"}

// checkOptions refuses a subject, a body or an agent that the message cannot
// carry (step 5).
func checkOptions(o Options, subject string) Refusals {
	var out Refusals
	if strings.ContainsAny(subject, "\r\n") || strings.TrimSpace(subject) == "" {
		out = append(out, Refusal{Code: "bad-message", Text: fmt.Sprintf("the subject %q is empty or holds a line break", subject)})
	}
	for _, l := range strings.Split(o.Body, "\n") {
		for _, key := range trailerKeys {
			if len(l) > len(key) && strings.EqualFold(l[:len(key)+1], key+":") {
				out = append(out, Refusal{Code: "body-trailer", Text: fmt.Sprintf("the body line %q reads as a trailer, which a blank line parts from the trailers", l)})
			}
		}
	}
	a := o.Agent
	switch {
	case a.CoAuthor != "" && a.Model == "":
		out = append(out, Refusal{Code: "bad-field", Text: "a co-author needs a model"})
	case strings.ContainsFunc(a.Model, func(r rune) bool { return r <= ' ' || r > '~' }):
		out = append(out, Refusal{Code: "bad-field", Text: fmt.Sprintf("the model %q holds a character that is not printable ASCII without a space", a.Model)})
	}
	if c := a.CoAuthor; c != "" && (strings.ContainsAny(c, "\r\n") || !strings.HasSuffix(c, ">") || !strings.Contains(c, " <")) {
		out = append(out, Refusal{Code: "bad-field", Text: fmt.Sprintf("the co-author %q is not of the form 'Name <email>'", c)})
	}
	return out
}

// request is one task trailer a plan asks the Source for.
type request struct{ kind, id, gate string }

// finish completes a plan whose request is resolved: it composes the
// message, proves the trailers, builds the tree and proves it (steps 5 to 8).
func (s *session) finish(ctx context.Context, p *Plan, o Options, trailers []request) (*Plan, error) {
	p.Branch, p.Base, p.Author, p.Committer = s.branch, s.base, s.author, s.committer
	if o.Subject != "" {
		p.Subject = o.Subject
	}
	p.Body = strings.TrimRight(o.Body, "\n")
	if refusals := checkOptions(o, p.Subject); len(refusals) > 0 {
		return nil, refusals
	}
	for _, t := range trailers {
		line, err := s.src.Trailer(ctx, t.kind, t.id, t.gate)
		if err != nil {
			return nil, err
		}
		key, value, _ := strings.Cut(line, ": ")
		p.Trailers = append(p.Trailers, Trailer{Key: key, Value: value})
	}
	if o.Agent.Model != "" {
		p.Trailers = append(p.Trailers, Trailer{Key: "Model", Value: o.Agent.Model})
	}
	if o.Agent.CoAuthor != "" {
		p.Trailers = append(p.Trailers, Trailer{Key: "Co-Authored-By", Value: o.Agent.CoAuthor})
	}
	if err := s.proveTrailers(ctx, p.Message(), p.Trailers); err != nil {
		return nil, err
	}
	sort.Slice(p.Files, func(i, j int) bool { return p.Files[i].Path < p.Files[j].Path })
	sort.Slice(p.Pins, func(i, j int) bool { return p.Pins[i].Path < p.Pins[j].Path })
	if len(p.Files) == 0 && len(p.Pins) == 0 {
		tree, err := s.git(ctx, "rev-parse", s.base+"^{tree}")
		p.Tree = tree
		return p, err
	}
	if err := s.buildTree(ctx, p); err != nil {
		return nil, err
	}
	if err := s.proveTree(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// proveTrailers runs git interpret-trailers --parse on the message, which
// must return exactly the planned trailers (step 6).
func (s *session) proveTrailers(ctx context.Context, message string, want []Trailer) error {
	got, err := parseTrailers(ctx, s.r, s.top, message)
	if err != nil {
		return err
	}
	planned := make([]string, len(want))
	for i, t := range want {
		planned[i] = t.Key + ": " + t.Value
	}
	if strings.Join(got, "\n") != strings.Join(planned, "\n") {
		return &ExecError{Tool: "git", Args: []string{"interpret-trailers", "--parse"}, Exit: 0,
			Stderr: fmt.Sprintf("Git reads the trailers %q where the plan has %q", got, planned)}
	}
	return nil
}

// parseTrailers returns the trailer lines Git reads in a message.
func parseTrailers(ctx context.Context, r *Repo, dir, message string) ([]string, error) {
	out, err := r.git(ctx, dir, []byte(message), nil, "interpret-trailers", "--parse")
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, l := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

// buildTree checks that each file the plan writes is clean, then writes the
// planned tree through a private index and takes git's diff to it (step 7).
// It writes unreachable objects and nothing else.
func (s *session) buildTree(ctx context.Context, p *Plan) error {
	var dirty Refusals
	for _, f := range p.Files {
		if f.Old == nil {
			if _, err := os.Lstat(filepath.Join(s.top, filepath.FromSlash(f.Path))); err == nil {
				dirty = append(dirty, Refusal{Code: "dirty", Text: f.Path + " exists in the working tree; move it away"})
			}
			continue
		}
		out, err := s.git(ctx, "status", "--porcelain", "--", f.Path)
		if err != nil {
			return err
		}
		if out != "" {
			dirty = append(dirty, Refusal{Code: "dirty", Text: f.Path + " holds changes HEAD lacks; commit or restore them"})
		}
	}
	if len(dirty) > 0 {
		return dirty
	}
	tmp, err := os.MkdirTemp("", "tabloio-index-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	index := []string{"GIT_INDEX_FILE=" + filepath.Join(tmp, "index")}
	run := func(stdin []byte, args ...string) (string, error) {
		out, err := s.r.git(ctx, s.top, stdin, index, args...)
		return strings.TrimSuffix(out, "\n"), err
	}
	if _, err := run(nil, "read-tree", "HEAD"); err != nil {
		return err
	}
	for _, f := range p.Files {
		blob, err := run(f.New, "hash-object", "-w", "--stdin", "--path="+f.Path)
		if err != nil {
			return err
		}
		if _, err := run(nil, "update-index", "--add", "--cacheinfo", "100644,"+blob+","+f.Path); err != nil {
			return err
		}
	}
	for _, pin := range p.Pins {
		if _, err := run(nil, "update-index", "--add", "--cacheinfo", "160000,"+pin.New+","+pin.Path); err != nil {
			return err
		}
	}
	if p.Tree, err = run(nil, "write-tree"); err != nil {
		return err
	}
	diff, err := s.r.git(ctx, s.top, nil, nil, "diff", "--no-color", "--no-ext-diff", "--no-textconv", "--submodule=short",
		"--abbrev=7", "--src-prefix=a/", "--dst-prefix=b/", "-U3", s.base+"^{tree}", p.Tree)
	p.Diff = diff
	return err
}

// proveTree validates the tree of Base and the planned tree: an error the
// second has and the first lacks is the refusal invalid, S11 apart, since a
// bare tree has no history (step 8).
func (s *session) proveTree(ctx context.Context, p *Plan) error {
	baseTree, err := s.git(ctx, "rev-parse", s.base+"^{tree}")
	if err != nil {
		return err
	}
	before, err := s.src.Validate(ctx, baseTree)
	if err != nil {
		return err
	}
	after, err := s.src.Validate(ctx, p.Tree)
	if err != nil {
		return err
	}
	type key struct{ code, path, message string }
	had := map[key]bool{}
	for _, d := range before {
		had[key{d.Code, d.Path, d.Message}] = true
	}
	var out Refusals
	for _, d := range after {
		if d.Severity != "error" || d.Code == "S11" || had[key{d.Code, d.Path, d.Message}] {
			continue
		}
		text := d.Message
		if d.Path != "" {
			text = d.Path + ": " + text
		}
		out = append(out, Refusal{Code: "invalid", Task: strings.Join(p.Tasks, " "), Rule: d.Code, Text: text})
	}
	if len(out) > 0 {
		return out
	}
	return nil
}
