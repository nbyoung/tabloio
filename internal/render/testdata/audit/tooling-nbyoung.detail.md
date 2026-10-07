# Audit

**Where do files and history disagree?** Tableaux tooling · ref `main` · person nbyoung@nbyoung.com · stale 7 days · level detail

Legend: `tabloio gates`.

**0 errors, 3 warnings, 0 items of information**: 3 findings in 2 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H3 | warning | `437e` `bc63` | 📝 defined | `tasks/437e.yaml`, `tasks/bc63.yaml` | `320cf2b` edits both task files under `Model: claude-fable-5-1`; the junction states `claude-haiku` | Revise the file: state the model that ran at the junction | **nbyoung@nbyoung.com** |
| H3 | warning | `ac33` Agent identity | 📝 defined | `tasks/ac33.yaml` | `018814a` edits the task file under `Model: claude-fable-5-1`; the junction states `claude-haiku` | Revise the file: state the model that ran at the junction | **nbyoung@nbyoung.com** |

**Most to resolve:** **nbyoung@nbyoung.com**, 3 of 3.

Command: `tabloio audit --person nbyoung@nbyoung.com --ref main --stale 7 --level detail`
