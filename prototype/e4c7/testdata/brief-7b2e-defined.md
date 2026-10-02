# Brief: 7b2e Gateway at defined

This brief stands alone. It covers one item of the work queue of ada@example.org on `main`: work ready.

## The item

- Contributor: ada@example.org  (default)
- Reviewer: none
- Gate: 📝 defined: Title, description, assignee and references exist

## The task

A mains-powered box indoors that receives the node's readings over LoRa and stores them in a database the dashboard reads.

References:

- Gateway overview, docs/overview.md#gateway

Under: Weather station (`a1c0`)

## Requirements

Requires: none.

Also needed by: `3c5d` Dashboard from function to integrate, Readings API

## Status

undefined undefined, 2026-09-15, ada@example.org

## When done

Commit the work and `.tableaux/status/7b2e.yaml`. Author and committer are the contributor, `ada@example.org` (D3). Commit on a branch off `main`; do not merge.

No one reviews this junction:

```
gate: defined
state: nominal
note: <one line: what is on the branch>
```

Authorisation stands as the review of `defined`, so the commit needs no `Reviewed:` trailer.

The contributor is a person, so the commit needs no trailer.

## Reproduce

```
tabloio queue --person ada@example.org --brief 7b2e defined
tabloio task 7b2e
```
