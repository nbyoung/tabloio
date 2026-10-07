# Task definition: `c07d` Node firmware

**What is this task and where does it stand?** Weather station · ref `main` · task `c07d` · person ben@example.org · level detail

Legend: `tabloio gates`.

| Field | Value |
|---|---|
| Task | `c07d` Node firmware |
| Assignee | **ben@example.org** |
| Parent | `4e2b` Sensor node, order 2 |
| Status | 📐 design 🟢 nominal, 🪆 from the subproject `firmware` |
| Note | Sleep scheduler in progress |
| Recorded | 2026-09-17, from the subproject |

## Description

Reads the sensors, sleeps between readings and publishes to the gateway.

## Place in the tree

| Field | Value |
|---|---|
| Path | `a1c0` Weather station › `4e2b` Sensor node › `c07d` Node firmware |
| Order | 2 of 2 under `4e2b` |
| Children | None |
| Authorities | **ben@example.org** (`4e2b`); then ada@example.org (`a1c0`) |
| Authorised | Yes |

## Requires: 1

| Task | Edge | What passes | Its status | Condition |
|---|---|---|---|---|
| `9f31` Sensor board | design → implementation | Pin map and sensor bus | ⚙️ function 🔴 stalled ⛔ blocked | unmet |

## Dependents: none

## Junctions

| Gate | Marks | Contributor | Model | Reviewer | Subproject | Stands |
|---|:-:|---|---|---|---|---|
| ❔ undefined |  |  |  |  |  | passed |
| 📝 defined | 🧑 | **ben@example.org** |  |  |  | passed |
| 📌 mockup | 🧑 | **ben@example.org** |  | **ben@example.org** |  | passed, reviewed |
| ⚙️ function | 🧑 | **ben@example.org** |  |  |  | passed |
| ⚡ performance | 🧑 | **ben@example.org** |  |  |  | passed |
| ⚓ reliability | — |  |  |  |  | does not apply |
| 📐 design | 🧑 | **ben@example.org** |  |  |  | passed; the status stands here |
| 🛠️ implementation | 🪆 |  |  |  | firmware `f1a0` Weather node image | next |
| 🧩 unit | 🤖👀 | opus@example.org | `claude-opus-5-5` | the assignee, **ben@example.org** |  | later |
| 🖼️ integrate | 🧑 | **ben@example.org** |  |  |  | later |
| 🌍 validate | 🧑 | **ben@example.org** |  |  |  | later |
| 🚀 release | 🧑 | **ben@example.org** |  |  |  | later |

🧑 a person contributes · 🤖 an agent contributes · 👀 a reviewer accepts · 🪆 a subproject does the work · — the gate does not apply.

## Subproject snapshot

| Field | Value |
|---|---|
| Subproject | `firmware` |
| Task | `f1a0` Weather node image |
| Assignee | **ben@example.org** |
| Status | 📐 design 🟢 nominal |
| Note | Sleep scheduler in progress |
| Recorded | 2026-09-17 |
| Children | None |

Command: `tabloio task c07d --person ben@example.org --ref main --level detail`
