// Package doc is the format-neutral document a view builds.
//
// A document holds headings, paragraphs, tables, lists and preformatted
// blocks of inline spans. It carries no format: a writer, such as the
// Markdown writer, turns it into bytes.
package doc

// Doc is a sequence of blocks.
type Doc struct{ Blocks []Block }

// Block is one of Heading, Para, Table, List, Pre.
type Block interface{ block() }

// Heading is a title or a section heading.
type Heading struct {
	Level int // 1 for the title, 2 and 3 below it
	Text  []Inline
}

// Para is a paragraph.
type Para struct{ Text []Inline }

// Table is a table with one header row.
type Table struct {
	Align []Align // one per column
	Head  []Cell
	Rows  [][]Cell
}

// Cell is a table cell.
type Cell struct {
	Indent int // tree depth, in steps
	Text   []Inline
}

// List is a bulleted or numbered list.
type List struct {
	Ordered bool
	Items   []Item
}

// Item is a list item: its text, then nested lists and preformatted blocks.
type Item struct {
	Text   []Inline
	Blocks []Block // nested lists and preformatted blocks
}

// Pre is a preformatted block, drawn as given: a tree, a brief, commands.
type Pre struct{ Lines []string }

func (Heading) block() {}
func (Para) block()    {}
func (Table) block()   {}
func (List) block()    {}
func (Pre) block()     {}

// Inline is one of Text, Code, Symbol, Strong, Link.
type Inline interface{ inline() }

// Text is project text or fixed text; a writer escapes it.
type Text string

// Code is an id, a key, a model, a hash, a path or a trailer.
type Code string

// Symbol is a gate, state, reason or mark symbol, as the legend gives it.
type Symbol string

// Strong is emphasised text.
type Strong []Inline

// Link is text that leads to a URL.
type Link struct {
	Text []Inline
	URL  string
}

func (Text) inline()   {}
func (Code) inline()   {}
func (Symbol) inline() {}
func (Strong) inline() {}
func (Link) inline()   {}

// Align is the alignment of a table column.
type Align int

// The alignments of a column.
const (
	Left Align = iota
	Centre
	Right
)
