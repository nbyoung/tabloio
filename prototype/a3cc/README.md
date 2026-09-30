# Prototype a3cc: status renderers in Markdown

## Question

Can one renderer turn tablo's status view data into readable Markdown for all four status views (global tableau, contextual tableau, work-blockage tree, work queue) without reading `.tableaux`? The risky part is the tableau: a wide table of emoji cells, folded columns and an indented tree that a plain Markdown renderer must still show as a table.

## Run

```
go run ./prototype/a3cc [-level glance|detail] prototype/a3cc/testdata/global-tableau.json
go test ./prototype/a3cc
```

The renderer dispatches on the `view` field of the JSON, read from a file argument or standard input. It uses the standard library only.

## Test data

Every file in `testdata/` is a copy from `/home/nbyoung/Projects/Tableaux/tablo/prototype/886d/output/` (tablo prototype 886d, corpus entry `weather-station` at `main`), made by `go run ./prototype/886d` in tablo:

| File | 886d command |
|------|--------------|
| `global-tableau.json` | `tableau` |
| `global-tableau-window0.json` | `tableau -window 0` |
| `contextual-task-4e2b.json` | `contextual -task 4e2b` |
| `contextual-person-ben.json` | `contextual -person ben@example.org` |
| `work-blockage-tree.json` | `blockage` |
| `queue-ada.json`, `queue-ben.json`, `queue-dan.json` | `queue -person <person>` |

## What it shows

Global tableau at detail, window 0 (`global-tableau-window0.json`). The column before the window folds to a count of 1 (`7b2e` stands at undefined):

```markdown
## Global tableau on `main`

| Id | Task | ❔ ×1 | 📝 | 📌 | ⚙️ | ⚡ | ⚓ | 📐 | 🛠️ | 🧩…🚀 ×0 | Status |
|---|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|---|
| `a1c0` | **Weather station** |  | 🟢 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  | rolls up from `3c5d`; 2026-09-17 |
| `4e2b` | &nbsp;&nbsp;**Sensor node** |  | 🧑 | 🧑 | 🔴⛔ | 🧑 | 🧑 | 🧑 | 🧑 |  | Barometer ICs on 14-week backorder; rolls up from `9f31`; 2026-09-17 |
| `9f31` | &nbsp;&nbsp;&nbsp;&nbsp;Sensor board |  | 🧑 | 🧑👀 | 🔴⛔ | 🧑 | 🧑 | 🧑 | 🧑 |  | Barometer ICs on 14-week backorder; 2026-09-28 |
| `c07d` | &nbsp;&nbsp;&nbsp;&nbsp;Node firmware |  | 🧑 | 🧑 | 🧑 | 🧑 | — | 🟢 | 🪆 |  | Sleep scheduler in progress; subproject at `F2`; 2026-09-17 |
| `7b2e` | &nbsp;&nbsp;Gateway |  | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  | 2026-09-15 |
| `3c5d` | &nbsp;&nbsp;Dashboard |  | 🟢 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  | 2026-09-22 |

Folded columns show the count of leaf tasks whose current gate lies in them.
```

Global tableau at glance (depth one; the parenthesis counts the hidden rows):

```markdown
## Global tableau on `main`

| Id | Task | ❔ | 📝 | 📌 | ⚙️ | ⚡ | ⚓ | 📐 | 🛠️ | 🧩 | 🖼️…🚀 ×0 | Status |
|---|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|---|
| `a1c0` | **Weather station** |  | 🟢 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  | 2026-09-17 |
| `4e2b` | &nbsp;&nbsp;**Sensor node** (2) |  | 🧑 | 🧑 | 🔴⛔ | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  | 2026-09-17 |
| `7b2e` | &nbsp;&nbsp;Gateway | ⚪ | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  | 2026-09-15 |
| `3c5d` | &nbsp;&nbsp;Dashboard |  | 🟢 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  | 2026-09-22 |

Folded columns show the count of leaf tasks whose current gate lies in them.
```

Work-blockage tree at detail:

```markdown
## Work-blockage tree on `main`

- **9f31 Sensor board is stalled at function: Barometer ICs on 14-week backorder** (ada@example.org, holds 2)
  - action: clear the blocked reason, or record a new status
  - `9f31` Sensor board at ⚡ performance
  - `c07d` Node firmware at 🛠️ implementation
- **9f31 Sensor board has not passed design** (ada@example.org, holds 1)
  - action: record 9f31 through design
  - `c07d` Node firmware at 🛠️ implementation
- **3c5d Dashboard is proposed** (ada@example.org, holds 1)
  - action: commit a trailer: Authorised: 3c5d
  - `3c5d` Dashboard at 📌 mockup

Next, not yet due:

- `3c5d` requires `7b2e` (Readings API) from function to integrate

Not derived: review outstanding, subproject pin not advanced.
```

Work queue for ada@example.org at detail:

```markdown
## Work queue for ada@example.org on `main`

| Kind | Task | Gate | Detail | Dependents | Status date |
|---|---|---|---|:-:|---|
| authorisation owed | `3c5d` Dashboard |  | proposed; the deciding commit is not by an authority | 0 |  |
| work ready | `7b2e` Gateway | 📝 defined |  | 1 | 2026-09-15 |
| reaffirmation | `9f31` Sensor board | ⚙️ function |  | 0 | 2026-09-28 |
| work waiting | `9f31` Sensor board | ⚡ performance | blocked: Barometer ICs on 14-week backorder | 1 | 2026-09-28 |
```

The contextual tableau for a person (`contextual-person-ben.json`) marks the corner with ◀; the task form (`contextual-task-4e2b.json`) renders the same way. The tests check each view at both levels, that every table row has the header's column count, and that an empty queue (`queue-ben.json`) reads as empty.

## Layout choices VIEWS.md leaves open

For the mockups and the design gate to settle:

- Tableau as a GFM table: Id, Task, one column per gate in the window, then a Status column. The example in VIEWS.md has no note column; the prototype adds one (note, roll-up source, subproject pin, date) at detail and only the date at glance.
- Indent by `&nbsp;&nbsp;` per depth, as the VIEWS.md examples do; parents bold; ids in code spans.
- A folded run is one extra column whose header carries the count (`🖼️…🚀 ×0`), and its cells stay empty. VIEWS.md shows the count in the root row's cell (`0`).
- Glance hides rows below depth one and appends the hidden row count to a depth-one task.
- Contextual tableau: `◀` marks the corner; spine and siblings carry no mark; a legend line follows the table.
- Blockage tree as a nested bullet list, not a code block: cause in bold with resolver and count; at detail the action, then each held task with its gate symbol and name. Not-yet-due requirements and the not-derived causes follow at detail.
- Queue as a table (Kind, Task, Gate, Detail); detail adds dependents and the status date. The agent's Model column and the brief are not rendered.
- Headings name the view, the task or person, and the ref.

## Gaps in the view data

- The view data lists symbols only for the columns in the window. The renderer carries its own gate-to-symbol table for folded columns. The data should carry the symbol of each folded gate.
- The fold `count` counts leaf tasks (886d README); VIEWS.md says tasks.
- The data holds no reason text for a queue item of kind `work ready`; its Detail cell stays empty.
- The queue data has no model and no brief, so those parts of VIEWS.md are not rendered.
- The data holds no recorder or commit for a parent row (only `rolled_up_from`), so provenance is not rendered at all.
- The queue's `3c5d` authorisation item has no gate and no date.

## What it leaves out

The `--text` Unicode format, the provenance level, the brief, the historical-junctions parameter, `person` marks on the global tableau, alignment padding of the Markdown source, and every command-line flag beyond `-level`. Emoji width is not handled (Markdown tables need no padding).

## What the design gate must decide

- The layout choices above, chiefly where a folded column's count goes and whether the Status column belongs in the table.
- Whether the tableau's Markdown form uses `&nbsp;` indentation (renders everywhere, unreadable as source) or a nested list.
- Whether view data carries the gate symbols and the row text, so renderers hold no table of their own.
