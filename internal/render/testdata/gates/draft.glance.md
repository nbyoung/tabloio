# Gate definition

**What do the columns and symbols mean?** Tableaux tooling · ref `main` · level glance

## Gates

| # | Symbol | Gate | Key |
|--:|:-:|---|---|
| 1 | ❔ | Undefined | `undefined` |
| 2 | 📝 | Defined | `defined` |
| 3 | 📌 | Mockup | `mockup` |
| 4 | 🔧 | Functional prototype | `function` |
| 5 | 📐 | Design | `design` |
| 6 | 🧱 | Implementation | `implementation` |
| 7 | 📏 | Unit test | `unit` |
| 8 | 🔗 | Integration | `integrate` |
| 9 | 🌍 | Validation | `validate` |
| 10 | 🚀 | Release | `release` |

## States

| Symbol | State |
|:-:|---|
| ⚪ | `undefined` |
| 🟢 | `nominal` |
| 🟡 | `at_risk` |
| 🔴 | `stalled` |
| ✅ | `complete` |

## Reasons

| Symbol | Reason |
|:-:|---|
| 🪫 | `overloaded` |
| ⛔ | `blocked` |
| 👓 | `review` |

## Junction marks

| Mark | Meaning |
|:-:|---|
| 🧑 | A person contributes |
| 🤖 | An agent contributes |
| 👀 | A reviewer accepts |
| 🪆 | A subproject does the work |
| — | The gate does not apply |

Command: `tabloio gates --ref main --level glance`
