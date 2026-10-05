# Task definition: `e9c6` Abstract views

**What is this task and where does it stand?** Tableaux tooling · ref `main` · task `e9c6` · person nbyoung@nbyoung.com · level detail

Legend: `tabloio gates`.

| Field | Value |
|---|---|
| Task | `e9c6` Abstract views |
| Assignee | noreply@anthropic.com |
| Parent | `2034` Views, order 1 |
| Status | 📐 design 🟢 nominal |
| Note | VIEWS.md is on the trunk at language 0.3.1; the twenty mockups may start, and implementation puts the text in place |
| Recorded | 2026-09-30 by noreply@anthropic.com |

## Description

VIEWS.md defines each view independently of any format: its question, the roles it serves, the data it draws from the project and the history, the parameters that focus it (a task, a person, a gate window or a list of columns, a ref) and its progressive-disclosure levels. The views are the gate definition, the task definition, authority delegation, task assignment, the contributor work queue, the work-blockage tree, the global tableau, the contextual tableau, the history and the audit.

References:

- [Proposed views](PLAN.md#views)

## Place in the tree

| Field | Value |
|---|---|
| Path | `437e` Tableaux tooling › `bc63` Method › `2034` Views › `e9c6` Abstract views |
| Order | 1 of 3 under `2034` |
| Children | None |
| Authorities | **nbyoung@nbyoung.com** (`2034`, `bc63`, `437e`) |
| Authorised | Yes |

## Requires: 1

| Task | Edge | What passes | Its status | Condition |
|---|---|---|---|---|
| `c2ad` Roles | design → design | The role names the views refer to | 🧱 implementation 🟢 nominal | met |

## Dependents: 21

| Task | Edge | What passes | Condition |
|---|---|---|---|
| `99f0` Gate definition view in Markdown | design → mockup | The abstract definition of the view | met |
| `05a9` Task definition view in Markdown | design → mockup | The abstract definition of the view | met |
| `a9ce` Authority delegation view in Markdown | design → mockup | The abstract definition of the view | met |
| … and 17 more: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` | design → mockup | The abstract definition of the view | met |
| `77b2` tablo: backend library and plumbing | implementation → design | The abstract views it derives | unmet, not yet due |

## Junctions

| Gate | Marks | Contributor | Model | Reviewer | Stands |
|---|:-:|---|---|---|---|
| ❔ undefined |  |  |  |  | passed |
| 📝 defined | 🤖 | noreply@anthropic.com | `claude-haiku` | the assignee, noreply@anthropic.com | passed |
| 📌 mockup | — |  |  |  | does not apply |
| 🔧 function | — |  |  |  | does not apply |
| 📐 design | 🤖👀 | noreply@anthropic.com | `claude-fable` | **nbyoung@nbyoung.com** | passed, reviewed; the status stands here |
| 🧱 implementation | 🤖 | noreply@anthropic.com | `claude-sonnet` | the assignee, noreply@anthropic.com | next |
| 📏 unit | — |  |  |  | does not apply |
| 🔗 integrate | — |  |  |  | does not apply |
| 🌍 validate | 🤖👀 | noreply@anthropic.com | `claude-opus` | **nbyoung@nbyoung.com** | later |
| 🚀 release | 🧑 | **nbyoung@nbyoung.com** |  |  | later |

🧑 a person contributes · 🤖 an agent contributes · 👀 a reviewer accepts · — the gate does not apply.

Command: `tabloio task e9c6 --ref main --for nbyoung@nbyoung.com --level detail`
