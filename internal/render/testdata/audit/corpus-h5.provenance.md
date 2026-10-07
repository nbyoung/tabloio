# Audit

**Where do files and history disagree?** Base · ref `main` at `3844b60`, 2026-09-03 · stale 7 days · level provenance

Legend: `tabloio gates`.

**0 errors, 1 warning, 0 items of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H5 | warning | `b2c9` Leaf | 📐 design | `status/b2c9.yaml` | The status still states review after olive accepted design in S3 | Record the gate, or the next hand-off, so that the status no longer states review | pat@example.org |

**Most to resolve:** pat@example.org, 1 of 1.

## H5, warning: `b2c9` at 📐 design

**Rule.** "it reports a status that still states `review` after the reviewer's `Reviewed:` commit for that junction" [README.md, Status](README.md#status)

**Junction.** pat@example.org contributes at 📐 design. olive@example.org reviews it.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `1ff946d` | 2026-09-02 | pat@example.org |  | Hand the design to its reviewer |  |
| `3844b60` | 2026-09-03 | olive@example.org |  | Accept the design | `Reviewed: b2c9 design` |

```
# shows the finding
git log --format='%h %(trailers:key=Reviewed)' -E --grep='^Reviewed: b2c9 '
```

Command: `tabloio audit --ref 3844b60 --stale 7 --level provenance`
