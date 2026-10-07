# Authority delegation

**Who may accept what?** Tableaux tooling · ref `main` at `3cdae52`, 2026-10-05 · level provenance

Legend: `tabloio gates`.

```
Task                                                  Assignee
437e Tableaux tooling                                 nbyoung@nbyoung.com
  bc63 Method                                           nbyoung@nbyoung.com
    c2ad Roles                                            noreply@anthropic.com
    2034 Views                                            nbyoung@nbyoung.com
      e9c6 Abstract views                                   noreply@anthropic.com
      bc86 Markdown views                                   noreply@anthropic.com
        99f0 Gate definition view in Markdown                 noreply@anthropic.com
        05a9 Task definition view in Markdown                 noreply@anthropic.com
        a9ce Authority delegation view in Markdown            noreply@anthropic.com
        bb7c Task assignment view in Markdown                 noreply@anthropic.com
        c74a Contributor work queue view in Markdown          noreply@anthropic.com
        efff Work-blockage tree view in Markdown              noreply@anthropic.com
        ab8e Global tableau view in Markdown                  noreply@anthropic.com
        c545 Contextual tableau view in Markdown              noreply@anthropic.com
        d615 History view in Markdown                         noreply@anthropic.com
        b14e Audit view in Markdown                           noreply@anthropic.com
      5fe3 HTML views                                       noreply@anthropic.com
        a6f7 Gate definition view in Html                     noreply@anthropic.com
        7dff Task definition view in Html                     noreply@anthropic.com
        b1b7 Authority delegation view in Html                noreply@anthropic.com
        ea51 Task assignment view in Html                     noreply@anthropic.com
        7783 Contributor work queue view in Html              noreply@anthropic.com
        5471 Work-blockage tree view in Html                  noreply@anthropic.com
        32e7 Global tableau view in Html                      noreply@anthropic.com
        69eb Contextual tableau view in Html                  noreply@anthropic.com
        3194 History view in Html                             noreply@anthropic.com
        a8b4 Audit view in Html                               noreply@anthropic.com
    e3cb Schema files                                     noreply@anthropic.com
    fcec Conformance corpus                               noreply@anthropic.com
    ac33 Agent identity                                   nbyoung@nbyoung.com
    9f3f Subproject linkage                               nbyoung@nbyoung.com
    7861 Productivity evidence                            noreply@anthropic.com
    7166 Language clarifications                          nbyoung@nbyoung.com
  77b2 tablo: backend library and plumbing              nbyoung@nbyoung.com
  6103 tabloio: command line, output and input          nbyoung@nbyoung.com
  c6e8 tablotui: terminal user interface                nbyoung@nbyoung.com
  595e tableaud: local daemon and HTML                  nbyoung@nbyoung.com
```

37 tasks, 7 delegations, 37 authorised, 0 proposed.

## Chains and defaults: 5

### `437e` Tableaux tooling · nbyoung@nbyoung.com

Authorities of its 5 children, nearest first: nbyoung@nbyoung.com (`437e`).

Children: `bc63` `77b2` `6103` `c6e8` `595e`.

| Gate | Marks | Contributor | Model | Reviewer |
|---|:-:|---|---|---|
| 📝 defined | 🤖 | noreply@anthropic.com | `claude-haiku` |  |
| 📌 mockup | 🤖👀 | noreply@anthropic.com | `claude-opus` | nbyoung@nbyoung.com |
| 🔧 function | 🤖 | noreply@anthropic.com | `claude-sonnet` |  |
| 📐 design | 🤖👀 | noreply@anthropic.com | `claude-opus` | nbyoung@nbyoung.com |
| 🧱 implementation | 🤖 | noreply@anthropic.com | `claude-sonnet` |  |
| 📏 unit | 🤖 | noreply@anthropic.com | `claude-sonnet` |  |
| 🔗 integrate | 🤖 | noreply@anthropic.com | `claude-sonnet` |  |
| 🌍 validate | 🤖👀 | noreply@anthropic.com | `claude-sonnet` | nbyoung@nbyoung.com |
| 🚀 release | 🧑 | nbyoung@nbyoung.com |  |  |

🧑 a person contributes · 🤖 an agent contributes · 👀 a reviewer accepts.

### `bc63` Method · nbyoung@nbyoung.com

Authorities of its 8 children, nearest first: nbyoung@nbyoung.com (`bc63`, `437e`).

Children: `c2ad` `2034` `e3cb` `fcec` `ac33` `9f3f` `7861` `7166`.

| Gate | Marks | Contributor | Model | Reviewer |
|---|:-:|---|---|---|
| 📐 design | 🤖👀 | noreply@anthropic.com | `claude-fable` | nbyoung@nbyoung.com, from `437e` |
| 🌍 validate | 🤖👀 | noreply@anthropic.com | `claude-opus` | nbyoung@nbyoung.com, from `437e` |

🤖 an agent contributes · 👀 a reviewer accepts.

### `2034` Views · nbyoung@nbyoung.com

Authorities of its 3 children, nearest first: nbyoung@nbyoung.com (`2034`, `bc63`, `437e`).

Children: `e9c6` `bc86` `5fe3`.

`2034` states no junction defaults.

### `bc86` Markdown views · noreply@anthropic.com

Authorities of its 10 children, nearest first: noreply@anthropic.com (`bc86`); then nbyoung@nbyoung.com (`2034`, `bc63`, `437e`).

Children: `99f0` `05a9` `a9ce` `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e`.

| Gate | Marks | Contributor | Model | Reviewer |
|---|:-:|---|---|---|
| 🔧 function | — |  |  |  |
| 📐 design | — |  |  |  |
| 🧱 implementation | — |  |  |  |
| 📏 unit | — |  |  |  |
| 🔗 integrate | — |  |  |  |
| 🌍 validate | — |  |  |  |
| 🚀 release | — |  |  |  |

— the gate does not apply.

### `5fe3` HTML views · noreply@anthropic.com

Authorities of its 10 children, nearest first: noreply@anthropic.com (`5fe3`); then nbyoung@nbyoung.com (`2034`, `bc63`, `437e`).

Children: `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4`.

| Gate | Marks | Contributor | Model | Reviewer |
|---|:-:|---|---|---|
| 🔧 function | — |  |  |  |
| 📐 design | — |  |  |  |
| 🧱 implementation | — |  |  |  |
| 📏 unit | — |  |  |  |
| 🔗 integrate | — |  |  |  |
| 🌍 validate | — |  |  |  |
| 🚀 release | — |  |  |  |

— the gate does not apply.

## Deciding commits: 4

| Commit | Date | Author | Committer | Accepts by | The authority is | Tasks |
|---|---|---|---|---|---|---|
| `8eb0cae` | 2026-09-30 | nbyoung@nbyoung.com | nbyoung@nbyoung.com | a merge of `320cf2b` | the author and the committer | 2: `437e` `bc63` |
| `16fe62f` | 2026-09-30 | nbyoung@nbyoung.com | nbyoung@nbyoung.com | a merge of `018814a` | the author and the committer | 1: `ac33` |
| `8d8b070` | 2026-09-30 | nbyoung@nbyoung.com | nbyoung@nbyoung.com | a merge of `9d9fb88` | the author and the committer | 1: `7166` |
| `6b6c99a` | 2026-09-29 | nbyoung@nbyoung.com | nbyoung@nbyoung.com | a change on the trunk | the author and the committer | 33: `c2ad` `2034` `e9c6` `bc86` `99f0` `05a9` `a9ce` `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `5fe3` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` `e3cb` `fcec` `9f3f` `7861` `77b2` `6103` `c6e8` `595e` |

```
git log --first-parent -1 --format='%h %as %ae %ce %p' main -- .tableaux/tasks/437e.yaml
git log --first-parent -1 --format='%h %as %ae %ce %p' -E --grep='^Authorised: 437e$' main
```

Command: `tabloio authority --ref 3cdae52 --level provenance`
