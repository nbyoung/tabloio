# Audit

**Where do files and history disagree?** Tableaux tooling · ref `corpus/tooling` at `50a373b`, 2026-10-17 · stale 7 days · level provenance

Legend: `tabloio gates`.

**0 errors, 13 warnings, 23 items of information**: 36 findings in 7 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| R9 | warning | `6103` `c6e8` `595e` |  | `.tableaux/tasks/6103.yaml`, `.tableaux/tasks/c6e8.yaml`, `.tableaux/tasks/595e.yaml` | the requirement on 77b2 from implementation is due at implementation and unmet: 77b2 stands at function | Await `77b2` at 🧱 implementation | nbyoung@nbyoung.com |
| H3 | warning | `e9c6` `7166` | 🧱 implementation |  | Model claude-opus-5-5 is outside the stated model claude-sonnet |  | noreply@anthropic.com |
|  | warning | `ac33` `9f3f` | 📐 design | `.tableaux/status/ac33.yaml`, `.tableaux/status/9f3f.yaml` | The status dates from 2026-09-30, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: <id>` | noreply@anthropic.com |
|  | warning | `7166` Language clarifications | 📐 design | `.tableaux/status/7166.yaml` | The status dates from 2026-10-07, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: 7166` | noreply@anthropic.com |
|  | warning | `c2ad` `fcec` `7861` | 🧱 implementation | `.tableaux/status/c2ad.yaml`, `.tableaux/status/fcec.yaml`, `.tableaux/status/7861.yaml` | The status dates from 2026-09-30, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: <id>` | noreply@anthropic.com |
|  | warning | `e9c6` Abstract views | 🧱 implementation | `.tableaux/status/e9c6.yaml` | The status dates from 2026-10-05, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: e9c6` | noreply@anthropic.com |
|  | warning | `e3cb` Schema files | 📏 unit | `.tableaux/status/e3cb.yaml` | The status dates from 2026-09-30, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: e3cb` | noreply@anthropic.com |

**Most to resolve:** noreply@anthropic.com, 16 of 36.

## R9, warning: `6103` `c6e8` `595e`

**Rule.** "A validator warns of an unmet requirement" [README.md, Tasks](https://github.com/nbyoung/tableaux/blob/main/README.md#tasks)

**The requirement.** `6103`: `77b2` tablo: backend library and plumbing, implementation → implementation, The library and plumbing it calls: unmet.

**The requirement.** `c6e8` `595e`: `77b2` tablo: backend library and plumbing, implementation → implementation, The library it embeds: unmet.

```
# shows the finding
git log -1 --format='%as %H %ae' -- .tableaux/tasks/6103.yaml
```

## H3, warning: `e9c6` `7166` at 🧱 implementation

**Rule.** "the audit reports a commit at the junction whose trailer names a model outside the one stated" [README.md, Junctions](https://github.com/nbyoung/tableaux/blob/main/README.md#junctions)

**The junction.** `e9c6` at 🧱 implementation: contributor noreply@anthropic.com, model `claude-sonnet`, reviewer noreply@anthropic.com, by default, from `437e`.

**The junction.** `7166` at 🧱 implementation: contributor noreply@anthropic.com, model `claude-sonnet`, reviewer nbyoung@nbyoung.com, by default, from `437e`.

**The model check.** `e9c6`: `ccb3947` at 🧱 implementation, the junction states `claude-sonnet`, the trailer `Model: claude-opus-5-5`: Differs.

**The model check.** `7166`: `56103af` at 🧱 implementation, the junction states `claude-sonnet`, the trailer `Model: claude-opus-5-5`: Differs.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `ccb3947` | 2026-10-05 | Claude Opus 5.5, noreply@anthropic.com |  | Record the abstract views task at its implementation gate | `Reviewed: e9c6 implementation`, `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` |
| `56103af` | 2026-10-07 | Claude Opus 5.5, noreply@anthropic.com |  | Note that F7 is resolved in the language clarifications' status | `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' ccb3947e3deb2f476a8c9c886266b2b4da66c6f6
```

## A stale status, warning: `ac33` `9f3f` at 📐 design

**Rule.** "A review that finds no change reaffirms with an empty commit, so a status is never older than its last confirmation." [README.md, Status](https://github.com/nbyoung/tableaux/blob/main/README.md#status)

**The status.** `ac33`: 📐 design 🟢 nominal: The model plan-and-record rule (F20) is on the trunk; F1, F2 and F10 remain for design, recorded 2026-09-30 by noreply@anthropic.com.

**The status.** `9f3f`: 📐 design 🟢 nominal: The subproject linkage rule (F3, F5, F12) is on the trunk at language 0.3.0; tablo and the subprojects follow, recorded 2026-09-30 by noreply@anthropic.com.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `6762222` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com | Norman Young, nbyoung@nbyoung.com | Advance the agent identity task to its design gate | `Model: claude-fable-5-1`, `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>` |
| `43d68e4` | 2026-09-30 | Claude Fable 5.1, noreply@anthropic.com |  | Advance the subproject linkage task to its design gate | `Model: claude-fable-5-1`, `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>` |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' 6762222ddffef76008daad27ebe26069ac67cc5d

# resolves it
git commit --allow-empty --trailer 'Reaffirmed: ac33'
```

## A stale status, warning: `7166` at 📐 design

**Rule.** "A review that finds no change reaffirms with an empty commit, so a status is never older than its last confirmation." [README.md, Status](https://github.com/nbyoung/tableaux/blob/main/README.md#status)

**The status.** `7166`: 📐 design 🟢 nominal: The trunk clarification (F19), with the order for a linked repository, and the exemption rule (F7) are on the trunk; F9 remains for design, recorded 2026-10-07 by noreply@anthropic.com.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `56103af` | 2026-10-07 | Claude Opus 5.5, noreply@anthropic.com |  | Note that F7 is resolved in the language clarifications' status | `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' 56103af8cbaf561a90f9fa67fea02053372a4214

# resolves it
git commit --allow-empty --trailer 'Reaffirmed: 7166'
```

## A stale status, warning: `c2ad` `fcec` `7861` at 🧱 implementation

**Rule.** "A review that finds no change reaffirms with an empty commit, so a status is never older than its last confirmation." [README.md, Status](https://github.com/nbyoung/tableaux/blob/main/README.md#status)

**The status.** `c2ad`: 🧱 implementation 🟢 nominal: The Roles text is on the trunk and covers all seven roles; the validate gate awaits the owner, recorded 2026-09-30 by noreply@anthropic.com.

**The status.** `fcec`: 🧱 implementation 🟢 nominal: The corpus builds and checks clean, with 85 entries on the trunk; all three corrections from the tablo prototypes are in; the unit gate runs the tablo conformance test against it, recorded 2026-09-30 by noreply@anthropic.com.

**The status.** `7861`: 🧱 implementation 🟢 nominal: The full report, its 23 tests and an example over five repositories are on the trunk, recorded 2026-09-30 by noreply@anthropic.com.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `3d8ce0e` | 2026-09-30 | Claude Sonnet 5.5, noreply@anthropic.com |  | Record the Roles task at its implementation gate | `Reviewed: c2ad implementation`, `Model: claude-sonnet-5-5`, `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>` |
| `c6f2655` | 2026-09-30 | Claude Sonnet 5.5, noreply@anthropic.com |  | Complete the productivity evidence report | `Reviewed: 7861 implementation`, `Model: claude-sonnet-5-5`, `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>` |
| `60495c2` | 2026-09-30 | Claude Sonnet 5.5, noreply@anthropic.com |  | Record all three corpus corrections at implementation | `Reviewed: fcec implementation`, `Model: claude-sonnet-5-5`, `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>` |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' 3d8ce0eee1e869c75453abba91cae03dec6587ce

# resolves it
git commit --allow-empty --trailer 'Reaffirmed: c2ad'
```

## A stale status, warning: `e9c6` at 🧱 implementation

**Rule.** "A review that finds no change reaffirms with an empty commit, so a status is never older than its last confirmation." [README.md, Status](https://github.com/nbyoung/tableaux/blob/main/README.md#status)

**The status.** `e9c6`: 🧱 implementation 🟢 nominal: VIEWS.md defines the ten views on the trunk and agrees with the twenty approved mockups on the gate window; validation checks that tablo and the front ends follow the text, recorded 2026-10-05 by noreply@anthropic.com.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `ccb3947` | 2026-10-05 | Claude Opus 5.5, noreply@anthropic.com |  | Record the abstract views task at its implementation gate | `Reviewed: e9c6 implementation`, `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' ccb3947e3deb2f476a8c9c886266b2b4da66c6f6

# resolves it
git commit --allow-empty --trailer 'Reaffirmed: e9c6'
```

## A stale status, warning: `e3cb` at 📏 unit

**Rule.** "A review that finds no change reaffirms with an empty commit, so a status is never older than its last confirmation." [README.md, Status](https://github.com/nbyoung/tableaux/blob/main/README.md#status)

**The status.** `e3cb`: 📏 unit 🟢 nominal: schemas/check.sh (files agree with SYNTAX.md, version 0.2.1) and schemas/load.py (files load as JSON Schema 2020-12) pass; every .tableaux file validates, recorded 2026-09-30 by noreply@anthropic.com.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `5cbec7c` | 2026-09-30 | Claude Sonnet 5.5, noreply@anthropic.com |  | Record e3cb at unit, nominal | `Reviewed: e3cb unit`, `Model: claude-sonnet-5-5`, `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>` |

```
# shows the finding
git show --no-patch --format='%as %H %ae %ce%n%(trailers)' 5cbec7cce7c3cb3bf99f2145713d13ccb0b23d0f

# resolves it
git commit --allow-empty --trailer 'Reaffirmed: e3cb'
```

Command: `tabloio audit --ref 50a373b --stale 7 --level provenance`
