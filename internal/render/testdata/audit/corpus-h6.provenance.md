# Audit

**Where do files and history disagree?** Base · ref `main` at `df6b60c`, 2026-09-04 · stale 7 days · level provenance

Legend: `tabloio gates`.

**0 errors, 1 warning, 0 items of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H6 | warning | `b2c9` Leaf | 📐 design | `status/b2c9.yaml` | The contributor's commit at the design junction carries no Model: trailer; the junction states claude-fable | Name the model in the next commit at the junction | bot@example.org |

**Most to resolve:** bot@example.org, 1 of 1.

## H6, warning: `b2c9` at 📐 design

**Rule.** "a commit at the junction by its contributor that carries no trailer, unless the project's `version.yaml` at that commit states a language before 0.2.1" [README.md, Junctions](README.md#junctions)

**Junction.** `b2c9` takes 📐 design from itself: contributor bot@example.org, model `claude-fable`.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `df6b60c` | 2026-09-04 | bot@example.org |  | Draft the design |  |

```
# shows the finding
git log --format='%h %(trailers:key=Model,valueonly)' -- .tableaux/status/b2c9.yaml
```

Command: `tabloio audit --ref df6b60c --stale 7 --level provenance`
