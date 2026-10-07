# Authority delegation

**Who may accept what?** Weather station · ref `main` · proposed only · level detail

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

Command: `tabloio authority --ref main --proposed --level detail`
