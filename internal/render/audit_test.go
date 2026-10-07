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
	v.Project = view.TaskRef{ID: "437e", Title: "Tableaux tooling"}
	v.Ref = view.Ref{Name: "main", OnTrunk: true}
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
	v.Ref = view.Ref{Name: "main", OnTrunk: true}
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

// TestAuditWordsNothing: the kind, the message, the action, the sentence and
// the facts print as tablo words them, split at backticks and no further.
func TestAuditWordsNothing(t *testing.T) {
	v := readIn(t, "audit", "tooling").(*view.Audit)
	v.Findings = v.Findings[:1]
	f := &v.Findings[0]
	f.Kind, f.Message, f.Action, f.Sentence = "A `kind` of *finding*", "A `message` with a | bar", "Do `this`", "The `sentence`."
	f.Facts = []view.Fact{{Name: "Odd", Value: "a `b` c"}}
	out := renderString(t, v, Options{Level: view.Provenance}) + renderString(t, v, Options{Level: view.Glance})
	for _, want := range []string{
		"| A `kind` of \\*finding\\* |",
		"| A `message` with a \\| bar | Do `this` |",
		"**Rule.** \"The `sentence`.\" [README.md, Status](README.md#status)\n",
		"**Odd.** a `b` c\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output lacks %q", want)
		}
	}
}
