package action

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Apply makes the planned commit. It returns a Refusal with the code stale
// when the checkout no longer stands where the plan read it, and the Result
// with the refusal altered when a hook rewrote the commit. On a commit that
// fails it writes each file back, removes each file it created, and returns
// the *ExecError with Git's own text. It never amends.
func (r *Repo) Apply(ctx context.Context, p *Plan) (Result, error) {
	dir := r.Dir
	if dir == "" {
		dir = "."
	}
	out, err := r.git(ctx, dir, nil, nil, "rev-parse", "--show-toplevel")
	if err != nil {
		return Result{}, err
	}
	top := strings.TrimSuffix(out, "\n")
	git := func(args ...string) (string, error) {
		out, err := r.git(ctx, top, nil, nil, args...)
		return strings.TrimSuffix(out, "\n"), err
	}
	if err := r.standing(ctx, top, p); err != nil {
		return Result{}, err
	}

	var created, dirs []string
	restore := func() {
		for _, f := range p.Files {
			full := filepath.Join(top, filepath.FromSlash(f.Path))
			if f.Old != nil {
				_ = os.WriteFile(full, f.Old, 0o644)
			}
		}
		for _, path := range created {
			_, _ = git("reset", "-q", "--", path)
			_ = os.Remove(filepath.Join(top, filepath.FromSlash(path)))
		}
		for i := len(dirs) - 1; i >= 0; i-- {
			_ = os.Remove(dirs[i]) // only when empty
		}
	}
	for _, f := range p.Files {
		full := filepath.Join(top, filepath.FromSlash(f.Path))
		dirs = append(dirs, missingDirs(filepath.Dir(full))...)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			restore()
			return Result{}, err
		}
		if err := os.WriteFile(full, f.New, 0o644); err != nil {
			restore()
			return Result{}, err
		}
		if f.Old == nil {
			created = append(created, f.Path)
			if _, err := git("add", "--intent-to-add", "--", f.Path); err != nil {
				restore()
				return Result{}, err
			}
		}
	}

	args := []string{"commit", "--cleanup=verbatim", "--file=-", "--only"}
	var paths []string
	for _, f := range p.Files {
		paths = append(paths, f.Path)
	}
	for _, pin := range p.Pins {
		paths = append(paths, pin.Path)
	}
	if len(paths) == 0 {
		args = append(args, "--allow-empty")
	} else {
		args = append(append(args, "--"), paths...)
	}
	if _, err := r.git(ctx, top, []byte(p.Message()), nil, args...); err != nil {
		restore()
		return Result{}, err
	}

	commit, err := git("rev-parse", "HEAD")
	if err != nil {
		return Result{}, err
	}
	res := Result{Commit: commit}
	tree, err := git("rev-parse", "HEAD^{tree}")
	if err != nil {
		return res, err
	}
	message, err := git("log", "-1", "--format=%B", "HEAD")
	if err != nil {
		return res, err
	}
	got, err := parseTrailers(ctx, r, top, message)
	if err != nil {
		return res, err
	}
	want := make([]string, len(p.Trailers))
	for i, t := range p.Trailers {
		want[i] = t.Key + ": " + t.Value
	}
	switch {
	case tree != p.Tree:
		return res, Refusal{Code: "altered", Text: fmt.Sprintf("a hook changed the tree of commit %s from the planned %s", short(commit), short(p.Tree))}
	case strings.Join(got, "\n") != strings.Join(want, "\n"):
		return res, Refusal{Code: "altered", Text: fmt.Sprintf("a hook changed the trailers of commit %s to %q", short(commit), got)}
	}
	return res, nil
}

// standing checks that HEAD is the plan's base, that each file still holds
// its old bytes or does not exist, and that each pinned submodule still
// stands at its new commit.
func (r *Repo) standing(ctx context.Context, top string, p *Plan) error {
	stale := func(format string, args ...any) error {
		return Refusal{Code: "stale", Text: fmt.Sprintf(format, args...) + "; plan again"}
	}
	head, err := r.git(ctx, top, nil, nil, "rev-parse", "--verify", "--quiet", "HEAD^{commit}")
	if head = strings.TrimSuffix(head, "\n"); err != nil || head != p.Base {
		return stale("HEAD stands at %s, not at %s", short(head), short(p.Base))
	}
	for _, f := range p.Files {
		b, err := os.ReadFile(filepath.Join(top, filepath.FromSlash(f.Path)))
		switch {
		case f.Old == nil && !errors.Is(err, fs.ErrNotExist):
			return stale("%s exists", f.Path)
		case f.Old != nil && (err != nil || !bytes.Equal(b, f.Old)):
			return stale("%s changed", f.Path)
		}
	}
	for _, pin := range p.Pins {
		at, err := r.git(ctx, filepath.Join(top, filepath.FromSlash(pin.Path)), nil, nil, "rev-parse", "HEAD")
		if at = strings.TrimSuffix(at, "\n"); err != nil || at != pin.New {
			return stale("the submodule %s stands at %s, not at %s", pin.Path, short(at), short(pin.New))
		}
	}
	return nil
}

// missingDirs returns the directories from the first that does not exist
// down to dir, outermost first.
func missingDirs(dir string) []string {
	var out []string
	for {
		if _, err := os.Stat(dir); err == nil {
			break
		}
		out = append([]string{dir}, out...)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return out
}
