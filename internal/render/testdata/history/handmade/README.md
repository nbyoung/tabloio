# Hand-made fixtures

The fixtures below are not tablo's output. No ref of the tooling plan or of
the corpus entry `weather-station` holds their data, so `tablo history
--json` cannot draw them. A regeneration of this view's fixtures from tablo
leaves them as they stand.

| File | Holds | Why it stays hand-made |
|---|---|---|
| `../in/proposal.json` | A line the branch `dan/display` of `weather-station` adds, proposing `3c5d` | One of the mockup's cases outside its pictures (design 9167, T4); the corpus entry has no such branch, and no range of it gives a line with `proposal` |
| `../in/reaffirmed.json` | A reaffirmation of `e9c6` on 2026-10-07, with no commit | One of the mockup's cases outside its pictures (design 9167, T4); no commit of the tooling plan carries `Reaffirmed: e9c6` |

`TestRealMockups` in `internal/cli/mockups_test.go` names them in
`handMade` and runs no command line for them. Change them only with the
case each serves.
