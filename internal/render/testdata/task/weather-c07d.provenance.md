# Task definition: `c07d` Node firmware

**What is this task and where does it stand?** Weather station · ref `main` at `edb30d2`, 2026-09-28 · task `c07d` · person ben@example.org · level provenance

Legend: `tabloio gates`.

| Field | Value |
|---|---|
| Task | `c07d` Node firmware |
| Assignee | **ben@example.org** |
| Parent | `4e2b` Sensor node, order 2 |
| Status | 📐 design 🟢 nominal, 🪆 from the subproject `firmware` |
| Note | Sleep scheduler in progress |
| Recorded | 2026-09-17, from the subproject |

## Status commit

| Field | Value |
|---|---|
| Commit | `97e4231` Pin the firmware at its design |
| Date | 2026-09-26 |
| Author | **ben@example.org** |
| Committer | **ben@example.org** |
| Trailers | `Reviewed: c07d mockup` |

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

## Authorisation commit

| Field | Value |
|---|---|
| Commit | `22a758b` Define the node firmware task |
| Date | 2026-09-16 |
| Author | **ben@example.org** |
| Committer | **ben@example.org** |
| Accepts by | a change on the trunk |
| The authority is | the author and the committer |

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

## Linkages

| Gate | Name | Value |
|---|---|---|
| 🛠️ implementation | url | `firmware`, stated in `.tableaux/tasks/c07d.yaml` |
|  | Form of the url | A path that is a Git submodule, so the submodule's pin fixes the commit |
|  | Pin | `cf5b169`, in full `cf5b16964dba7f8f1c8842a8202878fd572ad835` |
|  | id | `f1a0`, stated in `.tableaux/tasks/c07d.yaml` |

## Where each junction field comes from

| Gate | Field | Value | Supplied by |
|---|---|---|---|
| 📌 mockup | reviewer | **ben@example.org** | `4e2b` Sensor node |
| ⚓ reliability | applies | false | `c07d` Node firmware |
| 🛠️ implementation | subproject | firmware | `c07d` Node firmware |
| 🧩 unit | contributor | opus@example.org | `c07d` Node firmware |
|  | model | `claude-opus-5-5` | `c07d` Node firmware |
|  | reviewer | **ben@example.org** | the plain default |

## Reviews: 1

| Gate | Reviewer | Accepted by | Effect |
|---|---|---|---|
| 📌 mockup | **ben@example.org** | `97e4231` Pin the firmware at its design, 2026-09-26 | The reviewer accepts |

## Models: none

## Newest events

7 events; the newest 5, newest first.

| Date | Commit | By | Event | What it records |
|---|---|---|---|---|
| 2026-09-26 | `97e4231` | **ben@example.org** | pin | firmware → `cf5b169` |
| 2026-09-26 | `97e4231` | **ben@example.org** | reviewed | 📌 mockup |
| 2026-09-26 | `97e4231` | **ben@example.org** | status | 📐 design |
| 2026-09-17 | `cf5b169` | **ben@example.org** | status | 📐 design 🟢 nominal: Sleep scheduler in progress |
| 2026-09-16 | `22a758b` | **ben@example.org** | task |  |
| … and 2 earlier |  |  |  |  |

```
git log --format='%as %h %ae' main -- .tableaux/tasks/c07d.yaml .tableaux/status/c07d.yaml firmware
```

Command: `tabloio task c07d --ref edb30d2 --person ben@example.org --level provenance`
