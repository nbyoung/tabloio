# Audit

**Where do files and history disagree?** Base · ref `main` · stale 7 days · level detail

Legend: `tabloio gates`.

**0 errors, 1 warning, 0 items of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H6 | warning | `b2c9` Leaf | 📐 design | `status/b2c9.yaml` | The contributor's commit at the design junction carries no Model: trailer; the junction states claude-fable | Name the model in the next commit at the junction | bot@example.org |

**Most to resolve:** bot@example.org, 1 of 1.

Command: `tabloio audit --ref main --stale 7 --level detail`
