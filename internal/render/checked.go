package render

import (
	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// buildFailure is the panic value of a builder that returns an error.
//
// The registry of render.go holds builders that cannot fail, and the status
// renderers have two errors to report: a grid row with the wrong number of
// cells, and a queue whose brief the hook of A6 cannot yet write. The change
// that lets a builder return its error belongs to render.go, which this task
// does not edit; until then registerStatus adapts a builder that returns an
// error to the registry by panicking with this value. View data that came
// through view.Decode never reaches the first error, and no command sets a
// queue's brief before the Agent briefs task wires the hook.
type buildFailure struct{ err error }

func (f buildFailure) Error() string { return f.err.Error() }
func (f buildFailure) Unwrap() error { return f.err }

// registerStatus adds a builder that returns an error to Render's registry.
func registerStatus[V view.View](build func(V, Options) (doc.Doc, error)) {
	register(func(v V, o Options) doc.Doc {
		d, err := build(v, o)
		if err != nil {
			panic(buildFailure{err})
		}
		return d
	})
}
