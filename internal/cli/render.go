package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/nbyoung/tabloio/brief"
	"github.com/nbyoung/tabloio/internal/render"
	rtext "github.com/nbyoung/tabloio/internal/render/text"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// renderFunc renders one view's data at one level. linkBase is the path
// from the directory the output lands in to the project directory, with
// forward slashes, and "" when they are the same.
type renderFunc func(data any, level string, linkBase string) ([]byte, error)

// textFunc renders one view's data at one level as Unicode text. linkBase
// is as renderFunc has it.
type textFunc func(data any, level string, linkBase string, o rtext.Options) ([]byte, error)

// markdown and text map a view to its renderer. The key "brief", in the
// markdown table alone, holds the renderer of the Agent briefs task (e4c7):
// the brief is the same bytes under both formats, so --text --brief reads
// the markdown entry too.
var markdown = map[string]renderFunc{
	"brief":      viaBrief,
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

var text = map[string]textFunc{
	"gates":      viaText[*view.Gates]("gates", nil),
	"task":       viaText[*view.Task]("task", nil),
	"authority":  viaText[*view.Delegation]("authority", nil),
	"assignment": viaText[*view.Assignment]("assignment", nil),
	"queue":      viaText[*view.Queue]("queue", nil),
	"blockage":   viaText[*view.Blockage]("blockage", nil),
	"tableau":    viaText("tableau", func(t *view.Tableau) bool { return !t.Contextual }),
	"context":    viaText("context", func(t *view.Tableau) bool { return t.Contextual }),
	"history":    viaText[*view.History]("history", nil),
	"audit":      viaText[*view.Audit]("audit", nil),
}

// viaBrief renders a brief: data is tablo's whole envelope of the queue view
// with the brief parameter, as JSON bytes (Result), which brief.Decode reads.
// The request comes from the envelope's params, which state the person and
// the item asked for. The level and the link base change nothing: a brief is
// the provenance of one item and holds no file link. A *brief.NoWorkError
// and a *brief.WaitingError come back as they stand, for the flow to report.
func viaBrief(data any, _, _ string) ([]byte, error) {
	b, ok := data.([]byte)
	if !ok {
		return nil, fmt.Errorf("the brief cannot render %T", data)
	}
	var env struct {
		Params struct {
			Person string `json:"person"`
			Brief  struct {
				Task string `json:"task"`
				Gate string `json:"gate"`
			} `json:"brief"`
		} `json:"params"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("brief: envelope: %w", err)
	}
	d, err := brief.Decode(b, brief.Request{
		Person: env.Params.Person, Task: env.Params.Brief.Task, Gate: env.Params.Brief.Gate,
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := brief.Render(&buf, d); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// viaRender returns the Markdown renderer of one view. Package render
// dispatches by the concrete type of the data, so every entry reaches the
// same function, render.Render; the entry only checks that the data is the
// type its view returns, and, where fits is set, that it is the right one of
// two views that share a type (tableau and context).
func viaRender[V view.View](name string, fits func(V) bool) renderFunc {
	return func(data any, level, linkBase string) ([]byte, error) {
		return draw(name, fits, data, level, linkBase, render.Options{Format: render.Markdown})
	}
}

// viaText returns the text renderer of one view: draw with Format Text and
// the options.
func viaText[V view.View](name string, fits func(V) bool) textFunc {
	return func(data any, level, linkBase string, o rtext.Options) ([]byte, error) {
		return draw(name, fits, data, level, linkBase, render.Options{Format: render.Text, Text: o})
	}
}

// draw is what viaRender and viaText share: the type check, the level, the
// recovery of a builder's panic, the link base as Links.Project ("." for
// ""), and the call of render.Render. It sets Level and Links on o.
//
// A builder of package render that cannot return its error panics with it;
// draw turns that panic back into the error.
func draw[V view.View](name string, fits func(V) bool, data any, level, linkBase string, o render.Options) (out []byte, err error) {
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
	o.Level = l
	o.Links = render.Links{Project: project}
	var buf bytes.Buffer
	if err := render.Render(&buf, v, o); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
