# Contributor work queue

**What do I do next?** Weather station · ref `main` · person ada@example.org · level detail

Legend: `tabloio gates`.

4 items: 0 reviews owed, 1 authorisation owed, 1 work ready, 1 reaffirmation, 1 work waiting.

| Kind | Task | Gate | Model | Since |
|---|---|---|---|---|
| Authorisation owed | `3c5d` Dashboard; proposed; the deciding commit is not by an authority | 📝 defined |  | 2026-09-22 |
| Work ready | `7b2e` Gateway | 📝 defined |  | 2026-09-15 |
| Reaffirmation | `9f31` Sensor board | ⚙️ function |  | 2026-09-28, 0 days |
| Work waiting | `9f31` Sensor board; blocked: Barometer ICs on 14-week backorder | ⚡ performance |  | 2026-09-28 |

## 1. Reviews owed: none

## 2. Authorisations owed: 1

**`3c5d` Dashboard** at 📝 defined

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- Contributor: 🧑 dan@example.org
- References: [Dashboard overview](docs/overview.md#dashboard)
- Cause: proposed; the deciding commit is not by an authority
- Requires: `7b2e` Gateway, function → integrate, Readings API: unmet, not yet due
- Status: 📝 defined 🟢 nominal, 2026-09-22, dan@example.org, `55f57c1`
- To authorise: `tabloio authorise 3c5d`

## 3. Work ready: 1

**`7b2e` Gateway** at 📝 defined

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- Contributor: 🧑 **ada@example.org**
- References: [Gateway overview](docs/overview.md#gateway)
- Also required by: `3c5d` Dashboard, function → integrate, Readings API: unmet, not yet due
- Status: ❔ undefined ⚪ undefined, 2026-09-15, `68d9021`
- Brief: `tabloio queue --person ada@example.org --ref main --brief 7b2e defined`

## 4. Reaffirmations: 1

**`9f31` Sensor board**, ⚙️ function, 2026-09-28, 0 days

- Gate: ⚙️ function, Functional prototype: A technical demonstration of function exists
- Contributor: 🧑 **ada@example.org**
- References: [Power budget](docs/sensor-board.md#power-budget), [docs/sensor-board.md](docs/sensor-board.md), [Barometer datasheet](https://example.org/datasheets/bmp390.pdf)
- Also required by: `c07d` Node firmware, design → implementation, Pin map and sensor bus: unmet
- Status: ⚙️ function 🔴 stalled ⛔ blocked, 2026-09-28, **ada@example.org**, `edb30d2`: Barometer ICs on 14-week backorder
- To reaffirm: `tabloio reaffirm 9f31`

## 5. Work waiting: 1

**`9f31` Sensor board** at ⚡ performance

- Gate: ⚡ performance, Performance prototype: Target performance shown in the deliverable's technology
- Contributor: 🧑 **ada@example.org**
- References: [Power budget](docs/sensor-board.md#power-budget), [docs/sensor-board.md](docs/sensor-board.md), [Barometer datasheet](https://example.org/datasheets/bmp390.pdf)
- Waits for: blocked: Barometer ICs on 14-week backorder
- Also required by: `c07d` Node firmware, design → implementation, Pin map and sensor bus: unmet
- Status: ⚙️ function 🔴 stalled ⛔ blocked, 2026-09-28, **ada@example.org**, `edb30d2`: Barometer ICs on 14-week backorder
- Do not start: the cause stands.

Command: `tabloio queue --person ada@example.org --ref main --level detail`
