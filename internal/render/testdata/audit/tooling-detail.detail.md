# Audit

**Where do files and history disagree?** Tableaux tooling · ref `corpus/tooling` · stale 7 days · level detail

Legend: `tabloio gates`.

**0 errors, 17 warnings, 23 items of information**: 40 findings in 11 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| R9 | warning | `6103` `c6e8` `595e` |  | `.tableaux/tasks/6103.yaml`, `.tableaux/tasks/c6e8.yaml`, `.tableaux/tasks/595e.yaml` | the requirement on 77b2 from implementation is due at implementation and unmet: 77b2 stands at function | Await the task the requirement names | nbyoung@nbyoung.com |
| J19 | warning | `77b2` tablo: backend library and plumbing |  | `.tableaux/tasks/77b2.yaml` | the project at subprojects/tablo states tableaux 0.3.1, before this project's 0.3.2 | Raise the subproject's language | nbyoung@nbyoung.com |
| J19 | warning | `6103` tabloio: command line, output and input |  | `.tableaux/tasks/6103.yaml` | the project at subprojects/tabloio states tableaux 0.3.1, before this project's 0.3.2 | Raise the subproject's language | nbyoung@nbyoung.com |
| J19 | warning | `c6e8` tablotui: terminal user interface |  | `.tableaux/tasks/c6e8.yaml` | the project at subprojects/tablotui states tableaux 0.3.1, before this project's 0.3.2 | Raise the subproject's language | nbyoung@nbyoung.com |
| J19 | warning | `595e` tableaud: local daemon and HTML |  | `.tableaux/tasks/595e.yaml` | the project at subprojects/tableaud states tableaux 0.3.1, before this project's 0.3.2 | Raise the subproject's language | nbyoung@nbyoung.com |
| H3 | warning | `e9c6` `7166` | 🧱 implementation |  | Model claude-opus-5-5 is outside the stated model claude-sonnet |  | noreply@anthropic.com |
|  | warning | `ac33` `9f3f` | 📐 design | `.tableaux/status/ac33.yaml`, `.tableaux/status/9f3f.yaml` | The status dates from 2026-09-30, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: <id>` | noreply@anthropic.com |
|  | warning | `7166` Language clarifications | 📐 design | `.tableaux/status/7166.yaml` | The status dates from 2026-10-07, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: 7166` | noreply@anthropic.com |
|  | warning | `c2ad` `fcec` `7861` | 🧱 implementation | `.tableaux/status/c2ad.yaml`, `.tableaux/status/fcec.yaml`, `.tableaux/status/7861.yaml` | The status dates from 2026-09-30, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: <id>` | noreply@anthropic.com |
|  | warning | `e9c6` Abstract views | 🧱 implementation | `.tableaux/status/e9c6.yaml` | The status dates from 2026-10-05, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: e9c6` | noreply@anthropic.com |
|  | warning | `e3cb` Schema files | 📏 unit | `.tableaux/status/e3cb.yaml` | The status dates from 2026-09-30, more than 7 days before 2026-10-17 | Reaffirm: commit `Reaffirmed: e3cb` | noreply@anthropic.com |

**Most to resolve:** noreply@anthropic.com, 16 of 40.

Command: `tabloio audit --ref corpus/tooling --stale 7 --level detail`
