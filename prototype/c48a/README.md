# Prototype c48a: structural renderers

Task `c48a` Structural renderers, gate `function`.

## The question

Can a renderer turn the JSON that tablo's structural-view prototype emits into Markdown for the authority delegation and task assignment views, at each of the three nested levels, without reading `.tableaux` and with the standard library alone?

The riskiest parts are the level nesting (provenance holds detail holds glance, and the JSON splits the facts across the three keys), the join of the assignment view's provenance sources onto its detail junctions, and a tree drawn in plain Markdown that survives a plain renderer.

## Run

```
export PATH=$PATH:/usr/local/go/bin
go test ./prototype/c48a
go run ./prototype/c48a -level glance|detail|provenance prototype/c48a/testdata/authority.json
```

The command reads the file, or standard input, and writes Markdown. The `view` key of the JSON chooses the renderer. The code is `main.go`; `Render` is the entry point.

## Test data

All files come from tablo's prototype `493e` (`cd tablo && go run ./prototype/493e -repo <corpus>/weather-station ...`), with the corpus at `tableaux/corpus/build/weather-station`, at tablo commit `7e94d58`:

| File                     | Flags                                          |
|--------------------------|------------------------------------------------|
| `authority.json`         | `-ref main -view authority`                    |
| `authority-person.json`  | `-ref main -view authority -person ada@example.org` |
| `authority-branch.json`  | `-ref 91aa774 -view authority` (off the trunk)  |
| `assignment.json`        | `-ref main -view assignment`                   |
| `assignment-person.json` | `-ref main -view assignment -person ada@example.org` |

## What it shows

Authority delegation at glance on `main`. A dot marks the parent's assignee; the root shows its own.

```
a1c0 Weather station    ada@example.org
  4e2b Sensor node      ben@example.org
    9f31 Sensor board   ada@example.org
    c07d Node firmware  ·
  7b2e Gateway          ·
  3c5d Dashboard        dan@example.org  proposed
```

At provenance the view adds a table of deciding commits:

```
| Task   | Commit    | Date       | Accepted as | Authority is | Author          | Committer       |
| ------ | --------- | ---------- | ----------- | ------------ | --------------- | --------------- |
| `9f31` | `f9e8746` | 2026-09-19 | merge       | author       | ben@example.org | ben@example.org |
| `3c5d` | `1b8cfb1` | 2026-09-27 | commit      | —            | dan@example.org | dan@example.org |
```

Task assignment at glance:

```
| Email            | Assigned | Contributes next | Reviews next | Models            |
| ---------------- | -------: | ---------------: | -----------: | ----------------- |
| ada@example.org  |        3 |                2 |            0 | —                 |
| ben@example.org  |        2 |                0 |            0 | —                 |
| dan@example.org  |        1 |                1 |            0 | —                 |
| opus@example.org |        0 |                0 |            0 | `claude-opus-5-5` |
```

At detail each person has tables of assigned tasks with status, subtrees under their authority, and junctions to contribute and to review with the next gate marked; at provenance a column names the task that states each field, or the plain default (`reviewer` `4e2b`, `contributor` default). On a ref off the trunk the tree marks every task proposed, as VIEWS.md asks. The tests in `main_test.go` read only `testdata/`.

## What it leaves out

- The `Text` format (Unicode boxes) and the other views.
- The parameters themselves: the JSON arrives filtered by tablo and the renderer only names `ref`, `task`, `person` and `proposed` in a header line.
- Folded subtrees ("10 children"): 493e does not fold, so the renderer shows none.
- Junction `window` and `columns` beyond what the JSON holds.
- The integration with `cmd/tabloio` and the `tablo` plumbing envelope (prototype `4ed9`).

## Gaps between the JSON and VIEWS.md

- The root row carries its assignee at glance in VIEWS.md's example; 493e marks `delegated: false` for the root and gives the assignee anyway. The renderer prints the root's assignee.
- The authority provenance `by` is empty for a proposed task (`3c5d`); the renderer shows a dash. VIEWS.md's provenance names the author and committer only, so the column "Authority is" is an addition that mirrors the JSON.
- VIEWS.md's detail level of the assignment view says "junctions by gate, with the model and the reviewer"; the JSON lists them flat, in task and gate order, and the renderer keeps that order.
- Reviews appear only for persons that review (`reviews` is null otherwise).
- `assigned` is a count at glance and a list at detail; the renderer takes either.
- The version of the language and the corpus differ from the VIEWS.md example, which describes the Tableaux project itself, so no figure there compares.

## Layout choices for the mockups and the design gate

1. The tree is a fenced code block, so a plain renderer keeps the indentation; a nested list would reflow and lose the assignee column.
2. The proposed mark is the word `proposed` after the assignee. A proposed task whose file differs from the trunk's also reads `(differs from trunk)`. The test data holds no such task, so the mark is untested against real data.
3. The tree shows no authorised mark; only the exception is marked.
4. Detail and provenance follow the tree as further sections with a table each, not as extra columns of the tree.
5. One heading per level section, and one `##` per person in the assignment view.
6. Numbers right-aligned; ids and models in code spans; absent values as an em dash.
7. A header line names the ref (a hash shortens to seven characters), the trunk status and the focusing parameters.
8. Whether the glance table of the assignment view belongs above per-person sections when a `person` focuses the view.
9. Whether "Next" is a column word or a mark, and whether all seven gates of every junction belong in detail or only the window.

## What the design gate must decide

- Whether the renderer takes the JSON of 493e as its data contract, or a typed structure from `tablo`.
- The Markdown layouts above, once the mockups exist.
- Whether the model check (the `Model:` trailer) shows in the assignment view.
