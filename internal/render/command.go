package render

import (
	"strconv"
	"strings"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/text"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// The names of the command line's flags, as tabloio and tablo spell them
// (design 5ca9, decision 2; design e3ed, decision 13).
const (
	flagTask       = "--task"
	flagPerson     = "--person"
	flagRef        = "--ref"
	flagWindow     = "--window"
	flagColumns    = "--columns"
	flagHistorical = "--historical"
	flagProposed   = "--proposed"
	flagStale      = "--stale"
	flagLevel      = "--level"
	flagText       = "--text"
	flagWidth      = "--width"
)

// refName is the ref as given, with a range as <from>..<name>; HEAD stands
// for an empty name.
func refName(r view.Source) string {
	name := r.Ref
	if name == "" {
		name = "HEAD"
	}
	if r.From != nil && r.From.Ref != "" {
		return r.From.Ref + ".." + name
	}
	return name
}

// command is the canonical command that reproduces a rendering (R12): not
// the words the user typed, but `tabloio`, the view's command, its
// positional argument, then the flags in force, each only when in force.
//
// The flags follow the order R12 states, which the queue mockup, the agent
// briefs and the Status renderers share (the owner's ruling of 2026-10-06;
// gates.md and task.md draw --ref first): --task, --person, --ref, --window
// or --columns, --historical, --proposed, --stale, --level. The --ref flag
// is absent when the name is HEAD. At provenance the flag names the commit,
// and a range its two commits, so the command reproduces the bytes; below
// it, the ref as given, so a snapshot does not change with every push. A view with a positional
// argument never repeats it as --task.
func command(name, positional string, h *view.Head, o Options) string {
	parts := []string{"tabloio", name}
	if positional != "" {
		parts = append(parts, positional)
	}
	par := h.Params
	if task := taskInForce(name, h); task != "" && positional == "" {
		parts = append(parts, flagTask, task)
	}
	if par.Person != "" {
		parts = append(parts, flagPerson, par.Person)
	}
	ref := refName(h.Source)
	if o.Level >= view.Provenance && h.Source.Commit != "" {
		ref = short(h.Source.Commit)
		if f := h.Source.From; f != nil && f.Commit != "" {
			ref = short(f.Commit) + ".." + ref
		} else if f != nil && f.Ref != "" {
			ref = f.Ref + ".." + ref
		}
	}
	if ref != "HEAD" {
		parts = append(parts, flagRef, ref)
	}
	if par.Window != nil {
		parts = append(parts, flagWindow, strconv.Itoa(*par.Window))
	} else if len(par.Columns) > 0 {
		parts = append(parts, flagColumns, strings.Join(par.Columns, ","))
	}
	if par.Historical {
		parts = append(parts, flagHistorical)
	}
	if par.Proposed {
		parts = append(parts, flagProposed)
	}
	if par.Stale > 0 {
		parts = append(parts, flagStale, strconv.Itoa(par.Stale))
	}
	parts = append(parts, flagLevel, o.Level.String())
	if o.Format == Text {
		parts = append(parts, flagText)
		if w := o.Text.Width; w != 0 && w != text.DefaultWidth {
			parts = append(parts, flagWidth, strconv.Itoa(w))
		}
	}
	return strings.Join(parts, " ")
}

// rootByDefault holds the views whose task defaults to the project's root:
// tablo echoes the root in params when no task is given, so the root there
// focuses nothing.
var rootByDefault = map[string]bool{
	"authority": true, "assignment": true, "queue": true,
	"blockage": true, "history": true, "audit": true,
}

// taskInForce returns the task parameter of a view, or "" where it is the
// project's root and the view's default.
func taskInForce(name string, h *view.Head) string {
	task := h.Params.Task
	if rootByDefault[name] && task != "" && task == h.Project.Root {
		return ""
	}
	return task
}

// refSpans prints the ref of the parameter line: ref `<name>`, with the
// commit and date at provenance or with Options.Stamp, and `, off the
// trunk` for a ref off the trunk.
func (p *page) refSpans() []doc.Inline {
	r := p.h.Source
	out := one(txt("ref "), code(refName(r)))
	if (p.o.Level >= view.Provenance || p.o.Stamp) && r.Commit != "" {
		out = append(out, txt(" at "), code(short(r.Commit)))
		if r.Date != "" {
			out = append(out, txt(", "+r.Date))
		}
	}
	if !r.OnTrunk {
		out = append(out, txt(", off the trunk"))
	}
	return out
}

// paramSpans prints the parameters in force in the order of VIEWS.md's
// table: task, person, window or columns, historical junctions, proposed,
// stale.
func (p *page) paramSpans(name string) [][]doc.Inline {
	var out [][]doc.Inline
	par := p.h.Params
	if task := taskInForce(name, p.h); task != "" {
		out = append(out, one(txt("task "), code(task)))
	}
	if par.Person != "" {
		out = append(out, one(txt("person "+par.Person)))
	}
	if par.Window != nil {
		out = append(out, one(txt("window "+strconv.Itoa(*par.Window))))
	} else if len(par.Columns) > 0 {
		out = append(out, one(txt("columns "+strings.Join(par.Columns, ","))))
	}
	if par.Historical {
		out = append(out, one(txt("historical junctions on")))
	}
	if par.Proposed {
		out = append(out, one(txt("proposed only")))
	}
	if par.Stale > 0 {
		out = append(out, one(txt("stale "+strconv.Itoa(par.Stale)+" days")))
	}
	return out
}
