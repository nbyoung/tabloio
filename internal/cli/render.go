package cli

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/nbyoung/tabloio/internal/render"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// renderFunc renders one view's data at one level. linkBase is the path
// from the directory the output lands in to the project directory, with
// forward slashes, and "" when they are the same.
type renderFunc func(data any, level string, linkBase string) ([]byte, error)

// markdown and text map a view to its renderer. The key "brief" is for the
// renderer of the Agent briefs task (e4c7), which fills it with the key of
// the markdown table. text starts empty; the Text renderers task (e0f7)
// fills it.
var markdown = map[string]renderFunc{
	"gates":      viaRender[*view.Gates]("gates", nil),
	"task":       viaRender[*view.Task]("task", nil),
	"authority":  viaRender[*view.Delegation]("authority", nil),
	"assignment": viaRender[*view.Assignment]("assignment", nil),
	"queue":      viaRender[*view.Queue]("queue", nil),
	"blockage":   viaRender[*view.Blockage]("blockage", nil),
	"tableau":    viaRender("tableau", func(t *view.Tableau) bool { return !t.Contextual }),
	"context":    viaRender("context", func(t *view.Tableau) bool { return t.Contextual }),
	"history":    viaRender[*view.History]("history", nil),
	"audit":      viaRender[*view.Audit]("audit", nil),
}

var text = map[string]renderFunc{}

// viaRender returns the Markdown renderer of one view. Package render
// dispatches by the concrete type of the data, so every entry reaches the
// same function, render.Render; the entry only checks that the data is the
// type its view returns, and, where fits is set, that it is the right one of
// two views that share a type (tableau and context).
//
// A builder of package render that cannot return its error panics with it;
// the entry turns that panic back into the error.
func viaRender[V view.View](name string, fits func(V) bool) renderFunc {
	return func(data any, level, linkBase string) (out []byte, err error) {
		v, ok := data.(V)
		if !ok || (fits != nil && !fits(v)) {
			return nil, fmt.Errorf("the %s view cannot render %T", name, data)
		}
		l, err := view.ParseLevel(level)
		if err != nil {
			return nil, err
		}
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			if e, ok := r.(error); ok {
				if _, runtimeErr := r.(runtime.Error); !runtimeErr {
					out, err = nil, e
					return
				}
			}
			panic(r)
		}()
		// Links.Project "" writes no file link, where a link base "" says
		// the output stands in the project directory.
		project := linkBase
		if project == "" {
			project = "."
		}
		var buf bytes.Buffer
		err = render.Render(&buf, v, render.Options{
			Format: render.Markdown,
			Level:  l,
			Links:  render.Links{Project: project},
		})
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}
