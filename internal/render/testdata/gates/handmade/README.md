# Hand-made fixtures

The fixture below is not tablo's output. No ref of the tooling plan or of
the corpus entry `weather-station` holds its data, so `tablo gates --json`
cannot draw it. A regeneration of this view's fixtures from tablo leaves it
as it stands.

| File | Holds | Why it stays hand-made |
|---|---|---|
| `../draft.json` | The data of design e3ed's draft `design/e3ed/gates.json`, in the envelope's shape (e3ed decision 17) | T2 compares its goldens with the design's drafts byte for byte, and the draft's data is the mockup's, not a commit's |

`TestRealMockups` in `internal/cli/mockups_test.go` names it in `handMade`
and runs no command line for it. Change it only with the case it serves.
