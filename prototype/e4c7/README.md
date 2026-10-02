# Prototype e4c7: agent briefs

Task `e4c7` Agent briefs, gate `function`.

## The question

Can a tool write the brief of one work queue item, in the six parts VIEWS.md lists, from the view data `tablo` emits, so that an agent starts from the brief alone, with the junction's `model` (F20) and the commit the agent makes?

The riskiest part is the data. The work queue item carries a kind, task, title, gate, cause, date and dependents count. The brief needs more: the gate's criteria, the resolved junction with its model and reviewer, the references, the requirement texts and the status note. The prototype joins the queue item with the task view and the gate view and shows the join suffices, except for the gaps below.

## Run

```
export PATH=$PATH:/usr/local/go/bin
go test ./prototype/e4c7
go run ./prototype/e4c7 -queue prototype/e4c7/testdata/queue-opus.json -task c07d -gate unit \
   -taskview prototype/e4c7/testdata/task-c07d.json -gateview prototype/e4c7/testdata/gate-c07d.json
```

Flags: `-queue` (work queue JSON), `-task`, `-gate` (empty matches the first item of the task), `-taskview`, `-gateview` (optional). The brief goes to stdout.

## Test data

All in `testdata/`, copied or generated from the `weather-station` corpus at `main` (commit `edb30d2`):

| File                       | Produced by                                                                        |
|----------------------------|------------------------------------------------------------------------------------|
| `queue-ada.json`           | tablo prototype 886d, `prototype/886d/output/queue-ada.json`, copied               |
| `task-<id>.json`           | tablo prototype 493e, `-view task -task <id>` (`7b2e`, `c07d`, `9f31`)             |
| `gate-<id>.json`           | tablo prototype 493e, `-view gate -task <id> -level detail`                        |
| `queue-opus.json`          | written by hand: see below                                                         |
| `brief-<id>-<gate>.md`     | this prototype's output, the golden files of the test                              |

Only `c07d` at `unit` has a `model` in the corpus (`opus@example.org`, `claude-opus-5-5`, reviewer `ben@example.org` from the assignee). Its next junction is the recursive `implementation`, so no real queue holds it, and `queue-opus.json` is a synthetic item that puts it there. It shows the model and the reviewer hand-off; the real items of `queue-ada.json` show a person's brief and a waiting item.

## What it shows

```
# Brief: c07d Node firmware at unit

## The item

- Contributor: opus@example.org, model `claude-opus-5-5`  (from `c07d`)
- Reviewer: ben@example.org  (assignee)
- Gate: 🧩 unit: All prescribed tests pass
...
## When done

Commit the work and `.tableaux/status/c07d.yaml`. Author and committer are the agent, `opus@example.org` (D3). ...

gate: design
state: nominal
reason: review
note: <one line: what is on the branch>

Model: <the identifier the harness reports, matching claude-opus-5-5>
Co-Authored-By: <your model's display name> <noreply@anthropic.com>

The reviewer merges the branch and passes the gate with a commit carrying `Reviewed: c07d unit`.
```

The commit rule follows VIEWS.md. A reviewer other than the contributor gives the hand-off: the status stays at its gate with the reason `review`, and the reviewer adds `Reviewed:`. An agent that reviews itself (junction reviewer equals contributor) records the gate and carries `Reviewed: <id> <gate>` above `Model:` (a test covers it). A person with no reviewer needs no trailer. The kinds other than work ready get their own commit: `Reaffirmed:`, `Reviewed:` and `Authorised:` as empty commits, and no commit for work waiting.

## What it leaves out

- The `Model:` value itself: the agent writes the identifier its harness reports, so the brief holds a placeholder that names the plan's model.
- The `--brief` flag, the work queue command and the Markdown renderers, which the real tasks build; this is a function over JSON.
- The glance and detail levels of the queue; the `task` parameter and the rest.
- The parent chain by title beyond the immediate parent, and the description of the task whose entry supplies the contributor (the data gives only the source's id).
- Text output, and a brief for a cross-project requirement (the code prints `url` and `commit` when present; the corpus has none in these tasks).

## Gaps in the tablo data

- The queue item names no model, reviewer, criteria, references or requirement texts. The brief needs the task view and the gate view beside the queue item, or `tablo view queue --brief` should emit them in one envelope.
- The task view gives the resolved reviewer, from the assignee for an agent, but the junction's `reviewer` field is empty and the source map says `assignee`. The prototype reads both.
- `dependents` in the task view carry `from` and `to`; the prototype says "Unblocks" for those whose `from` is the brief's gate and "Also needed by" for the rest, since VIEWS.md says "the dependents this junction unblocks" without defining the match.
- The parent chain is one level, the `parent` of the glance.

## What the design gate must decide

- The data contract: one `tablo` brief view, or the join of queue, task and gate views in the front end.
- Whether `Model:` is a placeholder in the brief, as here, or the dispatcher fills in the identifier it spawned the agent on.
- The branch name. The brief says "a branch off `main`"; the hand-written briefs name `wave/<id>`.
- The status the hand-off writes: VIEWS.md shows the gate the task stands at with `reason: review`, so the status never names the gate under review; the prototype follows it.
- The flag spelling: README.md writes `tabloio queue --for <person> --brief`, VIEWS.md `--person <p> --brief <task> <gate>`. The prototype prints the VIEWS.md form.
- Whether a brief for a person (no model) exists. The queue serves an agent as a brief, but the code renders a person's item too, without the `Model:` trailer.
- Whether the brief carries the environment rules the hand-written briefs carry (worktree, identity through `GIT_AUTHOR_*` variables, checks to run). Those belong to the project or the harness, not to the tasks, and the prototype leaves them out.
