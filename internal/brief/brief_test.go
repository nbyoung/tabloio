package brief

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// envelopeData returns the data member of the envelope in testdata/name.
func envelopeData(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	return env.Data
}

// load decodes the brief of testdata/name.json.
func load(t *testing.T, name string) Data {
	t.Helper()
	d, err := Decode(envelopeData(t, name+".json"), Request{})
	if err != nil {
		t.Fatalf("Decode %s: %v", name, err)
	}
	return d
}

// golden reads testdata/name.
func golden(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// render returns the brief of d.
func render(t *testing.T, d Data) string {
	t.Helper()
	var b bytes.Buffer
	if err := Render(&b, d); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return b.String()
}

// sameText fails with the first differing line when got and want differ.
func sameText(t *testing.T, got, want string) {
	t.Helper()
	if got == want {
		return
	}
	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		var gl, wl string
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			t.Fatalf("line %d differs\n got: %q\nwant: %q", i+1, gl, wl)
		}
	}
}

// lacks and has check a substring of a rendering.
func has(t *testing.T, out string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(out, s) {
			t.Errorf("output lacks %q", s)
		}
	}
}

// hasFlat checks a substring of a rendering with its white space joined, so
// that the check ignores wrapping.
func hasFlat(t *testing.T, out string, subs ...string) {
	t.Helper()
	has(t, flat(out), subs...)
}

func lacks(t *testing.T, out string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if strings.Contains(out, s) {
			t.Errorf("output holds %q", s)
		}
	}
}

// flat joins the white space of a text, so that a check ignores wrapping.
func flat(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestT1HandoffBrief(t *testing.T) {
	sameText(t, render(t, load(t, "e4c7-design")), golden(t, "e4c7-design.txt"))
}

func TestT2SelfReviewBrief(t *testing.T) {
	sameText(t, render(t, load(t, "e9c6-implementation")), golden(t, "e9c6-implementation.txt"))
}

func TestT3MockupConformance(t *testing.T) {
	got, want := render(t, load(t, "c74a-mockup")), golden(t, "mockup-c74a.txt")
	sameText(t, got, golden(t, "c74a-mockup.txt"))
	cut := func(s string) (before, from string) {
		i, j := strings.Index(s, "\n5. "), strings.Index(s, "\n6. ")
		if i < 0 || j < i {
			t.Fatal("a brief lacks the headings of parts 5 and 6")
		}
		return s[:i], s[j:]
	}
	gb, gf := cut(got)
	wb, wf := cut(want)
	sameText(t, gb, wb)
	sameText(t, gf, wf)
}

func TestT4WaitingItem(t *testing.T) {
	d := load(t, "1422-implementation")
	var b bytes.Buffer
	err := Render(&b, d)
	var w *WaitingError
	if !errors.As(err, &w) {
		t.Fatalf("Render error = %v, want *WaitingError", err)
	}
	want := "1422 at implementation waits and no brief starts it: waits for 5ca9 Read commands at implementation"
	if err.Error() != want {
		t.Errorf("message = %q, want %q", err, want)
	}
	if b.Len() != 0 {
		t.Errorf("Render wrote %d bytes of a waiting item", b.Len())
	}
}

func TestT5NoWork(t *testing.T) {
	req := Request{Person: "p@example.org", Task: "e4c7", Gate: "design"}
	for _, data := range []string{
		`{"view": "work-queue", "params": {}}`,
		`{"brief": null}`,
		`{"brief": {}}`,
	} {
		_, err := Decode([]byte(data), req)
		var nw *NoWorkError
		if !errors.As(err, &nw) || *nw != (NoWorkError{"p@example.org", "e4c7", "design"}) {
			t.Fatalf("Decode(%s) error = %v, want *NoWorkError", data, err)
		}
		want := "no work for p@example.org at e4c7 design; tabloio queue --person p@example.org lists the items"
		if err.Error() != want {
			t.Errorf("message = %q, want %q", err, want)
		}
	}

	var raw map[string]any
	if err := json.Unmarshal(envelopeData(t, "e4c7-design.json"), &raw); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		field string
		clear func(b map[string]any)
	}{
		{"task.id", func(b map[string]any) { b["task"].(map[string]any)["id"] = "" }},
		{"gate.key", func(b map[string]any) { delete(b["gate"].(map[string]any), "key") }},
		{"junction.contributor.value", func(b map[string]any) {
			b["junction"].(map[string]any)["contributor"].(map[string]any)["value"] = ""
		}},
		{"status.path", func(b map[string]any) { b["status"].(map[string]any)["path"] = "" }},
		{"ref.commit", func(b map[string]any) { b["ref"].(map[string]any)["commit"] = "" }},
	} {
		var fresh map[string]any
		if err := json.Unmarshal(envelopeData(t, "e4c7-design.json"), &fresh); err != nil {
			t.Fatal(err)
		}
		c.clear(fresh["brief"].(map[string]any))
		data, err := json.Marshal(fresh)
		if err != nil {
			t.Fatal(err)
		}
		_, err = Decode(data, req)
		want := "brief: " + c.field + " is missing in tablo's data"
		if err == nil || err.Error() != want {
			t.Errorf("Decode without %s: error = %v, want %q", c.field, err, want)
		}
	}
	if _, err := Decode([]byte(`{"brief": [`), req); err == nil {
		t.Error("Decode accepted invalid JSON")
	}
}

// variant returns a copy of d changed by f. The slices it changes are fresh.
func variant(d Data, f func(*Data)) Data {
	d.Junction.References = append([]Reference(nil), d.Junction.References...)
	d.Dependents = append([]Edge(nil), d.Dependents...)
	f(&d)
	return d
}

func TestT6WaysOfTheCommit(t *testing.T) {
	base := load(t, "e9c6-implementation")
	const trailers = "\n    Model: <the identifier your harness reports, matching claude-sonnet>\n" +
		"    Co-Authored-By: <your model's display name> <noreply@anthropic.com>"

	defined := render(t, variant(base, func(d *Data) { d.Gate.Key = "defined" }))
	has(t, defined, trailers, "    gate: defined\n")
	hasFlat(t, defined, "Authorisation stands as the review of defined", "carries no Reviewed: trailer")
	lacks(t, defined, "\n    Reviewed:", "hand the work off", "You are the reviewer", "A person contributes")

	handoff := render(t, variant(base, func(d *Data) { d.Junction.Reviewer.Value = "nbyoung@nbyoung.com" }))
	has(t, handoff, trailers,
		"    gate: design\n    state: nominal\n    reason: review\n    note: <one line that says what waits on the branch>\n")
	hasFlat(t, handoff, "The reviewer nbyoung@nbyoung.com accepts this gate, so hand the work off",
		"passes the gate with a commit that carries the trailer Reviewed: e9c6 implementation")
	lacks(t, handoff, "\n    Reviewed:", "Authorisation stands", "You are the reviewer", "Whoever dispatches")

	self := render(t, base)
	has(t, self, "\n    Reviewed: e9c6 implementation"+trailers, "Whoever dispatches you merges the branch.")
	hasFlat(t, self, "You are the reviewer of this gate as well as its contributor")
	lacks(t, self, "Authorisation stands", "hand the work off", "A person contributes")

	none := render(t, variant(base, func(d *Data) {
		d.Junction.Model = Field{}
		d.Junction.Reviewer = Field{}
	}))
	has(t, none, "A person contributes here, so the commit carries no trailer.", "  Reviewer     none\n")
	hasFlat(t, none, "No one reviews this gate, so the commit that records the status passes it.")
	lacks(t, none, "Reviewed:", "Model:", "Co-Authored-By:", "hand the work off", "You are the reviewer")
}

func TestT7CompleteAndRecursive(t *testing.T) {
	base := load(t, "e9c6-implementation")

	last := render(t, variant(base, func(d *Data) { d.Gate.Follows = "" }))
	has(t, last, "    gate: implementation\n    state: complete\n    note: <one line that says what the gate now holds>\n")
	lacks(t, last, "state: nominal")

	plain := render(t, base)
	has(t, plain, "    gate: implementation\n    state: nominal\n    note: ")

	rec := render(t, variant(base, func(d *Data) { d.Gate.Follows = "recursive" }))
	has(t, rec, "\n    gate: implementation\n\n  The next junction is recursive, so the file holds the gate alone. Edit it\n"+
		"  by hand, and change no other file under .tableaux.\n")
	lacks(t, rec, "    state:", "    note:", "so that it reads as above")

	// A hand-off keeps its own file, whatever follows.
	hand := render(t, variant(base, func(d *Data) {
		d.Gate.Follows = "recursive"
		d.Junction.Reviewer.Value = "nbyoung@nbyoung.com"
	}))
	has(t, hand, "    reason: review\n", "so that it reads as above")
	lacks(t, hand, "The next junction is recursive")
}

func TestT8Person(t *testing.T) {
	base := load(t, "e4c7-design")
	person := render(t, variant(base, func(d *Data) {
		d.Junction.Model = Field{}
		d.Junction.Reviewer = Field{}
	}))
	lacks(t, person, "  Model ", "The model is the plan's statement", "export ", "Model:", "git log", "Co-Authored-By")
	has(t, person, "A person contributes here, so the commit carries no trailer.")

	// A person who reviews their own gate carries the Reviewed: trailer alone.
	self := render(t, variant(base, func(d *Data) {
		d.Junction.Model = Field{}
		d.Junction.Reviewer.Value = d.Junction.Contributor.Value
	}))
	has(t, self, "paragraph of one trailer:\n\n    Reviewed: e4c7 design\n")
	lacks(t, self, "Model:", "export ", "Co-Authored-By")
}

func TestT9Wrap(t *testing.T) {
	x := func(n int) string { return strings.Repeat("x", n) }
	for _, c := range []struct {
		name   string
		text   string
		indent int
		want   []string
	}{
		{"fits at 76, indent 0", x(37) + " " + x(38), 0, []string{x(37) + " " + x(38)}},
		{"one over 76, indent 0", x(37) + " " + x(39), 0, []string{x(37), x(39)}},
		{"fits at 76, indent 2", x(37) + " " + x(36), 2, []string{"  " + x(37) + " " + x(36)}},
		{"one over 76, indent 2", x(37) + " " + x(37), 2, []string{"  " + x(37), "  " + x(37)}},
		{"indent 4", x(35) + " " + x(36), 4, []string{"    " + x(35) + " " + x(36)}},
		{"indent 4 breaks", x(36) + " " + x(36), 4, []string{"    " + x(36), "    " + x(36)}},
		{"symbol counts as one rune", "🧱 " + x(74), 0, []string{"🧱 " + x(74)}},
		{"symbol and one more rune", "🧱 " + x(75), 0, []string{"🧱", x(75)}},
		{"symbol mid line", "ab " + "🧱 " + x(69), 2, []string{"  ab 🧱 " + x(69)}},
		{"word longer than the line", "a " + x(100) + " b", 2, []string{"  a", "  " + x(100), "  b"}},
		{"two paragraphs", "one two\n\nthree", 4, []string{"    one two", "", "    three"}},
		{"white space collapses", " one \t two\nthree ", 0, []string{"one two three"}},
		{"empty", "  \n ", 2, nil},
	} {
		got := wrap(c.text, c.indent)
		if fmt.Sprintf("%q", got) != fmt.Sprintf("%q", c.want) {
			t.Errorf("%s: wrap = %q, want %q", c.name, got, c.want)
		}
		for _, l := range got {
			if n := utf8.RuneCountInString(l); n > width && !strings.Contains(c.name, "longer") {
				t.Errorf("%s: line of %d runes", c.name, n)
			}
		}
	}
}

// edges returns n edges alike but for the task, ids from first.
func edges(n, first int, text string) []Edge {
	var es []Edge
	for i := 0; i < n; i++ {
		es = append(es, Edge{ID: fmt.Sprintf("t%03d", first+i), Title: "T", From: "design", To: "mockup",
			Text: text, Condition: "met"})
	}
	return es
}

func sizes(runs [][]Edge) []int {
	var s []int
	for _, r := range runs {
		s = append(s, len(r))
	}
	return s
}

func TestT10Fold(t *testing.T) {
	for _, c := range []struct {
		name string
		in   []Edge
		want []int
	}{
		{"none", nil, nil},
		{"four stay four", edges(4, 0, "a"), []int{1, 1, 1, 1}},
		{"five stay five", edges(5, 0, "a"), []int{1, 1, 1, 1, 1}},
		{"eight stay eight", edges(8, 0, "a"), []int{1, 1, 1, 1, 1, 1, 1, 1}},
		{"nine fold to one", edges(9, 0, "a"), []int{9}},
		{"twenty fold to one", edges(20, 0, "a"), []int{20}},
		{"an unlike entry breaks the run",
			append(append(edges(5, 0, "a"), edges(1, 5, "b")...), edges(4, 6, "a")...),
			[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		{"two long runs around an unlike entry",
			append(append(edges(9, 0, "a"), edges(1, 9, "b")...), edges(9, 10, "a")...),
			[]int{9, 1, 9}},
		{"a long run then a short one",
			append(edges(9, 0, "a"), edges(3, 9, "b")...),
			[]int{9, 1, 1, 1}},
	} {
		if got := sizes(fold(c.in)); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%s: fold sizes = %v, want %v", c.name, got, c.want)
		}
	}

	// In the brief: the fold names the first and the last task of the run.
	base := load(t, "e9c6-implementation")
	base.Dependents = nil
	for n, want := range map[int]string{8: "", 9: "Also required by: 9 tasks, t000 to t008, from design to mockup, a: met"} {
		out := render(t, variant(base, func(d *Data) { d.Dependents = edges(n, 0, "a") }))
		if want == "" {
			if got := strings.Count(out, "Also required by: t0"); got != n {
				t.Errorf("%d alike dependents give %d lines", n, got)
			}
			lacks(t, out, " tasks, ")
		} else {
			has(t, out, want)
			lacks(t, out, "Also required by: t0")
		}
	}
}

func TestT11SourceLabels(t *testing.T) {
	for _, c := range []struct {
		s    Source
		want string
	}{
		{Source{Kind: "task", ID: "437e", Title: "Tableaux tooling"}, "(Tableaux tooling, 437e)"},
		{Source{Kind: "task", ID: "e4c7", Title: "Agent briefs"}, "(this task)"},
		{Source{Kind: "assignee"}, "(the assignee; the junction states none)"},
		{Source{Kind: "default"}, "(the assignee, by default)"},
		{Source{}, ""},
	} {
		if got := sourceLabel(c.s, "e4c7"); got != c.want {
			t.Errorf("sourceLabel(%+v) = %q, want %q", c.s, got, c.want)
		}
	}

	base := load(t, "e4c7-design")
	out := render(t, variant(base, func(d *Data) {
		d.Junction.References = []Reference{
			{URL: "docs/gates/design.md", Text: "What the design gate asks for", Source: Source{Kind: "task", ID: "07e0", Title: "tabloio"}},
			{URL: "docs/gates/own.md", Source: Source{Kind: "task", ID: "e4c7", Title: "Agent briefs"}},
		}
	}))
	has(t, out,
		"\n  Junction reference: docs/gates/design.md, What the design gate asks for   (tabloio, 07e0)\n",
		"\n  Junction reference: docs/gates/own.md   (this task)\n  A junction reference ")
	hasFlat(t, out, "(this task) A junction reference expands the gate's criteria for this task. Read each one before you start.")
	lacks(t, out, "Junction references: none")
	has(t, render(t, base), "\n  Junction references: none\n")
}

func TestT12RefForms(t *testing.T) {
	base := load(t, "e4c7-design")
	const hash = "4593882"
	for _, c := range []struct {
		name, lead, branch string
	}{
		{"main", "of tabloio@example.org on main at " + hash, "on a branch off main."},
		{"HEAD", "of tabloio@example.org at " + hash, "on a branch off " + hash + "."},
		{"4593882", "of tabloio@example.org at " + hash, "on a branch off " + hash + "."},
		{"4593882a7b506cd9461fef8f9ccc6a6d653f6390", "of tabloio@example.org at " + hash, "on a branch off " + hash + "."},
	} {
		out := flat(render(t, variant(base, func(d *Data) {
			d.Ref.Name = c.name
			d.Person = "tabloio@example.org"
		})))
		has(t, out, c.lead+", 2026-10-02", c.branch)
		// Part 6 names the short hash whatever the ref was.
		has(t, out, "--ref "+hash+" --brief e4c7 design", "tabloio task e4c7 --ref "+hash)
	}
}

func TestT13CrossProjectRequirement(t *testing.T) {
	base := load(t, "e4c7-design")
	out := render(t, variant(base, func(d *Data) {
		d.Requires = []Edge{{ID: "886d", Title: "Status views", URL: "https://github.com/nbyoung/tablo",
			Commit: "0123456789abcdef0123456789abcdef01234567", From: "integrate", To: "implementation",
			Text: "The brief object", Condition: "unmet"}}
	}))
	has(t, flat(out), "Requires: 886d Status views in https://github.com/nbyoung/tablo at 0123456 "+
		"from integrate to implementation, The brief object: unmet")
	has(t, flat(render(t, base)), "Requires: 5ca9 Read commands from implementation to implementation,")
	lacks(t, render(t, base), " in https://")
}

func TestPartsOfRequirements(t *testing.T) {
	base := load(t, "e4c7-design")
	none := render(t, variant(base, func(d *Data) { d.Requires = nil }))
	has(t, none, "  Requires: nothing\n", "  Unblocks: nothing; no task requires e4c7\n")

	other := render(t, variant(base, func(d *Data) {
		d.Dependents = []Edge{{ID: "aaaa", Title: "A", From: "unit", To: "unit", Text: "x", Condition: "pending"}}
	}))
	has(t, other, "  Unblocks: nothing at this gate\n", "  Also required by: aaaa A from unit to unit, x: pending\n")

	now := render(t, variant(base, func(d *Data) {
		d.Dependents = []Edge{{ID: "bbbb", Title: "B", From: "design", To: "mockup", Text: "y", Condition: "met"}}
		d.Parent = &Parent{ID: "07e0", Title: "tabloio", Dependents: []Edge{
			{ID: "cccc", Title: "C", From: "design", To: "design", Text: "z", Condition: "met"},
			{ID: "dddd", Title: "D", From: "unit", To: "unit", Text: "w", Condition: "met"},
		}}
	}))
	has(t, now, "  Unblocks: bbbb B at mockup, y\n",
		"  Its parent 07e0 tabloio at design is required by:\n    cccc C at design, z\n")
	lacks(t, now, "dddd", "Also required by")
}

func TestPersonRecordedWithNote(t *testing.T) {
	base := load(t, "e4c7-design")
	out := render(t, variant(base, func(d *Data) {
		d.Status.Note = ""
		d.Status.Reason = "review"
		d.Status.ReasonSymbol = "🔎"
	}))
	has(t, out, "\n4. The status\n\n  🔧 function 🟢 nominal 🔎 review, 2026-09-30, noreply@anthropic.com, baab0e5\n\n")
}

// outputs returns every brief the tests above render.
func outputs(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, n := range []string{"e4c7-design", "e9c6-implementation", "c74a-mockup"} {
		d := load(t, n)
		out[n] = render(t, d)
		out[n+"/defined"] = render(t, variant(d, func(d *Data) { d.Gate.Key = "defined" }))
		out[n+"/handoff"] = render(t, variant(d, func(d *Data) { d.Junction.Reviewer.Value = "r@example.org" }))
		out[n+"/person"] = render(t, variant(d, func(d *Data) { d.Junction.Model, d.Junction.Reviewer = Field{}, Field{} }))
		out[n+"/complete"] = render(t, variant(d, func(d *Data) { d.Gate.Follows = "" }))
		out[n+"/recursive"] = render(t, variant(d, func(d *Data) { d.Gate.Follows = "recursive" }))
		out[n+"/head"] = render(t, variant(d, func(d *Data) { d.Ref.Name = "HEAD" }))
		out[n+"/folded"] = render(t, variant(d, func(d *Data) { d.Dependents = edges(12, 0, "a") }))
		out[n+"/junction"] = render(t, variant(d, func(d *Data) {
			d.Junction.References = []Reference{{URL: "docs/gates/g.md", Text: "Standing text", Source: Source{Kind: "task", ID: "07e0", Title: "tabloio"}}}
		}))
		out[n+"/cross"] = render(t, variant(d, func(d *Data) {
			d.Requires = []Edge{{ID: "886d", Title: "Status views", URL: "https://example.org/tablo", Commit: "0123456789abcdef", From: "a", To: "b", Text: "t", Condition: "unmet"}}
		}))
	}
	return out
}

// literal reports whether a line is one the text never wraps and so may run
// past the width: a part 1 row, a reference line, a literal block line or a
// command.
func literal(l string) bool {
	for _, p := range []string{
		"  Task ", "  Gate ", "  Contributor ", "  Model ", "  Reviewer ", "  Reference: ", "  Junction reference: ",
		"    export ", "    git log ", "    Reviewed: ", "    Model: ", "    Co-Authored-By: ",
		"    gate: ", "    state: ", "    reason: ", "    note: ", "  tabloio ",
	} {
		if strings.HasPrefix(l, p) {
			return true
		}
	}
	return false
}

func TestT14Shape(t *testing.T) {
	for name, out := range outputs(t) {
		if !strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\n\n") {
			t.Errorf("%s: the text does not end in exactly one newline", name)
		}
		if strings.Contains(out, "\r") {
			t.Errorf("%s: the text holds a carriage return", name)
		}
		if !utf8.ValidString(out) {
			t.Errorf("%s: the text is not UTF-8", name)
		}
		for i, l := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
			if strings.HasSuffix(l, " ") {
				t.Errorf("%s: line %d ends in a space: %q", name, i+1, l)
			}
			if n := utf8.RuneCountInString(l); n > width && !literal(l) {
				t.Errorf("%s: line %d has %d runes: %q", name, i+1, n, l)
			}
		}
	}
	d := load(t, "e9c6-implementation")
	if a, b := render(t, d), render(t, d); a != b {
		t.Error("two renderings of one brief differ")
	}
}

func TestShort(t *testing.T) {
	for in, want := range map[string]string{"": "", "abc": "abc", "abcdefg": "abcdefg", "abcdefgh1234": "abcdefg"} {
		if got := short(in); got != want {
			t.Errorf("short(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWay(t *testing.T) {
	mk := func(gate, contributor, reviewer string) Data {
		var d Data
		d.Gate.Key = gate
		d.Junction.Contributor.Value = contributor
		d.Junction.Reviewer.Value = reviewer
		return d
	}
	for _, c := range []struct {
		d    Data
		want string
	}{
		{mk("defined", "a", "b"), "defined"},
		{mk("defined", "a", "a"), "defined"},
		{mk("design", "a", "b"), "handoff"},
		{mk("design", "a", "a"), "self"},
		{mk("design", "a", ""), "none"},
	} {
		if got := way(c.d); got != c.want {
			t.Errorf("way(%s, %s, %s) = %q, want %q", c.d.Gate.Key, c.d.Junction.Contributor.Value, c.d.Junction.Reviewer.Value, got, c.want)
		}
	}
}
