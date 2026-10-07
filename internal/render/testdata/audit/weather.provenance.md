# Audit

**Where do files and history disagree?** Weather station · ref `main` at `edb30d2`, 2026-09-28 · stale 3 days · level provenance

Legend: `tabloio gates`.

**0 errors, 2 warnings, 0 items of information**: 2 findings in 2 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
|  | warning | `9f31` Sensor board |  | `tasks/9f31.yaml` | The task stands proposed: no trailer or owner commit on the trunk authorises it | Have an authority merge the task or commit the `Authorised:` trailer | ada@example.org |
|  | warning | `3c5d` Dashboard | 📝 defined | `status/3c5d.yaml` | The status dates from 2026-09-22, more than 3 days before 2026-09-28, with no later reaffirmation | Reaffirm the status with `Reaffirmed: 3c5d`, or record the next event | dan@example.org |

**Most to resolve:** ada@example.org, 1 of 2.

## A proposed task, warning: `9f31`

**Authorisation.** No `Authorised:` trailer and no commit of an authority on the trunk names `9f31`.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `91aa774` | 2026-09-18 | ada@example.org |  |  |  |

```
# shows the finding
git log --format='%as %h %ae' -E --grep='^Authorised: 9f31$'
```

## A status older than the days stated, warning: `3c5d` at 📝 defined

**Status.** `3c5d` stands at 📝 defined 🟢 nominal, recorded on 2026-09-22 by `55f57c1`.

| Commit | Date | Author | Committer | Subject | Trailers |
|---|---|---|---|---|---|
| `55f57c1` | 2026-09-22 | dan@example.org |  |  |  |

```
# shows the finding
git log -1 --format='%as %h %ae' -- .tableaux/status/3c5d.yaml
```

Command: `tabloio audit --ref edb30d2 --stale 3 --level provenance`
