# Audit

**Where do files and history disagree?** Weather station · ref `main` · stale 3 days · level glance

Legend: `tabloio gates`.

**0 errors, 2 warnings, 0 items of information**: 2 findings.

| Severity | Rule | Kind | Findings | Tasks | Resolver |
|---|---|---|--:|---|---|
| warning |  | A proposed task | 1 | `9f31` Sensor board | ada@example.org |
| warning |  | A status older than the days stated | 1 | `3c5d` Dashboard | dan@example.org |

**Most to resolve:** ada@example.org, 1 of 2.

Command: `tabloio audit --ref main --stale 3 --level glance`
