# Hand-made fixtures

The fixture below is not tablo's output. No ref of the tooling plan or of
the corpus entry `weather-station` holds its data, so `tablo blockage
--json` cannot draw it. A regeneration of this view's fixtures from tablo
leaves it as it stands.

| File | Holds | Why it stays hand-made |
|---|---|---|
| `../illustration.json` | The illustration of the mockup `blockage.md`, which design a3cc's draft `design/a3cc/blockage.detail.md` rewords | T10 compares its goldens with the draft, and the causes it draws stand in no commit of either project |

`TestRealMockups` in `internal/cli/mockups_test.go` names it in `handMade`
and runs no command line for it. Change it only with the case it serves.
