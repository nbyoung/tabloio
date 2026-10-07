# Authority delegation

**Who may accept what?** Weather station · ref `main` at `edb30d2`, 2026-09-28 · level provenance

Legend: `tabloio gates`.

```
Task                    Assignee
a1c0 Weather station    ada@example.org
  4e2b Sensor node        ben@example.org
    9f31 Sensor board       ada@example.org
    c07d Node firmware      ben@example.org
  7b2e Gateway            ada@example.org
  3c5d Dashboard          dan@example.org  proposed
```

6 tasks, 3 delegations, 5 authorised, 1 proposed.

## Chains and defaults: 2

### `a1c0` Weather station · ada@example.org

Authorities of its 3 children, nearest first: ada@example.org (`a1c0`).

Children: `4e2b` `7b2e` `3c5d`.

`a1c0` states no junction defaults.

### `4e2b` Sensor node · ben@example.org

Authorities of its 2 children, nearest first: ben@example.org (`4e2b`); then ada@example.org (`a1c0`).

Children: `9f31` `c07d`.

| Gate | Marks | Contributor | Model | Reviewer |
|---|:-:|---|---|---|
| 📌 mockup | 👀 |  |  | ben@example.org |

👀 a reviewer accepts.

## Deciding commits: 4

| Commit | Date | Author | Committer | Accepts by | The authority is | Tasks |
|---|---|---|---|---|---|---|
| `1b8cfb1` | 2026-09-27 | dan@example.org | dan@example.org | a change on the trunk |  | 1: `3c5d` |
| `f9e8746` | 2026-09-19 | ben@example.org | ben@example.org | a merge | the author | 1: `9f31` |
| `22a758b` | 2026-09-16 | ben@example.org | ben@example.org | a change on the trunk | the author | 1: `c07d` |
| `68d9021` | 2026-09-15 | ada@example.org | ada@example.org | a change on the trunk | the author | 3: `a1c0` `4e2b` `7b2e` |

```
git log --first-parent -1 --format='%h %as %ae %ce %p' main -- .tableaux/tasks/3c5d.yaml
```

Command: `tabloio authority --ref edb30d2 --level provenance`
