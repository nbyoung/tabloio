# Audit

**Where do files and history disagree?** Tableaux tooling · ref `corpus/tooling` · person nbyoung@nbyoung.com · stale 7 days · level detail

Legend: `tabloio gates`.

**0 errors, 7 warnings, 15 items of information**: 22 findings in 5 rows.

| Rule | Severity | Task | Gate | File | Message | Action | Resolver |
|---|---|---|---|---|---|---|---|
| R9 | warning | `6103` `c6e8` `595e` |  | `.tableaux/tasks/6103.yaml`, `.tableaux/tasks/c6e8.yaml`, `.tableaux/tasks/595e.yaml` | the requirement on 77b2 from implementation is due at implementation and unmet: 77b2 stands at function | Await the task the requirement names | **nbyoung@nbyoung.com** |
| J19 | warning | `77b2` tablo: backend library and plumbing |  | `.tableaux/tasks/77b2.yaml` | the project at subprojects/tablo states tableaux 0.3.1, before this project's 0.3.2 | Raise the subproject's language | **nbyoung@nbyoung.com** |
| J19 | warning | `6103` tabloio: command line, output and input |  | `.tableaux/tasks/6103.yaml` | the project at subprojects/tabloio states tableaux 0.3.1, before this project's 0.3.2 | Raise the subproject's language | **nbyoung@nbyoung.com** |
| J19 | warning | `c6e8` tablotui: terminal user interface |  | `.tableaux/tasks/c6e8.yaml` | the project at subprojects/tablotui states tableaux 0.3.1, before this project's 0.3.2 | Raise the subproject's language | **nbyoung@nbyoung.com** |
| J19 | warning | `595e` tableaud: local daemon and HTML |  | `.tableaux/tasks/595e.yaml` | the project at subprojects/tableaud states tableaux 0.3.1, before this project's 0.3.2 | Raise the subproject's language | **nbyoung@nbyoung.com** |

**Most to resolve:** **nbyoung@nbyoung.com**, 7 of 22.

Command: `tabloio audit --person nbyoung@nbyoung.com --ref corpus/tooling --stale 7 --level detail`
