# Prototype e3ed: the legend and task renderers

Task `e3ed` Legend and task renderers, gate `function`.

## The question

Can one small renderer turn the gate definition and task definition view data, as `tablo view` will emit it, into Markdown that follows VIEWS.md at its three levels, with no knowledge of the project beyond the JSON?

The risky parts are the nesting of the levels (each includes the one before), the facts the task view leaves to the legend (state, reason and gate symbols), and the expansion of the legend for one task.

## Run

```
go run ./prototype/e3ed [-level glance|detail|provenance] [-gates testdata/gate.json] testdata/task-c07d.json
go test ./prototype/e3ed
```

The file's `view` key selects the renderer. `-gates` gives a task view the symbols of a gate view; without it the method's default symbols apply.

## Test data

All files come from tablo's prototype `493e` on the corpus entry `weather-station` at `main`, with `cd tablo && go run ./prototype/493e -repo <corpus>/weather-station -ref main ...`:

| File                        | Flags                   |
|-----------------------------|-------------------------|
| `testdata/gate.json`        | `-view gate`            |
| `testdata/gate-c07d.json`   | `-view gate -task c07d` |
| `testdata/task-c07d.json`   | `-view task -task c07d` (a recursive junction, an unmet requirement, a snapshot) |
| `testdata/task-9f31.json`   | `-view task -task 9f31` (dependents, references, a stalled status) |
| `testdata/task-4e2b.json`   | `-view task -task 4e2b` (a parent with a rolled-up status) |
| `testdata/task-a1c0.json`   | `-view task -task a1c0` (the root) |

Each holds all three levels; the renderer picks by `-level`.

## What it shows

The legend at glance:

```
| ❔ undefined | 📝 defined | 📌 mockup | ⚙️ function | ⚡ performance | ⚓ reliability | 📐 design | ... |
|:-:|:-:|:-:|...

States ⚪ undefined · 🟢 nominal · 🟡 at_risk · 🔴 stalled · ✅ complete.

Reasons 🪫 overloaded · ⛔ blocked.

Marks 🤖 an agent contributes · 👀 a reviewer accepts · 🪆 a subproject does the work · — the gate does not apply · 🧑 a person contributes.
```

A parent task at glance (`4e2b`), with the roll-up named:

```
> `4e2b` **Sensor node** — ben@example.org — under `a1c0` Weather station, order 1 — ⚙️ function 🔴 stalled ⛔ blocked, 2026-09-17, rolled up from `9f31`: Barometer ICs on 14-week backorder
```

A leaf at detail (`c07d`, excerpt):

```
| Task | From | To | Text | Condition |
| `9f31` Sensor board | design | implementation | Pin map and sensor bus | unmet, due |

| Gate | Marks | Contributor | Model | Reviewer | Subproject |
| 🛠️ implementation | 🪆 |   |   |   | `f1a0` at `firmware` |
| 🧩 unit | 🤖👀 | opus@example.org | claude-opus-5-5 | ben@example.org |   |
```

At provenance the page adds the authorisation commit and way, the review commits, the fields that resolve from another task, the events and the command that lists them all. The legend for a task (`gate-c07d.json`) marks the gates that do not apply with —.

The tests check each level, that the levels nest by prefix, the expanded legend, the roll-up and root lines, dependents and references, and the errors. They read only `testdata/`.

## What it leaves out

The text format. The `ref`, `window`, `columns` and `person` parameters: tablo applies them and the renderer shows what arrives, except that it reads `folded` and prints a count. A config for marks. Wrapping and width. Links between views. The README's "What it does" promises `--markdown` and `--text`; only Markdown is here.

## Gaps between the JSON and VIEWS.md

- **State and reason symbols.** The task view carries gate symbols (in its junctions) but not state or reason symbols. The renderer takes them from a gate view via `-gates`, else from built-in defaults. The data contract could carry them.
- **Reserved reason `review`.** The legend's JSON lists overloaded and blocked only. The renderer adds the row 👓 review from a built-in text.
- **Children.** The task view lists child ids without titles or statuses, so a parent shows ids only.
- **Junction defaults a parent states** (Authority delegation, not these views) and the `how its url fixes` a snapshot commit: not in the data.
- **Status commit** at provenance is `null` for the tasks shown, so the status commit line does not appear.
- **Recorder** is absent from a rolled-up status; the line names the child instead.
- **Legend name vs key.** The JSON gives both; the prototype prints the key in the header and the name in the detail table.
- **Staleness.** `493e` predates corpus changes: `4e2b` rolls up from `9f31` here as the corrected README rule gives.

## Layout choices VIEWS.md leaves open, for the mockups and the design gate

1. The legend's gate row: a header-only table, symbol and key, centred, with no body row. Whether symbol and name, or key, and whether a body row belongs there.
2. States, reasons and marks as prose lines at glance and tables at detail. The VIEWS example prints severities on the glance line; here they wait for detail.
3. The task at glance: one block quote line, as the VIEWS example shows, with the parent and order inline, plus a second quote line for a subproject snapshot.
4. The task at detail: second-level headings for description, references, children, requirements, dependents and junctions, each a table or list. The requirement condition reads "unmet, due" (the word from the data, then whether due), while VIEWS lists "met, due, unmet".
5. A junction row shows marks joined without spaces, with the model and reviewer in their own columns and a subproject as `id` at `url`.
6. Provenance as a bullet list plus an events table, with fields resolved from `default` left out.
7. A title line: `# Legend` for the gate view; none for the task view, which opens with the quote.
8. Reference links: `[text](url)` or `<url>` when the text is absent.
9. Output is UTF-8 with the symbols as the data gives them; no alt text for a plain renderer.

## What the design gate must decide

- Whether the view data carries state and reason symbols and the `review` reason, or the front end owns them.
- Whether the renderer ships as a library function per view (as here) or a table of view renderers keyed by the `view` value.
- How the mockups settle the layout choices above, and whether the Unicode text renderer reuses this structure.
- Whether `tabloio` calls `tablo` in process or reads JSON from a pipe, since this prototype reads files.
