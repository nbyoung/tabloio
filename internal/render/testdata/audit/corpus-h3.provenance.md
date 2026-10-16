# Audit

**Where do files and history disagree?** Base · ref `main` at `f20514e`, 2026-09-04 · stale 7 days · level provenance

Legend: `tabloio gates`.

**0 errors, 1 warning, 0 items of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H3 | warning | `b2c9` Leaf | 📐 design | `status/b2c9.yaml` | Model claude-sonnet-5 is outside the stated model claude-fable | Revise the file | olive@example.org |

**Most to resolve:** olive@example.org, 1 of 1.

## H3, warning: `b2c9` at 📐 design

**Rule.** "the audit reports a commit at the junction whose trailer names a model outside the one stated" [README.md, Junctions](README.md#junctions)

**The junction.** `b2c9` at 📐 design: contributor bot@example.org, model `claude-fable`, from `b2c9`.

**The model check.** `b2c9`: `b16817c` at 📐 design, the junction states `claude-fable`, the trailer `Model: claude-sonnet-5`: Differs.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `b16817c` | 2026-09-02 | bot@example.org |  | Record the leaf at design | `Model: claude-sonnet-5` |

```
# shows the finding
git log --format='%h %(trailers:key=Model,valueonly)' -- .tableaux/status/b2c9.yaml
```

Command: `tabloio audit --ref f20514e --stale 7 --level provenance`
