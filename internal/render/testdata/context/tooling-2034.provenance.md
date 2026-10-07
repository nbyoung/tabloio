# Contextual tableau: `2034` Views

**How does my corner stand?** Tableaux tooling · ref `main` at `3cdae52`, 2026-10-05 · task `2034` · window 1 · level provenance

Legend: `tabloio gates`.

| Id | Task | ❔ ×0 | 📝 | 📌 | 🔧 | 📐 | 🧱 | 📏 | 🔗…🚀 ×0 |
|---|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|
| `2034` | **Views** |  | 🟢 | 🤖👀 | 🤖 | 🤖👀 | 🤖 | 🤖 |  |
| `e9c6` | &nbsp;&nbsp;Abstract views |  |  |  |  | 🟢 | 🤖 | — |  |
| `bc86` | &nbsp;&nbsp;**Markdown views** (10) |  | 🟢 | 🤖👀 | — | — | — | — |  |
| `5fe3` | &nbsp;&nbsp;**HTML views** (10) |  | 🟢 | 🤖👀 | — | — | — | — |  |

❔ undefined · 📝 defined · 📌 mockup · 🔧 function · 📐 design · 🧱 implementation · 📏 unit · 🔗 integrate · 🚀 release · 🟢 nominal · 🤖 an agent contributes · 👀 a reviewer accepts · — the gate does not apply.

## Notes

| Rows | Date | Note | From |
|---|---|---|---|
| `2034` `bc86` | 2026-09-29 | Mockup waits for Abstract views (e9c6) at design | rolled up from `99f0` Gate definition view in Markdown |
| `e9c6` | 2026-09-30 | VIEWS.md is on the trunk at language 0.3.1; the twenty mockups may start, and implementation puts the text in place |  |
| `5fe3` | 2026-09-29 | Mockup waits for Abstract views (e9c6) at design | rolled up from `a6f7` Gate definition view in Html |

## Folded columns

| Folded column | Gate | Tasks |
|:-:|---|--:|
| ❔ | ❔ undefined | 0 |
| 🔗…🚀 | 🔗 integrate | 0 |
| 🔗…🚀 | 🌍 validate | 0 |
| 🔗…🚀 | 🚀 release | 0 |

## Status cells: 2

| Rows | Cell | Date | Recorder | Deciding commit | Committer | Trailers |
|---|---|---|---|---|---|---|
| `e9c6` | 📐 design | 2026-09-30 | noreply@anthropic.com | `879b447` Advance the abstract views task to its design gate | noreply@anthropic.com | `Model: claude-fable-5-1` |
| `99f0` | 📝 defined | 2026-09-29 | noreply@anthropic.com | `1a17bfc` Record the defined gate for the views and mockups | nbyoung@nbyoung.com |  |

## Roll-ups: 3

| Parent | Shows | Rolls up from | Date |
|---|---|---|---|
| `2034` Views | 📝 defined 🟢 nominal | `bc86` Markdown views | 2026-09-29 |
| `bc86` Markdown views | 📝 defined 🟢 nominal | `99f0` Gate definition view in Markdown | 2026-09-29 |
| `5fe3` HTML views | 📝 defined 🟢 nominal | `a6f7` Gate definition view in Html | 2026-09-29 |

## Marks: 3

| Gate | Mark | Rows | Contributor and model | Reviewer |
|---|:-:|---|---|---|
| 📌 mockup | 🤖👀 | `437e` `bc63` `2034` `bc86` `5fe3` `99f0` `05a9` `a9ce` `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` | noreply@anthropic.com, `claude-opus`, from `437e` | nbyoung@nbyoung.com, from `437e` |
| 📐 design | 🤖👀 | `437e` | noreply@anthropic.com, `claude-opus`, from `437e` | nbyoung@nbyoung.com, from `437e` |
| 📐 design | 🤖👀 | `bc63` `2034` | noreply@anthropic.com, `claude-fable`, from `bc63` | nbyoung@nbyoung.com, from `437e` |

🤖 an agent contributes · 👀 a reviewer accepts.

## Snapshots: none

```
git log -1 --format='%as %ae %h' -- .tableaux/status/e9c6.yaml # date, recorder and deciding commit
```

Command: `tabloio context --task 2034 --ref 3cdae52 --window 1 --level provenance`
