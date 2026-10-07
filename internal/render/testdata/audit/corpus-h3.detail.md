# Audit

**Where do files and history disagree?** Base · ref `main` · stale 7 days · level detail

Legend: `tabloio gates`.

**0 errors, 1 warning, 0 items of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H3 | warning | `b2c9` Leaf | 📐 design | `status/b2c9.yaml` | Model claude-sonnet-5 is outside the stated model claude-fable | Revise the file: state the model that ran at the junction | olive@example.org |

**Most to resolve:** olive@example.org, 1 of 1.

Command: `tabloio audit --ref main --stale 7 --level detail`
