# Contextual tableau: nbyoung@nbyoung.com

**How does my corner stand?** Tableaux tooling · ref `corpus/tooling` · person nbyoung@nbyoung.com · window 1 · level detail

Legend: `tabloio gates`.

| Id | Task | ❔…📌 ×0 | 🔧 | 📐 | 🧱 | 📏 | 🔗 | 🌍…🚀 ×0 |
|---|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|
| `437e` | **Tableaux tooling** |  | 🟢 | 🤖👀 | 🤖 | 🤖 | 🤖 |  |
| `bc63` | &nbsp;&nbsp;**Method** |  |  | 🟢 | 🤖 | 🤖 | 🤖 |  |
| `c2ad` | &nbsp;&nbsp;&nbsp;&nbsp;Roles · sibling |  |  |  | 🟢 | — | — |  |
| `2034` | &nbsp;&nbsp;&nbsp;&nbsp;**Views** (23) |  |  |  | 🟢 | 🤖 | 🤖 |  |
| `e3cb` | &nbsp;&nbsp;&nbsp;&nbsp;Schema files · sibling |  |  |  |  | 🟢 | — |  |
| `fcec` | &nbsp;&nbsp;&nbsp;&nbsp;Conformance corpus · sibling |  |  |  | 🟢 | 🤖 | 🤖 |  |
| `ac33` | &nbsp;&nbsp;&nbsp;&nbsp;Agent identity |  |  | 🟢 | \[🤖👀\] | — | — |  |
| `9f3f` | &nbsp;&nbsp;&nbsp;&nbsp;Subproject linkage |  |  | 🟢 | \[🤖👀\] | — | — |  |
| `7861` | &nbsp;&nbsp;&nbsp;&nbsp;Productivity evidence · sibling |  |  |  | 🟢 | — | — |  |
| `7166` | &nbsp;&nbsp;&nbsp;&nbsp;Language clarifications |  |  | 🟢 | \[🤖👀\] | — | — |  |
| `77b2` | &nbsp;&nbsp;tablo: backend library and plumbing |  | 🟢 | 🪆 | 🪆 | 🪆 | 🪆 |  |
| `6103` | &nbsp;&nbsp;tabloio: command line, output and input |  |  | 🟢 | 🪆 | 🪆 | 🪆 |  |
| `c6e8` | &nbsp;&nbsp;tablotui: terminal user interface |  |  | 🟢 | 🪆 | 🪆 | 🪆 |  |
| `595e` | &nbsp;&nbsp;tableaud: local daemon and HTML |  |  | 🟢 | 🪆 | 🪆 | 🪆 |  |

❔ undefined · 📌 mockup · 🔧 function · 📐 design · 🧱 implementation · 📏 unit · 🔗 integrate · 🌍 validate · 🚀 release · 🟢 nominal · 🤖 an agent contributes · 👀 a reviewer accepts · 🪆 a subproject does the work · — the gate does not apply.

## Notes

| Rows | Date | Note | From |
|---|---|---|---|
| `437e` | 2026-09-30 | prototype/8ed1 derives the history and a stand-in audit between two refs from one git log pass and one cat-file batch, on the weather station | rolled up from `77b2` tablo: backend library and plumbing |
| `bc63` | 2026-09-30 | The model plan-and-record rule (F20) is on the trunk; F1, F2 and F10 remain for design | rolled up from `ac33` Agent identity |
| `c2ad` | 2026-09-30 | The Roles text is on the trunk and covers all seven roles; the validate gate awaits the owner |  |
| `2034` | 2026-10-05 | VIEWS.md defines the ten views on the trunk and agrees with the twenty approved mockups on the gate window; validation checks that tablo and the front ends follow the text | rolled up from `e9c6` Abstract views |
| `e3cb` | 2026-09-30 | schemas/check.sh (files agree with SYNTAX.md, version 0.2.1) and schemas/load.py (files load as JSON Schema 2020-12) pass; every .tableaux file validates |  |
| `fcec` | 2026-09-30 | The corpus builds and checks clean, with 85 entries on the trunk; all three corrections from the tablo prototypes are in; the unit gate runs the tablo conformance test against it |  |
| `ac33` | 2026-09-30 | The model plan-and-record rule (F20) is on the trunk; F1, F2 and F10 remain for design |  |
| `9f3f` | 2026-09-30 | The subproject linkage rule (F3, F5, F12) is on the trunk at language 0.3.0; tablo and the subprojects follow |  |
| `7861` | 2026-09-30 | The full report, its 23 tests and an example over five repositories are on the trunk |  |
| `7166` | 2026-10-07 | The trunk clarification (F19), with the order for a linked repository, and the exemption rule (F7) are on the trunk; F9 remains for design |  |
| `77b2` | 2026-09-30 | prototype/8ed1 derives the history and a stand-in audit between two refs from one git log pass and one cat-file batch, on the weather station | `b2c1` tablo in `subprojects/tablo` at `f7c3190`, at 🔧 function |
| `6103` | 2026-09-30 | design/e0f7.md is on the trunk with the owner's decisions of 2026-10-07, after two revisions; implementation follows, with one writer over doc.Doc, golang.org/x/text for the width table, one rule for the variation selector and no option for it, and --width as the one new option; no terminal draws the output before validate | `07e0` tabloio in `subprojects/tabloio` at `4bb4300`, at 📐 design |
| `c6e8` | 2026-09-30 | design/518e.md is on the trunk with the owner's decisions of 2026-10-07; implementation waits for tabloio's action package (b618) at implementation and a tabloio tag that go.mod can require | `40e8` tablotui in `subprojects/tablotui` at `ce24c31`, at 📐 design |
| `595e` | 2026-09-30 | design/db74.md is on the trunk with the owner's decisions of 2026-10-07, after two revisions; implementation follows, and it meets the accessibility task (5a2f) in base.html, shared/grid.html, the sheet and the script; the adapters' focus and the audit's control wait for tablo's views and for the temporal templates (9a9c) | `8608` tableaud in `subprojects/tableaud` at `0da180c`, at 📐 design |

## Folded columns

| Folded column | Gate | Tasks |
|:-:|---|--:|
| ❔…📌 | ❔ undefined | 0 |
| ❔…📌 | 📝 defined | 0 |
| ❔…📌 | 📌 mockup | 0 |
| 🌍…🚀 | 🌍 validate | 0 |
| 🌍…🚀 | 🚀 release | 0 |

Command: `tabloio context --person nbyoung@nbyoung.com --ref corpus/tooling --window 1 --level detail`
