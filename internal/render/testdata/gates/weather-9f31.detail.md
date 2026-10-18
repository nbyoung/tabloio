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

## Junction marks

| Mark | Meaning | Junction |
|:-:|---|---|
| 🧑 | A person contributes | plain |
| 🤖 | An agent contributes | plain |
| 👀 | A reviewer accepts | plain |
| 🪆 | A subproject does the work | recursive |
| — | The gate does not apply | not\_applicable |

Command: `tabloio gates --task 9f31 --ref main --level detail`
