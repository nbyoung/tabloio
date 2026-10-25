# Global tableau

**How does the whole project stand?** tabloio · ref `main` · window 1 · level glance

Legend: `tabloio gates`.

| Id | Task | ❔…📐 ×0 | 🧱 | 📏 | 🔗 | 🌍 | 🚀 |
|---|---|:-:|:-:|:-:|:-:|:-:|:-:|
| [`07e0`](.tableaux/tasks/07e0.yaml) | **tabloio** |  | 🟢 | 🤖 | 🤖 | 🤖👀 | 🧑 |
| [`0a2f`](.tableaux/tasks/0a2f.yaml) | &nbsp;&nbsp;Bootstrap |  | 🟢 | — | — | — | 🧑 |
| [`1093`](.tableaux/tasks/1093.yaml) | &nbsp;&nbsp;**Renderers** (5) |  |  |  | 🟢👓 | 🤖👀 | 🧑 |
| [`5ca9`](.tableaux/tasks/5ca9.yaml) | &nbsp;&nbsp;Read commands |  |  |  | 🟢 | 🤖👀 | 🧑 |
| [`b618`](.tableaux/tasks/b618.yaml) | &nbsp;&nbsp;Write commands |  |  |  | 🟢 | 🤖👀 | 🧑 |
| [`1422`](.tableaux/tasks/1422.yaml) | &nbsp;&nbsp;Snapshot workflow |  |  |  |  | 🟢 | 🧑 |
| [`e4c7`](.tableaux/tasks/e4c7.yaml) | &nbsp;&nbsp;Agent briefs |  |  |  | 🟢 | 🤖👀 | 🧑 |

❔ undefined · 📐 design · 🧱 implementation · 📏 unit · 🔗 integrate · 🌍 validate · 🚀 release · 🟢 nominal · 👓 review · 🧑 a person contributes · 🤖 an agent contributes · 👀 a reviewer accepts · — the gate does not apply.

## Notes

| Rows | Date | Note | From |
|---|---|---|---|
| `07e0` | 2026-09-30 | The skeleton is on the trunk and passes vet and test locally and in CI run 36673447631 on d6eaed1; the tablo require line waits for the Renderers task, as the README states | rolled up from `0a2f` Bootstrap |
| `0a2f` | 2026-09-30 | The skeleton is on the trunk and passes vet and test locally and in CI run 36673447631 on d6eaed1; the tablo require line waits for the Renderers task, as the README states |  |
| `1093` | 2026-10-23 | validate awaits the owner's review. The field tests show TABLEAUX.md on GitHub at main equal to tableau --level glance on a clone at that commit, and the gate and task views render at detail with exit 0. The owner reads TABLEAUX.md and a task page in a GitHub comment preview. | rolled up from `e3ed` Legend and task renderers |
| `5ca9` | 2026-10-23 | validate waits on the owner's checks and on a person's own plan through the close of Parcelang's migration, with the cost and the views of the read commands on that plan as its evidence. The field tests now show the ten read commands exit 0 at three levels in both formats on nine plans in 540 runs, the slowest command at 390 ms on the largest plan under a load of about 4. |  |
| `b618` | 2026-10-23 | validate waits on a repair. The five write commands pass the field tests with standard input a pipe, which exits 2 without --yes and commits with it, and a hand the method refuses exits 1 with no commit. They fail with standard input on /dev/null, which the terminal check counts as a terminal, so the prompt shows and the command exits 1 where decision 6 and T28 state exit 2. The prompt on a real terminal and tablotui's use wait on that repair. |  |
| `1422` | 2026-10-23 | The tools follow design 1422 under the owner's review (8ebdc09), on 7 of 7 points of the text, T1 to T14 on tabloio v0.2.0 and tablo v0.2.0 with the hook in a median of 66 ms, and the plan valid with the tablo clone mapped; GitHub shows the trunk-only regenerate skipped on the branch push 28444e5, contents write against the default read, the GITHUB\_TOKEN push f5d2558 starting no run, and the deploy key's push 71f5ab7 onto the ruleset-protected main followed by one run at the fixed point |  |
| `e4c7` | 2026-10-23 | validate waits on the owner's checks and on a person's own plan through the close of Parcelang's migration, with an agent dispatched from a generated brief for a real item as its evidence. The field tests now show 110 items of work ready on nine plans, each with a generated brief that holds A1 to A7, its six parts and the reproduction by its last command. |  |

Command: `tabloio tableau --ref main --window 1 --level glance`
