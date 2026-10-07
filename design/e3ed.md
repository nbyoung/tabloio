# Design e3ed: Legend and task renderers

This task delivers the architecture the four renderer tasks share under `internal/render/`, and on it the two views that every other view links to: the gate definition and the task definition. Four choices shape it. The renderers take input types of their own, which mirror tablo's view data as JSON and import nothing from tablo (D1). A view builds a document of headings, tables, lists and preformatted blocks, and a writer per format turns that document into bytes, so that the text format (`e0f7`) adds a writer and no view (D2). One set of rules, R1 to R17 below, holds across the ten views where the ten mockups differ (D3 to D8). The renderer emits what the data states and no commentary (D9).

The structural (`c48a`), status (`a3cc`) and temporal (`9167`) designs refer here for everything under [The shared architecture](#the-shared-architecture); each carries its own views, tests and decisions.

| Draft                                              | Holds                                                                                   |
|----------------------------------------------------|-----------------------------------------------------------------------------------------|
| [`e3ed/gates.json`](e3ed/gates.json)               | A fixture: the gate definition data of the Tableaux tooling plan, in the input types' JSON |
| [`e3ed/gates.glance.md`](e3ed/gates.glance.md)     | The golden output for that fixture at glance                                            |
| [`e3ed/gates.detail.md`](e3ed/gates.detail.md)     | The golden output at detail                                                             |
| [`e3ed/task.detail.md`](e3ed/task.detail.md)       | The golden output of the task definition of `e9c6` at detail, with a person marked      |

A draft fixes bytes: where a draft and a rule below seem to differ, the implementation reports it and changes neither.

## The shared architecture

### Packages

| Path                        | Holds                                                                                          |
|-----------------------------|------------------------------------------------------------------------------------------------|
| `internal/render/`          | `Render`, `Options`, the rules R1 to R17 as helpers, and one builder file per view (`gates.go`, `task.go`, and so on) |
| `internal/render/view/`     | The input types, one struct per view, and `Decode`                                             |
| `internal/render/doc/`      | The document model: blocks and inline spans, with no format in it                              |
| `internal/render/markdown/` | The Markdown writer: escaping and table layout                                                 |
| `internal/render/testdata/` | Fixtures `<view>/<case>.json` and golden files `<view>/<case>.<level>.md`                      |

Data flows one way: `view.Decode` reads JSON into a view struct; the view's builder applies the level, the folds and the wording and returns a `doc.Doc`; the writer prints it. A builder never writes bytes and a writer never sees view data.

```go
// Package render writes a view of a Tableaux project in a textual format.
package render

// Format names an output form.
type Format int

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

// Render writes one view at one level. It is a pure function of v and o.
func Render(w io.Writer, v view.View, o Options) error

// The errors Render returns; a write error passes through unwrapped.
var (
	ErrFormat = errors.New("render: format not built")
	ErrLevel  = errors.New("render: the view data holds less than the level asked for")
	ErrView   = errors.New("render: unknown view")
)
```

`Render` switches on the concrete type of `v`, calls that view's unexported builder, `func buildGates(v *view.Gates, o Options) doc.Doc`, and hands the result to `markdown.Write`. The package holds no state and reads no environment, file or clock. Exit codes belong to the Read commands (`5ca9`); a render error is its exit 3 (A6).

### The input types

The types are tabloio's own. Each mirrors the `data` object that `tablo view` emits for the view; the JSON tags are the contract (A1), and every field the function prototypes do not emit is an assumption that each design lists against its view. Decoding ignores unknown fields and leaves absent ones zero, so tablo may add a field within a minor series.

```go
// Package view holds the data of each view as tabloio's renderers read it.
package view

// Level orders the three levels; each includes the one before.
type Level int

const (
	Glance Level = iota
	Detail
	Provenance
)

// ParseLevel reads "glance", "detail" or "provenance"; Level also implements
// encoding.TextMarshaler and TextUnmarshaler with the same words.
func ParseLevel(s string) (Level, error)

// View is the data of one view.
type View interface{ Head() *Head }

// Decode reads the data object of one view. The names are tabloio's command
// names: gates, task, authority, assignment, queue, blockage, tableau,
// context, history, audit.
func Decode(name string, data []byte) (View, error)

// Head is what every view's data carries; each view struct embeds it.
type Head struct {
	Level   Level   `json:"level"`   // the deepest level the data holds
	Project TaskRef `json:"project"` // the root task
	Ref     Ref     `json:"ref"`
	Params  Params  `json:"params"` // the parameters in force, as tablo resolves them
	Legend  Legend  `json:"legend"`
}

// Ref is the commit in view.
type Ref struct {
	Name    string `json:"name"`     // as given, "HEAD" by default
	From    string `json:"from"`     // the start of a range; history only
	Commit  string `json:"commit"`   // the full hash
	Date    string `json:"date"`     // its author date, YYYY-MM-DD
	OnTrunk bool   `json:"on_trunk"`
}

// Params echoes the focusing parameters; a zero field is not in force.
type Params struct {
	Task       string   `json:"task"` // empty when the view covers the root
	Person     string   `json:"person"`
	Window     *int     `json:"window"`
	Columns    []string `json:"columns"`
	Historical bool     `json:"historical"`
	Proposed   bool     `json:"proposed"`
	Stale      int      `json:"stale"`
}

// Legend is gates.yaml and the method's marks, in file order.
type Legend struct {
	Gates   []Gate   `json:"gates"`
	States  []State  `json:"states"`
	Reasons []Reason `json:"reasons"`
	Marks   []Mark   `json:"marks"`
}

type Gate struct{ Key, Symbol, Name, Criteria string }        // tags: key, symbol, name, criteria
type State struct {
	Key, Symbol, Synopsis string
	Severity              int
}
type Reason struct {
	Key, Symbol, Synopsis string
	Reserved              bool // true for review
}
type Mark struct{ Key, Symbol, Meaning, Junction string }     // key: person, agent, reviewer, subproject, exempt

// TaskRef names a task.
type TaskRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Person is an actor; Name is empty where only the email is known.
type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Status is a status as a view shows it: stated, rolled up, or from a snapshot.
type Status struct {
	Gate     string  `json:"gate"`
	State    string  `json:"state"`
	Reason   string  `json:"reason"`
	Note     string  `json:"note"`
	Date     string  `json:"date"`
	Recorder string  `json:"recorder"`
	From     *Origin `json:"from"` // nil for a status file
}

// Origin says where a derived status comes from.
type Origin struct {
	Kind string  `json:"kind"` // "rollup" or "snapshot"
	Task TaskRef `json:"task"` // the child, or the subproject's task
	URL  string  `json:"url"`  // snapshot: the subproject
	Pin  string  `json:"pin"`  // snapshot: the commit read
	Gate string  `json:"gate"` // snapshot: where the subproject's task stands
}

// Commit is a Git fact.
type Commit struct {
	Hash          string   `json:"hash"`
	Subject       string   `json:"subject"`
	Date          string   `json:"date"` // author date
	Author        Person   `json:"author"`
	Committer     Person   `json:"committer"`
	AuthorTime    string   `json:"author_time"`    // RFC 3339 with offset
	CommitterTime string   `json:"committer_time"`
	Trailers      []string `json:"trailers"`
	Files         []File   `json:"files"`
}

type File struct{ Path, Change string } // change: added, changed, removed, pin

// Command is a line a reader runs to reproduce or resolve a fact.
type Command struct {
	Text    string `json:"text"`
	Comment string `json:"comment"`
}

// Marked pairs a junction's marks, in the legend's order, with whether the person in view acts there.
type Marked struct {
	Marks []string `json:"marks"` // keys of Legend.Marks
	Acts  bool     `json:"acts"`
}
```

Every field carries a JSON tag, its name in snake case; a declaration here and in the three sibling designs shows the tag only where it differs or where the line has room. The prototype's question, whether the view data or the front end owns the symbols, is settled: every view's data carries `legend`, and the renderer holds no table of symbols (A2). A key with no symbol prints as the key alone.

### The document model

```go
// Package doc is the format-neutral document a view builds.
package doc

type Doc struct{ Blocks []Block }

// Block is one of Heading, Para, Table, List, Pre.
type Block interface{ block() }

type Heading struct {
	Level int // 1 for the title, 2 and 3 below it
	Text  []Inline
}
type Para struct{ Text []Inline }
type Table struct {
	Align []Align // one per column: Left, Centre, Right
	Head  []Cell
	Rows  [][]Cell
}
type Cell struct {
	Indent int // tree depth, in steps
	Text   []Inline
}
type List struct {
	Ordered bool
	Items   []Item
}
type Item struct {
	Text   []Inline
	Blocks []Block // nested lists and preformatted blocks
}
type Pre struct{ Lines []string } // drawn as given: a tree, a brief, commands

// Inline is one of Text, Code, Symbol, Strong, Link.
type Inline interface{ inline() }

type Text string   // project or fixed text; the writer escapes it
type Code string   // an id, a key, a model, a hash, a path, a trailer
type Symbol string // a gate, state, reason or mark symbol, as the legend gives it
type Strong []Inline
type Link struct {
	Text []Inline
	URL  string
}
type Align int
```

### The Markdown writer

`markdown.Write(w io.Writer, d doc.Doc) error` prints blocks with one blank line between them, `\n` line ends, no trailing space and one final newline.

- **Tables.** A row is `| ` + the cells joined by ` | ` + ` |`; an empty cell is empty, so two bars stand two spaces apart. The delimiter row is `|` + one of `---`, `:-:`, `--:` per column + `|`. Cells carry no padding: padding needs the display width of an emoji, which media disagree on (prototype `e0f7`), and it rewrites every row of `STATUS.md` when one cell grows, where an unpadded table changes one line. A step of `Cell.Indent` prints `&nbsp;&nbsp;`.
- **Headings** print `#` signs; **lists** print `- ` or `1. ` and indent a nested block by the marker's width; **Pre** prints a fence of three backticks, or one longer than the longest run of backticks inside it.
- **Escaping.** In `Text`, a backslash precedes each of ``\ ` * _ [ ] < |``, `&` prints `&amp;`, and a run of white space, a line break included, prints as one space. At the start of a paragraph or an item, a backslash also precedes `#`, `>`, `+`, `-`, and the dot of a leading number. `Code` prints between backticks, with two and a space each side when the content holds one, and `|` inside it prints `\|`. A `Link` URL percent-encodes spaces and parentheses. `Symbol` prints as given.

### The page frame

Every rendering has this frame, in this order; the drafts show it.

1. `# <view title>`, with a subject where the view has one: ``# Task definition: `e9c6` Abstract views``.
2. The parameter line: the view's question in bold, then, joined by ` · `: the project's title; ``ref `<name>` `` (a range reads ``ref `<from>..<name>` ``); the parameters in force in the order of VIEWS.md's table, each as its name and value (``task `e9c6` ``, `person <email>`, `window 1` or `columns <keys>`, `historical junctions on`, `proposed only`, `stale 7 days`); and `level <level>`. Off the trunk the ref reads ``ref `<name>`, off the trunk``.
3. The legend line, in every view but the gate definition: `Legend: <the gates view>.` (R13).
4. The view's blocks for the level, with headings from `##`.
5. At detail and provenance, when `Links.Views` names any of the view's related views: `See also: <views>.`, in the order the view's design lists.
6. `Command: ` and the command that reproduces the rendering, in a code span (R12).

### The rules

- **R1 Levels nest.** A deeper level adds blocks, and adds columns to a table the level above shows; it removes nothing. Glance is one table or one list; a heading at glance stands at every level.
- **R2 A symbol carries its name.** A gate, state or reason prints as symbol, space, key: `📐 design`. A symbol stands alone only in a grid cell, in a grid's column header and in a Marks cell, and a key line follows each such table: the symbols the table shows and no others, in legend order, each with its key or its meaning with a lower-case initial, joined by ` · `, with a final full stop.
- **R3 The status phrase** is gate, state and reason by R2, joined by spaces: `📝 defined 🟢 nominal 👓 review`. A status with no state prints the gate alone. An origin follows after a comma: ``rolled up from `a6f7` <title>``, or ``🪆 from the subproject `<url>` ``.
- **R4 A task** prints as its id in a code span and its title: `` `e9c6` Abstract views``. A list of several tasks in one cell prints ids alone, separated by spaces.
- **R5 A person** prints as the email. A commit fact at provenance prints `Name, email` when the data holds the name. When `Params.Person` is set, a list view prints that email in bold wherever it holds a position, and a grid encloses in `[` `]` each cell where `Marked.Acts` is true.
- **R6 A commit** prints as its first seven characters in a code span, followed by its subject where a table gives it a cell of its own. The history's provenance alone prints the full hash.
- **R7 An absent value** is an empty cell. `—` is the exempt mark and nothing else.
- **R8 Counts** print as digits. A heading over a list whose length depends on the project's state carries the count after a colon, `## Dependents: 21`. A count line over a view that sorts its items into fixed kinds names every kind, zeros included.
- **R9 Empty forms.** A view with no item prints its whole frame and, in place of its first table or list, one bold sentence that each design fixes: `**No cause holds any task.**`. An empty section inside a view prints its heading with `none` for the count and no body: `## Requires: none`. No table has zero rows, and an empty view is no error.
- **R10 The run fold.** In a table or a list of tasks, a run of more than eight consecutive rows that agree in every cell but the task (decided at review on 2026-10-06, decision 8) shows its first three rows, then one row whose task cell reads `… and <n> more: ` with the ids of all the rest, and whose other cells repeat the shared values. A row that differs from its neighbours never folds. `Options.Unfold` turns the rule off. The two grids never fold rows this way.
- **R11 Collapsed children.** A tree shows a parent's hidden descendants as a count on the parent's row: `(31)` after the title in a grid, where tablo's data hides the rows by level, and `<n> children` in the authority tree, where the renderer hides them (design `c48a`).
- **R12 Commands.** The closing command is canonical, not the words the user typed: `tabloio`, the view's command, its positional argument, then `--task`, `--person`, `--ref`, `--window` or `--columns`, `--historical`, `--proposed`, `--stale`, `--level`, each only when in force, and `--ref` only when the name is not `HEAD` (A3, A4). Below provenance the frame and the command name the ref as given and no commit, so a snapshot does not change with every push (design `1422`, decision 3); `Options.Stamp` adds ``at `<commit>`, <date>`` to the ref. At provenance the ref always reads ``ref `<name>` at `<commit>`, <date>`` and the command reads `--ref <commit>`, so the command reproduces the bytes. A `git` command comes from the data as a `Command` and prints in a `Pre`, its comment after ` # `.
- **R13 Links.** With `Links.Project` set, the defining mention of a task links to its file: the Id cell of a grid row, the Task cell of a row that is about that task, the title line of the task view. The target is `<Project>/.tableaux/tasks/<id>.yaml`, cleaned as a slash path. A reference of a task or a junction links by its `url`: an absolute URL as given, a path relative to the repository root as `<Project>/<path>` (A5). A view named in the legend line or the See also line prints as a link, `[gates](GATES.md)`, when `Links.Views` holds it, and as its command in a code span, `` `tabloio gates` ``, when not.
- **R14 No commentary.** The renderer prints data, headings, column heads, count lines, key lines and the fixed sentences a design lists. It prints no sentence that explains why the data reads as it does.
- **R15 Determinism.** The same view data and options give the same bytes: no clock, no environment, no map order (the See also line follows the design's order), and UTF-8 as the data gives it, with no normalisation.
- **R16 Words for a requirement.** An edge prints `<from> → <to>`. A condition prints `met` or `unmet`, with `, not yet due` after it when the requirement is not due.
- **R17 Two kinds of text.** Text that the project's files and commits hold (a title, a note, a description, a subject, a reference's text, a requirement's text) is one `Text`, escaped whole. Text that tablo words (a mark's meaning and junction, a `Fact` value, a finding's message and action, a cause's action) is worded text: the builder splits it at backticks into `Text` and `Code`, so tablo names an id, a key or a trailer in backticks and no front end parses more than that.

### A second output form

The text format of `e0f7` is `text.Write(w, d, text.Options{Width, Selector})` over the same `doc.Doc`. Everything it needs is in the model already: `Table.Align` and `Cell.Indent` to draw a box table and a tree, `Symbol` apart from `Text` so that a width function and a variation-selector policy apply to symbols alone, `Strong` and `Link` to drop or underline, and `Pre` to pass through. No builder changes; `Render` gains one case. The width at which a table breaks is the text writer's to design, since Markdown needs none.

## The gate definition view

```go
// Gates is the gate definition view.
type Gates struct {
	Head
	Task     *TaskRef     `json:"task"`      // set with the task parameter
	ForTask  []GateFor    `json:"for_task"`  // detail, with a task: one per gate, in order
	Language string       `json:"language"`  // provenance: the language version
	Trunk    string       `json:"trunk"`     // provenance
	Files    []FileCommit `json:"files"`     // provenance: gates.yaml and version.yaml
	Commands []Command    `json:"commands"`  // provenance
}

// GateFor says how one task's junction expands a gate.
type GateFor struct {
	Gate       string      `json:"gate"`
	Applies    bool        `json:"applies"`
	References []Reference `json:"references"`
	StatedBy   string      `json:"stated_by"` // the task whose entry states them
}

type Reference struct{ Text, URL string }
type FileCommit struct {
	Path   string `json:"path"`
	Commit Commit `json:"commit"`
}
```

| Level      | Blocks                                                                                                                                 |
|------------|----------------------------------------------------------------------------------------------------------------------------------------|
| glance     | `## Gates`: `#`, Symbol, Gate, Key. `## States`: Symbol, State. `## Reasons`: Symbol, Reason. `## Junction marks`: Mark, Meaning         |
| detail     | Gates gains Criteria, and with a task a column headed ``For `<id>` ``; States gains Severity and Synopsis; Reasons gains Synopsis; Junction marks gains Junction. A fixed sentence follows the Reasons table when a reason is reserved; the draft holds it, with the two symbols taken from the legend |
| provenance | A `Source:` line after each table (the key and `.tableaux/gates.yaml`; for the marks, `README.md`, Junctions, at the language version). `## Provenance`: a table Fact, Value, Source with the language version, the trunk and the ref; a table File, Last changed by, Date, Author, Committer, Subject; then the commands |

The four headings never carry a count, so `#gates`, `#states`, `#reasons` and `#junction-marks` resolve at every level, as the mockup promises. The title reads ``# Gate definition for `e9c6` Abstract views`` with a task. A ``For `<id>` `` cell reads `Applies`, `— The gate does not apply`, or `Applies; reference: <text>, <url in code>, stated by <id in code>`, one reference after another. The empty form: a project with no reason prints `## Reasons: none`.

## The task definition view

```go
// Task is the task definition view.
type Task struct {
	Head
	Task          TaskRef        `json:"task"`
	Assignee      string         `json:"assignee"`
	Parent        *Parent        `json:"parent"` // nil for the root
	Status        Status         `json:"status"`
	Description   string         `json:"description"`    // detail from here
	References    []Reference    `json:"references"`
	Path          []TaskRef      `json:"path"`           // the root first, the task last
	Siblings      int            `json:"siblings"`       // children of the parent, the task included
	Children      []ChildRow     `json:"children"`
	Authorities   []Authority    `json:"authorities"`    // nearest first
	Authorised    bool           `json:"authorised"`
	Requires      []Requirement  `json:"requires"`
	Dependents    []Requirement  `json:"dependents"`
	Junctions     []Junction     `json:"junctions"`      // every gate, in order
	Snapshot      *Snapshot      `json:"snapshot"`       // a recursive next junction
	StatusCommit  *Commit        `json:"status_commit"`  // provenance from here
	Authorisation *Authorisation `json:"authorisation"`
	Rollup        []Fact         `json:"rollup"`         // a parent: the steps of its roll-up
	Linkages      []Linkage      `json:"linkages"`
	Sources       []FieldSource  `json:"sources"`
	Reviews       []Review       `json:"reviews"`
	Models        []ModelCheck   `json:"models"`
	Events        []Event        `json:"events"`         // newest first, at most five
	EventCount    int            `json:"event_count"`
	Commands      []Command      `json:"commands"`
}

type Parent struct {
	TaskRef
	Order int `json:"order"`
}
type ChildRow struct {
	TaskRef
	Status Status `json:"status"`
}
type Authority struct {
	Email string `json:"email"`
	By    string `json:"by"` // the ancestor's id
}
type Requirement struct {
	Task       TaskRef `json:"task"`       // the other task
	From, To   string  // tags: from, to
	Text       string  `json:"text"`
	Met, Due   bool    // tags: met, due
	Status     *Status `json:"status"`     // of the required task; nil for a dependent
	Subproject string  `json:"subproject"` // a cross-project entry: its url
	Commit     string  `json:"commit"`     // and the commit it reads
}
type Junction struct {
	Gate        string      `json:"gate"`
	Marked                              // the marks, and whether the person acts
	Contributor string      `json:"contributor"`
	Model       string      `json:"model"`
	Reviewer    string      `json:"reviewer"`
	ByDefault   bool        `json:"reviewer_by_default"` // the assignee reviews, since no entry states one
	Subproject  *SubRef     `json:"subproject"`
	References  []Reference `json:"references"`
	Stands      string      `json:"stands"`   // passed, here, next, later, exempt
	Reviewed    bool        `json:"reviewed"`
}
type SubRef struct {
	URL  string  `json:"url"`
	Task TaskRef `json:"task"`
}
type Snapshot struct {
	SubRef
	Assignee string     `json:"assignee"`
	Status   Status     `json:"status"`
	Children []ChildRow `json:"children"`
}
type Authorisation struct {
	Commit Commit `json:"commit"`
	Way    string `json:"way"` // change, merge, trailer
	By     string `json:"by"`  // author, committer, both, or empty for a proposed task
}
type Fact struct{ Name, Value string } // tags: name, value
type Linkage struct {
	Gate  string `json:"gate"` // the first gate that states it
	Facts []Fact `json:"facts"`
}
type FieldSource struct {
	Gate, Field, Value string // tags: gate, field, value
	By                 string `json:"by"` // an id, or "default"
}
type Review struct {
	Gate, Reviewer string
	Commit         *Commit `json:"commit"` // nil where the authorisation stands as the review
	Effect         string  `json:"effect"` // accepts, authorisation, none
}
type ModelCheck struct {
	Commit                  string
	Gate, Stated, Trailer   string
	Reading                 string `json:"reading"` // agrees, differs, missing, exempt
}
type Event struct {
	Date, Commit, By string
	Kinds            []string `json:"kinds"` // task, authorised, status, reaffirmed, reviewed, pin
	Status           *Status  `json:"status"`
	Gate             string   `json:"gate"`
	Pin              *PinMove `json:"pin"`
}
type PinMove struct{ URL, Old, New string }
```

| Level      | Blocks, in order                                                                                                                                  |
|------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| glance     | The field table, Field and Value: Task, Assignee, Parent (`None` for the root), Status (R3), Note, Recorded                                        |
| detail     | `## Description` with `References:` and a list; `## Place in the tree` (Path, Order, Children, Authorities, Authorised); `## Requires: n` (Task, Edge, What passes, Its status, Condition); `## Dependents: n` (Task, Edge, What passes, Condition); `## Junctions` (Gate, Marks, Contributor, Model, Reviewer, Stands) with its key line; `## Subproject snapshot` when the data holds one |
| provenance | `## Status commit` and `## Authorisation commit` as field tables, placed after the field table and after Place in the tree; `## Roll-up` for a parent; `## Linkages`; `## Where each junction field comes from` (Gate, Field, Value, Supplied by); `## Reviews: n` (Gate, Reviewer, Accepted by, Effect); `## Models: n` (Commit, Junction, The junction states, `Model:` trailer, Reading); `## Newest events` (Date, Commit, By, Event, What it records); the commands |

The draft fixes the detail form. The remaining wording:

- **Recorded** reads `<date> by <recorder>`; for a roll-up `<date>, the oldest among its children`; for a snapshot `<date>, from the subproject`.
- **Path** joins tasks (R4) by ` › `. **Order** reads ``<order> of <siblings> under `<parent>` ``. **Children** reads `None`, or the children's ids with their shared status phrase when all agree, else one `id status` pair per child, joined by `; `. **Authorities** groups consecutive ancestors with one email: ``<email> (`2034`, `bc63`)``, groups joined by `; then `. **Authorised** reads `Yes` or `No, proposed`.
- **Junctions** shows every gate, one row each, for a leaf, a parent and a task with recursive junctions alike. A recursive junction adds a Subproject column, ``<url> `<id>` <title>``, to the table. Reviewer reads `the assignee, <email>` when `ByDefault`. Stands reads `passed`, `passed; the status stands here`, `next`, `later` or `does not apply`, with `, reviewed` after `passed` when `Reviewed`.
- **Status commit** rows: Commit (R6), Date, Author, Committer, Trailers. **Authorisation commit** adds Accepts by (`a change on the trunk`, `a merge`, `an Authorised: trailer`) and The authority is (`the author`, `the committer`, `the author and the committer`, or empty).
- **Supplied by** reads a task (R4), or `the plain default`. **Effect** reads `The reviewer accepts`, `The authorisation stands as the review`, or `None`. **Reading** reads `Agrees`, `Differs`, `No trailer`, or `Exempt`.
- **Newest events** opens with `<EventCount> events; the newest <n>, newest first.` and closes, when events remain, with a row `… and <m> earlier`. What it records is a status phrase with its note after a colon, a gate for a review, or ``<url> `<old>` → `<new>` `` for a pin.
- **Roll-up**, **Linkages** and **Subproject snapshot** are tables of the data's facts, Name and Value, in the data's order.

Requires and Dependents take the run fold (R10). Fixed related views, for the See also line: history, blockage, authority, assignment, queue, context, tableau, audit.

## Conformance to the texts

| View            | VIEWS.md section                    | Golden reference                 | glance                | detail                          | provenance                           |
|-----------------|-------------------------------------|----------------------------------|-----------------------|---------------------------------|--------------------------------------|
| Gate definition | [Gate definition](https://github.com/nbyoung/tableaux/blob/main/VIEWS.md#gate-definition) | `docs/mockups/gates.md`  | The four tables       | The criteria, severities, synopses; the task column | The source lines, the two tables, the commands |
| Task definition | [Task definition](https://github.com/nbyoung/tableaux/blob/main/VIEWS.md#task-definition) | `docs/mockups/task.md`   | The field table       | The six sections                | The commits, the sources, the reviews, the models, the events |

Both views show every gate and take no window, as the owner ruled. Where the mockup cannot be reproduced:

| # | Mockup | What it shows | What the renderer does | Kind |
|---|--------|---------------|------------------------|------|
| G1 | both | One file holds three levels under `## Glance`, `## Detail`, `## Provenance`, and an opening that explains the file | One rendering holds one level; its headings rise one step; the frame replaces the opening | By design: gates.md says so itself |
| G2 | both | Paragraphs that explain the project's data ("No junction of `e9c6` states a reference…", "The fold keeps every row that differs…", "The edge reads…") | Omits them (R14) | D9 |
| G3 | gates | A sentence under each heading at detail | Omits them, and keeps the one that states the reserved meaning of `review` | D9 |
| G4 | gates | The reserved reason `review`, the Junction column of the marks, author names, the reproduce commands | Reads `legend.reasons[].reserved`, `legend.marks[].junction`, `Person.Name`, `commands`; prototype `493e` emits none of them | A7 |
| G5 | gates | "Illustration: a gate with a reference" | The same column, from `for_task[].references`; the corpus entry `weather-station` holds the one junction reference | T5 |
| G6 | task | `Status` as `📐 design · 🟢 nominal` | R3: no dot | D3 |
| G7 | task | Counts in words ("Twenty-one tasks require…"), and a breakdown by gate | R8: `## Dependents: 21` | D3 |
| G8 | task | A parent's junctions as "The children inherit", with seven gates in one row; recursive junctions in a third table | One table, one row per gate | D6 |
| G9 | task | "Stands" reads `later; the last gate`; a condition reads `not met, not yet due` | The fixed words of the model and R16 | D3 |
| G10 | task | State and reason symbols on the status; the children's titles and statuses; the snapshot's facts; the status commit; the reviews' effect; the models table | Reads `legend`, `children`, `snapshot`, `status_commit`, `reviews`, `models`; prototype `493e` emits child ids only, a null status commit and no models | A8 |
| G11 | task | Prose that explains a linkage and a roll-up step by step | A table of `Fact` pairs that tablo words | A9 |

## Assumed interfaces

| #  | Of                    | Assumption                                                                                                                              |
|----|-----------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| A1 | tablo `4ed9`, and the view tasks `493e`, `886d`, `8ed1` | The envelope's `data` for a view is one JSON object whose keys are the tags above, flat across the levels, with `level` naming the deepest level it holds; the Read commands pass `data` to `view.Decode`. If `4ed9` has tabloio link tablo as a library, `5ca9` marshals tablo's value to JSON and decodes it, and the renderers do not change |
| A2 | tablo `493e`, `886d`, `8ed1` | Every view's data carries `level`, `project`, `ref` (name, commit, date, on_trunk), `params` and `legend`, the legend with every gate, state, reason and mark of the project |
| A3 | tabloio `5ca9`        | The commands are `gates`, `task <id>`, `authority`, `assignment`, `queue`, `blockage`, `tableau`, `context`, `history`, `audit`, and the flags `--task`, `--person`, `--ref`, `--window`, `--columns`, `--historical`, `--proposed`, `--stale`, `--level` |
| A4 | tabloio `5ca9`        | The person flag is `--person`, as the owner ruled at review on 2026-10-06; the constant in `internal/render/command.go` and the drafts, which were written with `--for`, follow at implementation |
| A5 | tabloio `5ca9`        | It fills `Options`: `Level` from `--level` or the role, `Format` from `--markdown` or `--text`, `Links.Project` from the output path and `-C` (`.` when the output goes to stdout or into the project directory), `Links.Views` empty unless a flag names targets, `Stamp` and `Unfold` false unless a flag sets them |
| A6 | tabloio `5ca9`        | It handles the envelope's diagnostics and every exit code; it maps a `Render` error to its failure code and prints nothing partial to a file |
| A7 | tablo `493e`          | The gate view's data holds `for_task` with `stated_by`, `language`, `trunk`, `files` with full commits, `commands`, and the legend fields G4 names |
| A8 | tablo `493e`          | The task view's data holds the fields of `view.Task`; `stands`, `reviewed`, `reviewer_by_default`, `met`, `due`, `effect`, `reading`, `way` and `by` take exactly the values the comments list |
| A9 | tablo `493e`          | `rollup` and `linkages` arrive as ordered name and value pairs in tablo's words; a value is worded text (R17) |
| A10 | tablo `493e`, `886d` | A junction's `marks` are tablo's: the renderer derives no mark. 👀 stands in the list only when the junction's reviewer differs from its contributor (D4) |

## Tests

Tests are table-driven Go tests under `internal/render/...`. A golden test decodes `testdata/<view>/<case>.json`, renders it, and compares bytes with `testdata/<view>/<case>.<level>.md`; `go test ./internal/render/... -update` rewrites the golden files, and a reviewer reads the diff. Fixtures come from three sources: the drafts here; the Tableaux tooling plan as the mockups show it at `3cdae52`, which the implementation transcribes into the input types; and the corpus entry `weather-station`, transcribed from `corpus/entries/weather-station/expected.yaml` and the prototypes' `testdata`.

| #   | Proves                                                                 | How                                                                                                   |
|-----|------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| T1  | `Decode` reads each view, ignores an unknown key, rejects an unknown level and an unknown view name | `e3ed/gates.json`; the same with an added key; `"level": "deep"`; `Decode("nope", …)` |
| T2  | The gate definition reproduces the mockup at glance and detail          | `e3ed/gates.json` against `e3ed/gates.glance.md` and `e3ed/gates.detail.md`, byte for byte             |
| T3  | The task definition reproduces the mockup at detail, with a person in bold | The fixture `task/e9c6.json`, transcribed from `task.md`, against `e3ed/task.detail.md`             |
| T4  | Provenance of both views, and the parent and recursive shapes           | Fixtures `gates/tooling.json`, `task/e9c6.json`, `task/5fe3.json`, `task/595e.json` at provenance; golden files the implementation generates and the reviewer reads against `gates.md` and `task.md` |
| T5  | A junction reference, a stalled status with a reason, a proposed task, an unmet requirement | `weather-station`: `gates/weather-9f31.json`, `task/weather-9f31.json`, `task/weather-c07d.json`, `task/weather-3c5d.json` |
| T6  | R1: levels nest                                                         | For every fixture: each heading and each table header cell at one level appears at the next, in order  |
| T7  | R2 and A2: the renderer invents no symbol                               | For every fixture and level: each rune outside ASCII in the output occurs in the fixture, or is one of `· … → › — ×` |
| T8  | The table layout                                                        | For every golden file: each row of a table has the header's cell count; no line ends in a space; the file ends in one newline |
| T9  | Escaping                                                                | A table of cases: each escaped character in a paragraph, a cell, a list item and a heading; a line break in a note; a backtick in a code span; a fence inside a `Pre`; a URL with a space |
| T10 | R15: determinism                                                        | Render each fixture twice, the second time with `Links.Views` built in another order; equal bytes     |
| T11 | R12: the ref and the commit                                             | One fixture four ways: glance, glance with `Stamp`, provenance, and `ref.name` `HEAD`; the frame and the command differ as R12 states |
| T12 | R13: links                                                              | `task/e9c6.json` with `Links.Project` `""`, `.` and `../..`, and with `Links.Views` holding `gates`    |
| T13 | R9 and R10: the empty forms and the fold                                | A task with no requirement and no dependent; a run of eight alike, of nine, of nine with one that differs in the middle; `Unfold` |
| T14 | `ErrLevel`, `ErrFormat`, `ErrView`, and a failing writer                | A glance fixture asked for detail; `Format: Text`; a foreign `View`; a writer that fails after n bytes |
| T15 | The document model serves a second writer                               | A test-only writer prints a `doc.Doc` as plain lines; every block and inline kind the ten builders emit passes through it |

Already tried, outside the repository: a script parses every table of the ten mockups, 123 tables, into the canonical row form with equal cell counts throughout, and the three Markdown drafts here are its output over `gates.md` and `task.md` with the edits G1 to G9. Not checked: how any Markdown viewer draws the output. The unit gate proves T1 to T15. The integrate gate replaces each transcribed fixture with the output of `tablo view` over `tableaux/corpus/build/weather-station` and over the tooling plan, which is the golden test against the corpus that task `1093` asks for; it waits for tablo's view tasks. The validate gate is the owner reading `STATUS.md` and a task page on GitHub.

## Implementation notes

Order: `view` (types and `Decode`), `doc`, `markdown` with T8 and T9, the rules as helpers in `internal/render/rules.go` and `command.go`, then `gates.go` and `task.go`. The three sibling tasks add builder files and view structs and change no shared file, so they proceed in parallel once this task's shared packages land. `internal/render/doc.go` keeps its package comment. `prototype/e3ed/` stays until the unit gate passes, then goes; the design keeps its nesting of levels, its `-update`-free tests as a model, and its code-span and link forms, and drops its block-quote glance line, its built-in symbols and its `-gates` flag.

The implementation adds no dependency. tabloio's README says the Renderers task adds the `require` on tablo to `go.mod`; with input types of their own the renderers import nothing from tablo, so that line moves to the Read commands (`5ca9`), the first task that calls tablo. The plan gives the four renderer tasks one requirement, on the module (`0a2f`), which fits this design. It gives no task of tabloio a requirement on tablo's view tasks; the golden test against real view data needs them at the integrate gate, and the umbrella's requirement of `6103` on `77b2` covers it only at the level of the whole subproject.

## Decisions at review

The owner decided each question this design raised on 2026-10-06, together with the cross-design rulings of that review; a decision the review did not reach stands as the design states it, accepted with the design.

1. **Input types of tabloio's own.** Accepted with the design, as stated: the renderers read JSON into `internal/render/view` and import nothing from tablo. The owner ruled that tabloio links tablo as a library (5ca9 decision 1), so the Read commands hand the renderers a decoded value from the library rather than from a process; the types here stay tabloio's, and the implementation decides whether they are satisfied by conversion or by tablo's types directly. The alternative imports tablo's Go types, which ties the four tasks to tablo's first tag.
2. **A document model between the views and the formats.** Accepted with the design, as stated: builders return a `doc.Doc`; Markdown and, later, text are writers. The alternative writes Markdown straight from each view, as the prototypes do, and has `e0f7` write each view a second time.
3. **One wording across the views** (R2, R3, R7, R8, R16). Accepted with the design, as stated: symbol with key, the status phrase without dots, an empty cell for an absent value, digits for counts, `met` and `unmet`. The alternative keeps each mockup's own words. Binds `c48a`, `a3cc`, `9167`, and tableaud's templates and tablotui's panes if the owner wants the three front ends to agree.
4. **The self-review mark.** Decided by the owner on 2026-10-06, in the general form tableaud's 438a decision 2 states: a junction shows the contributor's mark alone where its reviewer is its contributor, so 🤖 alone for an agent that reviews its own work and 🧑 alone for a person's plain junction with no other reviewer; 👀 shows only where the reviewer differs from the contributor. tablo derives the marks and no renderer does. README.md's agent-only reviewer default stands, so a person's plain junction with no stated reviewer has none. Binds tablo's `493e` and `886d`, tableaud `438a` to `9a9c`, tablotui `679b`. The alternative showed 🤖👀 wherever a junction had any reviewer.
5. **The fold notation** of a tableau. Decided by the owner on 2026-10-06, confirming all three front ends: the count stands in the column header, `❔ ×0`, and a run of folded columns shares one header, `🔗…🚀 ×0`, as `tableau.md` draws it; the count includes parents by their roll-up, as VIEWS.md states. The alternative put the count in the first row, as `context.md` and the VIEWS.md example do; VIEWS.md's log records the ruling and the mockup stays as drawn. Binds `a3cc`, tableaud `44bf`, tablotui `679b`.
6. **Every gate, one row each,** in every list of junctions. Accepted with the design, as stated: no row stands for a run of gates. The alternative keeps the folded rows of `task.md`'s parent and recursive shapes and of `authority.md`.
7. **The empty form** (R9). Accepted with the design, as stated: the whole frame, one bold sentence in place of the first table, `none` on an empty section's heading, zeros in a count line, no error. The alternative, from `queue.md`, keeps a table with one `none` row per kind. Binds all four designs and, for the words, tableaud's templates.
8. **The fold of a long list** (R10). Decided by the owner on 2026-10-06, for tableaud's threshold (438a decision 5): more than eight alike rows show three and fold the rest with every id; the front end folds, tablo's data arrives whole; a flag unfolds. R10, T13 and `c48a`'s S8 now read so. The alternatives were this design's five or more, as `queue.md` has it, a summary by parent, and the children of a parent in one row.
9. **No commentary** (R14). Accepted with the design, as stated: the renderer prints no sentence that explains the data, so most prose of the mockups does not appear; the gate definition keeps one sentence, on `review`. The alternative has tablo emit explanatory text per view.
10. **No commit below provenance** (R12). Accepted with the design, as stated: glance and detail name the ref as given, so `STATUS.md` holds a fixed point, as design `1422` decided; provenance names the commit. The mockups name the commit at every level; a flag behind `Options.Stamp` gives that form.
11. **Links** (R13). Accepted with the design, as stated: the defining mention of a task links to its task file, references link by their `url`, and a view links to another only when the command line names the target file. The alternative links views to fixed names such as `gates.md`, which breaks for a rendering that stands alone.
12. **Unpadded tables.** Accepted with the design, as stated. The alternative pads cells for a reader of the source, at the cost of emoji-width guesses and of diffs that touch every row.
13. **The command vocabulary.** Decided by the owner on 2026-10-06: `--person` for the person and `--ref` for the ref, in every tool, and one subcommand per view in tablo as in tabloio (4ed9 decisions 2 and 3). A3, A4 and R12 read so. The task files of `c48a`, `a3cc` and `9167` now require this task from implementation to implementation, as the review found the plan lacked.
14. **The implementation review.** Decided by the owner on 2026-10-06, over the branch that implements this design. The method of `View` is `Header() *Head`, since a struct that embeds `Head` cannot also promote a method named `Head`; `Head` itself has the method and every view struct gets it by embedding. `Decode` and `Render` dispatch through registries that each view file and each builder file fill in `init`, so a sibling task adds files and changes no shared file, as the implementation notes promise; `Decode` returns `view.ErrUnknown` for a name it does not know. The printed command follows R12 as written, `--task`, `--person`, `--ref`, which the queue mockup, the agent briefs (e4c7) and the Status renderers (a3cc) share, where `gates.md`, `task.md` and the drafts here put `--ref` first; the golden files under `testdata/` read so and the drafts stay as drawn. `prototype/e3ed/` goes at this gate, not after the unit gate: a prototype is removed when its task records `implementation`, and the README states the rule. Binds c48a, a3cc, 9167, e0f7 and 5ca9.
