# Work-blockage tree

**What waits on what?** Tableaux tooling · ref `main` at `3cdae52`, 2026-10-05 · level provenance

Legend: `tabloio gates`.

0 causes: 0 unmet requirements, 0 statuses off nominal, 0 reviews outstanding, 0 authorisations outstanding, 0 snapshots not advanced.

**No cause holds any task.**

## Next: 10 requirements not yet due

- `77b2` **tablo: backend library and plumbing** stands at 📝 defined 🟢 nominal; next 📌 mockup.
  - 🧱 implementation: requires `e3cb` Schema files at 🧱 implementation, The schema files it embeds: met, not yet due.
  - 📐 design: requires `e9c6` Abstract views at 🧱 implementation, The abstract views it derives: unmet, not yet due.
  - 📏 unit: requires `fcec` Conformance corpus at 🧱 implementation, The corpus it conforms to: met, not yet due.
- `6103` **tabloio: command line, output and input** stands at 📝 defined 🟢 nominal; next 📌 mockup.
  - 🧱 implementation: requires `77b2` tablo: backend library and plumbing at 🧱 implementation, The library and plumbing it calls: unmet, not yet due.
  - 📐 design: requires `bc86` Markdown views at 📌 mockup, The Markdown mockups it reproduces: unmet, not yet due.
- `c6e8` **tablotui: terminal user interface** stands at 📝 defined 🟢 nominal; next 📌 mockup.
  - 🧱 implementation: requires `77b2` tablo: backend library and plumbing at 🧱 implementation, The library it embeds: unmet, not yet due.
  - 📐 design: requires `bc86` Markdown views at 📌 mockup, The Markdown mockups that fix the textual layout: unmet, not yet due.
  - 📐 design: requires `5fe3` HTML views at 📌 mockup, The HTML mockups that fix the disclosure levels: unmet, not yet due.
- `595e` **tableaud: local daemon and HTML** stands at 📝 defined 🟢 nominal; next 📌 mockup.
  - 🧱 implementation: requires `77b2` tablo: backend library and plumbing at 🧱 implementation, The library it embeds: unmet, not yet due.
  - 📐 design: requires `5fe3` HTML views at 📌 mockup, The HTML mockups it reproduces: unmet, not yet due.

```
git ls-tree 3cdae52 subprojects/ # the four pins
git -C subprojects/tableaud rev-parse main # the trunk tip a pin compares with
```

Command: `tabloio blockage --ref 3cdae52 --level provenance`
