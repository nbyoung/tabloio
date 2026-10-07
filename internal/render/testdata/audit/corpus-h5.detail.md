# Audit

**Where do files and history disagree?** Base · ref `main` · stale 7 days · level detail

Legend: `tabloio gates`.

**0 errors, 1 warning, 0 items of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H5 | warning | `b2c9` Leaf | 📐 design | `status/b2c9.yaml` | The status still states review after olive accepted design in S3 | Record the gate, or the next hand-off, so that the status no longer states review | pat@example.org |

**Most to resolve:** pat@example.org, 1 of 1.

Command: `tabloio audit --ref main --stale 7 --level detail`
