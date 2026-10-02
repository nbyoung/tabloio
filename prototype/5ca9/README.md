# Prototype 5ca9: the read commands' surface

Task `5ca9` Read commands, gate `function`.

## The question

Can one subcommand per view take the focusing parameters VIEWS.md defines as flags, pass exactly those to `tablo view` in the plumbing envelope's shape, and write the view as Markdown or Unicode text to stdout or a file, with usage errors caught before `tablo` runs?

The riskiest part is the mapping: VIEWS.md has seven shared parameters plus `brief` and `stale`, each view names its own subset, and `tablo view` has no spelling for them yet.

## Run

```
export PATH=$PATH:/usr/local/go/bin
go test ./prototype/5ca9
go run ./prototype/5ca9 <command> --testdata prototype/5ca9/testdata [flags]
```

Without `--testdata` the prototype runs `tablo` (`--tablo PATH` names it). With it, a stand-in plays `tablo view`: it parses the very argument vector tabloio composed, checks its shape, and answers from one fixed testdata file wrapped in the envelope. It applies no focusing, so `--task` changes the vector, not the answer (except `context`, where `--task 4e2b` picks the task form's file, and `task`, which has data for `9f31` only).

## The surface

| Command      | `tablo view`           | Flags beyond the common ones                          |
|--------------|------------------------|-------------------------------------------------------|
| `gates`      | `gate-definition`      | `--task --window --columns`                           |
| `task <id>`  | `task-definition`      | `--for --window --columns`                            |
| `authority`  | `authority-delegation` | `--task --for --proposed`                             |
| `assignment` | `task-assignment`      | `--task --for --window --columns`                     |
| `queue`      | `work-queue`           | `--task --for --window --columns --brief TASK:GATE`   |
| `blockage`   | `work-blockage-tree`   | `--task --for --window --columns`                     |
| `tableau`    | `global-tableau`       | `--for --window --columns --historical`               |
| `context`    | `contextual-tableau`   | `--task --for --window --columns --historical`        |
| `history`    | `history`              | `--task --for --window --columns`                     |
| `audit`      | `audit`                | `--task --for --stale DAYS`                           |

Common to all: `--markdown` (default) or `--text`, `-o FILE` (`-` is stdout), `-C DIR`, `--ref REF`, `--role ROLE`, `--level glance|detail|provenance`, `--tablo PATH`, `--testdata DIR`.

The vector: `tablo [-C dir] [--ref r] --json view <name> [--task t] [--person p] [--window n | --columns a,b] [--historical-junctions] [--proposed] [--stale n] [--brief t:g] [--role r] [--level l]`. `--for` becomes `--person`: README.md writes `--for`, VIEWS.md writes person. A flag a view does not name is a usage error (exit 2), so a typo never reaches `tablo`. Exit codes follow 4ed9: 0, 1 (an error diagnostic), 2 (usage), 3 (a read failure, `tablo`'s own code passes through).

Checked before `tablo` runs: `--window` excludes `--columns`; `--markdown` excludes `--text`; `context` takes `--task` or `--for`, not both; a task is four lowercase hex digits; `--for` holds an email; level and role come from the fixed lists; a range (`a..b`) belongs to `history`; `--brief` reads `TASK:GATE`. The `task` subcommand takes its id as a positional argument, before or after the flags. Nothing writes to the project: the only write is `-o`.

## What it shows

```
$ tabloio queue --for ada@example.org --text --testdata prototype/5ca9/testdata
Queue
═════

work-queue at HEAD, level glance

- order: kind
- person: ada@example.org

items
─────

┌──────┬─────────────┬────────────────────┬──────────────┬──────────────────────────────────────────────────────┬────────────┬─────────────┐
│ task │ gate        │ kind               │ title        │ cause                                                │ dependents │ status_date │
├──────┼─────────────┼────────────────────┼──────────────┼──────────────────────────────────────────────────────┼────────────┼─────────────┤
│ 3c5d │             │ authorisation owed │ Dashboard    │ proposed; the deciding commit is not by an authority │ 0          │             │
│ 7b2e │ defined     │ work ready         │ Gateway      │                                                      │ 1          │ 2026-09-15  │
...
```

`TestArgvMapsFlags` pins the vector for nine commands. `TestExecRunnerPlumbing` runs a shell script as `tablo` and checks the arguments it receives, a warning diagnostic on stderr, and exit-code pass-through.

## Testdata

All from the `weather-station` corpus entry; tests read only `testdata/`.

| File                               | Produced by                                                              |
|------------------------------------|--------------------------------------------------------------------------|
| `global-tableau.json`, `contextual-task-4e2b.json`, `contextual-person-ben.json`, `work-blockage-tree.json`, `queue-ada.json` | copied from `tablo/prototype/886d/output/` |
| `gate-detail.json`, `authority-detail.json`, `assignment-detail.json` | `go run ./prototype/493e -repo <corpus>/weather-station -ref main -view gate\|authority\|assignment -level detail` |
| `task-detail.json`                 | the same with `-view task -task 9f31`                                    |
| `history-w9-w13.json`              | `go run ./prototype/8ed1 -repo <corpus>/weather-station -view history -range W9..W13` |
| `audit-w3-w13.json`                | `go run ./prototype/8ed1 -repo <corpus>/weather-station -view audit -range W3..W13 -stale 3` |

The 493e files nest `glance`, `detail` and `provenance`; 886d files are flat; 8ed1 prints a bare array. The stand-in flattens the first to the requested level and wraps the others in a `data` object with `events` or `findings`. The prototypes do not emit the 4ed9 envelope. The tablo prototypes also differ on names (`view` is `global-tableau` in 886d and `gate` in 493e), so the stand-in keys on the `tablo view` name, not on the file.

## Stand-ins and omissions

- The renderer is generic: a heading, the scalar fields, each list of objects as a table, each object as a section, keys sorted. The real renderers (task 1093) draw each view's own layout; nothing here proposes one. The prototype shows the two formats and the level plumbing, not a layout.
- The stand-in maps a role to a level (observer or none: glance; owner: provenance; other: detail). `tablo` resolves that, and the viewer's role at each item.
- No viewer default: tabloio passes `--person` only when `--for` is given and leaves the default (the email Git commits with) to `tablo`.
- No YAML input, no `--brief` rendering, no git or `.tableaux` access, no colour.

## For the design gate

- **The spelling of `tablo view`.** Names above are this prototype's proposal (`gate-definition` and so on, `--historical-junctions`, `--person`). The plumbing task owns the final names; tabloio follows.
- **Who resolves role and level.** Here tabloio forwards `--role` and `--level`, and `tablo` picks the level. The alternative is tabloio asking for provenance and dropping detail itself. The first keeps the rule in one place.
- **`--for` or `--person`.** README.md's example uses `--for`; the views call it person.
- **Ranges.** A range rides in `--ref` and only `history` takes it. A `--range` flag is the alternative.
- **`--brief`.** The README example shows a bare `--brief`; a brief needs a task and a gate (VIEWS.md), so the flag takes `TASK:GATE`. A bare flag could brief the first work-ready item.
- **Strictness.** A flag a view does not name fails here; VIEWS.md says it has no effect. Keep the error, or accept silently for one command line across views (for example `--task` on `tableau`).
- **Defaults of format.** Markdown is the default, as README.md says; a file named `*.txt` could imply `--text`, and this prototype does not.
- **Key order.** Go maps lose JSON key order. The renderers need ordered data from `tablo` (arrays) or a typed structure.
