# Audit

**Where do files and history disagree?** Base · ref `main` at `f20514e`, 2026-09-04 · stale 7 days · level provenance

Legend: `tabloio gates`.

**0 errors, 1 warning, 1 item of information**: 2 findings in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H3 | warning | `b2c9` Leaf | 📐 design |  | Model claude-sonnet-5 is outside the stated model claude-fable |  | bot@example.org |

## H3, warning: `b2c9` at 📐 design

**Rule.** "the audit reports a commit at the junction whose trailer names a model outside the one stated" [README.md, Junctions](https://github.com/nbyoung/tableaux/blob/main/README.md#junctions)

**The junction.** `b2c9` at 📐 design: contributor bot@example.org, model `claude-fable`, reviewer pat@example.org, the assignee, from `b2c9`.

**The model check.** `b2c9`: `b16817c` at 📐 design, the junction states `claude-fable`, the trailer `Model: claude-sonnet-5`: Differs.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `b16817c` | 2026-09-02 | Corpus Agent, bot@example.org |  | Record the leaf at design | `Model: claude-sonnet-5` |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' b16817cda3528588ee33642676fbb623a598baed
```

Command: `tabloio audit --ref f20514e --stale 7 --level provenance`
