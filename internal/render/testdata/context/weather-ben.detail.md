# Contextual tableau: ben@example.org

**How does my corner stand?** Weather station · ref `main` · person ben@example.org · window 1 · level detail

Legend: `tabloio gates`.

| Id | Task | ❔ ×0 | 📝 | 📌 | ⚙️ | ⚡ | ⚓ | 📐 | 🛠️ | 🧩 | 🖼️…🚀 ×0 |
|---|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|
| `a1c0` | **Weather station** · spine |  | 🟢 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  |
| `4e2b` | &nbsp;&nbsp;**Sensor node** · spine |  | 🧑 | 🧑 | 🔴⛔ | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  |
| `9f31` | &nbsp;&nbsp;&nbsp;&nbsp;Sensor board · sibling |  | 🧑 | 🧑👀 | 🔴⛔ | 🧑 | 🧑 | 🧑 | 🧑 | 🧑 |  |
| `c07d` | &nbsp;&nbsp;&nbsp;&nbsp;Node firmware · corner |  | \[🧑\] | \[🧑\] | \[🧑\] | \[🧑\] | — | 🟢 | 🪆 | 🤖👀 |  |

❔ undefined · 📝 defined · 📌 mockup · ⚙️ function · ⚡ performance · ⚓ reliability · 📐 design · 🛠️ implementation · 🧩 unit · 🖼️ integrate · 🚀 release · 🟢 nominal · 🔴 stalled · ⛔ blocked · 🧑 a person contributes · 🤖 an agent contributes · 👀 a reviewer accepts · 🪆 a subproject does the work · — the gate does not apply.

## Notes

| Rows | Date | Note | From |
|---|---|---|---|
| `a1c0` | 2026-09-17 |  | rolled up from `3c5d` Dashboard |
| `4e2b` | 2026-09-17 | Barometer ICs on 14-week backorder | rolled up from `9f31` Sensor board |
| `9f31` | 2026-09-28 | Barometer ICs on 14-week backorder |  |
| `c07d` | 2026-09-17 | Sleep scheduler in progress | `f1a0` Weather node image in `firmware` at `F2`, at 📐 design |

## Folded columns

| Folded column | Gate | Tasks |
|:-:|---|--:|
| ❔ | ❔ undefined | 0 |
| 🖼️…🚀 | 🖼️ integrate | 0 |
| 🖼️…🚀 | 🌍 validate | 0 |
| 🖼️…🚀 | 🚀 release | 0 |

Command: `tabloio context --person ben@example.org --ref main --window 1 --level detail`
