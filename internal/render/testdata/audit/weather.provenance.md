# Audit

**Where do files and history disagree?** Weather station · ref `main` at `edb30d2`, 2026-09-28 · stale 3 days · level provenance

Legend: `tabloio gates`.

**0 errors, 3 warnings, 32 items of information**: 35 findings in 3 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| R9 | warning | `c07d` Node firmware |  | `.tableaux/tasks/c07d.yaml` | the requirement on 9f31 from design is due at implementation and unmet: 9f31 stands at function | Await `9f31` at 📐 design | ada@example.org |
|  | warning | `3c5d` Dashboard |  | `.tableaux/tasks/3c5d.yaml` | The task stands proposed: dan@example.org, the author of its deciding commit, is none of its authorities | Authorise: commit `Authorised: 3c5d` | ada@example.org |
|  | warning | `3c5d` Dashboard | 📝 defined | `.tableaux/status/3c5d.yaml` | The status dates from 2026-09-22, more than 3 days before 2026-09-28 | Reaffirm: commit `Reaffirmed: 3c5d` | dan@example.org |

**Most to resolve:** ada@example.org, 2 of 35.

## R9, warning: `c07d`

**Rule.** "A validator warns of an unmet requirement" [README.md, Tasks](https://github.com/nbyoung/tableaux/blob/main/README.md#tasks)

**The requirement.** `c07d`: `9f31` Sensor board, design → implementation, Pin map and sensor bus: unmet.

```
# shows the finding
git log -1 --format='%as %H %ae' -- .tableaux/tasks/c07d.yaml
```

## A proposed task, warning: `3c5d`

**Rule.** "The task is authorised when the deciding commit's author or committer is one of its authorities, and proposed otherwise." [README.md, Proposed and authorised tasks](https://github.com/nbyoung/tableaux/blob/main/README.md#proposed-and-authorised-tasks)

**The authorisation.** `3c5d`: `1b8cfb1` Revise the dashboard scope, 2026-09-27, author and committer Dan Reyes, dan@example.org, not by an authority, accepting by a change on the trunk.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `1b8cfb1` | 2026-09-27 | Dan Reyes, dan@example.org |  | Revise the dashboard scope |  |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' 1b8cfb16a248f50eeefb76e0994234693ff4f4a6

# resolves it
git commit --allow-empty --trailer 'Authorised: 3c5d'
```

## A stale status, warning: `3c5d` at 📝 defined

**Rule.** "A review that finds no change reaffirms with an empty commit, so a status is never older than its last confirmation." [README.md, Status](https://github.com/nbyoung/tableaux/blob/main/README.md#status)

**The status.** `3c5d`: 📝 defined 🟢 nominal, recorded 2026-09-22 by dan@example.org.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `55f57c1` | 2026-09-22 | Dan Reyes, dan@example.org |  | Record the dashboard definition |  |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' 55f57c1ea69a0e16894408dfbd32a50d3dbdb0a9

# resolves it
git commit --allow-empty --trailer 'Reaffirmed: 3c5d'
```

Command: `tabloio audit --ref edb30d2 --stale 3 --level provenance`
