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
tabloio context --for ben@example.org          # a contributor's neighbourhood
tabloio queue --for opus@example.org --brief   # a brief for an agent session
tabloio review 9f31 design                     # commit 'Reviewed: 9f31 design'
tabloio authorise 9f31 c07d                    # accept two proposals at once
```

## Plan

The project's plan is the Tableaux project in [`.tableaux/`](.tableaux/). The
[Tableaux tooling plan](https://github.com/nbyoung/tableaux/blob/main/PLAN.md)
in the `tableaux` repository pins this project as the submodule
`subprojects/tabloio` and tracks its root task through a recursive junction, states the review policy every task here inherits, and
proposes Go as the implementation language.

## Licence

[MIT](LICENSE).
