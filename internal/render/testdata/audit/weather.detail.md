# Audit

**Where do files and history disagree?** Weather station · ref `main` · stale 3 days · level detail

Legend: `tabloio gates`.

**0 errors, 2 warnings, 0 items of information**: 2 findings in 2 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
|  | warning | `9f31` Sensor board |  | `tasks/9f31.yaml` | The task stands proposed: no trailer or owner commit on the trunk authorises it | Have an authority merge the task or commit the `Authorised:` trailer | ada@example.org |
|  | warning | `3c5d` Dashboard | 📝 defined | `status/3c5d.yaml` | The status dates from 2026-09-22, more than 3 days before 2026-09-28, with no later reaffirmation | Reaffirm the status with `Reaffirmed: 3c5d`, or record the next event | dan@example.org |

**Most to resolve:** ada@example.org, 1 of 2.

Command: `tabloio audit --ref main --stale 3 --level detail`
