# Task definition: `9f31` Sensor board

**What is this task and where does it stand?** Weather station · ref `main` · task `9f31` · level detail

Legend: `tabloio gates`.

| Field | Value |
|---|---|
| Task | `9f31` Sensor board |
| Assignee | ada@example.org |
| Parent | `4e2b` Sensor node, order 1 |
| Status | ⚙️ function 🔴 stalled ⛔ blocked |
| Note | Barometer ICs on 14-week backorder |
| Recorded | 2026-09-28 by ada@example.org |

## Description

The PCB that carries the barometer, hygrometer and thermometer and powers them from the solar cell within the power budget.

References:

- [docs/sensor-board.md](docs/sensor-board.md)
- [Barometer datasheet](https://example.org/datasheets/bmp390.pdf)

## Place in the tree

| Field | Value |
|---|---|
| Path | `a1c0` Weather station › `4e2b` Sensor node › `9f31` Sensor board |
| Order | 1 of 2 under `4e2b` |
| Children | None |
| Authorities | ben@example.org (`4e2b`); then ada@example.org (`a1c0`) |
| Authorised | Yes |

## Requires: none

## Dependents: 1

| Task | Edge | What passes | Condition |
|---|---|---|---|
| `c07d` Node firmware | design → implementation | Pin map and sensor bus | unmet |

## Junctions

| Gate | Marks | Contributor | Model | Reviewer | Stands |
|---|:-:|---|---|---|---|
| ❔ undefined |  |  |  |  | passed |
| 📝 defined | 🧑 | ada@example.org |  |  | passed |
| 📌 mockup | 🧑👀 | ada@example.org |  | ben@example.org | passed, reviewed |
| ⚙️ function | 🧑 | ada@example.org |  |  | passed; the status stands here |
| ⚡ performance | 🧑 | ada@example.org |  |  | next |
| ⚓ reliability | 🧑 | ada@example.org |  |  | later |
| 📐 design | 🧑 | ada@example.org |  |  | later |
| 🛠️ implementation | 🧑 | ada@example.org |  |  | later |
| 🧩 unit | 🧑 | ada@example.org |  |  | later |
| 🖼️ integrate | 🧑 | ada@example.org |  |  | later |
| 🌍 validate | 🧑👀 | ada@example.org |  | ben@example.org | later |
| 🚀 release | 🧑 | ada@example.org |  |  | later |

🧑 a person contributes · 👀 a reviewer accepts.

Command: `tabloio task 9f31 --ref main --level detail`
