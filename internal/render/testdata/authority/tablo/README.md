# tablo's golden data

The JSON files here are tablo's golden data of the `authority` view, copied
unchanged from `github.com/nbyoung/tablo` at commit
`620473c299bd693af258804146c44d28bd3d737b`, the module version
`v0.0.0-20261017000822-620473c299bd`. Each holds the `data` of one
envelope: the view of the corpus entry `weather-station` at `main`.

| File | In tablo | Parameters |
|---|---|---|
| `authority.provenance.json` | `internal/views/testdata/weather-station/` | task `a1c0` |

`TestTabloGoldens` in `tablo_test.go` wraps each file in the envelope that
tablo's own schema test builds around its golden data (the source `main`
with no commit, no project, an observer), with the parameters above and
the level the name ends in. It decodes the envelope, renders the view at
each level the data holds in both formats, and checks the rules every
fixture keeps. These files are the oracle of the view types: change none
of them by hand. When tablo's golden data changes, copy the files again
and name the new commit here.
