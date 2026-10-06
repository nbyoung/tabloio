package render

import (
	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// frame is what a view says about its page frame; the builder fills it in.
type frame struct {
	Name       string       // the view's command name: gates, task, ...
	Positional string       // the command's positional argument, if any
	Title      []doc.Inline // after `# `
	Question   string       // the view's question, in bold
	Legend     bool         // the legend line: every view but the gate definition
	SeeAlso    []string     // related views, in the order the view's design lists
	Body       []doc.Block  // the blocks of the level, with headings from ##
}

// frame builds the page frame (design e3ed, "The page frame"), in this order:
// the title, the parameter line, the legend line, the blocks, the See also
// line at detail and provenance, and the closing command.
func (p *page) frame(f frame) doc.Doc {
	blocks := []doc.Block{heading(1, f.Title...), p.paramLine(f.Question)}
	if f.Legend {
		blocks = append(blocks, para(spans(one(txt("Legend: ")), p.viewRef("gates"), one(txt(".")))...))
	}
	blocks = append(blocks, f.Body...)
	if p.o.Level >= view.Detail && p.namesAny(f.SeeAlso) {
		var runs [][]doc.Inline
		for _, name := range f.SeeAlso {
			runs = append(runs, p.viewRef(name))
		}
		blocks = append(blocks, para(spans(one(txt("See also: ")), joined(", ", runs...), one(txt(".")))...))
	}
	blocks = append(blocks, para(txt("Command: "), code(command(f.Name, f.Positional, p.h, p.o))))
	return doc.Doc{Blocks: blocks}
}

// namesAny reports whether Links.Views names any of the views.
func (p *page) namesAny(names []string) bool {
	for _, n := range names {
		if _, ok := p.o.Links.Views[n]; ok {
			return true
		}
	}
	return false
}

// paramLine is the second line of the frame: the question in bold, then,
// joined by " · ", the project, the ref, the parameters in force and the level.
func (p *page) paramLine(question string) doc.Para {
	proj := p.h.Project
	var runs [][]doc.Inline
	switch {
	case proj.Title != "":
		runs = append(runs, one(txt(proj.Title)))
	case proj.ID != "":
		runs = append(runs, one(code(proj.ID)))
	}
	runs = append(runs, p.refSpans())
	runs = append(runs, p.paramSpans()...)
	runs = append(runs, one(txt("level "+p.o.Level.String())))
	return para(spans(one(doc.Strong{txt(question)}, txt(" ")), joined(" · ", runs...))...)
}
