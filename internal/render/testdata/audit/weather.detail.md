# Audit

**Where do files and history disagree?** Weather station · ref `main` · stale 3 days · level detail

Legend: `tabloio gates`.

**0 errors, 3 warnings, 32 items of information**: 35 findings in 3 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| R9 | warning | `c07d` Node firmware |  | `.tableaux/tasks/c07d.yaml` | the requirement on 9f31 from design is due at implementation and unmet: 9f31 stands at function | Await `9f31` at 📐 design | ada@example.org |
|  | warning | `3c5d` Dashboard |  | `.tableaux/tasks/3c5d.yaml` | The task stands proposed: dan@example.org, the author of its deciding commit, is none of its authorities | Authorise: commit `Authorised: 3c5d` | ada@example.org |
|  | warning | `3c5d` Dashboard | 📝 defined | `.tableaux/status/3c5d.yaml` | The status dates from 2026-09-22, more than 3 days before 2026-09-28 | Reaffirm: commit `Reaffirmed: 3c5d` | dan@example.org |

**Most to resolve:** ada@example.org, 2 of 35.

Command: `tabloio audit --ref main --stale 3 --level detail`
