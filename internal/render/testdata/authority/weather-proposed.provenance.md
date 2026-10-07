# Authority delegation

**Who may accept what?** Weather station · ref `main` at `edb30d2`, 2026-09-28 · proposed only · level provenance

Legend: `tabloio gates`.

```
Task                  Assignee
a1c0 Weather station  ada@example.org
  3c5d Dashboard        dan@example.org  proposed
```

2 tasks, 1 delegations, 1 authorised, 1 proposed.

## Chains and defaults: 1

### `a1c0` Weather station · ada@example.org

Authorities of its 3 children, nearest first: ada@example.org (`a1c0`).

Children: `4e2b` `7b2e` `3c5d`.

`a1c0` states no junction defaults.

## Deciding commits: 1

| Commit | Date | Author | Committer | Accepts by | The authority is | Tasks |
|---|---|---|---|---|---|---|
| `1b8cfb1` | 2026-09-27 | dan@example.org | dan@example.org | a change on the trunk |  | 1: `3c5d` |

```
git log --first-parent -1 --format='%h %as %ae %ce %p' main -- .tableaux/tasks/3c5d.yaml
```

Command: `tabloio authority --ref edb30d2 --proposed --level provenance`
