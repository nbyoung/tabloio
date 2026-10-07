# Audit

**Where do files and history disagree?** Base · ref `main` at `7064289`, 2026-09-02 · stale 7 days · level provenance

Legend: `tabloio gates`.

**0 errors, 1 warning, 0 items of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H2 | warning | `b2c9` Leaf | 📐 design |  | pat is not the reviewer of the design junction; olive is | Have olive@example.org accept the gate with `Reviewed: b2c9 design` | olive@example.org |

**Most to resolve:** olive@example.org, 1 of 1.

## H2, warning: `b2c9` at 📐 design

**Rule.** "A `Reviewed:` commit from anyone other than the junction's reviewer has no effect, and the audit reports it." [README.md, Status](README.md#status)

**Junction.** pat@example.org contributes at 📐 design. olive@example.org reviews it.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `7064289` | 2026-09-02 | pat@example.org |  | Accept the design | `Reviewed: b2c9 design` |

```
# shows the finding
git log --format='%h %ae %(trailers:key=Reviewed)' -E --grep='^Reviewed: b2c9 '
```

Command: `tabloio audit --ref 7064289 --stale 7 --level provenance`
