# Design a3cc: Status renderers

This task renders the four views of where the work stands: the global tableau and the contextual tableau, which share one grid; the work-blockage tree; and the contributor work queue. All four stand on the architecture of [design e3ed](e3ed.md#the-shared-architecture) and cite its rules R1 to R17. Four choices are this task's own. One grid serves both tableaux, with the fold count in the column header (e3ed, decision 5) and the dates and notes in a table of their own beneath it (D1). The grid's cells arrive from tablo ready to draw, so the renderer applies no cell rule, no window and no roll-up. The blockage tree is a nested list whose sentences the renderer words from structured causes (D3). The queue's provenance level is the brief, which belongs to the Agent briefs task (`e4c7`) behind one hook (A6).

The Text renderers task (`e0f7`) waits for this task at implementation: it draws the same `doc.Doc` this task's builders return, and the grid is the table its width rules size first.

| Draft                                                  | Holds                                                                                  |
|--------------------------------------------------------|----------------------------------------------------------------------------------------|
| [`a3cc/tableau.glance.md`](a3cc/tableau.glance.md)     | The golden output of the global tableau of the Tableaux tooling plan at glance          |
| [`a3cc/context.glance.md`](a3cc/context.glance.md)     | The contextual tableau of `2034` at glance: a folded run, collapsed children            |
| [`a3cc/blockage.detail.md`](a3cc/blockage.detail.md)   | The work-blockage tree at detail, over the state the mockup's illustration assumes      |
| [`a3cc/queue.glance.md`](a3cc/queue.glance.md)         | The agent's work queue at glance, with two run folds                                    |
| [`a3cc/queue.empty.md`](a3cc/queue.empty.md)           | The owner's work queue: the empty form                                                  |

## The model

The task adds `internal/render/grid.go` (both tableaux), `blockage.go`, `queue.go` and `count.go`, three view files in `internal/render/view/`, and test data under `testdata/tableau/`, `testdata/context/`, `testdata/blockage/` and `testdata/queue/`. `count.go` holds one helper, `func count(n int, one, many string) string`, which prints `1 cause`, `2 causes`; the temporal renderers use it too.

### The grid of both tableaux

```go
// Tableau is the global and the contextual tableau; Decode("tableau") and
// Decode("context") both return it, with Contextual set for the second.
type Tableau struct {
	Head
	Contextual bool         `json:"contextual"`
	Subject    *TaskRef     `json:"subject"`   // the contextual tableau's task; nil for a person
	Columns    []Column     `json:"columns"`   // every gate once, in order, alone or in a folded run
	Rows       []GridRow    `json:"rows"`      // the rows of the level, in display order
	Statuses   []StatusCell `json:"statuses"`  // provenance from here
	Rollups    []Rollup     `json:"rollups"`
	MarkGroups []MarkGroup  `json:"mark_groups"`
	Snapshots  []SnapshotRow `json:"snapshots"`
	Commands   []Command    `json:"commands"`
}

// Column is one column of the grid.
type Column struct {
	Gates  []string `json:"gates"`  // one gate, or the run a folded column stands for
	Folded bool     `json:"folded"`
	Count  int      `json:"count"`   // folded: the tasks in view whose current gate lies in the run
	ByGate []int    `json:"by_gate"` // folded: the same count per gate of the run
}

// GridRow is one task.
type GridRow struct {
	TaskRef
	Depth  int        `json:"depth"`
	Parent bool       `json:"parent"`
	Hidden int        `json:"hidden"` // descendants the level does not draw
	Role   string     `json:"role"`   // contextual: "", spine or sibling
	Status Status     `json:"status"` // with its origin, for the notes
	Cells  []GridCell `json:"cells"`  // one per column
}

// GridCell is one junction as the tableau shows it.
type GridCell struct {
	Kind   string `json:"kind"`   // empty, status, marks, exempt
	State  string `json:"state"`  // status
	Reason string `json:"reason"` // status
	Marked                        // marks; Acts encloses any kind of cell
}

type StatusCell struct {
	Tasks  []string `json:"tasks"` // the rows one commit records at one gate
	Gate   string   `json:"gate"`
	Commit Commit   `json:"commit"`
}
type Rollup struct {
	Parent TaskRef `json:"parent"`
	Status Status  `json:"status"`
	Child  TaskRef `json:"child"` // the child it takes its status from
}
type MarkGroup struct {
	Gate            string   `json:"gate"`
	Marks           []string `json:"marks"`
	Tasks           []string `json:"tasks"`
	Contributor     string   `json:"contributor"`
	Model           string   `json:"model"`
	Reviewer        string   `json:"reviewer"`
	ContributorFrom string   `json:"contributor_from"` // a task id, or "default"
	ReviewerFrom    string   `json:"reviewer_from"`
	ExemptFrom      string   `json:"exempt_from"`
}
type SnapshotRow struct {
	Task   string  `json:"task"`
	SubRef                          // the subproject's url and task (design e3ed)
	Pin    string  `json:"pin"`
	SetBy  Commit  `json:"set_by"`
	Status Status  `json:"status"` // of the subproject's task, with its origin
}
```

| Level      | Global tableau                                                         | Contextual tableau                                                        |
|------------|------------------------------------------------------------------------|---------------------------------------------------------------------------|
| glance     | The grid of the rows to depth one, its key line; `## Notes`             | The grid of the task and its children, or of the person's tasks; its key line |
| detail     | The grid of every row; `## Notes`; `## Folded columns` when a column folds | The grid with the spine and the siblings; `## Notes`; `## Folded columns`  |
| provenance | `## Status cells: n`, `## Roll-ups: n`, `## Marks: n`, `## Snapshots: n`, the commands | The same                                                      |

**The grid** is one table: `Id`, `Task`, then one centred column per `Column`.

- A column header is the gate's symbol; a folded column's is the symbol, or `first…last` for a run, then ` ×` and the count: `❔ ×0`, `🔗…🚀 ×0`. A folded column's cells are empty.
- The Id cell is the id in a code span, and the link to the task file (R13). The Task cell indents by `Depth`, prints a parent's title in bold, then ` (<Hidden>)` when rows hide beneath it, then ` · spine` or ` · sibling`.
- A cell prints by kind: `empty` nothing; `status` the state's symbol and the reason's, with no space between; `marks` the marks' symbols in the legend's order, with no space; `exempt` `—`. Brackets enclose a cell where `Acts` is true (R5).
- The key line (R2) follows the table: the gates the headers show, then the states, reasons and marks the cells show.

tablo decides every cell: the historical cells arrive `empty` unless the parameter is on, the window and the list of columns arrive as `Columns`, and the rows of a level arrive as `Rows` with `Hidden` counts. The renderer checks only that each row holds one cell per column, and returns an error that names the row when it does not.

**The notes** table has the columns Rows, Date, Note, From. Rows that share a date, a note and an origin share one line, placed where the first of them stands in display order; a row with no date and no note gives no line. Rows lists the ids (R4). From is empty for a status file; for a roll-up it reads ``rolled up from <task>``; for a snapshot ``<task> in `<url>` at `<pin>`, at <gate>``, the task and the gate being the subproject's. When no row has a date or a note the section does not print.

**Folded columns** has the columns Folded column, Gate, Tasks: one line per gate of each folded column, the first cell repeating the column's header without its count.

**Provenance.** Status cells: Rows, Cell (the gate's symbol), Date, Recorder, Deciding commit (R6 with its subject), Committer, Trailers. Roll-ups: Parent, Shows (gate and state symbols), Rolls up from, Date. Marks: Gate, Mark, Rows, Contributor and model, Reviewer, one line per `MarkGroup` and no merging across gates; a source prints ``, from `<id>` `` after the value, `default` prints `, by default`, and an exempt group prints ``Exempt, from `<id>` `` in the fourth cell. Snapshots: Row, Subproject, Pin, Pin set by (R6, date, author), Root task at the pin (the task and its status phrase), Rolls up from.

**Titles and empty forms.** The global title is `# Global tableau`. The contextual title carries its subject: ``# Contextual tableau: `2034` Views`` or `# Contextual tableau: <email>`. A person with no task prints `**No task stands in this corner.**` in place of the grid. The related views, for the See also line: gates, context or tableau, task, blockage, queue, history.

### The work-blockage tree

```go
// Blockage is the work-blockage tree.
type Blockage struct {
	Head
	Causes   []Cause     `json:"causes"`    // largest first, then display order
	Kinds    []KindCount `json:"kinds"`     // the five kinds, in VIEWS.md's order
	NotDue   []Waiting   `json:"not_due"`   // detail: by task, in display order
	NotDueN  int         `json:"not_due_count"`
	NotDueUnmet int      `json:"not_due_unmet"`
	Commands []Command   `json:"commands"`  // provenance: what reproduces the facts
}

type KindCount struct {
	Kind  string `json:"kind"` // requirement, status, review, authorisation, snapshot
	Count int    `json:"count"`
}

// Cause is one root of the tree.
type Cause struct {
	Kind     string    `json:"kind"`
	Task     TaskRef   `json:"task"`
	Gate     string    `json:"gate"`     // requirement, review: the gate not passed
	Status   *Status   `json:"status"`   // status: what the task states
	URL, Pin string                      // snapshot
	Resolver string    `json:"resolver"`
	Mark     string    `json:"mark"`     // the resolver's mark key, or empty
	Act      string    `json:"act"`      // contributes, reviews, authorises, records, advances
	Holds    int       `json:"holds"`    // distinct tasks beneath, at every depth
	Action   string    `json:"action"`   // detail: worded text (R17)
	Held     []Held    `json:"held"`     // detail
	Facts    []Fact    `json:"facts"`    // provenance: worded pairs, in order
	Resolve  []Command `json:"resolve"`  // provenance
}

// Held is a task a cause holds, and what it holds in turn.
type Held struct {
	Task     TaskRef      `json:"task"`
	Gate     string       `json:"gate"`
	Parent   bool         `json:"parent"`
	Via      string       `json:"via"`      // self, parent, requirement
	Child    string       `json:"child"`    // parent: the child that holds it
	Requires *Requirement `json:"requires"` // requirement: the entry (design e3ed)
	Also     []int        `json:"also"`     // the other causes it stands under, counted from 1
	Held     []Held       `json:"held"`
}

// Waiting is a task with requirements not yet due.
type Waiting struct {
	Task     TaskRef       `json:"task"`
	Status   Status        `json:"status"`
	Next     string        `json:"next"`
	Requires []Requirement `json:"requires"`
}
```

| Level      | Blocks, in order                                                                                                                             |
|------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| glance     | The count line; the table Cause, Who acts, Holds; the line `Next: <n> requirements are not yet due, <m> of them unmet.` when n is not zero     |
| detail     | The count line; the causes as an ordered list with their trees; `## Next: <n> requirement(s) not yet due`, by the helper, a nested list                        |
| provenance | Under each cause, after its tree: one item per fact, `**<name>.** <value>`, then the cause's commands in a `Pre`; at the end the view's commands |

- **The count line** reads `<n> causes: ` and the five kinds, each by the helper: `unmet requirement(s)`, `status(es) off nominal`, `review(s) outstanding`, `authorisation(s) outstanding`, `snapshot(s) not advanced`.
- **A cause** reads, by kind: ``<task> has not passed <gate>``; ``<task> stands at <status phrase>: <note>``; ``<task> awaits review at <gate>``; ``<task> is proposed``; ``<task> waits on `<url>` at `<pin>` ``. **Who acts** reads the mark's symbol, the email and the act: `👀 nbyoung@nbyoung.com reviews`. At glance these fill the table's cells; at detail the item's first line joins them in bold and plain: `**<cause>** · <who acts> · holds <n>`.
- **A held task** reads ``<task> at <gate>: `` and, by `Via`: `the cause holds the task itself.`; ``the parent of `<child>`.``; ``requires `<id>` at <from gate>, <text>.`` A parent's title prints in bold. `Also under cause <n>.` follows for each number in `Also`.
- **A task with requirements not yet due** reads ``<task, title in bold> stands at <status phrase>; next <gate>.`` and beneath it, per entry: ``<to gate>: requires <task> at <from gate>, <text>: <condition>.`` (R16).
- **The empty form (R9).** With no cause the count line stands, and `**No cause holds any task.**` replaces the table or the list; the requirements not yet due still follow.

Related views: queue, audit, task, tableau, history.

### The contributor work queue

```go
// Queue is the contributor work queue of one person.
type Queue struct {
	Head
	Items []QueueItem     `json:"items"` // kind order, then display order; reaffirmations oldest first
	Brief json.RawMessage `json:"brief"` // set with the brief parameter; task e4c7 decodes it
}

// QueueItem is one thing the person does.
type QueueItem struct {
	Kind   string  `json:"kind"`  // review, authorisation, ready, reaffirmation, waiting
	Task   TaskRef `json:"task"`
	Gate   string  `json:"gate"`  // the next gate; a reaffirmation: the gate the status names
	Model  string  `json:"model"`
	Since  string  `json:"since"` // the date of the status
	Age    int     `json:"age"`   // reaffirmation: days from the status to the ref's commit
	Cause  string  `json:"cause"` // waiting, authorisation: worded text (R17)

	Junction       *Junction     `json:"junction"`        // detail from here (design e3ed)
	Sources        []FieldSource `json:"sources"`         // who supplies contributor, model, reviewer
	TaskReferences []Reference   `json:"task_references"`
	Requires       []Requirement `json:"requires"`
	Unblocks       []Requirement `json:"unblocks"`         // dependents this junction's gate meets
	AlsoRequiredBy []Requirement `json:"also_required_by"` // the other dependents
	ParentUnblocks []ParentEdge  `json:"parent_unblocks"`
	Status         *Status       `json:"status"`
	StatusCommit   *Commit       `json:"status_commit"`
}

type ParentEdge struct {
	Parent   TaskRef     `json:"parent"`
	Gate     string      `json:"gate"`
	Requires Requirement `json:"requires"` // Task is the dependent
}
```

| Level      | Blocks, in order                                                                                                                             |
|------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| glance     | The count line; the table Kind, Task, Gate, Model, Since                                                                                      |
| detail     | The count line; five sections, `## 1. Reviews owed: n` to `## 5. Work waiting: n`, each item a bold paragraph and a list                       |
| provenance | With a brief: the brief, by the hook of A6. Without one: the detail, since the queue's provenance is the brief and nothing else                 |

- **The count line** reads `<n> items: <n> review(s) owed, <n> authorisation(s) owed, <n> work ready, <n> reaffirmation(s), <n> work waiting.`
- **The glance table.** Kind reads `Review owed`, `Authorisation owed`, `Work ready`, `Reaffirmation`, `Work waiting`. Task is the task (R4), with `; ` and the cause after it where the item has one. Since is the date, and for a reaffirmation `<date>, <age> days`. The table takes the run fold (R10). A kind with no item gives no row: the count line names it.
- **The sections** always number five, in VIEWS.md's order, and an empty one reads `## 1. Reviews owed: none` (R9).
- **An item at detail** opens ``**<task>** at <gate>, model `<model>` `` (a reaffirmation: ``**<task>**, <gate>, <date>, <age> days``) and lists, each line only when the data holds it: `Gate:` the gate, its name and criteria from the legend; `Contributor:` the mark, the email, the model and ``from <task>``; `Reviewer:` 👀's symbol from the legend, the email, and ``from <task>`` or `the assignee`; `References:` the junction's, then the task's, as links (R13); `Requires:`, `Waits for:` (the cause, in a waiting item), `Unblocks:`, `Also required by:` and `Its parent unblocks:`, each entry as ``<task>, <from> → <to>, <text>: <condition>`` and nested when more than one; `Status:` the phrase, the date, the recorder, the commit (R6) and the note after a colon; and one command line by kind: `Brief:` for work ready, `To accept:` for a review, `To authorise:` for an authorisation, `To reaffirm:` for a reaffirmation (A5). A waiting item ends `Do not start: the cause stands.`
- **The fold at detail.** Items that fold at glance fold here: after the first three, one paragraph reads ``**… and <n> more** at <gate>, model `<model>`: <ids>.``
- **The empty form.** `**The queue is empty.**` follows the count line at glance, as the draft shows; at detail the five headings read `none`.

Related views: task, blockage, assignment, context, tableau, authority.

### What the design keeps from the prototype

From `prototype/a3cc` it keeps the GFM table with `&nbsp;` indentation, bold parents, the count in the folded header, the hidden-row count after a title, and the nested list for the tree. It drops the Status column inside the grid (the notes table replaces it), the `◀` mark of the corner (the role labels replace it), the renderer's own table of gate symbols (the legend arrives in the data), and the "Not derived" line. The prototype's question of `&nbsp;` against a nested list is settled for `&nbsp;`: a list cannot hold the gate columns.

## Conformance to the texts

| View               | VIEWS.md section                                                                                         | Golden reference            | glance                         | detail                                    | provenance                             |
|--------------------|----------------------------------------------------------------------------------------------------------|-----------------------------|--------------------------------|-------------------------------------------|----------------------------------------|
| Global tableau     | [Global tableau](https://github.com/nbyoung/tableaux/blob/main/VIEWS.md#global-tableau)                   | `docs/mockups/tableau.md`   | Rows to depth one, the notes   | Every row, the folded counts              | The commit, child, ancestor and pin behind each cell |
| Contextual tableau | [Contextual tableau](https://github.com/nbyoung/tableaux/blob/main/VIEWS.md#contextual-tableau)           | `docs/mockups/context.md`   | The task and its children      | The spine, the siblings, the notes, the folded counts | As the global tableau       |
| Work-blockage tree | [Work-blockage tree](https://github.com/nbyoung/tableaux/blob/main/VIEWS.md#work-blockage-tree)           | `docs/mockups/blockage.md`  | One line per cause             | The tree, then what is not yet due        | The facts and the command per cause    |
| Work queue         | [Contributor work queue](https://github.com/nbyoung/tableaux/blob/main/VIEWS.md#contributor-work-queue)   | `docs/mockups/queue.md`     | One line per item              | Each item expanded, in five sections      | The brief (`e4c7`)                     |

The owner's rulings stand: the window belongs to the two tableaux, and the queue and the tree list every gate they name; a historical cell arrives empty unless its parameter is on, and the renderer draws what arrives. Where a mockup cannot be reproduced:

| #   | Mockup   | What it shows                                                                                              | What the renderer does                                                              | Kind |
|-----|----------|------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------|------|
| V1  | context  | The fold count in the first row, a run headed `🔗 … 🚀`, empty cells of three spaces                        | The count in the header, `🔗…🚀 ×0`, as `tableau.md`                                 | e3ed D5 |
| V2  | context  | Date and Note as two columns of the grid at detail                                                          | The notes table beneath the grid, as `tableau.md`                                   | D1 |
| V3  | tableau  | A paragraph that explains rows, columns and cells                                                           | The key line (R2), which names each symbol the grid shows                           | e3ed D9 |
| V4  | tableau  | Notes in an order no rule gives (the twenty leaves before the root); `From` reads `Its status file`          | Display order of each group's first row; an empty From for a status file            | Finding |
| V5  | both     | Bold labels (`**Status cells.**`) over the provenance tables; a "Why that child" column of reasoning         | `##` headings with counts; Parent, Shows, Rolls up from, Date                        | e3ed D9 |
| V6  | both     | Marks rows that merge gates (`📏 🔗`, `📌 to 🚀`) and name rows in words ("the five parents and the twenty mockups") | One line per gate and source, every id listed                               | e3ed D6 |
| V7  | both     | Folded headers for gates outside the window; state and reason symbols; the recorder and commit of each cell; roll-up children; snapshots | Reads `legend`, `columns`, `statuses`, `rollups`, `mark_groups`, `snapshots`; prototype `886d` emits symbols for the window's columns only, no provenance, and counts leaves in a fold | A2, A3 |
| V8  | tableau  | Four "Variations" sections                                                                                  | Each is one rendering with another parameter; T4 covers them                         | T4 |
| V9  | blockage | At detail with no cause, a table "Kind of cause, The project at the ref, Causes"                            | The count line, which gives the five counts; the middle column is commentary         | e3ed D9 |
| V10 | blockage | The step from a child to its parent in the illustration, which VIEWS.md does not define                     | Draws `via: parent` when tablo emits it                                             | Finding, A4 |
| V11 | blockage | Provenance as headed prose and fenced task-file lines                                                       | A list of worded facts per cause, and commands                                       | A4 |
| V12 | blockage | Review outstanding and snapshot not advanced                                                                | Draws them; prototype `886d` derives neither and says so                             | A4 |
| V13 | queue    | A `none` row per empty kind in the glance table; a sentence under each empty section                        | The count line and `none` on the heading (R9)                                        | e3ed D7 |
| V14 | queue    | `Reaffirmations` beside `Reaffirmation` as the Kind of one table                                            | The singular in every row                                                            | Finding |
| V15 | queue    | `Requires: nothing`, `Unblocks: nothing`; "from design to design"                                           | Omits an empty line; `design → design` (R16)                                         | e3ed D3 |
| V16 | queue    | The model, the since date, the junction and its sources, the references, the unblock lists, the age          | Reads them; prototype `886d` emits kind, task, title, cause and a dependents count   | A5 |
| V17 | queue    | Two "Illustrative items"                                                                                    | The same forms from fixtures: a review owed and a work waiting item                  | T8 |

## Assumed interfaces

Design e3ed's A1 to A10 hold here. This task adds:

| #  | Of                     | Assumption                                                                                                                              |
|----|------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| A1 | tablo `886d`           | Both tableaux arrive in one shape, `view.Tableau`; `params.window` or `params.columns` echoes what is in force, the default window included |
| A2 | tablo `886d`           | `columns` covers every gate once; a folded column carries the count of tasks, parents included or not as VIEWS.md settles, and the renderer prints the number it gets. `rows` holds the rows of the level with `hidden` counts; each row holds one cell per column; a cell's `kind` is `empty`, `status`, `marks` or `exempt` |
| A3 | tablo `886d`           | A row's `status.from` names the leaf a roll-up takes its note from, or the subproject's task; provenance adds `statuses`, `rollups` (the direct child), `mark_groups` and `snapshots` |
| A4 | tablo `886d`           | The blockage data holds all five kinds of cause, `kinds` with a count for each, `holds` as distinct tasks, `held` as a tree with `via` in `self`, `parent`, `requirement`, the worded `action`, and at provenance worded `facts` and `resolve` commands |
| A5 | tablo `886d`           | The queue data holds the fields of `view.QueueItem`; `age` counts days to the ref's commit date, not to the clock; `kind` takes the five listed values |
| A6 | tabloio `e4c7`         | Agent briefs adds `internal/render/brief.go` with `func buildBrief(q *view.Queue, o Options) (doc.Doc, error)` and decodes `Queue.Brief` itself. Until it lands, this task's stub returns `ErrFormat` for a queue with a brief at provenance |
| A7 | tabloio `5ca9`, `e4c7` | The brief's command reads `tabloio queue --for <email> --ref <ref> --brief <task> <gate>`, as the mockup writes it; `tableau` and `context` take `--window`, `--columns` and `--historical`; `context` takes `--task` or `--for`; `queue` and `blockage` take `--task` and `--for` |
| A8 | tabloio `b618`         | The write commands read `tabloio review <task> <gate>`, `tabloio authorise <task>` and `tabloio reaffirm <task>`; the queue prints them as the command of an item |

## Tests

The "for every fixture" tests of design e3ed (T6, T7, T8, T10) cover this task's fixtures.

| #   | Proves                                                                   | How                                                                                                             |
|-----|--------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------|
| T1  | The global tableau reproduces the mockup at glance                        | The fixture `tableau/tooling-glance.json`, transcribed from `tableau.md`, against `a3cc/tableau.glance.md`       |
| T2  | The contextual tableau reproduces the mockup at glance, with a folded run | `context/tooling-2034.json` against `a3cc/context.glance.md`                                                    |
| T3  | Detail and provenance of both, and the person form                        | `tableau/tooling.json`, `context/tooling-2034.json`, `context/tooling-nbyoung.json`; golden files the implementation generates and the reviewer reads against the mockups |
| T4  | The parameters: historical junctions, a person marked, window 0, a reason beside a state | Four fixtures transcribed from the Variations of `tableau.md`: cells before the status filled; `[🤖👀]`; the header `❔…📝 ×29`; the cell `🟢👓` |
| T5  | The grid on the corpus                                                    | `weather-station`, transcribed from `prototype/a3cc/testdata`: `tableau/weather.json`, `tableau/weather-window0.json`, `context/weather-4e2b.json`, `context/weather-ben.json`: a stalled and blocked cell, an undefined task, a recursive junction, a roll-up |
| T6  | The notes group and order; the section vanishes when empty                | Rows that share date, note and origin; a row with neither; a fixture with no note at all                         |
| T7  | A row with the wrong number of cells is an error that names the row       | A fixture with a cell removed                                                                                   |
| T8  | The queue: glance with folds, the empty form, each kind at detail         | `queue/tooling-agent.json` against `a3cc/queue.glance.md`; `queue/tooling-owner.json` against `a3cc/queue.empty.md`; `queue/review-owed.json` and `queue/waiting.json` from the mockup's illustrative items; `queue/weather-ada.json` for an authorisation owed and a blocked item |
| T9  | The queue's provenance                                                    | With `brief` set, the builder calls the hook, and the stub returns `ErrFormat`; without it, provenance equals detail but for the frame's level and ref |
| T10 | The blockage tree: glance, detail, the empty form                         | `blockage/illustration.json` against `a3cc/blockage.detail.md`, and at glance; `blockage/tooling.json`, with no cause and ten requirements not yet due; `blockage/weather.json` for a status cause and an authorisation cause |
| T11 | The words of each kind of cause, act and held task                        | A table of cases, one per kind and per `via`                                                                    |
| T12 | The count helper                                                          | 0, 1 and 2 of each label                                                                                        |

Already tried, outside the repository: the five drafts are script output over `tableau.md`, `context.md` and `queue.md` with the edits V1 to V15, and the blockage draft is the mockup's illustration reworded by the rules above; the run fold of the queue's glance gives the mockup's own two folded rows. Not checked: how a viewer draws a grid of eleven emoji columns, and the widths; no claim here rests on either. The unit gate proves T1 to T12. The integrate gate regenerates the fixtures from `tablo view` and joins the brief of `e4c7`. The validate gate is the owner reading `STATUS.md`, which design `1422` renders with `tableau --level glance`.

## Implementation notes

The task follows `e3ed` for the shared packages and then adds `grid.go`, `blockage.go`, `queue.go`, `count.go`, `view/tableau.go`, `view/blockage.go`, `view/queue.go` and its test data; it touches the switches in `Render` and `Decode` and no other shared file. `e0f7` requires this task at implementation for its design; what it needs is in `grid.go` and the `doc.Table` it returns. `prototype/a3cc/` stays until the unit gate passes, then goes. No dependency.

Two points for the plan and the method, not edited here. The queue's brief crosses three tasks, `a3cc`, `e4c7` and `5ca9`, and the plan links `e4c7` to `5ca9` alone; A6 is the seam, and a requirement of `e4c7` on `a3cc` at implementation would state it. VIEWS.md does not say whether a cause holds a parent whose child it holds (V10), nor whether a fold counts parents (A2); tablo's `886d` design settles both, and the renderer follows the data.

## Decisions at review

1. **The notes stand beneath the grid in both tableaux,** grouped by date, note and origin. The alternative keeps the Date and Note columns that `context.md` draws inside the grid, which repeats one note on twenty-three rows there and makes the two tableaux two layouts. Binds tableaud `44bf` and tablotui `679b` only if the owner wants one layout across front ends.
2. **The notes follow display order** of each group's first row (V4). The alternative reproduces the mockup's order, which needs a rule the mockup does not state.
3. **The renderer words a cause, an act and a held task** from structured data, and tablo words only the action and the provenance facts (R17). The alternative has tablo emit each line as text, which moves the wording of a view into the backend and out of reach of the text and HTML forms.
4. **The queue's provenance without a brief is its detail.** The alternative makes `--level provenance` without `--brief` a usage error, which is the Read commands' to raise; the renderer then never sees the case.
5. **`Reaffirmation` in the singular** as a Kind (V14), and `Review owed`, `Authorisation owed` likewise, with the plurals in the count line and the section headings.
6. **An empty line of an item does not print** (V15): no `Requires: nothing`. The alternative prints every line of every item, so that items align line for line.
