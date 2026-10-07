# History

**What happened, when, and who did it?** Tableaux tooling · ref `a7932ae..37a08e9` at `37a08e9`, 2026-09-29 · level provenance

Legend: `tabloio gates`.

4 events in 1 commit, 2026-09-29: 0 task, 0 authorised, 0 status, 0 reaffirmed, 0 reviewed, 4 pin.

## 2026-09-29 · 4 events in 1 commit: 4 pin

- **2026-09-29** · nbyoung@nbyoung.com · pin · `77b2` · `subprojects/tablo` new `dc669dd`
  - Status after: ❔ undefined ⚪ undefined, 🪆 from the subproject `subprojects/tablo`
  - Model: none recorded; exempt
  - Subproject events: 12 events in 1 commit of `subprojects/tablo`
  - Commit: `37a08e92e840ea3cb48bc7ca6a5041c44096ef8e`
  - Subject: Pin the subprojects
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-09-29 17:30:02 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-09-29 17:30:02 -0400
  - Trailers: `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' a7932ae..37a08e9 -- .tableaux/tasks/77b2.yaml .tableaux/status/77b2.yaml subprojects/tablo
    git diff --raw --abbrev=7 37a08e9^ 37a08e9 -- subprojects/tablo
    git -C subprojects/tablo log --format='%as %h %ae' dc669dd -- .tableaux/tasks .tableaux/status
    ```
- **2026-09-29** · nbyoung@nbyoung.com · pin · `6103` · `subprojects/tabloio` new `46521bc`
  - Status after: ❔ undefined ⚪ undefined, 🪆 from the subproject `subprojects/tabloio`
  - Model: none recorded; exempt
  - Subproject events: 12 events in 1 commit of `subprojects/tabloio`
  - Commit: `37a08e92e840ea3cb48bc7ca6a5041c44096ef8e`
  - Subject: Pin the subprojects
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-09-29 17:30:02 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-09-29 17:30:02 -0400
  - Trailers: `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' a7932ae..37a08e9 -- .tableaux/tasks/6103.yaml .tableaux/status/6103.yaml subprojects/tabloio
    git diff --raw --abbrev=7 37a08e9^ 37a08e9 -- subprojects/tabloio
    git -C subprojects/tabloio log --format='%as %h %ae' 46521bc -- .tableaux/tasks .tableaux/status
    ```
- **2026-09-29** · nbyoung@nbyoung.com · pin · `c6e8` · `subprojects/tablotui` new `4544042`
  - Status after: ❔ undefined ⚪ undefined, 🪆 from the subproject `subprojects/tablotui`
  - Model: none recorded; exempt
  - Subproject events: 7 events in 1 commit of `subprojects/tablotui`
  - Commit: `37a08e92e840ea3cb48bc7ca6a5041c44096ef8e`
  - Subject: Pin the subprojects
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-09-29 17:30:02 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-09-29 17:30:02 -0400
  - Trailers: `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' a7932ae..37a08e9 -- .tableaux/tasks/c6e8.yaml .tableaux/status/c6e8.yaml subprojects/tablotui
    git diff --raw --abbrev=7 37a08e9^ 37a08e9 -- subprojects/tablotui
    git -C subprojects/tablotui log --format='%as %h %ae' 4544042 -- .tableaux/tasks .tableaux/status
    ```
- **2026-09-29** · nbyoung@nbyoung.com · pin · `595e` · `subprojects/tableaud` new `59783a9`
  - Status after: ❔ undefined ⚪ undefined, 🪆 from the subproject `subprojects/tableaud`
  - Model: none recorded; exempt
  - Subproject events: 11 events in 1 commit of `subprojects/tableaud`
  - Commit: `37a08e92e840ea3cb48bc7ca6a5041c44096ef8e`
  - Subject: Pin the subprojects
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-09-29 17:30:02 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-09-29 17:30:02 -0400
  - Trailers: `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' a7932ae..37a08e9 -- .tableaux/tasks/595e.yaml .tableaux/status/595e.yaml subprojects/tableaud
    git diff --raw --abbrev=7 37a08e9^ 37a08e9 -- subprojects/tableaud
    git -C subprojects/tableaud log --format='%as %h %ae' 59783a9 -- .tableaux/tasks .tableaux/status
    ```

```
git log --format='%as %h %ae' a7932ae..37a08e9 -- .tableaux/tasks .tableaux/status subprojects/tablo subprojects/tabloio subprojects/tablotui subprojects/tableaud
```

Command: `tabloio history --ref 37a08e9 --level provenance`
