# Audit

**Where do files and history disagree?** Base · ref `main` at `041ae3b`, 2026-09-02 · stale 7 days · level provenance

Legend: `tabloio gates`.

**0 errors, 0 warnings, 1 item of information**: 1 finding in 1 row.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| H4 | information | `b2c9` Leaf | 📐 design | `status/b2c9.yaml` | The newest event is the contributor's and the status states no review; olive may have a design to review | Record the hand-off: the reason review when the work is done; or work on | pat@example.org |

**Most to resolve:** pat@example.org, 1 of 1.

## H4, information: `b2c9` at 📐 design

**Rule.** "it reports, as information, a task whose next junction has a reviewer, whose newest event is the contributor's and whose status states no `review`, since the work may be done or in progress" [README.md, Status](README.md#status)

**Junction.** pat@example.org contributes at 📐 design. olive@example.org reviews it.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `041ae3b` | 2026-09-02 | pat@example.org |  | Draft the design |  |

```
# shows the finding: the newest event and who made it
git log -1 --format='%as %h %ae' -- .tableaux/tasks/b2c9.yaml .tableaux/status/b2c9.yaml
```

Command: `tabloio audit --ref 041ae3b --stale 7 --level provenance`
