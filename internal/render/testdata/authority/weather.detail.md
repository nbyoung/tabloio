# Authority delegation

**Who may accept what?** Weather station · ref `main` · level detail

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

Command: `tabloio authority --ref main --level detail`
