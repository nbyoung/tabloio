package render

import (
	"strconv"
	"strings"

	"github.com/nbyoung/tabloio/internal/render/doc"
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
)

// refName is the ref as given, with a range as <from>..<name>; HEAD stands
// for an empty name.
func refName(r view.Ref) string {
	name := r.Name
	if name == "" {
		name = "HEAD"
	}
	if r.From != "" {
		return r.From + ".." + name
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
// so the command reproduces the bytes; below it, the ref as given, so a
// snapshot does not change with every push. A view with a positional
// argument never repeats it as --task.
func command(name, positional string, h *view.Head, o Options) string {
	parts := []string{"tabloio", name}
	if positional != "" {
		parts = append(parts, positional)
	}
	par := h.Params
	if par.Task != "" && positional == "" {
		parts = append(parts, flagTask, par.Task)
	}
	if par.Person != "" {
		parts = append(parts, flagPerson, par.Person)
	}
	ref := refName(h.Ref)
	if o.Level >= view.Provenance && h.Ref.Commit != "" {
		ref = short(h.Ref.Commit)
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
	return strings.Join(parts, " ")
}

// refSpans prints the ref of the parameter line: ref `<name>`, with the
// commit and date at provenance or with Options.Stamp, and `, off the
// trunk` for a ref off the trunk.
func (p *page) refSpans() []doc.Inline {
	r := p.h.Ref
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
func (p *page) paramSpans() [][]doc.Inline {
	var out [][]doc.Inline
	par := p.h.Params
	if par.Task != "" {
		out = append(out, one(txt("task "), code(par.Task)))
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
