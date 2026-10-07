package cli

import "context"

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
// (package render/view), and the diagnostics of the project.
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
