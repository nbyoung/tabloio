# tablo's golden data

The JSON files here are tablo's golden data of the `audit` view, copied
unchanged from `github.com/nbyoung/tablo` at commit
`03c5e35eac90d3da49fb63f29a4b3ca35b16c0e1`, the module version
`v0.0.0-20261017141209-03c5e35eac90` that `go.mod` requires; each file
there equals its copy here byte for byte. Each holds the `data` of one
envelope: the view of the corpus entry `weather-station` at `main`.

| File | In tablo | Parameters |
|---|---|---|
| `audit.provenance.json` | `view/testdata/` | stale 7, now 2026-09-28 |

`TestTabloGoldens` in `tablo_test.go` wraps each file in the envelope that
tablo's own schema test builds around its golden data (the source `main`
with no commit, no project, an observer), with the parameters above and
the level the name ends in. It decodes the envelope, renders the view at
each level the data holds in both formats, and checks the rules every
fixture keeps. These files are the oracle of the view types: change none
of them by hand. When tablo's golden data changes, copy the files again
and name the new commit here.
