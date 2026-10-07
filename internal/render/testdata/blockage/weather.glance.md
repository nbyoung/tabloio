# Work-blockage tree

**What waits on what?** Weather station · ref `main` · level glance

Legend: `tabloio gates`.

3 causes: 1 unmet requirement, 1 status off nominal, 0 reviews outstanding, 1 authorisation outstanding, 0 snapshots not advanced.

| Cause | Who acts | Holds |
|---|---|--:|
| `9f31` Sensor board stands at ⚙️ function 🔴 stalled ⛔ blocked: Barometer ICs on 14-week backorder | 🧑 ada@example.org records | 2 |
| `9f31` Sensor board has not passed 📐 design | 🧑 ada@example.org contributes | 1 |
| `3c5d` Dashboard is proposed | 🧑 ada@example.org authorises | 1 |

Next: 1 requirement is not yet due, 1 of them unmet.

Command: `tabloio blockage --ref main --level glance`
