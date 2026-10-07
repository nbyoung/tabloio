# Contributor work queue

**What do I do next?** Tableaux tooling · ref `main` · person noreply@anthropic.com · level detail

Legend: `tabloio gates`.

60 items: 0 reviews owed, 0 authorisations owed, 28 work ready, 32 reaffirmations, 0 work waiting.

| Kind | Task | Gate | Model | Since |
|---|---|---|---|---|
| Work ready | `c2ad` Roles | 🌍 validate | `claude-opus` | 2026-09-30 |
| Work ready | `e9c6` Abstract views | 🧱 implementation | `claude-sonnet` | 2026-09-30 |
| Work ready | `99f0` Gate definition view in Markdown | 📌 mockup | `claude-opus` | 2026-09-29 |
| Work ready | `05a9` Task definition view in Markdown | 📌 mockup | `claude-opus` | 2026-09-29 |
| Work ready | `a9ce` Authority delegation view in Markdown | 📌 mockup | `claude-opus` | 2026-09-29 |
| Work ready | … and 17 more: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` | 📌 mockup | `claude-opus` | 2026-09-29 |
| Work ready | `e3cb` Schema files | 🌍 validate | `claude-opus` | 2026-09-30 |
| Work ready | `fcec` Conformance corpus | 📏 unit | `claude-sonnet` | 2026-09-30 |
| Work ready | `ac33` Agent identity | 🧱 implementation | `claude-sonnet` | 2026-09-30 |
| Work ready | `9f3f` Subproject linkage | 🧱 implementation | `claude-sonnet` | 2026-09-30 |
| Work ready | `7861` Productivity evidence | 🌍 validate | `claude-opus` | 2026-09-30 |
| Work ready | `7166` Language clarifications | 🧱 implementation | `claude-sonnet` | 2026-09-30 |
| Reaffirmation | `99f0` Gate definition view in Markdown | 📝 defined |  | 2026-09-29, 6 days |
| Reaffirmation | `05a9` Task definition view in Markdown | 📝 defined |  | 2026-09-29, 6 days |
| Reaffirmation | `a9ce` Authority delegation view in Markdown | 📝 defined |  | 2026-09-29, 6 days |
| Reaffirmation | … and 17 more: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` | 📝 defined |  | 2026-09-29, 6 days |
| Reaffirmation | `7166` Language clarifications | 📐 design |  | 2026-09-30, 5 days |
| Reaffirmation | `ac33` Agent identity | 📐 design |  | 2026-09-30, 5 days |
| Reaffirmation | `c2ad` Roles | 🧱 implementation |  | 2026-09-30, 5 days |
| Reaffirmation | `e3cb` Schema files | 📏 unit |  | 2026-09-30, 5 days |
| Reaffirmation | `7861` Productivity evidence | 🧱 implementation |  | 2026-09-30, 5 days |
| Reaffirmation | `fcec` Conformance corpus | 🧱 implementation |  | 2026-09-30, 5 days |
| Reaffirmation | `9f3f` Subproject linkage | 📐 design |  | 2026-09-30, 5 days |
| Reaffirmation | `77b2` tablo: backend library and plumbing | 📝 defined |  | 2026-09-30, 5 days |
| Reaffirmation | `6103` tabloio: command line, output and input | 📝 defined |  | 2026-09-30, 5 days |
| Reaffirmation | `c6e8` tablotui: terminal user interface | 📝 defined |  | 2026-09-30, 5 days |
| Reaffirmation | `595e` tableaud: local daemon and HTML | 📝 defined |  | 2026-09-30, 5 days |
| Reaffirmation | `e9c6` Abstract views | 📐 design |  | 2026-09-30, 5 days |

## 1. Reviews owed: none

## 2. Authorisations owed: none

## 3. Work ready: 28

**`c2ad` Roles** at 🌍 validate, model `claude-opus`

- Gate: 🌍 validate, Validation: Passes user and field tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`, from `bc63`
- Reviewer: 👀 nbyoung@nbyoung.com, from `437e`
- References: [Proposed roles](PLAN.md#roles)
- Also required by: `e9c6` Abstract views, design → design, The role names the views refer to: met
- Status: 🧱 implementation 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `3d8ce0e`: The Roles text is on the trunk and covers all seven roles; the validate gate awaits the owner
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief c2ad validate`

**`e9c6` Abstract views** at 🧱 implementation, model `claude-sonnet`

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`, from `437e`
- Reviewer: 👀 **noreply@anthropic.com**, the assignee
- References: [Proposed views](PLAN.md#views)
- Requires: `c2ad` Roles, design → design, The role names the views refer to: met
- Unblocks: `77b2` tablo: backend library and plumbing, implementation → design, The abstract views it derives: unmet, not yet due
- Status: 📐 design 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `879b447`: VIEWS.md is on the trunk at language 0.3.1; the twenty mockups may start, and implementation puts the text in place
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief e9c6 implementation`

**`99f0` Gate definition view in Markdown** at 📌 mockup, model `claude-opus`

- Gate: 📌 mockup, Mockup: A non-technical mockup of the outcome exists
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`, from `437e`
- Reviewer: 👀 nbyoung@nbyoung.com, from `437e`
- References: [The mockup](docs/mockups/gates.md)
- Requires: `e9c6` Abstract views, design → mockup, The abstract definition of the view: met
- Its parent unblocks:
  - `bc86` Markdown views at 📌 mockup is required by `6103` tabloio: command line, output and input, mockup → design, The Markdown mockups it reproduces: unmet, not yet due
  - `bc86` Markdown views at 📌 mockup is required by `c6e8` tablotui: terminal user interface, mockup → design, The Markdown mockups that fix the textual layout: unmet, not yet due
- Status: 📝 defined 🟢 nominal, 2026-09-29, **noreply@anthropic.com**, `1a17bfc`: Mockup waits for Abstract views (e9c6) at design
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief 99f0 mockup`

**`05a9` Task definition view in Markdown** at 📌 mockup, model `claude-opus`

- Gate: 📌 mockup, Mockup: A non-technical mockup of the outcome exists
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief 05a9 mockup`

**`a9ce` Authority delegation view in Markdown** at 📌 mockup, model `claude-opus`

- Gate: 📌 mockup, Mockup: A non-technical mockup of the outcome exists
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief a9ce mockup`

**… and 17 more** at 📌 mockup, model `claude-opus`: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4`.

**`e3cb` Schema files** at 🌍 validate, model `claude-opus`

- Gate: 🌍 validate, Validation: Passes user and field tests
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief e3cb validate`

**`fcec` Conformance corpus** at 📏 unit, model `claude-sonnet`

- Gate: 📏 unit, Unit test: All prescribed tests pass
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief fcec unit`

**`ac33` Agent identity** at 🧱 implementation, model `claude-sonnet`

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief ac33 implementation`

**`9f3f` Subproject linkage** at 🧱 implementation, model `claude-sonnet`

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief 9f3f implementation`

**`7861` Productivity evidence** at 🌍 validate, model `claude-opus`

- Gate: 🌍 validate, Validation: Passes user and field tests
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief 7861 validate`

**`7166` Language clarifications** at 🧱 implementation, model `claude-sonnet`

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Brief: `tabloio queue --person noreply@anthropic.com --ref main --brief 7166 implementation`

## 4. Reaffirmations: 32

**`99f0` Gate definition view in Markdown**, 📝 defined, 2026-09-29, 6 days

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- To reaffirm: `tabloio reaffirm 99f0`

**`05a9` Task definition view in Markdown**, 📝 defined, 2026-09-29, 6 days

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- To reaffirm: `tabloio reaffirm 05a9`

**`a9ce` Authority delegation view in Markdown**, 📝 defined, 2026-09-29, 6 days

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- To reaffirm: `tabloio reaffirm a9ce`

**… and 17 more**, 📝 defined, 2026-09-29, 6 days: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4`.

**`7166` Language clarifications**, 📐 design, 2026-09-30, 5 days

- Gate: 📐 design, Design: A model and sufficient tests exist
- To reaffirm: `tabloio reaffirm 7166`

**`ac33` Agent identity**, 📐 design, 2026-09-30, 5 days

- Gate: 📐 design, Design: A model and sufficient tests exist
- To reaffirm: `tabloio reaffirm ac33`

**`c2ad` Roles**, 🧱 implementation, 2026-09-30, 5 days

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- To reaffirm: `tabloio reaffirm c2ad`

**`e3cb` Schema files**, 📏 unit, 2026-09-30, 5 days

- Gate: 📏 unit, Unit test: All prescribed tests pass
- To reaffirm: `tabloio reaffirm e3cb`

**`7861` Productivity evidence**, 🧱 implementation, 2026-09-30, 5 days

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- To reaffirm: `tabloio reaffirm 7861`

**`fcec` Conformance corpus**, 🧱 implementation, 2026-09-30, 5 days

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- To reaffirm: `tabloio reaffirm fcec`

**`9f3f` Subproject linkage**, 📐 design, 2026-09-30, 5 days

- Gate: 📐 design, Design: A model and sufficient tests exist
- To reaffirm: `tabloio reaffirm 9f3f`

**`77b2` tablo: backend library and plumbing**, 📝 defined, 2026-09-30, 5 days

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- To reaffirm: `tabloio reaffirm 77b2`

**`6103` tabloio: command line, output and input**, 📝 defined, 2026-09-30, 5 days

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- To reaffirm: `tabloio reaffirm 6103`

**`c6e8` tablotui: terminal user interface**, 📝 defined, 2026-09-30, 5 days

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- To reaffirm: `tabloio reaffirm c6e8`

**`595e` tableaud: local daemon and HTML**, 📝 defined, 2026-09-30, 5 days

- Gate: 📝 defined, Defined: Title, description, assignee and references exist
- To reaffirm: `tabloio reaffirm 595e`

**`e9c6` Abstract views**, 📐 design, 2026-09-30, 5 days

- Gate: 📐 design, Design: A model and sufficient tests exist
- To reaffirm: `tabloio reaffirm e9c6`

## 5. Work waiting: none

Command: `tabloio queue --person noreply@anthropic.com --ref main --level detail`
