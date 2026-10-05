# Audit

**Where do files and history disagree?** Tableaux tooling · ref `main` · stale 7 days · level glance

Legend: `tabloio gates`.

**1 error, 3 warnings, 28 items of information**: 32 findings.

| Severity | Rule | Kind | Findings | Tasks | Resolver |
|---|---|---|--:|---|---|
| error | S11 | A status past a reviewed junction with no review | 1 | `e3cb` Schema files | noreply@anthropic.com |
| warning | H3 | A `Model:` trailer outside the junction's model | 3 | `437e` `bc63` `ac33` | nbyoung@nbyoung.com |
| information | H4 | A hand-off the history implies | 28 | `99f0` `05a9` `a9ce` `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` `ac33` `9f3f` `7166` `e9c6` `fcec` `c2ad` `e3cb` `7861` | noreply@anthropic.com |

**Most to resolve:** noreply@anthropic.com, 29 of 32.

Command: `tabloio audit --ref main --stale 7 --level glance`
