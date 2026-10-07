# Work-blockage tree

**What waits on what?** Weather station · ref `main` · level detail

Legend: `tabloio gates`.

3 causes: 1 unmet requirement, 1 status off nominal, 0 reviews outstanding, 1 authorisation outstanding, 0 snapshots not advanced.

| Cause | Who acts | Holds |
|---|---|--:|
| `9f31` Sensor board stands at ⚙️ function 🔴 stalled ⛔ blocked: Barometer ICs on 14-week backorder | 🧑 ada@example.org records | 2 |
| `9f31` Sensor board has not passed 📐 design | 🧑 ada@example.org contributes | 1 |
| `3c5d` Dashboard is proposed | 🧑 ada@example.org authorises | 1 |

1. **`9f31` Sensor board stands at ⚙️ function 🔴 stalled ⛔ blocked: Barometer ICs on 14-week backorder** · 🧑 ada@example.org records · holds 2
   - Action: ada@example.org clears the `blocked` reason, or records a new status.
   - `9f31` Sensor board at ⚡ performance: the cause holds the task itself.
   - `c07d` Node firmware at 🛠️ implementation: requires `9f31` at 📐 design, Sensor interface.
2. **`9f31` Sensor board has not passed 📐 design** · 🧑 ada@example.org contributes · holds 1
   - Action: ada@example.org records `9f31` through design.
   - `c07d` Node firmware at 🛠️ implementation: requires `9f31` at 📐 design, Sensor interface.
3. **`3c5d` Dashboard is proposed** · 🧑 ada@example.org authorises · holds 1
   - Action: ada@example.org commits the trailer `Authorised: 3c5d`.
   - `3c5d` Dashboard at 📌 mockup: the cause holds the task itself.

## Next: 1 requirement not yet due

- `3c5d` **Dashboard** stands at 📝 defined 🟢 nominal; next 📌 mockup.
  - 🖼️ integrate: requires `7b2e` Gateway at ⚙️ function, Readings API: unmet, not yet due.

Command: `tabloio blockage --ref main --level detail`
