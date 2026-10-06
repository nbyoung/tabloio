# tabloio

The Tableaux command-line front end. Its output side renders every view in a textual format for static media; its input side operates the method through Git commits. No output ever changes project state: every change is a commit.

The name plays on *Tableaux*: it joins the pronunciation to *io*, input and output.

## Place in the family

| Project                                            | Role                                                     |
|----------------------------------------------------|----------------------------------------------------------|
| [tableaux](https://github.com/nbyoung/tableaux)    | The language: method, syntax, schemas, corpus, mockups   |
| [tablo](https://github.com/nbyoung/tablo)          | The backend: library and plumbing command                |
| [tabloio](https://github.com/nbyoung/tabloio)      | The command line: textual output and Git input           |
| [tablotui](https://github.com/nbyoung/tablotui)    | The terminal user interface                              |
| [tableaud](https://github.com/nbyoung/tableaud)    | The local daemon: HTML views with progressive disclosure |

The split follows Git's own: `tablo` is plumbing that reads a repository and
emits data, and the three front ends are porcelain that presents it. A front end
never reads a task file itself.

## What it does

- **Output** any view in one of two textual formats. `--markdown`, the default, reads in a plain renderer: a `STATUS.md` that CI regenerates, an attachment to email or chat, a brief pasted into an agent's session. `--text` draws the view with Unicode box characters for a terminal or a log with no Markdown viewer. Scripts that want the view data call `tablo view` for JSON or YAML.
- **Input** to the method: `propose`, `record`, `review`, `reaffirm` and `authorise` write the right file or compose the commit with the right trailers, and show the commit before making it.

```
tabloio tableau > STATUS.md                    # the global tableau in Markdown
tabloio tableau --text                         # the same drawn in Unicode
tabloio context --person ben@example.org        # a contributor's neighbourhood
tabloio queue --person opus@example.org --brief  # a brief for an agent session
tabloio review 9f31 design                     # commit 'Reviewed: 9f31 design'
tabloio authorise 9f31 c07d                    # accept two proposals at once
```

## Plan

The project's plan is the Tableaux project in [`.tableaux/`](.tableaux/). The
[Tableaux tooling plan](https://github.com/nbyoung/tableaux/blob/main/PLAN.md)
in the `tableaux` repository pins this project as the submodule
`subprojects/tabloio` and tracks its root task through a recursive junction, states the review policy every task here inherits, and
proposes Go as the implementation language.

## Layout

The module is `github.com/nbyoung/tabloio`. The four subprojects share this shape.

| Path                  | Holds                                                                                   |
|-----------------------|-----------------------------------------------------------------------------------------|
| `cmd/tabloio/`        | The command. `main.go` prints the version; the read and write subcommands land in their own tasks |
| `internal/render/`    | The renderers from tablo's view data: Markdown and Unicode text                         |
| `.github/workflows/`  | `ci.yml` checks every pull request and push to `main`; `release.yml` publishes a tag    |
| `.goreleaser.yaml`    | The release build: six static binaries, archives and checksums                          |
| `.golangci.yml`       | The lint configuration                                                                  |
| `.tableaux/`          | The plan                                                                                |

Packages under `internal/` stay private to this module. The actions behind the
write commands are the exception in prospect: `tablotui` reuses them, so the
Write commands task designs them as an exported package.

## Build and test

Development needs Go at the version `go.mod` names, or newer; the toolchain
fetches that version itself when the local one is older. Using `tabloio` needs
no Go at all: see [Release](#release).

```
go build ./...        # build
go test ./...         # test
gofmt -l .            # list files gofmt would change; CI fails on any
go vet ./...          # vet
golangci-lint run     # lint, with the configuration in .golangci.yml
```

`ci.yml` runs the same four checks, gofmt, vet, lint and test, on every pull
request and every push to `main`, with the Go version from `go.mod` and
golangci-lint from its official action.

## Release

A release is a tag on `main`:

```
git tag v0.1.0
git push origin v0.1.0
```

`release.yml` runs GoReleaser on the tag. It builds `cmd/tabloio` with
`CGO_ENABLED=0` for Linux, macOS and Windows on amd64 and arm64, so each
binary is static and depends on nothing on the host but `git`. It publishes a
`tar.gz` archive per target, `zip` on Windows, and a `checksums.txt`, to the
GitHub release for the tag. The build sets `main.version` from the tag, so
`tabloio` reports the version it was released as; a development build reports
`dev`, or the module version that `go install` records.

## Dependency on tablo

`tabloio` takes every view as data from
[`tablo`](https://github.com/nbyoung/tablo) and never reads a task file itself.
The dependency follows one policy:

- **The range.** A Go `require` line states a minimum, and minimal version
  selection builds with the highest minimum any module in the build asks for,
  so the effective range runs from that minimum to the next incompatible
  version. While `tablo` is at `v0`, a minor version may break compatibility,
  so the range is one minor series, `v0.N.x`, and a bump of `N` is a deliberate
  commit that names what changed. From `v1` the range is the major version.
- **When the line enters `go.mod`.** `tablo` has no release yet, and a
  `require` on an untagged module resolves nowhere, so this bootstrap adds none.
  The first task that imports `tablo`, Renderers, adds the line with
  `go get github.com/nbyoung/tablo@v0.N.M` against tablo's first tag, and
  `go.sum` pins the hash from then on.
- **Day-to-day development.** The family builds against local checkouts through
  a personal `go.work` in the directory above the `tableaux` clone, created
  with `go work init ./tablo ./tabloio ./tablotui ./tableaud` and never
  committed; Go finds it from any directory below. `go.mod` carries no
  `replace` directive, so the committed module builds against the tagged
  `tablo` on any machine, and `GOWORK=off` builds here as CI does. The
  `tableaux` repository commits its own `go.work` over the pinned submodules
  once each bootstrap lands.
- **CI and release.** Neither checkout holds a `go.work`, so both resolve
  `tablo` from the module proxy at the required version.

## Licence

[MIT](LICENSE).
