# b618: write commands prototype

## Question

Can tabloio compose a commit with the right trailers, and write a status or task file as text without reformatting the rest, in a real Git repository? SYNTAX.md "Commit trailers" prescribes `Authorised: <id>`, `Reviewed: <id> <gate>`, `Reaffirmed: <id>`, `Model:` and `Co-Authored-By:`. The status commit of a self-reviewing agent carries `Reviewed:` above `Model:` in one paragraph, and a blank line inside it leaves `Reviewed:` unparsed (F20).

## Run

```
go test -v ./prototype/b618                # every check, with the shown commits
go test -v -run Demo ./prototype/b618      # the shown commits only
```

The tests build a throwaway repository under `t.TempDir()` with `.tableaux/gates.yaml`, tasks and a status file. They read no other file and no corpus; the prototype needs no view data from tablo.

## What it shows

`Propose`, `Record`, `Review`, `Reaffirm` and `Authorise` each return a `Commit`: subject, trailers and the complete new text of each file. `Show` prints it, and `Apply` writes the files and runs `git commit`. The tests check that:

- `git interpret-trailers --parse` reads back exactly the trailers composed, in order, with `Reviewed:` above `Model:` and `Co-Authored-By:` in one final paragraph.
- `git log --grep '^Reviewed: b618 function$'` finds the commit, as the README's own query does.
- `Record` edits `gate:`, `state:`, `reason:` and `note:` lines in place. A comment, a blank line, odd spacing and an unchanged field stay byte for byte; a folded note is replaced whole; an empty reason or note removes the line; a missing field is inserted in canonical order.
- `Review`, `Reaffirm` and `Authorise` make empty commits that accept several tasks in one commit; they change no file.
- `Propose` allocates a fresh id, skipping one that a task or status file already holds, takes `order` one above the parent's largest child, and writes no `Authorised:` trailer: the task stays proposed.
- An unknown task or gate, or no task, is refused before anything is written.

Shown output (from `TestDemo`):

```
--- commit message ---
Record task b618 at the function gate

Reviewed: b618 function
Model: claude-sonnet-5-5
Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>
--- .tableaux/status/b618.yaml ---
@@ -1,7 +1,5 @@
 # kept verbatim
-gate: defined
+gate: function
 
 state:   nominal
-note: >
-  The prototype is
-  next
+note: Write commands shown
```

## Leaves out

- Command-line parsing, prompting for confirmation and the `tabloio` subcommands. The functions are the prototype of the exported package that tablotui will reuse.
- The trunk check. Authorisation counts only on the trunk, and the prototype does not check that the repository stands on it.
- Whether the committer is an authority or the reviewer, whether a gate applies to the task (a not-applicable junction), and whether the commit is the agent's or a person's. The prototype checks that the task exists and the gate is named in `gates.yaml`.
- The commit's identity: `Apply` takes it from the caller's Git configuration and environment.
- Recursive junctions (a status that holds only `gate`), junction and reference editing in `Propose`, and editing a task file.
- A status note that needs YAML block or escape forms beyond a double-quoted scalar.
- The diff uses `git diff --no-index`; the prototype has no fallback without `git`.

## For the design gate

- The exported API: `Commit` as data, with `Show` and `Apply`, or a function per command.
- The subject lines: they are the prototype's first proposal ("Accept tasks 1d2e, b618").
- Whether `propose` may add `Authorised:` when the proposer is an authority (the README's "proposal and acceptance at once"), and whether `record` adds `Reviewed:` by itself when the recorder is also the reviewer, or only on a flag as here.
- Whether `Model:` and `Co-Authored-By:` come from flags, the environment or configuration; the prototype takes them as an `Agent` value.
- Whether `reaffirm` and `review` refuse a task with no status, or one already at or past the gate.
- How `propose` picks `order` (the prototype appends) and offers references and junctions.
- Whether `Show` is the `git diff` text as here or the renderers' own layout.
