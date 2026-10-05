# Audit

**Where do files and history disagree?** Tableaux tooling · ref `main` · stale 7 days · level detail

Legend: `tabloio gates`.

**1 error, 3 warnings, 28 items of information**: 32 findings in 8 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| S11 | error | `e3cb` Schema files | 🧱 implementation | `status/e3cb.yaml` | The status stands at 📏 unit and passes 🧱 implementation, whose reviewer is noreply@anthropic.com; no commit carries the trailer `Reviewed: e3cb implementation`. `fbe3c97` writes that line in its body, not among its trailers | Review: commit `Reviewed: e3cb implementation`, or return the status to 📐 design, the last reviewed gate | noreply@anthropic.com |
| H3 | warning | `437e` `bc63` | 📝 defined | `tasks/437e.yaml`, `tasks/bc63.yaml` | `320cf2b` edits both task files under `Model: claude-fable-5-1`; the junction states `claude-haiku` | Revise the file: state the model that ran at the junction | nbyoung@nbyoung.com |
| H3 | warning | `ac33` Agent identity | 📝 defined | `tasks/ac33.yaml` | `018814a` edits the task file under `Model: claude-fable-5-1`; the junction states `claude-haiku` | Revise the file: state the model that ran at the junction | nbyoung@nbyoung.com |
| H4 | information | `99f0` `05a9` `a9ce` `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` | 📌 mockup | `status/<id>.yaml` | The contributor's `5958858` is the newest event and the status states no 👓 review; nbyoung@nbyoung.com may have a mockup to review | Record the hand-off: the reason 👓 review when the work is done; or work on | noreply@anthropic.com |
| H4 | information | `ac33` `9f3f` `7166` | 🧱 implementation | `status/<id>.yaml` | The contributor's status commit is the newest event (`6762222`, `43d68e4`, `d9396b4`) and the status states no 👓 review; nbyoung@nbyoung.com, the assignee, reviews | Record the hand-off: the reason 👓 review when the work is done; or work on | noreply@anthropic.com |
| H4 | information | `e9c6` Abstract views | 🧱 implementation | `status/e9c6.yaml` | The contributor's `879b447` is the newest event and the status states no 👓 review; the contributor is its own reviewer here | Record the gate with `Reviewed: e9c6 implementation` on the same commit; or work on | noreply@anthropic.com |
| H4 | information | `fcec` Conformance corpus | 📏 unit | `status/fcec.yaml` | The contributor's `60495c2` is the newest event and the status states no 👓 review; the contributor is its own reviewer here | Record the gate with `Reviewed: fcec unit` on the same commit; or work on | noreply@anthropic.com |
| H4 | information | `c2ad` `e3cb` `7861` | 🌍 validate | `status/<id>.yaml` | The contributor's status commit is the newest event (`3d8ce0e`, `5cbec7c`, `c6f2655`) and the status states no 👓 review; nbyoung@nbyoung.com may have a validation to review | Record the hand-off: the reason 👓 review when the work is done; or work on | noreply@anthropic.com |

**Most to resolve:** noreply@anthropic.com, 29 of 32.

Command: `tabloio audit --ref main --stale 7 --level detail`
