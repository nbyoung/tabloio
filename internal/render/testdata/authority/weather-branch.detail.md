# Authority delegation

**Who may accept what?** Weather station · ref `91aa774`, off the trunk · level detail

Legend: `tabloio gates`.

```
Task                    Assignee
a1c0 Weather station    ada@example.org  proposed
  4e2b Sensor node        ben@example.org  proposed
    9f31 Sensor board       ada@example.org  proposed (differs from trunk)
    c07d Node firmware      ben@example.org  proposed
  7b2e Gateway            ada@example.org  proposed
```

5 tasks, 2 delegations, 0 authorised, 5 proposed.

## Chains and defaults: 2

### `a1c0` Weather station · ada@example.org

Authorities of its 2 children, nearest first: ada@example.org (`a1c0`).

Children: `4e2b` `7b2e`.

`a1c0` states no junction defaults.

### `4e2b` Sensor node · ben@example.org

Authorities of its 2 children, nearest first: ben@example.org (`4e2b`); then ada@example.org (`a1c0`).

Children: `9f31` `c07d`.

| Gate | Marks | Contributor | Model | Reviewer |
|---|:-:|---|---|---|
| 📌 mockup | 👀 |  |  | ben@example.org |

👀 a reviewer accepts.

Command: `tabloio authority --ref 91aa774 --level detail`
