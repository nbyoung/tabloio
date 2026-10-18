# Audit

**Where do files and history disagree?** Base · ref `main` at `df6b60c`, 2026-09-04 · stale 7 days · level provenance

Legend: `tabloio gates`.

**0 errors, 1 warning, 2 items of information**: 3 findings in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H6 | warning | `b2c9` Leaf | 📐 design |  | the contributor's commit at the design junction carries no Model: trailer; the junction states claude-fable |  | bot@example.org |

**Most to resolve:** bot@example.org, 1 of 3.

## H6, warning: `b2c9` at 📐 design

**Rule.** "a commit at the junction by its contributor that carries no trailer, unless the project's `version.yaml` at that commit states a language before 0.2.1" [README.md, Junctions](https://github.com/nbyoung/tableaux/blob/main/README.md#junctions)

**The junction.** `b2c9` at 📐 design: contributor bot@example.org, model `claude-fable`, reviewer pat@example.org, the assignee, from `b2c9`.

**The model check.** `b2c9`: `df6b60c` at 📐 design, the junction states `claude-fable`: No trailer.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `df6b60c` | 2026-09-04 | Corpus Agent, bot@example.org |  | Draft the design |  |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' df6b60cbb2a98df9ff66fa8e4253022b06502979
```

Command: `tabloio audit --ref df6b60c --stale 7 --level provenance`
