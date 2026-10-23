# Global tableau

**How does the whole project stand?** tabloio · ref `main` · window 1 · level glance

Legend: `tabloio gates`.

| Id | Task | ❔…📐 ×0 | 🧱 | 📏 | 🔗 | 🌍 | 🚀 |
|---|---|:-:|:-:|:-:|:-:|:-:|:-:|
| [`07e0`](.tableaux/tasks/07e0.yaml) | **tabloio** |  | 🟢 | 🤖 | 🤖 | 🤖👀 | 🧑 |
| [`0a2f`](.tableaux/tasks/0a2f.yaml) | &nbsp;&nbsp;Bootstrap |  | 🟢 | — | — | — | 🧑 |
| [`1093`](.tableaux/tasks/1093.yaml) | &nbsp;&nbsp;**Renderers** (5) |  |  |  | 🟢 | 🤖👀 | 🧑 |
| [`5ca9`](.tableaux/tasks/5ca9.yaml) | &nbsp;&nbsp;Read commands |  |  |  | 🟢 | 🤖👀 | 🧑 |
| [`b618`](.tableaux/tasks/b618.yaml) | &nbsp;&nbsp;Write commands |  |  |  | 🟢 | 🤖👀 | 🧑 |
| [`1422`](.tableaux/tasks/1422.yaml) | &nbsp;&nbsp;Snapshot workflow |  |  |  | 🟢 | 🤖👀 | 🧑 |
| [`e4c7`](.tableaux/tasks/e4c7.yaml) | &nbsp;&nbsp;Agent briefs |  |  |  | 🟢 | 🤖👀 | 🧑 |

❔ undefined · 📐 design · 🧱 implementation · 📏 unit · 🔗 integrate · 🌍 validate · 🚀 release · 🟢 nominal · 🧑 a person contributes · 🤖 an agent contributes · 👀 a reviewer accepts · — the gate does not apply.

## Notes

| Rows | Date | Note | From |
|---|---|---|---|
| `07e0` | 2026-09-30 | The skeleton is on the trunk and passes vet and test locally and in CI run 36673447631 on d6eaed1; the tablo require line waits for the Renderers task, as the README states | rolled up from `0a2f` Bootstrap |
| `0a2f` | 2026-09-30 | The skeleton is on the trunk and passes vet and test locally and in CI run 36673447631 on d6eaed1; the tablo require line waits for the Renderers task, as the README states |  |
| `1093` | 2026-10-17 | T1 to T15 pass on fixtures drawn with tablo at the pin, the tooling plan's at corpus/tooling and weather-station's at main, and the Read commands give every golden of the gate and task views through Run (5ca9 T18, TestRealMockups); gates/draft stays as the draft T2 compares; the validate gate is the owner reading STATUS.md and a task page on GitHub | rolled up from `e3ed` Legend and task renderers |
| `5ca9` | 2026-10-17 | T1 to T18 pass, T16 and T17 again on a copy of weather-station with TABLEAUX\_CORPUS set, and T18 TestRealMockups reaches 231 golden files through Run, those of the tooling plan on tableaux-tooling through TABLO\_CORPUS on a clone of the umbrella at corpus/tooling; the design names no validate check, so at validate the owner runs the read commands on the plan; 1422 T14 waits on releases of tablo and tabloio |  |
| `b618` | 2026-10-17 | T1 to T29 pass, T29 TestCorpusEntries on copies of eight corpus entries with tablo's library, where each refusal fires from tablo's facts, each commit adds no validator error and shows in tablo's history, and every value of scalars.tsv reads back through tablo's YAML reader; tablo's 4ed9, 8118 and 493e stand at implementation, unit and unit at the pin 03c5e35; the validate gate is the prompt on a real terminal and tablotui's use (518e A9), which waits on a tabloio tag that holds action |  |
| `1422` | 2026-10-19 | T14 passes through snapshot/test.sh --real with the releases tabloio v0.1.1 and tablo v0.1.0, T3, T5, T6 and T11 on a copy of version-trunk-empty and the hook in a median of 59 ms on the Tableaux plan, and T1 to T13 pass as CI runs them; the validate gate needs the trunk-only condition, the permissions, a protected trunk with a bypass credential and loop avoidance seen on GitHub, which waits on a workflow of this repository that calls the action at a tag |  |
| `e4c7` | 2026-10-17 | T1 to T15 pass, and TestRealBrief gives from tablo's library the brief of e3ed at integrate, work ready in tabloio at the pin of the tooling tag, and confirms A1 to A7; part 5 quotes tabloio record and tabloio review (b618 A10); the root's junctions of design, implementation and unit each name their standing text under docs/gates/, and the six other gates name none; the validate gate is the owner dispatching an agent from a generated brief and checking the identity and the trailers of its commits |  |

Command: `tabloio tableau --ref main --window 1 --level glance`
