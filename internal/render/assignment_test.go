package render

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/markdown"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// assignmentOf builds an assignment view with the legend of the tooling plan
// and one section for the given junctions.
func assignmentOf(t *testing.T, email string, con, rev []view.JunctionRow) *view.Assignment {
	t.Helper()
	v := &view.Assignment{}
	v.Level = view.Provenance
	v.Project = view.TaskRef{ID: "r000", Title: "Root"}
	v.Ref = view.Ref{Name: "main", OnTrunk: true}
	v.Legend = readFixture(t, "assignment", "tooling").Header().Legend
	v.People = []view.PersonRow{{Email: email}}
	v.Sections = []view.PersonSection{{Email: email, Contributes: con, Reviews: rev}}
	return v
}

func junctions(gate string, n int, when func(i int) string) []view.JunctionRow {
	var out []view.JunctionRow
	for i := 0; i < n; i++ {
		out = append(out, view.JunctionRow{
			Task: view.TaskRef{ID: fmt.Sprintf("%s%02d", gate[:2], i), Title: fmt.Sprintf("Task %d", i)},
			Gate: gate, Contributor: "c@x.org", Model: "claude-opus", Reviewer: "r@x.org", When: when(i),
		})
	}
	return out
}

func always(w string) func(int) string { return func(int) string { return w } }

// linesBetween returns the lines after the first line that starts with from,
// up to the first later line that starts with to.
func linesBetween(out, from, to string) []string {
	var got []string
	in := false
	for _, ln := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(ln, from):
			in = true
		case in && strings.HasPrefix(ln, to):
			return got
		case in:
			got = append(got, ln)
		}
	}
	return got
}

// tableRows returns the lines of the tables among the lines.
func tableRows(lines []string) []string {
	var out []string
	for _, ln := range lines {
		if strings.HasPrefix(ln, "|") {
			out = append(out, ln)
		}
	}
	return out
}

// T8: gate groups and the run fold.
func TestGateGroupsAndFold(t *testing.T) {
	// Rows given out of gate order group in the legend's order, and the rows
	// keep their order inside a group.
	t.Run("legend order", func(t *testing.T) {
		var rows []view.JunctionRow
		for _, g := range []string{"validate", "design", "defined", "validate", "release", "design"} {
			rows = append(rows, junctions(g, 1, always("next"))...)
		}
		rows[3].Task.ID, rows[5].Task.ID = "va99", "de99"
		v := assignmentOf(t, "c@x.org", rows, nil)
		out := renderString(t, v, Options{Level: view.Detail})
		var groups []string
		for _, ln := range strings.Split(out, "\n") {
			if strings.HasPrefix(ln, "**") && strings.Contains(ln, " junctions · ") {
				groups = append(groups, ln)
			}
		}
		want := []string{
			"**📝 defined** · 1 junctions · 1 next",
			"**📐 design** · 2 junctions · 2 next",
			"**🌍 validate** · 2 junctions · 2 next",
			"**🚀 release** · 1 junctions · 1 next",
		}
		if !slices.Equal(groups, want) {
			t.Errorf("groups\n%s\nwant\n%s", strings.Join(groups, "\n"), strings.Join(want, "\n"))
		}
		if !strings.Contains(out, "### Contributes: 6 junctions, 6 next\n") {
			t.Errorf("heading:\n%s", out)
		}
		if i, j := strings.Index(out, "`va00`"), strings.Index(out, "`va99`"); i < 0 || j < i {
			t.Errorf("a group reorders its rows")
		}
	})

	cases := []struct {
		name   string
		rows   []view.JunctionRow
		unfold bool
		table  int    // table lines, header and delimiter included
		fold   string // the fold row, if any
	}{
		{"seven alike", junctions("design", 7, always("passed")), false, 2 + 7, ""},
		{"eight alike", junctions("design", 8, always("passed")), false, 2 + 8, ""},
		{"nine alike", junctions("design", 9, always("passed")), false, 2 + 4,
			"| … and 6 more: `de03` `de04` `de05` `de06` `de07` `de08` | `claude-opus` | r@x.org | passed |"},
		{"nine unfolded", junctions("design", 9, always("passed")), true, 2 + 9, ""},
		{"eight that alternate", junctions("design", 8, func(i int) string { return []string{"next", "later"}[i%2] }), false, 2 + 8, ""},
		{"twenty that alternate", junctions("design", 20, func(i int) string { return []string{"next", "later"}[i%2] }), false, 2 + 20, ""},
		{"twenty alike", junctions("design", 20, always("next")), false, 2 + 4,
			"| … and 17 more: `de03` `de04` `de05` `de06` `de07` `de08` `de09` `de10` `de11` `de12` `de13` `de14` `de15` `de16` `de17` `de18` `de19` | `claude-opus` | r@x.org | next |"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := assignmentOf(t, "c@x.org", c.rows, nil)
			out := renderString(t, v, Options{Level: view.Detail, Unfold: c.unfold})
			rows := tableRows(linesBetween(out, "**📐 design", "### "))
			if len(rows) != c.table {
				t.Errorf("%d table lines, want %d:\n%s", len(rows), c.table, strings.Join(rows, "\n"))
			}
			if c.fold == "" && strings.Contains(out, "more:") {
				t.Errorf("a fold stands:\n%s", strings.Join(rows, "\n"))
			}
			if c.fold != "" && !slices.Contains(rows, c.fold) {
				t.Errorf("the fold row is missing:\n%s", strings.Join(rows, "\n"))
			}
		})
	}

	// The same for reviews, whose table names the contributor.
	t.Run("reviews", func(t *testing.T) {
		v := assignmentOf(t, "r@x.org", nil, junctions("design", 9, always("passed")))
		out := renderString(t, v, Options{Level: view.Detail})
		rows := tableRows(linesBetween(out, "**📐 design", "### "))
		if rows[0] != "| Task | Contributor | Model | When |" || len(rows) != 2+4 ||
			rows[5] != "| … and 6 more: `de03` `de04` `de05` `de06` `de07` `de08` | c@x.org | `claude-opus` | passed |" {
			t.Errorf("reviews:\n%s", strings.Join(rows, "\n"))
		}
		if !strings.Contains(out, "### Contributes: none\n\n### Reviews: 9 junctions, 0 next\n") {
			t.Errorf("headings:\n%s", out)
		}
	})

	// The tasks assigned fold by the same rule, and the folded ids stay.
	t.Run("tasks assigned", func(t *testing.T) {
		v := assignmentOf(t, "c@x.org", nil, nil)
		for i := 0; i < 12; i++ {
			v.Sections[0].Assigned = append(v.Sections[0].Assigned, view.AssignedRow{
				TaskRef: view.TaskRef{ID: fmt.Sprintf("t%03d", i), Title: "T"}, Under: "r000",
				Status: view.Status{Gate: "defined", State: "nominal"},
				Next:   &view.Marked{Marks: []string{"agent"}}, Gate: "mockup",
			})
		}
		out := renderString(t, v, Options{Level: view.Detail})
		rows := tableRows(linesBetween(out, "### Tasks assigned", "### "))
		if len(rows) != 2+4 || !strings.HasPrefix(rows[5], "| … and 9 more: `t003`") || !strings.Contains(rows[5], "`t011` | `r000` | 📝 defined 🟢 nominal | 📌 mockup 🤖 |") {
			t.Errorf("tasks assigned:\n%s", strings.Join(rows, "\n"))
		}
	})
}

// T9: empty sections.
func TestEmptySections(t *testing.T) {
	out := renderString(t, readFixture(t, "assignment", "weather"), Options{Level: view.Detail})
	sec := func(email string) string {
		i := strings.Index(out, "\n## "+email+"\n")
		if i < 0 {
			t.Fatalf("no section for %s", email)
		}
		rest := out[i+1:]
		if j := strings.Index(rest[3:], "\n## "); j >= 0 {
			rest = rest[:j+3]
		}
		return rest
	}
	ada := sec("ada@example.org")
	if !strings.Contains(ada, "### Reviews: none\n\n### Counts by gate\n") {
		t.Errorf("an email that reviews nothing:\n%s", ada)
	}
	if strings.Contains(ada, "Models by gate") || strings.Contains(ada, "No contributor") {
		t.Errorf("ada holds a section she has no row for")
	}
	if !strings.Contains(ada, "### Authority\n\nOver `a1c0` Weather station and its 5 descendants.\n") {
		t.Errorf("ada's authority:\n%s", ada)
	}
	if dan := sec("dan@example.org"); !strings.Contains(dan, "### Authority\n\nNone.\n") {
		t.Errorf("an email with no subtree:\n%s", dan)
	}
	opus := sec("opus@example.org")
	for _, want := range []string{"### Tasks assigned: none\n\n### Contributes: none\n\n### Reviews: none\n\n### Models by gate\n"} {
		if !strings.Contains(opus, want) {
			t.Errorf("an email with nothing lacks %q:\n%s", want, opus)
		}
	}
	// Every gate of the legend stands in the counts, the zeros included.
	if n := len(tableRows(linesBetween(opus, "### Counts by gate", "### "))); n != 2+12+1 {
		t.Errorf("the counts hold %d lines", n)
	}
}

// The glance, the detail of a person and the provenance.
func TestAssignmentShapes(t *testing.T) {
	tooling := readFixture(t, "assignment", "tooling")

	t.Run("glance", func(t *testing.T) {
		out := renderString(t, tooling, Options{Level: view.Glance})
		for _, want := range []string{
			"|---|--:|--:|--:|---|\n",
			"| noreply@anthropic.com | 27 | 28 | 2 | `claude-haiku`, `claude-opus`, `claude-fable`, `claude-sonnet` |\n",
			"\n🪆 4 tasks hand their next junction to a subproject: `77b2` `6103` `c6e8` `595e`.\n",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("the glance lacks %q", want)
			}
		}
		if strings.Contains(out, "## ") {
			t.Errorf("a section stands at glance")
		}
		// Without a task that hands its junction over, the line is absent.
		v := *tooling.(*view.Assignment)
		v.Recursive = nil
		if out := renderString(t, &v, Options{Level: view.Glance}); strings.Contains(out, "🪆") {
			t.Errorf("the subproject line stands with no recursive task")
		}
	})

	t.Run("the person is bold where she holds a position", func(t *testing.T) {
		v := readFixture(t, "assignment", "tooling-nbyoung")
		out := renderString(t, v, Options{Level: view.Detail})
		if !strings.Contains(out, "· person nbyoung@nbyoung.com · level detail") ||
			!strings.Contains(out, "Command: `tabloio assignment --person nbyoung@nbyoung.com --ref main --level detail`") {
			t.Errorf("the frame:\n%s", out)
		}
		if strings.Contains(out, "**nbyoung") {
			t.Errorf("an email of the glance table or a heading is bold")
		}
		// A table of another email's section holds her as the reviewer.
		w := *tooling.(*view.Assignment)
		w.Params.Person = "nbyoung@nbyoung.com"
		w.Sections = w.Sections[1:]
		if out := renderString(t, &w, Options{Level: view.Detail}); !strings.Contains(out, "| `claude-opus` | **nbyoung@nbyoung.com** | next |") {
			t.Errorf("the reviewer is not bold:\n%s", out)
		}
	})

	t.Run("provenance", func(t *testing.T) {
		out := renderString(t, tooling, Options{Level: view.Provenance})
		for _, want := range []string{
			"\n## nbyoung@nbyoung.com, provenance\n",
			"\n## noreply@anthropic.com, provenance\n",
			"| Position | Tasks | Stated in |\n",
			"| Contributes at | Tasks | Contributor and model from | Reviewer | Reviewer from |\n",
			"| Reviews at | Tasks | Reviewer from | Contributor and model from |\n",
			"| Task | Stated in | Pinned at |\n",
			"| `77b2` | `junctions.mockup.subproject.url` in `.tableaux/tasks/77b2.yaml`, commit `6b6c99a` | `00f8f68` |\n",
			// A Stated reads the task and its title from the rows of the view.
			"| 🚀 release | 8: `c2ad` `e9c6` `e3cb` `fcec` `ac33` `9f3f` `7861` `7166` | `437e` Tableaux tooling, `junctions.release` in `.tableaux/tasks/437e.yaml`, commit `6b6c99a` |  |  |\n",
			"\n## Commits: 2\n\n| Commit | Date | Author | Committer |\n|---|---|---|---|\n| `320cf2b` Name a model family per gate (D7) | 2026-09-30 | noreply@anthropic.com | nbyoung@nbyoung.com |\n",
			"git ls-tree 3cdae52 subprojects/ # the four pins\n",
			"Command: `tabloio assignment --ref 3cdae52 --level provenance`\n",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("the provenance lacks %q", want)
			}
		}
		// The sections come first, then the provenance of each email.
		if strings.Index(out, "## noreply@anthropic.com\n") > strings.Index(out, "## nbyoung@nbyoung.com, provenance") {
			t.Errorf("a provenance section stands before the last email's detail")
		}
	})

	t.Run("links and the See also line", func(t *testing.T) {
		out := renderString(t, tooling, Options{Level: view.Detail, Links: Links{Project: "../..", Views: map[string]string{"authority": "A.md", "audit": "U.md"}}})
		for _, want := range []string{
			"| [`437e` Tableaux tooling](../../.tableaux/tasks/437e.yaml) |",
			"See also: `tabloio queue`, [authority](A.md), `tabloio task`, `tabloio tableau`, `tabloio context`, `tabloio blockage`, `tabloio history`, [audit](U.md).\n",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("the output lacks %q", want)
			}
		}
	})

	t.Run("the data holds less than the level", func(t *testing.T) {
		v := readFixture(t, "assignment", "tooling-nbyoung")
		if err := Render(new(strings.Builder), v, Options{Level: view.Provenance}); err == nil {
			t.Errorf("provenance of detail data")
		}
	})
}

// T10: a Stated words its source: a task, the default, a task with no title.
func TestStated(t *testing.T) {
	titles := map[string]string{"437e": "Tableaux tooling"}
	statedCases := []struct {
		name string
		in   view.Stated
		want string
	}{
		{"a task", view.Stated{By: "437e", Key: "junctions.release", File: ".tableaux/tasks/437e.yaml", Commit: "6b6c99a1234"},
			"`437e` Tableaux tooling, `junctions.release` in `.tableaux/tasks/437e.yaml`, commit `6b6c99a`"},
		{"the default", view.Stated{By: "default", Key: "assignee", File: ".tableaux/tasks/ID.yaml", Commit: "6b6c99a"},
			"the plain default, `assignee` in `.tableaux/tasks/ID.yaml`, commit `6b6c99a`"},
		{"a task with no title", view.Stated{By: "bc63", Key: "junctions.design"}, "`bc63`, `junctions.design`"},
		{"nothing", view.Stated{}, ""},
	}
	for _, c := range statedCases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := markdown.Write(&buf, doc.Doc{Blocks: []doc.Block{doc.Para{Text: stated(c.in, titles)}}}); err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(buf.String()); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
