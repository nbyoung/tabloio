# Audit

**Where do files and history disagree?** Base · ref `main` · stale 7 days · level detail

Legend: `tabloio gates`.

**0 errors, 0 warnings, 1 item of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H4 | information | `b2c9` Leaf | 📐 design | `status/b2c9.yaml` | The newest event is the contributor's and the status states no review; olive may have a design to review | Record the hand-off: the reason review when the work is done; or work on | pat@example.org |

**Most to resolve:** pat@example.org, 1 of 1.

Command: `tabloio audit --ref main --stale 7 --level detail`
