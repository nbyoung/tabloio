package render

import (
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/view"
)

var detailProvenance = []view.Level{view.Detail, view.Provenance}

// auditGoldens lists the golden cases of the audit: T8 to T11.
var auditGoldens = []golden{
	// T8 and T9
	{view: "audit", name: "tooling", levels: allLevels},
	{view: "audit", name: "tooling-nbyoung", levels: glanceDetail},
	// T10
	{view: "audit", name: "corpus-h2", levels: detailProvenance},
	{view: "audit", name: "corpus-h3", levels: detailProvenance},
	{view: "audit", name: "corpus-h4", levels: detailProvenance},
	{view: "audit", name: "corpus-h5", levels: detailProvenance},
	{view: "audit", name: "corpus-h6", levels: detailProvenance},
	{view: "audit", name: "weather", levels: allLevels},
	// T11
	{view: "audit", name: "empty", levels: allLevels},
}

func TestAuditGolden(t *testing.T) {
	for _, g := range auditGoldens {
		for _, level := range g.levels {
			t.Run(g.name+"."+level.String(), func(t *testing.T) {
				checkGolden(t, g.view, g.name, level, Options{})
			})
		}
	}
}

// T12: the glance merge. Findings of one rule and severity make one row, each
// task and each resolver once, in the data's order; a task in two findings
// counts in each; findings with no rule merge only with their own kind.
func TestAuditMerge(t *testing.T) {
	ref := func(ids ...string) []view.TaskRef {
		var out []view.TaskRef
		for _, id := range ids {
			out = append(out, view.TaskRef{ID: id})
		}
		return out
	}
	v := &view.Audit{}
	v.Level = view.Detail
	v.Project = view.Project{Root: "437e", Title: "Tableaux tooling"}
	v.Source = view.Source{Ref: "main", OnTrunk: true}
	v.Params.Stale = 7
	v.Findings = []view.Finding{
		{Rule: "H3", Severity: "warning", Kind: "First kind", Tasks: ref("aaaa", "bbbb"), Resolver: "x@example.org"},
		{Rule: "H3", Severity: "warning", Kind: "Other kind", Tasks: ref("bbbb", "cccc"), Resolver: "y@example.org"},
		{Rule: "H3", Severity: "warning", Tasks: ref("cccc"), Resolver: "x@example.org"},
		{Rule: "H3", Severity: "information", Kind: "Quiet", Tasks: ref("dddd"), Resolver: "x@example.org"},
		{Severity: "warning", Kind: "A proposed task", Tasks: ref("eeee")},
		{Severity: "warning", Kind: "A stale status", Tasks: ref("ffff")},
		{Severity: "warning", Kind: "A proposed task", Tasks: ref("gggg")},
	}
	v.Warnings, v.Information = 7, 1
	rows := auditRows(v.Findings)
	if len(rows) != 4 {
		t.Fatalf("%d rows, want 4", len(rows))
	}
	r := rows[0]
	if r.findings != 5 || r.kind != "First kind" || len(r.tasks) != 3 || len(r.resolvers) != 2 ||
		r.resolvers[0] != "x@example.org" || r.resolvers[1] != "y@example.org" {
		t.Errorf("the merged row is %+v", r)
	}
	out := renderString(t, v, Options{Level: view.Glance})
	for _, want := range []string{
		"| warning | H3 | First kind | 5 | `aaaa` `bbbb` `cccc` | x@example.org, y@example.org |\n",
		"| information | H3 | Quiet | 1 | `dddd` | x@example.org |\n",
		"| warning |  | A proposed task | 2 | `eeee` `gggg` |  |\n",
		"| warning |  | A stale status | 1 | `ffff` |  |\n",
		"**0 errors, 7 warnings, 1 item of information**: 8 findings.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Most to resolve") {
		t.Errorf("the line of the most to resolve prints with no one named:\n%s", out)
	}
	detail := renderString(t, v, Options{Level: view.Detail})
	if !strings.Contains(detail, ": 8 findings in 7 rows.\n") {
		t.Errorf("the detail count line lacks the rows:\n%s", detail)
	}
}

// TestAuditCounts covers the count line's singular forms and the empty form.
func TestAuditCounts(t *testing.T) {
	v := &view.Audit{}
	v.Level = view.Detail
	v.Source = view.Source{Ref: "main", OnTrunk: true}
	v.Params.Stale = 7
	v.Errors, v.Warnings, v.Information = 1, 1, 1
	v.Findings = []view.Finding{{Rule: "S11", Severity: "error", Tasks: []view.TaskRef{{ID: "aaaa"}}}}
	v.Most = &view.Resolving{Email: "a@example.org", Count: 1}
	out := renderString(t, v, Options{Level: view.Detail})
	if want := "**1 error, 1 warning, 1 item of information**: 3 findings in 1 row.\n"; !strings.Contains(out, want) {
		t.Errorf("want %q:\n%s", want, out)
	}
	if want := "**Most to resolve:** a@example.org, 1 of 3.\n"; !strings.Contains(out, want) {
		t.Errorf("want %q:\n%s", want, out)
	}
	v.Findings, v.Errors, v.Warnings, v.Information = nil, 0, 0, 0
	out = renderString(t, v, Options{Level: view.Detail})
	for _, want := range []string{"**0 errors, 0 warnings, 0 items of information**: 0 findings.\n", "\n**No finding.**\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("want %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Most to resolve") || strings.Contains(out, "| Rule") {
		t.Errorf("an empty audit prints a table or the line of the most to resolve:\n%s", out)
	}
}

// TestAuditWords: the kind, the message and the sentence print as tablo
// words them, split at backticks and no further; the action and the facts
// are the builder's words of the act and of the typed facts.
func TestAuditWords(t *testing.T) {
	v := readIn(t, "audit", "tooling").(*view.Audit)
	v.Findings = v.Findings[:1]
	f := &v.Findings[0]
	f.Kind, f.Message, f.Sentence = "A `kind` of *finding*", "A `message` with a | bar", "The `sentence`."
	f.Act = "await"
	f.Facts = []view.FindingFacts{{Task: "e3cb", Requirement: &view.Requirement{
		Task: view.TaskRef{ID: "e9c6"}, From: "design", To: "mockup", Text: "Views", Due: true}}}
	out := renderString(t, v, Options{Level: view.Provenance}) + renderString(t, v, Options{Level: view.Glance})
	for _, want := range []string{
		"| A `kind` of \\*finding\\* |",
		"| A `message` with a \\| bar | Await `e9c6` at 📐 design |",
		"**Rule.** \"The `sentence`.\" [README.md, Tasks](https://github.com/nbyoung/tableaux/blob/main/README.md#tasks)\n",
		"**The requirement.** `e3cb`: `e9c6`, design → mockup, Views: unmet.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output lacks %q", want)
		}
	}

	p := newPage(&v.Head, Options{Level: view.Detail})
	one := []view.TaskRef{{ID: "e3cb"}}
	for _, c := range []struct {
		act   string
		tasks []view.TaskRef
		want  string
	}{
		{"authorise", one, "Authorise: commit Authorised: e3cb"},
		{"authorise", []view.TaskRef{{ID: "aaaa"}, {ID: "bbbb"}}, "Authorise: commit Authorised: <id>"},
		{"review", one, "Review: commit Reviewed: e3cb implementation"},
		{"reaffirm", one, "Reaffirm: commit Reaffirmed: e3cb"},
		{"record_handoff", one, "Record the hand-off: the reason 👓 review when the work is done; or work on"},
		{"clear_handoff", one, "Clear the hand-off: record the gate, or the next hand-off"},
		{"revise", one, "Revise the file"},
		{"move_pin", one, "Move the pin or the commit"},
		{"checkout", one, "Check out the submodule"},
		{"await", one, "Await the task the requirement names"},
		{"none", one, ""},
	} {
		got := plainInline(p.auditAction(view.Finding{Act: c.act, Tasks: c.tasks, Gate: "implementation"}), map[string]bool{})
		if got != c.want {
			t.Errorf("%s: %q, want %q", c.act, got, c.want)
		}
	}
}
