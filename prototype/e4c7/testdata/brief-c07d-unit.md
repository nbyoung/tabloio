# Brief: c07d Node firmware at unit

This brief stands alone. It covers one item of the work queue of opus@example.org on `main`: work ready.

## The item

- Contributor: opus@example.org, model `claude-opus-5-5`  (from `c07d`)
- Reviewer: ben@example.org  (assignee)
- Gate: 🧩 unit: All prescribed tests pass

The `model` is the plan's statement (F20). Run on a model that matches it, as an identifier or by prefix, and record the model you ran in the `Model:` trailer.

## The task

Reads the sensors, sleeps between readings and publishes to the gateway.

Under: Sensor node (`4e2b`)

## Requirements

Requires: `9f31` Sensor board from design to implementation, Pin map and sensor bus: unmet

Unblocks: nothing.

## Status

design nominal, 2026-09-17, ben@example.org: Sleep scheduler in progress

## When done

Commit the work and `.tableaux/status/c07d.yaml`. Author and committer are the agent, `opus@example.org` (D3). Commit on a branch off `main`; do not merge.

The reviewer `ben@example.org` accepts this gate, so record the hand-off. The status stays at the gate it stands at, with the reason `review`:

```
gate: design
state: nominal
reason: review
note: <one line: what is on the branch>
```

Trailers, last, as one final paragraph with no blank line between the lines:

```
Model: <the identifier the harness reports, matching claude-opus-5-5>
Co-Authored-By: <your model's display name> <noreply@anthropic.com>
```

The reviewer merges the branch and passes the gate with a commit carrying `Reviewed: c07d unit`.

## Reproduce

```
tabloio queue --person opus@example.org --brief c07d unit
tabloio task c07d
```
