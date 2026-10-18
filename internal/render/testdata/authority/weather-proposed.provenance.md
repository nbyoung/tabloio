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

## Deciding commits: 2

| Commit | Date | Author | Committer | Accepts by | The authority is | Tasks |
|---|---|---|---|---|---|---|
| `1b8cfb1` | 2026-09-27 | Dan Reyes, dan@example.org | Dan Reyes, dan@example.org | a change on the trunk |  | 1: `3c5d` |
| `68d9021` | 2026-09-15 | Ada Byron, ada@example.org | Ada Byron, ada@example.org | a change on the trunk | the author and the committer | 1: `a1c0` |

```
git log --first-parent --format='%H %as %ae %ce %P' edb30d2baa5ea3c7e3fb8eef8ac6cce6e145221a -- .tableaux/tasks
```

Command: `tabloio authority --ref edb30d2 --proposed --level provenance`
