# Gate definition

**What do the columns and symbols mean?** Tableaux tooling · ref `corpus/tooling` at `50a373b`, 2026-10-17 · level provenance

## Gates

| # | Symbol | Gate | Key | Criteria |
|--:|:-:|---|---|---|
| 1 | ❔ | Undefined | `undefined` | No one has started work on the definition |
| 2 | 📝 | Defined | `defined` | Title, description, assignee and references exist |
| 3 | 📌 | Mockup | `mockup` | A non-technical mockup of the outcome exists |
| 4 | 🔧 | Functional prototype | `function` | A technical demonstration of function exists |
| 5 | 📐 | Design | `design` | A model and sufficient tests exist |
| 6 | 🧱 | Implementation | `implementation` | Artifacts suffice for unit, integration and validation tests |
| 7 | 📏 | Unit test | `unit` | All prescribed tests pass |
| 8 | 🔗 | Integration | `integrate` | Assembled with neighbouring components |
| 9 | 🌍 | Validation | `validate` | Passes user and field tests |
| 10 | 🚀 | Release | `release` | All variants documented and approved |

Source: `gates` in `.tableaux/gates.yaml`.

## States

| Symbol | State | Severity | Synopsis |
|:-:|---|--:|---|
| ⚪ | `undefined` | 0 | The work has not yet been defined |
| 🟢 | `nominal` | 1 | The work is proceeding as expected |
| 🟡 | `at_risk` | 2 | The work is at risk of stalling |
| 🔴 | `stalled` | 3 | Practically all progress has stalled |
| ✅ | `complete` | 0 | All deliverables satisfy their requirements |

Source: `states` in `.tableaux/gates.yaml`.

## Reasons

| Symbol | Reason | Synopsis |
|:-:|---|---|
| 🪫 | `overloaded` | The assigned resource is overloaded |
| ⛔ | `blocked` | An external resource is unavailable |
| 👓 | `review` | The work waits for its reviewer |

The method reserves `review`: the contributor has handed the next junction's work to its reviewer, and the work queue lists the review as owed. 👓 beside a state says the work waits for its reviewer; 👀 in a cell says a reviewer accepts the work there.

Source: `reasons` in `.tableaux/gates.yaml`; the reserved meaning of `review` in `README.md`, Gates.

## Junction marks

| Mark | Meaning | Junction |
|:-:|---|---|
| 🧑 | A person contributes | plain |
| 🤖 | An agent contributes | plain |
| 👀 | A reviewer accepts | plain |
| 🪆 | A subproject does the work | recursive |
| — | The gate does not apply | not\_applicable |

Source: `README.md`, Junctions, at language 0.3.2.

## Provenance

| Fact | Value | Source |
|---|---|---|
| Language version | 0.3.2 | `tableaux` in `.tableaux/version.yaml` |
| Trunk | `main` | `trunk` in `.tableaux/version.yaml` |
| Ref | `corpus/tooling` at `50a373b`, 2026-10-17, on the trunk | The `--ref` parameter |

| File | Last changed by | Date | Author | Committer | Subject |
|---|---|---|---|---|---|
| `.tableaux/gates.yaml` | `e91ba68` | 2026-10-02 | Claude Opus 5.5, noreply@anthropic.com | Claude Opus 5.5, noreply@anthropic.com | Mark four gates with symbols that need no variation selector |
| `.tableaux/version.yaml` | `50a373b` | 2026-10-17 | Norman Young, nbyoung@nbyoung.com | Norman Young, nbyoung@nbyoung.com | Say that a requirement whose from gate does not apply has no condition |

```
git show 50a373b2857a36e4ef1a754605433943d6bfe324:.tableaux/gates.yaml
git show 50a373b2857a36e4ef1a754605433943d6bfe324:.tableaux/version.yaml
git log -1 --format='%H %as %an <%ae> / %cn <%ce> %s' 50a373b2857a36e4ef1a754605433943d6bfe324 -- .tableaux/gates.yaml
git log -1 --format='%H %as %an <%ae> / %cn <%ce> %s' 50a373b2857a36e4ef1a754605433943d6bfe324 -- .tableaux/version.yaml
```

Command: `tabloio gates --ref 50a373b --level provenance`
