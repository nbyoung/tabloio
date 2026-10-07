# History

**What happened, when, and who did it?** Tableaux tooling · ref `0704a09..main` at `3cdae52`, 2026-10-05 · level provenance

Legend: `tabloio gates`.

13 events in 7 commits, 2026-09-30 to 2026-10-05: 0 task, 0 authorised, 1 status, 0 reaffirmed, 0 reviewed, 12 pin.

## 2026-09-30 · 1 event in 1 commit: 1 status

- **2026-09-30** · noreply@anthropic.com · status · `e9c6` Abstract views · 📐 design 🟢 nominal
  - Status after: 📐 design 🟢 nominal
  - Note: VIEWS.md is on the trunk at language 0.3.1; the twenty mockups may start, and implementation puts the text in place
  - `437e` Tableaux tooling after: 📝 defined 🟢 nominal, rolled up from `99f0` Gate definition view in Markdown
  - Model: `claude-fable-5-1`; the 📐 design junction states `claude-fable`
  - Commit: `879b447a920aad55e698906812350ea7d3785b4c`
  - Subject: Advance the abstract views task to its design gate
  - Author: Claude Fable 5.1 \<noreply@anthropic.com>, 2026-09-30 15:17:53 -0400
  - Committer: Claude Fable 5.1 \<noreply@anthropic.com>, 2026-09-30 15:17:53 -0400
  - Trailers: `Model: claude-fable-5-1`, `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`
  - Files: `.tableaux/status/e9c6.yaml` (changed)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/e9c6.yaml .tableaux/status/e9c6.yaml
    git show 879b447:.tableaux/status/e9c6.yaml
    ```

## 2026-10-02 · 10 events in 4 commits: 10 pin

- **2026-10-02** · nbyoung@nbyoung.com · pin · `6103` · `subprojects/tabloio` `62261d4` → `7eba1e5`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tabloio`
  - Model: `claude-opus-5-5`
  - Subproject events: 16 events in 8 commits of `subprojects/tabloio`
    - 2026-09-30 · noreply@anthropic.com · status, reviewed · `a3cc` Status renderers · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-09-30 · noreply@anthropic.com · status, reviewed · `c48a` Structural renderers · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-09-30 · noreply@anthropic.com · status, reviewed · `9167` Temporal renderers · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-09-30 · noreply@anthropic.com · status, reviewed · `e0f7` Text renderers · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-09-30 · noreply@anthropic.com · status, reviewed · `b618` Write commands · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-09-30 · noreply@anthropic.com · status, reviewed · `e3ed` Legend and task renderers · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-09-30 · noreply@anthropic.com · status, reviewed · `e4c7` Agent briefs · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-09-30 · noreply@anthropic.com · status, reviewed · `5ca9` Read commands · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
  - Commit: `55e32e32d65e9cc04da707b215f4b9009a4806e6`
  - Subject: Advance tabloio to its function prototypes
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:07:09 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:07:09 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tabloio` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/6103.yaml .tableaux/status/6103.yaml subprojects/tabloio
    git diff --raw --abbrev=7 55e32e3^ 55e32e3 -- subprojects/tabloio
    git -C subprojects/tabloio log --format='%as %h %ae' 62261d4..7eba1e5 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tabloio log --format='%as %h %ae' 62261d4..7eba1e5 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `77b2` · `subprojects/tablo` `7e94d58` → `816f2a0`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tablo`, unchanged
  - Model: `claude-opus-5-5`
  - Subproject events: 0 events in 0 commits of `subprojects/tablo`
  - Commit: `fb8c0b2b527789760956b0a238c5f07b97c7f277`
  - Subject: Advance the subprojects to the eyeglasses review symbol
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:12:40 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:12:40 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/77b2.yaml .tableaux/status/77b2.yaml subprojects/tablo
    git diff --raw --abbrev=7 fb8c0b2^ fb8c0b2 -- subprojects/tablo
    git -C subprojects/tablo log --format='%as %h %ae' 7e94d58..816f2a0 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tablo log --format='%as %h %ae' 7e94d58..816f2a0 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `6103` · `subprojects/tabloio` `7eba1e5` → `581ae98`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tabloio`, unchanged
  - Model: `claude-opus-5-5`
  - Subproject events: 0 events in 0 commits of `subprojects/tabloio`
  - Commit: `fb8c0b2b527789760956b0a238c5f07b97c7f277`
  - Subject: Advance the subprojects to the eyeglasses review symbol
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:12:40 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:12:40 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/6103.yaml .tableaux/status/6103.yaml subprojects/tabloio
    git diff --raw --abbrev=7 fb8c0b2^ fb8c0b2 -- subprojects/tabloio
    git -C subprojects/tabloio log --format='%as %h %ae' 7eba1e5..581ae98 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tabloio log --format='%as %h %ae' 7eba1e5..581ae98 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `c6e8` · `subprojects/tablotui` `159fdab` → `dbfb8e3`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tablotui`, unchanged
  - Model: `claude-opus-5-5`
  - Subproject events: 0 events in 0 commits of `subprojects/tablotui`
  - Commit: `fb8c0b2b527789760956b0a238c5f07b97c7f277`
  - Subject: Advance the subprojects to the eyeglasses review symbol
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:12:40 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:12:40 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/c6e8.yaml .tableaux/status/c6e8.yaml subprojects/tablotui
    git diff --raw --abbrev=7 fb8c0b2^ fb8c0b2 -- subprojects/tablotui
    git -C subprojects/tablotui log --format='%as %h %ae' 159fdab..dbfb8e3 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tablotui log --format='%as %h %ae' 159fdab..dbfb8e3 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `595e` · `subprojects/tableaud` `2061a7f` → `b06cad0`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tableaud`, unchanged
  - Model: `claude-opus-5-5`
  - Subproject events: 0 events in 0 commits of `subprojects/tableaud`
  - Commit: `fb8c0b2b527789760956b0a238c5f07b97c7f277`
  - Subject: Advance the subprojects to the eyeglasses review symbol
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:12:40 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:12:40 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/595e.yaml .tableaux/status/595e.yaml subprojects/tableaud
    git diff --raw --abbrev=7 fb8c0b2^ fb8c0b2 -- subprojects/tableaud
    git -C subprojects/tableaud log --format='%as %h %ae' 2061a7f..b06cad0 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tableaud log --format='%as %h %ae' 2061a7f..b06cad0 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `77b2` · `subprojects/tablo` `816f2a0` → `00f8f68`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tablo`, unchanged
  - Model: `claude-opus-5-5`
  - Subproject events: 0 events in 0 commits of `subprojects/tablo`
  - Commit: `552d38fd8adf40e41b39f65d5904e6fc6a88b265`
  - Subject: Advance the subprojects to the wrench, brick, ruler and link gate symbols
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:34:57 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:34:57 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/77b2.yaml .tableaux/status/77b2.yaml subprojects/tablo
    git diff --raw --abbrev=7 552d38f^ 552d38f -- subprojects/tablo
    git -C subprojects/tablo log --format='%as %h %ae' 816f2a0..00f8f68 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tablo log --format='%as %h %ae' 816f2a0..00f8f68 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `6103` · `subprojects/tabloio` `581ae98` → `cb9ff8f`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tabloio`, unchanged
  - Model: `claude-opus-5-5`
  - Subproject events: 0 events in 0 commits of `subprojects/tabloio`
  - Commit: `552d38fd8adf40e41b39f65d5904e6fc6a88b265`
  - Subject: Advance the subprojects to the wrench, brick, ruler and link gate symbols
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:34:57 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:34:57 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/6103.yaml .tableaux/status/6103.yaml subprojects/tabloio
    git diff --raw --abbrev=7 552d38f^ 552d38f -- subprojects/tabloio
    git -C subprojects/tabloio log --format='%as %h %ae' 581ae98..cb9ff8f -- .tableaux/tasks .tableaux/status
    git -C subprojects/tabloio log --format='%as %h %ae' 581ae98..cb9ff8f -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `c6e8` · `subprojects/tablotui` `dbfb8e3` → `334df71`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tablotui`, unchanged
  - Model: `claude-opus-5-5`
  - Subproject events: 0 events in 0 commits of `subprojects/tablotui`
  - Commit: `552d38fd8adf40e41b39f65d5904e6fc6a88b265`
  - Subject: Advance the subprojects to the wrench, brick, ruler and link gate symbols
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:34:57 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:34:57 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/c6e8.yaml .tableaux/status/c6e8.yaml subprojects/tablotui
    git diff --raw --abbrev=7 552d38f^ 552d38f -- subprojects/tablotui
    git -C subprojects/tablotui log --format='%as %h %ae' dbfb8e3..334df71 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tablotui log --format='%as %h %ae' dbfb8e3..334df71 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `595e` · `subprojects/tableaud` `b06cad0` → `cffbfa6`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tableaud`, unchanged
  - Model: `claude-opus-5-5`
  - Subproject events: 0 events in 0 commits of `subprojects/tableaud`
  - Commit: `552d38fd8adf40e41b39f65d5904e6fc6a88b265`
  - Subject: Advance the subprojects to the wrench, brick, ruler and link gate symbols
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:34:57 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 07:34:57 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablo` (pin), `subprojects/tabloio` (pin), `subprojects/tablotui` (pin), `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/595e.yaml .tableaux/status/595e.yaml subprojects/tableaud
    git diff --raw --abbrev=7 552d38f^ 552d38f -- subprojects/tableaud
    git -C subprojects/tableaud log --format='%as %h %ae' b06cad0..cffbfa6 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tableaud log --format='%as %h %ae' b06cad0..cffbfa6 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```
- **2026-10-02** · nbyoung@nbyoung.com · pin · `6103` · `subprojects/tabloio` `cb9ff8f` → `4593882`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tabloio`
  - Model: `claude-opus-5-5`
  - Subproject events: 3 events in 3 commits of `subprojects/tabloio`
    - 2026-09-30 · noreply@anthropic.com · status · `1422` Snapshot workflow · 📝 defined 🟢 nominal 👓 review · `claude-opus-5-5`
    - 2026-10-02 · nbyoung@nbyoung.com · reviewed · `1422` Snapshot workflow · 📐 design · `claude-opus-5-5`
    - 2026-10-02 · noreply@anthropic.com · status · `1422` Snapshot workflow · 📐 design 🟢 nominal · `claude-opus-5-5`
  - Commit: `1c0f04daa18af8b4140e8eab2948466981b65010`
  - Subject: Advance tabloio to its snapshot workflow design
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 13:48:21 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-02 13:48:21 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tabloio` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/6103.yaml .tableaux/status/6103.yaml subprojects/tabloio
    git diff --raw --abbrev=7 1c0f04d^ 1c0f04d -- subprojects/tabloio
    git -C subprojects/tabloio log --format='%as %h %ae' cb9ff8f..4593882 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tabloio log --format='%as %h %ae' cb9ff8f..4593882 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```

## 2026-10-04 · 1 event in 1 commit: 1 pin

- **2026-10-04** · nbyoung@nbyoung.com · pin · `c6e8` · `subprojects/tablotui` `334df71` → `a2878b5`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tablotui`
  - Model: `claude-opus-5-5`
  - Subproject events: 10 events in 5 commits of `subprojects/tablotui`
    - 2026-10-02 · noreply@anthropic.com · status, reviewed · `518e` Actions · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-02 · noreply@anthropic.com · status, reviewed · `f394` Role and context · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-02 · noreply@anthropic.com · status, reviewed · `679b` Tableau grid · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-02 · noreply@anthropic.com · status, reviewed · `171b` Live refresh · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-02 · noreply@anthropic.com · status, reviewed · `0afa` Detail panes · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
  - Commit: `fa4551ecdc8d106722ac41b25a9213abc453fab0`
  - Subject: Advance tablotui to its five function prototypes
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-04 16:38:02 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-04 16:38:02 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tablotui` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/c6e8.yaml .tableaux/status/c6e8.yaml subprojects/tablotui
    git diff --raw --abbrev=7 fa4551e^ fa4551e -- subprojects/tablotui
    git -C subprojects/tablotui log --format='%as %h %ae' 334df71..a2878b5 -- .tableaux/tasks .tableaux/status
    git -C subprojects/tablotui log --format='%as %h %ae' 334df71..a2878b5 -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```

## 2026-10-05 · 1 event in 1 commit: 1 pin

- **2026-10-05** · nbyoung@nbyoung.com · pin · `595e` · `subprojects/tableaud` `cffbfa6` → `e6ec4ec`
  - Status after: 📝 defined 🟢 nominal, 🪆 from the subproject `subprojects/tableaud`
  - Model: `claude-opus-5-5`
  - Subproject events: 21 events in 13 commits of `subprojects/tableaud`
    - 2026-10-05 · noreply@anthropic.com · status, reviewed · `9a9c` Temporal templates · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · status, reviewed · `44bf` Status templates · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · reviewed · `438a` Legend and task templates · 🔧 function · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · status, reviewed · `438a` Legend and task templates · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · reviewed · `e2b6` Structural templates · 🔧 function · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · status, reviewed · `e2b6` Structural templates · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · reviewed · `49ce` Server · 🔧 function · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · status, reviewed · `49ce` Server · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · reviewed · `db74` Progressive disclosure · 🔧 function · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · status, reviewed · `db74` Progressive disclosure · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · reviewed · `5a2f` Accessibility and theming · 🔧 function · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · status, reviewed · `5a2f` Accessibility and theming · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
    - 2026-10-05 · noreply@anthropic.com · status, reviewed · `6160` Static export · 🔧 function 🟢 nominal · `claude-sonnet-5-5`
  - Commit: `3cdae527d47d37fdff55bc302c76c017cf91c606`
  - Subject: Advance tableaud to the eight function prototypes
  - Author: Norman Young \<nbyoung@nbyoung.com>, 2026-10-05 12:59:45 -0400
  - Committer: Norman Young \<nbyoung@nbyoung.com>, 2026-10-05 12:59:45 -0400
  - Trailers: `Model: claude-opus-5-5`, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`
  - Files: `subprojects/tableaud` (pin)
  - Reproduce:
    ```
    git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks/595e.yaml .tableaux/status/595e.yaml subprojects/tableaud
    git diff --raw --abbrev=7 3cdae52^ 3cdae52 -- subprojects/tableaud
    git -C subprojects/tableaud log --format='%as %h %ae' cffbfa6..e6ec4ec -- .tableaux/tasks .tableaux/status
    git -C subprojects/tableaud log --format='%as %h %ae' cffbfa6..e6ec4ec -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
    ```

```
git log --format='%as %h %ae' 0704a09..main -- .tableaux/tasks .tableaux/status subprojects/tablo subprojects/tabloio subprojects/tablotui subprojects/tableaud
git log --format='%as %h %ae' 0704a09..main -E --grep='^(Authorised|Reaffirmed): [0-9a-f]{4}$' --grep='^Reviewed: [0-9a-f]{4} '
```

Command: `tabloio history --ref 3cdae52 --level provenance`
