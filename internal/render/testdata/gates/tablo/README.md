# tablo's golden data

The JSON files here are tablo's golden data of the `gates` view, copied
unchanged from `github.com/nbyoung/tablo` at commit
`0a38365e4e38d25774c00848829cbdcdaf938ad6`, the module version
`v0.0.0-20261016180757-0a38365e4e38`. Each holds the `data` of one
envelope: the view of the corpus entry `weather-station` at `main`.

| File | In tablo | Parameters |
|---|---|---|
| `gates-9f31.provenance.json` | `internal/views/testdata/weather-station/` | task `9f31` |

`TestTabloGoldens` in `tablo_test.go` wraps each file in the envelope that
tablo's own schema test builds around its golden data (the source `main`
with no commit, no project, an observer), with the parameters above and
the level the name ends in. It decodes the envelope, renders the view at
each level the data holds in both formats, and checks the rules every
fixture keeps. These files are the oracle of the view types: change none
of them by hand. When tablo's golden data changes, copy the files again
and name the new commit here.
