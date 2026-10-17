package cli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nbyoung/tablo"

	"github.com/nbyoung/tabloio/internal/render/view"
)

// Source gives view data. The adapter to tablo's library implements it; a
// test gives a fake.
type Source interface {
	// View derives one view. It returns no Data only when an error
	// diagnostic prevents the view.
	View(ctx context.Context, q Query) (Result, error)
	// Viewer returns the email Git commits with in dir, or "" when Git has none.
	Viewer(ctx context.Context, dir string) (string, error)
}

// Query names one view and every parameter in force. A zero field means
// the parameter is not set.
type Query struct {
	Dir        string // -C; "" is the current directory
	View       string // the command's name: "gates" … "audit"
	Ref        string // "" leaves tablo's default; "from..to" on history
	Task       string
	Person     string // the --person value or the viewer; "" is no person
	Window     *int
	Columns    []string
	Historical bool
	Proposed   bool
	Stale      *int // nil leaves tablo's default, seven days
	Brief      *Brief
	Level      string // always set
}

// Brief names the one queue item a brief is written for.
type Brief struct{ Task, Gate string }

// Result is what tablo returns: the view's data in tabloio's view types
// (package render/view), and the diagnostics of the project. For a brief,
// Data holds tablo's whole envelope as JSON bytes, which package brief
// decodes, since a brief reads the envelope's source and params beside its
// data.
type Result struct {
	Data        any
	Diagnostics []Diagnostic
}

// Diagnostic mirrors a diagnostic of tablo's envelope.
type Diagnostic struct {
	Severity, Code, Path string
	Line, Col            int
	Message              string
}

// UsageFault is an error of a Source that stands for a wrong command line:
// the library refuses parameters the command line checks did not catch. The
// read flow reports it as a usage error, with exit 2.
type UsageFault struct{ Msg string }

// Error returns the library's message.
func (e *UsageFault) Error() string { return e.Msg }

// today returns the host's local date, YYYY-MM-DD, which the audit counts
// staleness from; a test replaces it. It is the one place the clock enters.
var today = func() string { return time.Now().Format("2006-01-02") }

// tabloSource is the Source over tablo's library, run in this process
// (design 5ca9, D1). It is the one file of the module that imports tablo.
type tabloSource struct{}

// NewSource returns the Source that reads through tablo's library.
func NewSource() Source { return tabloSource{} }

// View opens the repository at q.Dir, loads q.Ref and derives the view. It
// decodes tablo's whole envelope with view.Decode, or, for a brief, hands
// the envelope's bytes on as they stand, and mirrors the envelope's
// diagnostics. An envelope with no data gives a Result without Data.
func (tabloSource) View(ctx context.Context, q Query) (Result, error) {
	dir := q.Dir
	if dir == "" {
		dir = "."
	}
	p, ref := tabloParams(q)

	repo, err := tablo.Open(ctx, dir, tablo.Options{})
	if err != nil {
		return Result{}, fault(err)
	}
	snap, err := repo.Load(ctx, ref)
	if err != nil {
		return Result{}, fault(err)
	}
	env, err := snap.View(ctx, tablo.View(q.View), p)
	if err != nil {
		return Result{}, fault(err)
	}
	res := Result{}
	for _, d := range env.Diagnostics {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Severity: d.Severity, Code: d.Code, Path: d.Path,
			Line: d.Line, Col: d.Col, Message: d.Message,
		})
	}
	if env.Data == nil {
		return res, nil
	}
	if q.Brief != nil {
		res.Data = env.JSON()
		return res, nil
	}
	v, err := view.Decode(q.View, env.JSON())
	if err != nil {
		return Result{}, err
	}
	res.Data = v
	return res, nil
}

// tabloParams maps a query to tablo's parameters and the ref to load: a
// history range gives its from side to Params.From and loads its to side.
func tabloParams(q Query) (tablo.Params, string) {
	p := tablo.Params{
		Task:       q.Task,
		Person:     q.Person,
		Window:     q.Window,
		Columns:    q.Columns,
		Historical: q.Historical,
		Proposed:   q.Proposed,
	}
	if q.Brief != nil {
		// tablo refuses a level beside a brief; the brief opens at
		// provenance in tablo's own rule.
		p.Brief = &tablo.Brief{Task: q.Brief.Task, Gate: q.Brief.Gate}
	} else {
		p.Level = tablo.Level(q.Level)
	}
	if q.Stale != nil {
		p.Stale = *q.Stale
	}
	ref := q.Ref
	if q.View == string(tablo.History) {
		if from, to, ok := strings.Cut(ref, ".."); ok {
			p.From, ref = from, to
		}
	}
	if q.View == string(tablo.Audit) {
		p.Now = today()
	}
	return p, ref
}

// fault turns a *tablo.UsageError into a *UsageFault and leaves any other
// error as it stands.
func fault(err error) error {
	var u *tablo.UsageError
	if errors.As(err, &u) {
		return &UsageFault{Msg: u.Msg}
	}
	return err
}

// Viewer reads GIT_AUTHOR_EMAIL when it is set and not empty, else
// `git -C dir config --get user.email`, else "": the rule tablo's own
// command applies. A value that is no email address counts as none. It is
// the one git command tabloio runs itself, and it reads.
func (tabloSource) Viewer(ctx context.Context, dir string) (string, error) {
	email := os.Getenv("GIT_AUTHOR_EMAIL")
	if email == "" {
		if dir == "" {
			dir = "."
		}
		out, err := exec.CommandContext(ctx, "git", "-C", dir, "config", "--get", "user.email").Output()
		var exit *exec.ExitError
		switch {
		case errors.As(err, &exit):
			// Git exits 1 for an unset key; any other status means no
			// repository or no usable configuration, which is no viewer.
			return "", nil
		case err != nil:
			return "", err
		}
		email = strings.TrimSpace(string(out))
	}
	if (tablo.Params{Viewer: email}).Check(tablo.Gates) != nil {
		return "", nil
	}
	return email, nil
}
