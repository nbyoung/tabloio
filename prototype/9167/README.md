# Prototype 9167: the temporal renderers

## The question

Can the history and audit view data that `tablo` emits render as Markdown at the three levels of VIEWS.md (language 0.3.1), with no access to the repository and no library but the standard one?

Yes. One small program reads the JSON records and writes a table per view. Events from one commit on one task merge into one row, findings sort errors first, and the level decides the columns and the detail.

## How to run

```
export PATH=$PATH:/usr/local/go/bin
go run ./prototype/9167 -view history|audit -in <file> [-level glance|detail|provenance] [-range R] [-task ID] [-person EMAIL] [-stale DAYS]
go test ./prototype/9167
```

`-range` and `-stale` only label the output: they say what the data covers. The tests read `testdata/` only.

## Test data

All files come from tablo's prototype 8ed1 on the corpus entry `weather-station` (`cd tablo && go run ./prototype/8ed1 -repo <corpus>/weather-station -view ... -range ...`). The 8ed1 prototype emits a JSON array, not the `tablo` envelope of prototype 4ed9; the renderer reads the array.

| File                            | Command arguments                                     |
|---------------------------------|-------------------------------------------------------|
| `history-all.json`              | `-view history -range ..W13`                          |
| `history-w9-w13.json`           | `-view history -range W9..W13`                        |
| `history-task-9f31.json`        | `-view history -range ..W13 -task 9f31`               |
| `audit-w3-w13.json`             | `-view audit -range W3..W13`                          |
| `audit-w3-w13-stale3.json`      | `-view audit -range W3..W13 -stale 3`                 |
| `audit-w4-w5.json`              | `-view audit -range W4..W5`                           |
| `audit-w13-w13.json`            | `-view audit -range W13..W13` (no findings)           |

## What it shows

`-view history -in testdata/history-w9-w13.json -level detail -range W9..W13`:

```
# History

For range `W9..W13`.

6 events in 4 commits, 2026-09-25 to 2026-09-28.

| Date | By | Task | Event | Detail |
| --- | --- | --- | --- | --- |
| 2026-09-25 | ada@example.org | `9f31` | status | function stalled (blocked): Barometer ICs on 14-week backorder |
| 2026-09-26 | ben@example.org | `c07d` | reviewed, status, pin | at mockup; design, no state; pin firmware at cf5b169 |
| 2026-09-27 | dan@example.org | `3c5d` | task |  |
| 2026-09-28 | ada@example.org | `9f31` | reaffirmed |  |
```

`-view audit -in testdata/audit-w3-w13-stale3.json -level detail`:

```
# Audit

2 warnings (1 introduced, 1 resolved).

| Rule | Severity | Task | Range | Message |
| --- | --- | --- | --- | --- |
| PROPOSED | warning | `9f31` | resolved | The task stands proposed: no trailer or owner commit on the trunk authorises it |
| STALE | warning | `3c5d` | introduced | The status dates from 2026-09-22, more than 3 days before 2026-09-28, with no later reaffirmation |
```

The levels:

- History glance: date, actor, task, event, and the gate and state of a status, the gate of a review, the note of a pin. Detail adds the status note. Provenance adds the commit column and the `git log` command.
- Audit glance: the counts by severity, then one row per rule and severity with its tasks. Detail: one row per finding. Provenance adds the `since` commit and the commands.

## What it leaves out

- The data gaps against VIEWS.md. The 8ed1 records carry no committer, model, status after each event, pin `url` or old and new commit, subject, trailers or files, so detail and provenance show less than VIEWS.md lists. The audit records carry no gate, file, resolver or "resolves when" text, so the findings table lacks those columns and "the person with the most to resolve" is absent.
- The pin event holds only a note ("pin firmware at cf5b169"), so the pin row reads as text, not as `url` with old and new commit.
- The subproject's events between two pins, the proposal marking at a branch ref, the `window` parameter and subtree replay.
- The `role` parameter and its default level; the caller names the level.
- The `tablo` envelope, the `--text` Unicode form, flag parsing for the real command, and CI regeneration. The code is a library shape (`History`, `Audit`) that `internal/render` can take.
- The audit's rule sentences: rules appear by the stand-in names of 8ed1 (PROPOSED, STALE), not by VIEWS.md's H1 to H6, R9 and J-numbers.

## What the design gate must decide

Layout choices VIEWS.md leaves open, for the mockups (bc86 and its leaves d615, b14e) to settle:

- One row per commit and task (events merged, as the VIEWS.md example does) or one row per event.
- The columns: Date, By, Task, Event, Detail, with Commit only at provenance. Whether the commit shows at glance, as the example does in its detail cell.
- Whether status reads "function stalled (blocked)" or as the gate symbols and state symbols the gate definition explains.
- Whether a note follows a colon in the detail cell or has its own column.
- The heading, the scope line ("For range ...") and the count line; whether the glance audit shows a table or one sentence.
- The audit glance table grouped by rule and severity; the ordering (errors first, then rule, then task) and whether `Range` (introduced, standing, resolved) stays a column, which depends on the open choice in 8ed1 about which findings a range lists.
- Where the reproduction commands go (a fenced block after the table) and their form.
- An empty view: "No events." and "No findings.".
- Whether provenance for a pin event lists the subproject's own events nested under the row.
