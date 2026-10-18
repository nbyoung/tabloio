# Contributor work queue

**What do I do next?** Tableaux tooling · ref `corpus/tooling` · person noreply@anthropic.com · level detail

Legend: `tabloio gates`.

36 items: 0 reviews owed, 0 authorisations owed, 8 work ready, 28 reaffirmations, 0 work waiting.

| Kind | Task | Gate | Model | Since |
|---|---|---|---|---|
| Work ready | `c2ad` Roles | 🌍 validate | `claude-opus` | 2026-09-30 |
| Work ready | `e9c6` Abstract views | 🌍 validate | `claude-opus` | 2026-10-05 |
| Work ready | `e3cb` Schema files | 🌍 validate | `claude-opus` | 2026-09-30 |
| Work ready | `fcec` Conformance corpus | 📏 unit | `claude-sonnet` | 2026-09-30 |
| Work ready | `ac33` Agent identity | 🧱 implementation | `claude-sonnet` | 2026-09-30 |
| Work ready | `9f3f` Subproject linkage | 🧱 implementation | `claude-sonnet` | 2026-09-30 |
| Work ready | `7861` Productivity evidence | 🌍 validate | `claude-opus` | 2026-09-30 |
| Work ready | `7166` Language clarifications | 🧱 implementation | `claude-sonnet` | 2026-10-07 |
| Reaffirmation | `ac33` Agent identity | 📐 design |  | 2026-09-30, 17 days |
| Reaffirmation | `c2ad` Roles | 🧱 implementation |  | 2026-09-30, 17 days |
| Reaffirmation | `e3cb` Schema files | 📏 unit |  | 2026-09-30, 17 days |
| Reaffirmation | `7861` Productivity evidence | 🧱 implementation |  | 2026-09-30, 17 days |
| Reaffirmation | `fcec` Conformance corpus | 🧱 implementation |  | 2026-09-30, 17 days |
| Reaffirmation | `9f3f` Subproject linkage | 📐 design |  | 2026-09-30, 17 days |
| Reaffirmation | `99f0` Gate definition view in Markdown | 📌 mockup |  | 2026-10-05, 12 days |
| Reaffirmation | `05a9` Task definition view in Markdown | 📌 mockup |  | 2026-10-05, 12 days |
| Reaffirmation | `a9ce` Authority delegation view in Markdown | 📌 mockup |  | 2026-10-05, 12 days |
| Reaffirmation | … and 17 more: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4` | 📌 mockup |  | 2026-10-05, 12 days |
| Reaffirmation | `e9c6` Abstract views | 🧱 implementation |  | 2026-10-05, 12 days |
| Reaffirmation | `7166` Language clarifications | 📐 design |  | 2026-10-07, 10 days |

## 1. Reviews owed: none

## 2. Authorisations owed: none

## 3. Work ready: 8

**`c2ad` Roles** at 🌍 validate, model `claude-opus`

- Gate: 🌍 validate, Validation: Passes user and field tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`, from `bc63`
- Reviewer: 👀 nbyoung@nbyoung.com, from `437e`
- References: [Proposed roles](PLAN.md#roles)
- Also required by: `e9c6` Abstract views, design → design, The role names the views refer to: met
- Status: 🧱 implementation 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `3d8ce0e`: The Roles text is on the trunk and covers all seven roles; the validate gate awaits the owner
- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief c2ad validate`

**`e9c6` Abstract views** at 🌍 validate, model `claude-opus`

- Gate: 🌍 validate, Validation: Passes user and field tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`, from `bc63`
- Reviewer: 👀 nbyoung@nbyoung.com, from `437e`
- References: [Proposed views](PLAN.md#views)
- Requires: `c2ad` Roles, design → design, The role names the views refer to: met
- Also required by:
  - `99f0` Gate definition view in Markdown, design → mockup, The abstract definition of the view: met
  - `05a9` Task definition view in Markdown, design → mockup, The abstract definition of the view: met
  - `a9ce` Authority delegation view in Markdown, design → mockup, The abstract definition of the view: met
  - … and 17 more: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4`, design → mockup, The abstract definition of the view: met
  - `77b2` tablo: backend library and plumbing, implementation → design, The abstract views it derives: met
- Status: 🧱 implementation 🟢 nominal, 2026-10-05, **noreply@anthropic.com**, `ccb3947`: VIEWS.md defines the ten views on the trunk and agrees with the twenty approved mockups on the gate window; validation checks that tablo and the front ends follow the text
- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief e9c6 validate`

**`e3cb` Schema files** at 🌍 validate, model `claude-opus`

- Gate: 🌍 validate, Validation: Passes user and field tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`, from `bc63`
- Reviewer: 👀 nbyoung@nbyoung.com, from `437e`
- References: [The schemas in place](SYNTAX.md)
- Also required by:
  - `fcec` Conformance corpus, implementation → implementation, The schema files the corpus validates against: met
  - `77b2` tablo: backend library and plumbing, implementation → implementation, The schema files it embeds: met, not yet due
- Status: 📏 unit 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `5cbec7c`: schemas/check.sh (files agree with SYNTAX.md, version 0.2.1) and schemas/load.py (files load as JSON Schema 2020-12) pass; every .tableaux file validates
- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief e3cb validate`

**`fcec` Conformance corpus** at 📏 unit, model `claude-sonnet`

- Gate: 📏 unit, Unit test: All prescribed tests pass
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`, from `437e`
- Reviewer: 👀 **noreply@anthropic.com**, the assignee
- References: [The examples the corpus starts from](README.md)
- Requires: `e3cb` Schema files, implementation → implementation, The schema files the corpus validates against: met
- Also required by: `77b2` tablo: backend library and plumbing, implementation → unit, The corpus it conforms to: met, not yet due
- Status: 🧱 implementation 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `60495c2`: The corpus builds and checks clean, with 85 entries on the trunk; all three corrections from the tablo prototypes are in; the unit gate runs the tablo conformance test against it
- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief fcec unit`

**`ac33` Agent identity** at 🧱 implementation, model `claude-sonnet`

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`, from `437e`
- Reviewer: 👀 nbyoung@nbyoung.com, the assignee
- References: [Findings F1, F2, F10 and F20](PLAN.md#findings)
- Status: 📐 design 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `6762222`: The model plan-and-record rule (F20) is on the trunk; F1, F2 and F10 remain for design
- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief ac33 implementation`

**`9f3f` Subproject linkage** at 🧱 implementation, model `claude-sonnet`

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`, from `437e`
- Reviewer: 👀 nbyoung@nbyoung.com, the assignee
- References: [Findings F3 to F5](PLAN.md#findings)
- Status: 📐 design 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `43d68e4`: The subproject linkage rule (F3, F5, F12) is on the trunk at language 0.3.0; tablo and the subprojects follow
- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief 9f3f implementation`

**`7861` Productivity evidence** at 🌍 validate, model `claude-opus`

- Gate: 🌍 validate, Validation: Passes user and field tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`, from `bc63`
- Reviewer: 👀 nbyoung@nbyoung.com, from `437e`
- References: [The review policy under test](PLAN.md#review-policy)
- Status: 🧱 implementation 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `c6f2655`: The full report, its 23 tests and an example over five repositories are on the trunk
- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief 7861 validate`

**`7166` Language clarifications** at 🧱 implementation, model `claude-sonnet`

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`, from `437e`
- Reviewer: 👀 nbyoung@nbyoung.com, the assignee
- References: [Findings F7 to F9 and F19](PLAN.md#findings)
- Status: 📐 design 🟢 nominal, 2026-10-07, **noreply@anthropic.com**, `56103af`: The trunk clarification (F19), with the order for a linked repository, and the exemption rule (F7) are on the trunk; F9 remains for design
- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief 7166 implementation`

## 4. Reaffirmations: 28

**`ac33` Agent identity**, 📐 design, 2026-09-30, 17 days

- Gate: 📐 design, Design: A model and sufficient tests exist
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`
- Reviewer: 👀 nbyoung@nbyoung.com, the assignee
- References: [Findings F1, F2, F10 and F20](PLAN.md#findings)
- Status: 📐 design 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `6762222`: The model plan-and-record rule (F20) is on the trunk; F1, F2 and F10 remain for design
- To reaffirm: `tabloio reaffirm ac33`

**`c2ad` Roles**, 🧱 implementation, 2026-09-30, 17 days

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`
- Reviewer: 👀 nbyoung@nbyoung.com
- References: [Proposed roles](PLAN.md#roles)
- Also required by: `e9c6` Abstract views, design → design, The role names the views refer to: met
- Status: 🧱 implementation 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `3d8ce0e`: The Roles text is on the trunk and covers all seven roles; the validate gate awaits the owner
- To reaffirm: `tabloio reaffirm c2ad`

**`e3cb` Schema files**, 📏 unit, 2026-09-30, 17 days

- Gate: 📏 unit, Unit test: All prescribed tests pass
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`
- Reviewer: 👀 nbyoung@nbyoung.com
- References: [The schemas in place](SYNTAX.md)
- Also required by:
  - `fcec` Conformance corpus, implementation → implementation, The schema files the corpus validates against: met
  - `77b2` tablo: backend library and plumbing, implementation → implementation, The schema files it embeds: met, not yet due
- Status: 📏 unit 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `5cbec7c`: schemas/check.sh (files agree with SYNTAX.md, version 0.2.1) and schemas/load.py (files load as JSON Schema 2020-12) pass; every .tableaux file validates
- To reaffirm: `tabloio reaffirm e3cb`

**`7861` Productivity evidence**, 🧱 implementation, 2026-09-30, 17 days

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`
- Reviewer: 👀 nbyoung@nbyoung.com
- References: [The review policy under test](PLAN.md#review-policy)
- Status: 🧱 implementation 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `c6f2655`: The full report, its 23 tests and an example over five repositories are on the trunk
- To reaffirm: `tabloio reaffirm 7861`

**`fcec` Conformance corpus**, 🧱 implementation, 2026-09-30, 17 days

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`
- Reviewer: 👀 **noreply@anthropic.com**, the assignee
- References: [The examples the corpus starts from](README.md)
- Requires: `e3cb` Schema files, implementation → implementation, The schema files the corpus validates against: met
- Also required by: `77b2` tablo: backend library and plumbing, implementation → unit, The corpus it conforms to: met, not yet due
- Status: 🧱 implementation 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `60495c2`: The corpus builds and checks clean, with 85 entries on the trunk; all three corrections from the tablo prototypes are in; the unit gate runs the tablo conformance test against it
- To reaffirm: `tabloio reaffirm fcec`

**`9f3f` Subproject linkage**, 📐 design, 2026-09-30, 17 days

- Gate: 📐 design, Design: A model and sufficient tests exist
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`
- Reviewer: 👀 nbyoung@nbyoung.com, the assignee
- References: [Findings F3 to F5](PLAN.md#findings)
- Status: 📐 design 🟢 nominal, 2026-09-30, **noreply@anthropic.com**, `43d68e4`: The subproject linkage rule (F3, F5, F12) is on the trunk at language 0.3.0; tablo and the subprojects follow
- To reaffirm: `tabloio reaffirm 9f3f`

**`99f0` Gate definition view in Markdown**, 📌 mockup, 2026-10-05, 12 days

- Gate: 📌 mockup, Mockup: A non-technical mockup of the outcome exists
- References: [The mockup](docs/mockups/gates.md)
- Requires: `e9c6` Abstract views, design → mockup, The abstract definition of the view: met
- Status: 📌 mockup ✅ complete, 2026-10-05, **noreply@anthropic.com**, `266cb8b`: The owner accepts docs/mockups/gates.md at the mockup gate, the last gate that applies to this task
- To reaffirm: `tabloio reaffirm 99f0`

**`05a9` Task definition view in Markdown**, 📌 mockup, 2026-10-05, 12 days

- Gate: 📌 mockup, Mockup: A non-technical mockup of the outcome exists
- References: [The mockup](docs/mockups/task.md)
- Requires: `e9c6` Abstract views, design → mockup, The abstract definition of the view: met
- Status: 📌 mockup ✅ complete, 2026-10-05, **noreply@anthropic.com**, `266cb8b`: The owner accepts docs/mockups/task.md at the mockup gate, the last gate that applies to this task
- To reaffirm: `tabloio reaffirm 05a9`

**`a9ce` Authority delegation view in Markdown**, 📌 mockup, 2026-10-05, 12 days

- Gate: 📌 mockup, Mockup: A non-technical mockup of the outcome exists
- References: [The mockup](docs/mockups/authority.md)
- Requires: `e9c6` Abstract views, design → mockup, The abstract definition of the view: met
- Status: 📌 mockup ✅ complete, 2026-10-05, **noreply@anthropic.com**, `266cb8b`: The owner accepts docs/mockups/authority.md at the mockup gate, the last gate that applies to this task
- To reaffirm: `tabloio reaffirm a9ce`

**… and 17 more**, 📌 mockup, 2026-10-05, 12 days: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4`.

**`e9c6` Abstract views**, 🧱 implementation, 2026-10-05, 12 days

- Gate: 🧱 implementation, Implementation: Artifacts suffice for unit, integration and validation tests
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`
- Reviewer: 👀 nbyoung@nbyoung.com
- References: [Proposed views](PLAN.md#views)
- Requires: `c2ad` Roles, design → design, The role names the views refer to: met
- Also required by:
  - `99f0` Gate definition view in Markdown, design → mockup, The abstract definition of the view: met
  - `05a9` Task definition view in Markdown, design → mockup, The abstract definition of the view: met
  - `a9ce` Authority delegation view in Markdown, design → mockup, The abstract definition of the view: met
  - … and 17 more: `bb7c` `c74a` `efff` `ab8e` `c545` `d615` `b14e` `a6f7` `7dff` `b1b7` `ea51` `7783` `5471` `32e7` `69eb` `3194` `a8b4`, design → mockup, The abstract definition of the view: met
  - `77b2` tablo: backend library and plumbing, implementation → design, The abstract views it derives: met
- Status: 🧱 implementation 🟢 nominal, 2026-10-05, **noreply@anthropic.com**, `ccb3947`: VIEWS.md defines the ten views on the trunk and agrees with the twenty approved mockups on the gate window; validation checks that tablo and the front ends follow the text
- To reaffirm: `tabloio reaffirm e9c6`

**`7166` Language clarifications**, 📐 design, 2026-10-07, 10 days

- Gate: 📐 design, Design: A model and sufficient tests exist
- Contributor: 🤖 **noreply@anthropic.com**, model `claude-sonnet`
- Reviewer: 👀 nbyoung@nbyoung.com, the assignee
- References: [Findings F7 to F9 and F19](PLAN.md#findings)
- Status: 📐 design 🟢 nominal, 2026-10-07, **noreply@anthropic.com**, `56103af`: The trunk clarification (F19), with the order for a linked repository, and the exemption rule (F7) are on the trunk; F9 remains for design
- To reaffirm: `tabloio reaffirm 7166`

## 5. Work waiting: none

Command: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --level detail`
