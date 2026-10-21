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

- **Output** any view in one of two textual formats. `--markdown`, the default, reads in a plain renderer: a `TABLEAUX.md` that CI regenerates, an attachment to email or chat, a brief pasted into an agent's session. `--text` draws the view with Unicode box characters for a terminal or a log with no Markdown viewer. Scripts that want the view data call `tablo <view>` for JSON or YAML.
- **Input** to the method: `propose`, `record`, `review`, `reaffirm` and `authorise` write the right file or compose the commit with the right trailers, and show the commit before making it.

```
tabloio tableau > TABLEAUX.md                  # the global tableau in Markdown
tabloio tableau --text                         # the same drawn in Unicode
tabloio context --person ben@example.org        # a contributor's neighbourhood
tabloio queue --person opus@example.org --ref 4f2a9c1 --brief 9f31 design  # a brief for an agent session
tabloio propose --parent 07e0 --title 'Rain gauge' --description '…'  # add a task file
tabloio record 9f31 --gate design --state nominal --model claude-opus-5-5 --yes  # record a status
tabloio review 9f31 design                     # commit 'Reviewed: 9f31 design'
tabloio reaffirm 9f31                          # commit 'Reaffirmed: 9f31'
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
| `cmd/tabloio/`        | The command. `main.go` builds the environment, with tablo's library as the source of view data, and calls `internal/cli`; it holds the version |
| `internal/cli/`       | The command line: the parser, the command table, `help` and `version`, the read commands, one per view, and the five write commands |
| `action/`             | The write actions: each plans one commit as data, shows it, and makes it; the facts come from tablo's library and the bytes from Git |
| `brief/`              | The brief composer: one work item of the queue as plain text, for an agent's session    |
| `internal/render/`    | The renderers from tablo's view data: Markdown and Unicode text                         |
| `.github/workflows/`  | `ci.yml` checks every pull request and push to `main`; `release.yml` publishes a tag    |
| `.goreleaser.yaml`    | The release build: six static binaries, archives and checksums                          |
| `.golangci.yml`       | The lint configuration                                                                  |
| `.tableaux/`          | The plan                                                                                |

Packages under `internal/` stay private to this module. Two packages stand
outside it for the other front ends: `brief`, which `tableaud` imports to print
the same brief text, and `action`, the plans and commits behind the write
commands, which `tablotui` imports to make the same commits.

A prototype under `prototype/` is the function gate's demonstration. It goes when its task records `implementation`: the design's account of what it kept from the prototype and `git log -- prototype/<id>` keep what it showed, and the trunk builds what it ships.

## Snapshot workflow

A project that keeps its plan in `.tableaux` validates it on every push and
regenerates `TABLEAUX.md` on its trunk with the composite action in
[`snapshot/`](snapshot/), and validates the staged plan before each commit
with the hook in [`hooks/`](hooks/). Both run the released `tablo` and
`tabloio` binaries and need no Go.

**The workflow.** The project adds `.github/workflows/tableaux.yml`, as
[`snapshot/consumer.yml`](snapshot/consumer.yml) gives it:

```yaml
name: Tableaux
on:
  push:
  pull_request:
permissions:
  contents: write
concurrency:
  group: tableaux-${{ github.ref }}
jobs:
  snapshot:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
        with:
          fetch-depth: 0          # history: authorisation, reviews, dates
          submodules: recursive   # pins: recursive junctions read subprojects
      - uses: nbyoung/tabloio/snapshot@v0.2.0
```

Every push and pull request runs `tablo validate` and annotates each
diagnostic at its file and line; an error fails the job. A push to the trunk
then renders `tabloio tableau --level glance` into `TABLEAUX.md` and commits it
as `github-actions[bot]` when it changed. The inputs `project`, `output`,
`view`, `trunk`, `tabloio-version` and `tablo-version` change the defaults;
[`snapshot/action.yml`](snapshot/action.yml) describes each.

**A protected trunk.** A trunk that requires pull requests rejects a push
made with `GITHUB_TOKEN`. The project then gives the checkout step a
credential that the trunk's ruleset lets bypass, and the action stays as it
is, since it pushes with whatever credential the checkout leaves. A
repository admin stores the credential once as a secret and adds its actor to
the ruleset's bypass list once. Either of two serves:

```yaml
      - uses: actions/checkout@v5
        with:
          fetch-depth: 0
          submodules: recursive
          token: ${{ steps.app.outputs.token }}       # a GitHub App installation token, or
          # ssh-key: ${{ secrets.TABLEAUX_DEPLOY_KEY }} # a deploy key with write access
```

A GitHub App serves many repositories of one organisation, and an earlier
step, such as `actions/create-github-app-token`, mints its token; a deploy
key serves one repository and needs no app. A ruleset's bypass list takes
both: GitHub Apps as its documentation of rulesets lists them, and deploy
keys as the actor type `DeployKey` of its REST API for rulesets.

**The hook.** `hooks/pre-commit` validates the project in the staged tree, so
an unstaged edit neither passes nor fails a commit, and a commit that stages
nothing under `.tableaux` skips it. It needs `tablo` on the path. Each
contributor copies it once per clone; each release archive holds it too:

```
curl -fsSL -o .git/hooks/pre-commit https://raw.githubusercontent.com/nbyoung/tabloio/v0.2.0/hooks/pre-commit
chmod +x .git/hooks/pre-commit
git config tableaux.project sub/dir   # only when .tableaux is not at the root
```

`git commit --no-verify` bypasses it once. `snapshot/test.sh` tests the
action's steps and the hook on stubs; CI runs it beside `go test`.
`snapshot/test.sh --real` installs the released binaries from GitHub instead,
reruns four of the cases with them on a corpus entry, and times the hook on
the Tableaux plan; CI does not run it.

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

`ci.yml` runs the same four checks, gofmt, vet, lint and test, then the
snapshot harness `snapshot/test.sh` and actionlint over the snapshot action
and its consumer example, on every pull request and every push to `main`,
with the Go version from `go.mod` and golangci-lint from its official
action. Its job `corpus` builds the tableaux corpus at the umbrella commit
that `.github/corpus.lock.yaml` names, through `.github/corpus.sh fetch`, and
runs `go test ./...` against it, so the tests that read the corpus run in CI
too.

## Release

A release is a tag on `main`:

```
git tag v0.2.0
git push origin v0.2.0
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
- **When the line enters `go.mod`.** The Read commands task adds the line,
  since `internal/cli` is the first package here to import `tablo`; it names
  the tag `v0.2.0`, the first tablo that reads language 0.4.0, and `go.sum`
  pins the hash.
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
