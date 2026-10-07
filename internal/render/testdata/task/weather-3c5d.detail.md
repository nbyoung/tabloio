# Task definition: `3c5d` Dashboard

**What is this task and where does it stand?** Weather station · ref `main` · task `3c5d` · level detail

Legend: `tabloio gates`.

| Field | Value |
|---|---|
| Task | `3c5d` Dashboard |
| Assignee | dan@example.org |
| Parent | `a1c0` Weather station |
| Status | 📝 defined 🟢 nominal |
| Note |  |
| Recorded | 2026-09-22 by dan@example.org |

## Description

A web page on the gateway that charts the last week of readings and the current values, with the battery state alongside.

References:

- [Dashboard overview](docs/overview.md#dashboard)

## Place in the tree

| Field | Value |
|---|---|
| Path | `a1c0` Weather station › `3c5d` Dashboard |
| Order | no order, among 3 under `a1c0` |
| Children | None |
| Authorities | ada@example.org (`a1c0`) |
| Authorised | No, proposed |

## Requires: 1

| Task | Edge | What passes | Its status | Condition |
|---|---|---|---|---|
| `7b2e` Gateway | function → integrate | Readings API | ❔ undefined ⚪ undefined | unmet, not yet due |

## Dependents: none

## Junctions

| Gate | Marks | Contributor | Model | Reviewer | Stands |
|---|:-:|---|---|---|---|
| ❔ undefined |  |  |  |  | passed |
| 📝 defined | 🧑 | dan@example.org |  |  | passed; the status stands here |
| 📌 mockup | 🧑 | dan@example.org |  |  | next |
| ⚙️ function | 🧑 | dan@example.org |  |  | later |
| ⚡ performance | 🧑 | dan@example.org |  |  | later |
| ⚓ reliability | 🧑 | dan@example.org |  |  | later |
| 📐 design | 🧑 | dan@example.org |  |  | later |
| 🛠️ implementation | 🧑 | dan@example.org |  |  | later |
| 🧩 unit | 🧑 | dan@example.org |  |  | later |
| 🖼️ integrate | 🧑 | dan@example.org |  |  | later |
| 🌍 validate | 🧑 | dan@example.org |  |  | later |
| 🚀 release | 🧑 | dan@example.org |  |  | later |

🧑 a person contributes.

Command: `tabloio task 3c5d --ref main --level detail`
