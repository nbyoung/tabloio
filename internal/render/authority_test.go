package render

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/view"
)

// renderLines renders a view and returns its lines.
func renderLines(t *testing.T, v view.View, o Options) []string {
	t.Helper()
	return strings.Split(renderString(t, v, o), "\n")
}

// fenced returns the lines of the first fenced block.
func fenced(t *testing.T, lines []string) []string {
	t.Helper()
	var out []string
	in := false
	for _, ln := range lines {
		switch {
		case strings.HasPrefix(ln, "```") && !in:
			in = true
		case strings.HasPrefix(ln, "```"):
			return out
		case in:
			out = append(out, ln)
		}
	}
	t.Fatal("no fenced block")
	return nil
}

// T5: the tree rule. Every assignee starts at the header's column plus twice
// the row's depth, and the column lies two characters past the longest task
// text, counted in code points.
func TestTreeRule(t *testing.T) {
	check := func(t *testing.T, v *view.Delegation, level view.Level) {
		t.Helper()
		tree := fenced(t, renderLines(t, v, Options{Level: level}))
		byID := map[string]view.DelegationRow{}
		for _, r := range v.Rows {
			byID[r.ID] = r
		}
		header := []rune(tree[0])
		column := strings.Index(tree[0], "Assignee")
		column = len([]rune(tree[0][:column]))
		if string(header[:4]) != "Task" || column < 6 {
			t.Fatalf("header %q", tree[0])
		}
		longest := 0
		for _, ln := range tree[1:] {
			runes := []rune(ln)
			depth := 0
			for depth < len(runes) && runes[depth] == ' ' {
				depth++
			}
			depth /= 2
			id := strings.Fields(ln)[0]
			r, ok := byID[id]
			if !ok {
				t.Fatalf("line %q names no row", ln)
			}
			if r.Depth != depth {
				t.Errorf("%s: indent %d, depth %d", id, depth, r.Depth)
			}
			at := column + 2*depth
			if at > len(runes) || !strings.HasPrefix(string(runes[at:]), r.Assignee) {
				t.Errorf("the assignee of %s does not start at column %d: %q", id, at, ln)
			}
			text := strings.Repeat("  ", depth) + id + " " + r.Title
			longest = max(longest, len([]rune(text)))
			if !strings.HasPrefix(ln, text) {
				t.Errorf("line %q does not open with %q", ln, text)
			}
		}
		if column != longest+2 {
			t.Errorf("the column is %d, two past the longest text is %d", column, longest+2)
		}
	}
	for _, f := range fixtures(t) {
		if f.view != "authority" {
			continue
		}
		v := readFixture(t, f.view, f.name).(*view.Delegation)
		if len(v.Rows) == 0 {
			continue
		}
		for _, l := range levelsOf(v) {
			t.Run(f.String()+"."+l.String(), func(t *testing.T) { check(t, v, l) })
		}
	}

	// A title with a non-ASCII letter keeps the column, by code points.
	t.Run("a non-ASCII letter", func(t *testing.T) {
		v := readFixture(t, "authority", "weather").(*view.Delegation)
		cp := *v
		cp.Rows = append([]view.DelegationRow(nil), v.Rows...)
		cp.Rows[2].Title = "Carte capteur météo à l’épreuve"
		cp.Rows[1].Title = "Nœud"
		for _, l := range []view.Level{view.Glance, view.Detail} {
			check(t, &cp, l)
		}
	})

	// The mockup's illustration of a proposed task, line for line.
	t.Run("the illustration of a proposed task", func(t *testing.T) {
		v := readFixture(t, "authority", "weather-proposed")
		got := fenced(t, renderLines(t, v, Options{Level: view.Glance}))
		want := []string{
			"Task                  Assignee",
			"a1c0 Weather station  ada@example.org",
			"  3c5d Dashboard        dan@example.org  proposed",
		}
		if !slices.Equal(got, want) {
			t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	})
}

// treeOf builds a delegation view: a root and the children it is given.
func treeOf(kids ...view.DelegationRow) *view.Delegation {
	v := &view.Delegation{}
	v.Level = view.Provenance
	v.Project = view.TaskRef{ID: "r000", Title: "Root"}
	v.Ref = view.Ref{Name: "main", OnTrunk: true}
	root := view.DelegationRow{TaskRef: view.TaskRef{ID: "r000", Title: "Root"}, Assignee: "a@x.org", Parent: true}
	v.Rows = append([]view.DelegationRow{root}, kids...)
	return v
}

func kid(n int, mod func(*view.DelegationRow)) view.DelegationRow {
	r := view.DelegationRow{
		TaskRef:  view.TaskRef{ID: fmt.Sprintf("k%03d", n), Title: fmt.Sprintf("Kid %d", n)},
		Depth:    1,
		Assignee: "a@x.org",
	}
	if mod != nil {
		mod(&r)
	}
	return r
}

func kids(n int, mod func(i int, r *view.DelegationRow)) []view.DelegationRow {
	var out []view.DelegationRow
	for i := 0; i < n; i++ {
		out = append(out, kid(i, func(r *view.DelegationRow) {
			if mod != nil {
				mod(i, r)
			}
		}))
	}
	return out
}

// T6: collapsed children.
func TestCollapsedChildren(t *testing.T) {
	cases := []struct {
		name   string
		kids   []view.DelegationRow
		level  view.Level
		unfold bool
		folded bool
		word   string
	}{
		{"four children", kids(4, nil), view.Glance, false, false, ""},
		{"five children", kids(5, nil), view.Glance, false, true, "5 children"},
		{"twelve children", kids(12, nil), view.Glance, false, true, "12 children"},
		{"five, a parent among them", kids(5, func(i int, r *view.DelegationRow) { r.Parent = i == 2 }), view.Glance, false, false, ""},
		{"five, another assignee", kids(5, func(i int, r *view.DelegationRow) {
			if i == 3 {
				r.Assignee = "b@x.org"
			}
		}), view.Glance, false, false, ""},
		{"five, one proposed", kids(5, func(i int, r *view.DelegationRow) { r.Proposed = i == 4 }), view.Glance, false, false, ""},
		{"five, unfolded", kids(5, nil), view.Glance, true, false, ""},
		{"five at detail", kids(5, nil), view.Detail, false, false, ""},
		{"five at provenance", kids(5, nil), view.Provenance, false, false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := treeOf(c.kids...)
			tree := fenced(t, renderLines(t, v, Options{Level: c.level, Unfold: c.unfold}))
			wantRows := 1 + 1 + len(c.kids)
			if c.folded {
				wantRows = 2
			}
			if len(tree) != wantRows {
				t.Errorf("%d tree lines, want %d:\n%s", len(tree), wantRows, strings.Join(tree, "\n"))
			}
			if c.folded && !strings.HasSuffix(tree[1], "a@x.org  "+c.word) {
				t.Errorf("the parent's row lacks %q: %q", c.word, tree[1])
			}
			if !c.folded && strings.Contains(strings.Join(tree, "\n"), " children") {
				t.Errorf("a count stands where nothing folds:\n%s", strings.Join(tree, "\n"))
			}
			// The count line counts every task, the hidden ones included.
			out := renderString(t, v, Options{Level: c.level, Unfold: c.unfold})
			if want := fmt.Sprintf("\n%d tasks, 0 delegations,", 1+len(c.kids)); !strings.Contains(out, want) {
				t.Errorf("the count line lacks %q:\n%s", want, out)
			}
		})
	}

	// The fold takes no part in the column: the hidden rows do not widen it.
	t.Run("hidden rows do not widen the column", func(t *testing.T) {
		long := kids(5, func(i int, r *view.DelegationRow) { r.Title = strings.Repeat("long ", 20) })
		tree := fenced(t, renderLines(t, treeOf(long...), Options{Level: view.Glance}))
		if tree[0] != "Task       Assignee" { // "r000 Root" and two more
			t.Errorf("header %q", tree[0])
		}
	})
}

// Levels, parameters and the shapes around the tree.
func TestDelegationFrame(t *testing.T) {
	v := readFixture(t, "authority", "weather-ada")
	out := renderString(t, v, Options{Level: view.Detail})
	for _, want := range []string{
		"**Who may accept what?** Weather station · ref `main` · person ada@example.org · level detail\n",
		"Authorities of its 3 children, nearest first: **ada@example.org** (`a1c0`).\n",
		"Authorities of its 2 children, nearest first: ben@example.org (`4e2b`); then **ada@example.org** (`a1c0`).\n",
		"Command: `tabloio authority --person ada@example.org --ref main --level detail`\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "See also") {
		t.Errorf("a See also line stands with no view named")
	}
	out = renderString(t, v, Options{Level: view.Detail, Links: Links{Project: ".", Views: map[string]string{"history": "H.md", "queue": "Q.md"}}})
	for _, want := range []string{
		"### [`a1c0` Weather station](.tableaux/tasks/a1c0.yaml) · **ada@example.org**\n",
		"See also: `tabloio task`, `tabloio assignment`, [history](H.md), `tabloio audit`, [queue](Q.md).\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(strings.Join(fenced(t, strings.Split(out, "\n")), "\n"), "[") {
		t.Errorf("the tree holds a link")
	}

	// Off the trunk the frame says so, and the word carries the difference.
	br := renderString(t, readFixture(t, "authority", "weather-branch"), Options{Level: view.Glance})
	if !strings.Contains(br, "ref `91aa774`, off the trunk") || !strings.Contains(br, "proposed (differs from trunk)") {
		t.Errorf("an off-trunk ref is not marked:\n%s", br)
	}
	// An empty view that is no proposal view says so in its own words.
	e := treeOf()
	e.Rows = nil
	if out := renderString(t, e, Options{Level: view.Glance}); !strings.Contains(out, "**No task is in view.**\n\n0 tasks, 0 delegations, 0 authorised, 0 proposed.") {
		t.Errorf("empty view:\n%s", out)
	}
	// The data holds less than the level asks for.
	g := readFixture(t, "authority", "weather-branch")
	if err := Render(io.Discard, g, Options{Level: view.Provenance}); !errors.Is(err, ErrLevel) {
		t.Errorf("provenance of detail data: %v", err)
	}
}

// T7: --proposed and its empty form.
func TestProposedOnly(t *testing.T) {
	out := renderString(t, readFixture(t, "authority", "tooling-proposed"), Options{Level: view.Glance})
	for _, want := range []string{
		"· proposed only · level glance\n",
		"\n**No task is proposed.**\n\n0 tasks, 0 delegations, 0 authorised, 0 proposed.\n",
		"Command: `tabloio authority --ref main --proposed --level glance`\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "```\nTask") {
		t.Errorf("a tree stands in the empty form")
	}
	w := renderString(t, readFixture(t, "authority", "weather-proposed"), Options{Level: view.Detail})
	if !strings.Contains(w, "  3c5d Dashboard        dan@example.org  proposed\n") ||
		!strings.Contains(w, "2 tasks, 1 delegations, 1 authorised, 1 proposed.") {
		t.Errorf("weather-proposed:\n%s", w)
	}
}

// T10: a Default words its sources: a field from an ancestor, an exempt gate,
// a field of the parent's own.
func TestDefault(t *testing.T) {
	v := treeOf(kid(0, nil))
	v.Parents = []view.Chain{{
		TaskRef: view.TaskRef{ID: "r000", Title: "Root"}, Assignee: "a@x.org",
		Authorities: []view.Authority{{Email: "a@x.org", By: "r000"}}, Children: []string{"k000"},
		Defaults: []view.Default{
			{Gate: "design", Marked: view.Marked{Marks: []string{"agent", "reviewer"}}, Contributor: "ag@x.org", Model: "claude-fable",
				Reviewer: "a@x.org", ReviewerFrom: "top1", ModelFrom: "top2"},
			{Gate: "unit", Marked: view.Marked{Marks: []string{"exempt"}}},
			{Gate: "release", Marked: view.Marked{Marks: []string{"person"}}, Contributor: "a@x.org", ContributorFrom: "top3"},
			{Gate: "implementation", Marked: view.Marked{Marks: []string{"agent"}}, ReviewerFrom: "top1"},
		},
	}}
	v.Legend = readFixture(t, "authority", "tooling").Header().Legend
	out := renderString(t, v, Options{Level: view.Detail})
	for _, want := range []string{
		"| 📐 design | 🤖👀 | ag@x.org | `claude-fable`, from `top2` | a@x.org, from `top1` |\n",
		"| 📏 unit | — |  |  |  |\n",
		"| 🚀 release | 🧑 | a@x.org, from `top3` |  |  |\n",
		"| 🧱 implementation | 🤖 |  |  |  |\n",
		"🧑 a person contributes · 🤖 an agent contributes · 👀 a reviewer accepts · — the gate does not apply.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the defaults lack %q:\n%s", want, out)
		}
	}
}
