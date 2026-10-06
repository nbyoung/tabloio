# Gate definition for `9f31` Sensor board

**What do the columns and symbols mean?** Weather station · ref `main` · task `9f31` · level detail

## Gates

| # | Symbol | Gate | Key | Criteria | For `9f31` |
|--:|:-:|---|---|---|---|
| 1 | ❔ | Undefined | `undefined` | No one has started work on the definition | Applies |
| 2 | 📝 | Defined | `defined` | Title, description, assignee and references exist | Applies |
| 3 | 📌 | Mockup | `mockup` | A non-technical mockup of the outcome exists | Applies |
| 4 | ⚙️ | Functional prototype | `function` | A technical demonstration of function exists | Applies |
| 5 | ⚡ | Performance prototype | `performance` | Target performance shown in the deliverable's technology | Applies; reference: Power budget, `docs/sensor-board.md#power-budget`, stated by `9f31` |
| 6 | ⚓ | Reliability prototype | `reliability` | Target reliability shown in the fundamental technology | Applies |
| 7 | 📐 | Design | `design` | A model and sufficient tests exist | Applies |
| 8 | 🛠️ | Implementation | `implementation` | Artifacts suffice for unit, integration and validation tests | Applies |
| 9 | 🧩 | Unit test | `unit` | All prescribed tests pass | Applies |
| 10 | 🖼️ | Integration | `integrate` | Assembled with neighbouring components | Applies |
| 11 | 🌍 | Validation | `validate` | Passes user and field tests | Applies |
| 12 | 🚀 | Release | `release` | All variants documented and approved | Applies |

## States

| Symbol | State | Severity | Synopsis |
|:-:|---|--:|---|
| ⚪ | `undefined` | 0 | The work has not yet been defined |
| 🟢 | `nominal` | 1 | The work is proceeding as expected |
| 🟡 | `at_risk` | 2 | The work is at risk of stalling |
| 🔴 | `stalled` | 3 | Practically all progress has stalled |
| ✅ | `complete` | 0 | All deliverables satisfy their requirements |

## Reasons

| Symbol | Reason | Synopsis |
|:-:|---|---|
| 🪫 | `overloaded` | The assigned resource is overloaded |
| ⛔ | `blocked` | An external resource is unavailable |
| 👓 | `review` | The work waits for its reviewer |

The method reserves `review`: the contributor has handed the next junction's work to its reviewer, and the work queue lists the review as owed. 👓 beside a state says the work waits for its reviewer; 👀 in a cell says a reviewer accepts the work there.

## Junction marks

| Mark | Meaning | Junction |
|:-:|---|---|
| 🧑 | A person contributes | Plain; the contributor is a person |
| 🤖 | An agent contributes | Plain; the junction states a `model` |
| 👀 | A reviewer accepts | Plain; the reviewer accepts the work at the gate |
| 🪆 | A subproject does the work | Recursive; another Tableaux project does the work |
| — | The gate does not apply | Not applicable; the entry exempts the task from the gate |

Command: `tabloio gates --ref main --task 9f31 --level detail`
