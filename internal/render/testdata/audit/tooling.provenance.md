# Audit

**Where do files and history disagree?** Tableaux tooling · ref `main` at `3cdae52`, 2026-10-05 · stale 7 days · level provenance

Legend: `tabloio gates`.

**1 error, 3 warnings, 28 items of information**: 32 findings in 8 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| S11 | error | `e3cb` Schema files | 🧱 implementation | `status/e3cb.yaml` | The status stands at 📏 unit and passes 🧱 implementation, whose reviewer is noreply@anthropic.com; no commit carries the trailer `Reviewed: e3cb implementation`. `fbe3c97` writes that line in its body, not among its trailers | Review: commit `Reviewed: e3cb implementation`, or return the status to 📐 design, the last reviewed gate | noreply@anthropic.com |
| H3 | warning | `437e` `bc63` | 📝 defined | `tasks/437e.yaml`, `tasks/bc63.yaml` | `320cf2b` edits both task files under `Model: claude-fable-5-1`; the junction states `claude-haiku` | Revise the file: state the model that ran at the junction | nbyoung@nbyoung.com |
| H3 | warning | `ac33` Agent identity | 📝 defined | `tasks/ac33.yaml` | `018814a` edits the task file under `Model: claude-fable-5-1`; the junction states `claude-haiku` | Revise the file: state the model that ran at the junction | nbyoung@nbyoung.com |
| H4 | information | `99f0` `05a9` `a9ce` `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` | 📌 mockup | `status/<id>.yaml` | The contributor's `5958858` is the newest event and the status states no 👓 review; nbyoung@nbyoung.com may have a mockup to review | Record the hand-off: the reason 👓 review when the work is done; or work on | noreply@anthropic.com |
| H4 | information | `ac33` `9f3f` `7166` | 🧱 implementation | `status/<id>.yaml` | The contributor's status commit is the newest event (`6762222`, `43d68e4`, `d9396b4`) and the status states no 👓 review; nbyoung@nbyoung.com, the assignee, reviews | Record the hand-off: the reason 👓 review when the work is done; or work on | noreply@anthropic.com |
| H4 | information | `e9c6` Abstract views | 🧱 implementation | `status/e9c6.yaml` | The contributor's `879b447` is the newest event and the status states no 👓 review; the contributor is its own reviewer here | Record the gate with `Reviewed: e9c6 implementation` on the same commit; or work on | noreply@anthropic.com |
| H4 | information | `fcec` Conformance corpus | 📏 unit | `status/fcec.yaml` | The contributor's `60495c2` is the newest event and the status states no 👓 review; the contributor is its own reviewer here | Record the gate with `Reviewed: fcec unit` on the same commit; or work on | noreply@anthropic.com |
| H4 | information | `c2ad` `e3cb` `7861` | 🌍 validate | `status/<id>.yaml` | The contributor's status commit is the newest event (`3d8ce0e`, `5cbec7c`, `c6f2655`) and the status states no 👓 review; nbyoung@nbyoung.com may have a validation to review | Record the hand-off: the reason 👓 review when the work is done; or work on | noreply@anthropic.com |

**Most to resolve:** noreply@anthropic.com, 29 of 32.

## S11, error: `e3cb` at 🧱 implementation

**Rule.** "A validator rejects a status whose gate passes a reviewed junction that has no such commit." [README.md, Status](README.md#status)

**Junction.** `e3cb` at 🧱 implementation takes its contributor noreply@anthropic.com and its model `claude-sonnet` from `437e`. It states no reviewer, so the assignee of `e3cb`, noreply@anthropic.com, reviews.

**Trailers.** Git reads a trailer only in the last paragraph of a message. In `fbe3c97` the last paragraph holds `Co-Authored-By:` alone, so the reviewer's line counts for nothing.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `0529170` | 2026-09-29 | Norman Young, nbyoung@nbyoung.com |  | Accept the schema files design | `Reviewed: e3cb design` |
| `fbe3c97` | 2026-09-29 | Claude Fable 5.1, noreply@anthropic.com | Norman Young, nbyoung@nbyoung.com | Advance Roles, schema files and evidence past their design reviews | `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>` |
| `5cbec7c` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com |  | Record e3cb at unit, nominal | `Reviewed: e3cb unit`, `Model: claude-sonnet-5-5` |

```
# shows the finding: three trailers, none for implementation
git log --format='%h %(trailers:key=Reviewed)' -E --grep='^Reviewed: e3cb '

# resolves it: noreply@anthropic.com, the reviewer, accepts the gate
git commit --allow-empty -m 'Accept the schema files implementation' \
    --trailer 'Reviewed: e3cb implementation' --trailer 'Model: <the model that runs>'
```

## H3, warning: `437e` `bc63` at 📝 defined

**Rule.** "the audit reports a commit at the junction whose trailer names a model outside the one stated" [README.md, Junctions](README.md#junctions)

**Junction.** Each of the two takes 📝 defined from `437e`: contributor noreply@anthropic.com, model `claude-haiku`. The model reads as a prefix; `claude-fable-5-1` lies outside it.

**Work.** A change to a task file is work at 📝 defined (PLAN.md, decision D8).

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `320cf2b` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com | Norman Young, nbyoung@nbyoung.com | Name a model family per gate (D7) | `Model: claude-fable-5-1` |

```
# shows the finding: the model on each commit that changes the task files
git log --format='%h %(trailers:key=Model,valueonly)' -- \
    .tableaux/tasks/437e.yaml .tableaux/tasks/bc63.yaml

# resolves it: nbyoung@nbyoung.com, an authority, states the model that ran
git commit .tableaux/tasks/437e.yaml -m 'State the model that defines the project'
```

## H3, warning: `ac33` at 📝 defined

**Rule.** "the audit reports a commit at the junction whose trailer names a model outside the one stated" [README.md, Junctions](README.md#junctions)

**Junction.** `ac33` takes 📝 defined from `437e`: contributor noreply@anthropic.com, model `claude-haiku`. The model reads as a prefix; `claude-fable-5-1` lies outside it.

**Work.** A change to a task file is work at 📝 defined (PLAN.md, decision D8).

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `018814a` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com | Norman Young, nbyoung@nbyoung.com | Record the model an agent runs (F20) | `Model: claude-fable-5-1` |

```
# shows the finding: the model on each commit that changes the task file
git log --format='%h %(trailers:key=Model,valueonly)' -- .tableaux/tasks/ac33.yaml

# resolves it: nbyoung@nbyoung.com, an authority, states the model that ran
git commit .tableaux/tasks/ac33.yaml -m 'State the model that defines the agent identity task'
```

## H4, information: `99f0` `05a9` `a9ce` `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` at 📌 mockup

**Rule.** "it reports, as information, a task whose next junction has a reviewer, whose newest event is the contributor's and whose status states no `review`, since the work may be done or in progress" [README.md, Status](README.md#status)

**Junction.** noreply@anthropic.com contributes at each. nbyoung@nbyoung.com reviews 📌 mockup, as `437e` states.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `5958858` | 2026-09-29 | Claude Fable 5.1, noreply@anthropic.com | Norman Young, nbyoung@nbyoung.com | Accept the defined gate of the views and mockups |  |

```
# shows the finding for one task: its newest event and who made it
git log -1 --format='%as %h %ae' -- .tableaux/tasks/99f0.yaml .tableaux/status/99f0.yaml
git log -1 --format='%as %h %ae' -E --grep='^(Authorised|Reaffirmed): 99f0$' --grep='^Reviewed: 99f0 '

# resolves it: noreply@anthropic.com, the contributor, hands the finished work off
git commit .tableaux/status/99f0.yaml -m 'Mark the task as awaiting review' \
    --trailer 'Model: <the model that runs>'
```

## H4, information: `ac33` `9f3f` `7166` at 🧱 implementation

**Rule.** "it reports, as information, a task whose next junction has a reviewer, whose newest event is the contributor's and whose status states no `review`, since the work may be done or in progress" [README.md, Status](README.md#status)

**Junction.** noreply@anthropic.com contributes at each. nbyoung@nbyoung.com reviews 🧱 implementation as the assignee of each.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `6762222` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com | Norman Young, nbyoung@nbyoung.com | Advance the agent identity task to its design gate |  |
| `43d68e4` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com |  | Advance the subproject linkage task to its design gate |  |
| `d9396b4` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com | Norman Young, nbyoung@nbyoung.com | Advance the language clarifications to their design gate |  |

```
# shows the finding for one task: its newest event and who made it
git log -1 --format='%as %h %ae' -- .tableaux/tasks/ac33.yaml .tableaux/status/ac33.yaml
git log -1 --format='%as %h %ae' -E --grep='^(Authorised|Reaffirmed): ac33$' --grep='^Reviewed: ac33 '

# resolves it: noreply@anthropic.com, the contributor, hands the finished work off
git commit .tableaux/status/ac33.yaml -m 'Mark the task as awaiting review' \
    --trailer 'Model: <the model that runs>'
```

## H4, information: `e9c6` at 🧱 implementation

**Rule.** "it reports, as information, a task whose next junction has a reviewer, whose newest event is the contributor's and whose status states no `review`, since the work may be done or in progress" [README.md, Status](README.md#status)

**Junction.** noreply@anthropic.com is the assignee of `e9c6` and so reviews its own work at 🧱 implementation.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `879b447` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com |  | Advance the abstract views task to its design gate |  |

```
# shows the finding for one task: its newest event and who made it
git log -1 --format='%as %h %ae' -- .tableaux/tasks/e9c6.yaml .tableaux/status/e9c6.yaml
git log -1 --format='%as %h %ae' -E --grep='^(Authorised|Reaffirmed): e9c6$' --grep='^Reviewed: e9c6 '

# resolves it: noreply@anthropic.com, the contributor, hands the finished work off
git commit .tableaux/status/e9c6.yaml -m 'Mark the task as awaiting review' \
    --trailer 'Model: <the model that runs>'
```

## H4, information: `fcec` at 📏 unit

**Rule.** "it reports, as information, a task whose next junction has a reviewer, whose newest event is the contributor's and whose status states no `review`, since the work may be done or in progress" [README.md, Status](README.md#status)

**Junction.** noreply@anthropic.com is the assignee of `fcec` and so reviews its own work at 📏 unit.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `60495c2` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com |  | Record all three corpus corrections at implementation |  |

```
# shows the finding for one task: its newest event and who made it
git log -1 --format='%as %h %ae' -- .tableaux/tasks/fcec.yaml .tableaux/status/fcec.yaml
git log -1 --format='%as %h %ae' -E --grep='^(Authorised|Reaffirmed): fcec$' --grep='^Reviewed: fcec '

# resolves it: noreply@anthropic.com, the contributor, hands the finished work off
git commit .tableaux/status/fcec.yaml -m 'Mark the task as awaiting review' \
    --trailer 'Model: <the model that runs>'
```

## H4, information: `c2ad` `e3cb` `7861` at 🌍 validate

**Rule.** "it reports, as information, a task whose next junction has a reviewer, whose newest event is the contributor's and whose status states no `review`, since the work may be done or in progress" [README.md, Status](README.md#status)

**Junction.** noreply@anthropic.com contributes at each. nbyoung@nbyoung.com reviews 🌍 validate, as `437e` states.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `3d8ce0e` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com |  | Record the Roles task at its implementation gate |  |
| `5cbec7c` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com |  | Record e3cb at unit, nominal |  |
| `c6f2655` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com |  | Complete the productivity evidence report |  |

```
# shows the finding for one task: its newest event and who made it
git log -1 --format='%as %h %ae' -- .tableaux/tasks/c2ad.yaml .tableaux/status/c2ad.yaml
git log -1 --format='%as %h %ae' -E --grep='^(Authorised|Reaffirmed): c2ad$' --grep='^Reviewed: c2ad '

# resolves it: noreply@anthropic.com, the contributor, hands the finished work off
git commit .tableaux/status/c2ad.yaml -m 'Mark the task as awaiting review' \
    --trailer 'Model: <the model that runs>'
```

Command: `tabloio audit --ref 3cdae52 --stale 7 --level provenance`
