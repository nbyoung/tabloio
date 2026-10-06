package render

import (
	"errors"
	"io"
	"reflect"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/markdown"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// Format names an output form.
type Format int

// The output forms.
const (
	Markdown Format = iota // the default
	Text                   // Unicode drawing; Render returns ErrFormat until task e0f7 lands
)

// Options holds what the command line fixes for one rendering.
type Options struct {
	Format Format
	Level  view.Level // the level to render; ErrLevel when the data holds less
	Stamp  bool       // name the ref's commit and date below provenance too (R12)
	Unfold bool       // draw every row: no run fold, no collapsed children (R10, R11)
	Links  Links
}

// Links says where a rendering's links lead (R13).
type Links struct {
	Project string            // the path from the output's directory to the directory that holds .tableaux; "" writes no file link
	Views   map[string]string // view name to link target, for example "gates": "GATES.md"
}

// The errors Render returns; a write error passes through unwrapped.
var (
	ErrFormat = errors.New("render: format not built")
	ErrLevel  = errors.New("render: the view data holds less than the level asked for")
	ErrView   = errors.New("render: unknown view")
)

// builders maps the concrete type of a view to its builder. Each view's file
// registers its builder in init, so a new view changes no shared file.
var builders = map[reflect.Type]func(view.View, Options) doc.Doc{}

// register adds the builder of the view type V to Render.
func register[V view.View](build func(V, Options) doc.Doc) {
	builders[reflect.TypeFor[V]()] = func(v view.View, o Options) doc.Doc {
		return build(v.(V), o)
	}
}

// Render writes one view at one level. It is a pure function of v and o.
func Render(w io.Writer, v view.View, o Options) error {
	d, err := buildDoc(v, o)
	if err != nil {
		return err
	}
	return markdown.Write(w, d)
}

// buildDoc checks the options against the view and returns the view's
// document. Render and the tests of the document model share it.
func buildDoc(v view.View, o Options) (doc.Doc, error) {
	if o.Format != Markdown {
		return doc.Doc{}, ErrFormat
	}
	if v == nil {
		return doc.Doc{}, ErrView
	}
	t := reflect.TypeOf(v)
	build, ok := builders[t]
	if !ok || (t.Kind() == reflect.Pointer && reflect.ValueOf(v).IsNil()) {
		return doc.Doc{}, ErrView
	}
	if o.Level < view.Glance || o.Level > v.Header().Level {
		return doc.Doc{}, ErrLevel
	}
	return build(v, o), nil
}
