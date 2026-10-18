# Audit

**Where do files and history disagree?** Weather station · ref `main` · stale 3 days · level glance

Legend: `tabloio gates`.

**0 errors, 3 warnings, 32 items of information**: 35 findings.

| Severity | Rule | Kind | Findings | Tasks | Resolver |
|---|---|---|--:|---|---|
| warning | R9 | An unmet requirement | 1 | `c07d` Node firmware | ada@example.org |
| warning |  | A proposed task | 1 | `3c5d` Dashboard | ada@example.org |
| warning |  | A stale status | 1 | `3c5d` Dashboard | dan@example.org |

**Most to resolve:** ada@example.org, 2 of 35.

Command: `tabloio audit --ref main --stale 3 --level glance`
