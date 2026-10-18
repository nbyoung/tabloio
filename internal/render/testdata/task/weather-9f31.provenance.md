# Task definition: `9f31` Sensor board

**What is this task and where does it stand?** Weather station · ref `main` at `edb30d2`, 2026-09-28 · task `9f31` · level provenance

Legend: `tabloio gates`.

| Field | Value |
|---|---|
| Task | `9f31` Sensor board |
| Assignee | ada@example.org |
| Parent | `4e2b` Sensor node, order 1 |
| Status | ⚙️ function 🔴 stalled ⛔ blocked |
| Note | Barometer ICs on 14-week backorder |
| Recorded | 2026-09-28 by ada@example.org |

## Status commit

| Field | Value |
|---|---|
| Commit | `edb30d2` Weekly review: no change |
| Date | 2026-09-28 |
| Author | Ada Byron, ada@example.org |
| Committer | Ada Byron, ada@example.org |
| Trailers | `Reaffirmed: 9f31` |

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

## Authorisation commit

| Field | Value |
|---|---|
| Commit | `f9e8746` Merge the sensor board task |
| Date | 2026-09-19 |
| Author | Ben Okafor, ben@example.org |
| Committer | Ben Okafor, ben@example.org |
| Accepts by | a merge |
| The authority is | the author and the committer |

## Requires: none

## Dependents: 1

| Task | Edge | What passes | Condition |
|---|---|---|---|
| `c07d` Node firmware | design → implementation | Pin map and sensor bus | unmet |

## Junctions

| Gate | Marks | Contributor | Model | Reviewer | Stands |
|---|:-:|---|---|---|---|
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

## Where each junction field comes from

| Gate | Field | Value | Supplied by |
|---|---|---|---|
| 📌 mockup | reviewer | ben@example.org | `4e2b` Sensor node |
| ⚡ performance | references | 1 | `9f31` Sensor board |
| 🌍 validate | reviewer | ben@example.org | `9f31` Sensor board |

## Reviews: 1

| Gate | Reviewer | Accepted by | Effect |
|---|---|---|---|
| 📌 mockup | ben@example.org | `85adb9b` Accept the sensor board mockup, 2026-09-24 | The reviewer accepts |

## Models: none

## Newest events

6 events; the newest 5, newest first.

| Date | Commit | By | Event | What it records |
|---|---|---|---|---|
| 2026-09-28 | `edb30d2` | ada@example.org | reaffirmed |  |
| 2026-09-25 | `a2393fb` | ada@example.org | status | ⚙️ function 🔴 stalled ⛔ blocked: Barometer ICs on 14-week backorder |
| 2026-09-24 | `85adb9b` | ben@example.org | reviewed | 📌 mockup |
| 2026-09-22 | `63a1f88` | ada@example.org | status | 📌 mockup 🟢 nominal |
| 2026-09-19 | `f9e8746` | ben@example.org | authorised |  |
| … and 1 earlier |  |  |  |  |

```
git log --format='%H %as %ae %s' edb30d2baa5ea3c7e3fb8eef8ac6cce6e145221a -- .tableaux/tasks/9f31.yaml .tableaux/status/9f31.yaml
```

Command: `tabloio task 9f31 --ref edb30d2 --level provenance`
