package render

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// readIn reads the fixture testdata/<view>/in/<name>.json. The temporal
// fixtures stand in their own directory: the tests of design e3ed that run
// for every fixture of testdata/*/*.json do not read them, and the tests of
// this file and of audit_test.go run in their stead.
func readIn(t testing.TB, viewName, name string) view.View {
	t.Helper()
	return readFixture(t, viewName, filepath.Join("in", name))
}

// checkGolden renders a fixture at a level and compares the bytes with
// testdata/<view>/<name>.<level>.md; with -update it rewrites the file.
func checkGolden(t *testing.T, viewName, name string, level view.Level, o Options) {
	t.Helper()
	o.Level = level
	got := renderString(t, readIn(t, viewName, name), o)
	file := filepath.Join("testdata", viewName, name+"."+level.String()+".md")
	if *update {
		if err := os.WriteFile(file, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("%s differs from its golden file\n--- got ---\n%s\n--- want ---\n%s", file, got, want)
	}
}

// historyGoldens lists the golden cases of the history: T1 to T5 and T11.
var historyGoldens = []golden{
	// T1 and T3
	{view: "history", name: "tooling-e9c6", levels: allLevels},
	// T2 and T3
	{view: "history", name: "tooling-range", levels: allLevels},
	// T4
	{view: "history", name: "first-pin", levels: allLevels},
	{view: "history", name: "reaffirmed", levels: glanceDetail},
	{view: "history", name: "proposal", levels: glanceDetail},
	// T5
	{view: "history", name: "weather-all", levels: []view.Level{view.Glance}},
	{view: "history", name: "weather-9f31", levels: glanceDetail},
	// T11
	{view: "history", name: "empty", levels: allLevels},
}

func TestHistoryGolden(t *testing.T) {
	for _, g := range historyGoldens {
		for _, level := range g.levels {
			t.Run(g.name+"."+level.String(), func(t *testing.T) {
				checkGolden(t, g.view, g.name, level, Options{})
			})
		}
	}
}

// temporalFixtures lists the fixtures of the two temporal views.
func temporalFixtures(t testing.TB) []fixtureName {
	t.Helper()
	var out []fixtureName
	for _, v := range []string{"history", "audit"} {
		paths, err := filepath.Glob(filepath.Join("testdata", v, "in", "*.json"))
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range paths {
			out = append(out, fixtureName{view: v, name: strings.TrimSuffix(filepath.Base(p), ".json")})
		}
	}
	if len(out) == 0 {
		t.Fatal("no temporal fixture")
	}
	return out
}

// headingsOf lists the headings of a document, in order.
func headingsOf(d doc.Doc) []string {
	var out []string
	for _, b := range d.Blocks {
		if h, ok := b.(doc.Heading); ok {
			out = append(out, plainInline(h.Text, map[string]bool{}))
		}
	}
	return out
}

// T6, T7 and T10 of design e3ed for the temporal fixtures. The levels nest by
// their headings: the history's detail replaces the glance's line table with
// lists by day, and the audit's detail table holds other columns than the
// glance's merged rows, so the columns do not nest as they do in the
// tables of the two definition views. Each rune outside ASCII occurs in the
// fixture or is one of the renderer's own. The same data and options give
// the same bytes, whatever the order of Links.Views.
func TestTemporalFixtures(t *testing.T) {
	const own = "·…→›—×"
	forward := map[string]string{"gates": "GATES.md", "history": "HISTORY.md", "audit": "AUDIT.md"}
	backward := map[string]string{"audit": "AUDIT.md", "history": "HISTORY.md", "gates": "GATES.md"}
	for _, f := range temporalFixtures(t) {
		t.Run(f.String(), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", f.view, "in", f.name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			v := readIn(t, f.view, f.name)
			var prev []string
			for _, l := range levelsOf(v) {
				heads := headingsOf(buildFor(t, v, Options{Level: l}))
				if l > view.Glance && !isSubsequence(prev, heads) {
					t.Errorf("%s does not hold the headings of %s:\n%v\n%v", l, l-1, heads, prev)
				}
				prev = heads
				out := renderString(t, v, Options{Level: l, Links: Links{Project: ".", Views: forward}})
				for _, r := range out {
					if r < 0x80 || strings.ContainsRune(own, r) || strings.ContainsRune(string(raw), r) {
						continue
					}
					t.Errorf("%s: the rune %q (U+%04X) is in the output and not in the fixture", l, r, r)
				}
				if again := renderString(t, v, Options{Level: l, Links: Links{Project: ".", Views: backward}}); again != out {
					t.Errorf("%s: two renderings differ", l)
				}
				if again := renderString(t, v, Options{Level: l, Links: Links{Project: ".", Views: forward}}); again != out {
					t.Errorf("%s: a repeated rendering differs", l)
				}
			}
		})
	}
}

// statusEvent is a status event that records defined, nominal.
func statusEvent(date, by string) view.Event {
	return view.Event{Date: date, By: by, Kind: "status", Recorded: &view.EventStatus{Gate: "defined", State: "nominal"}}
}

// historyOf makes a history of n status lines, one commit each, spread over
// the given number of days.
func historyOf(n, days int) *view.History {
	v := &view.History{}
	v.Level = view.Provenance
	v.Project = view.Project{Root: "437e", Title: "Tableaux tooling"}
	v.Subject = view.TaskRef{ID: "437e", Title: "Tableaux tooling"}
	v.Source = view.Source{Ref: "main", Commit: "3cdae527d47d37fdff55bc302c76c017cf91c606", OnTrunk: true}
	v.Legend = view.Legend{
		Gates:  []view.Gate{{Key: "defined", Symbol: "📝"}},
		States: []view.State{{Key: "nominal", Symbol: "🟢"}},
	}
	perDay := make([]int, days)
	for i := 0; i < n; i++ {
		d := i * days / n
		perDay[d]++
		v.Lines = append(v.Lines, view.Line{
			Task:   view.TaskRef{ID: fmt.Sprintf("%04x", i)},
			Events: []view.Event{statusEvent(fmt.Sprintf("2026-09-%02d", d+1), "a@example.org")},
		})
	}
	v.Events, v.Commits = n, n
	v.Kinds = []view.KindCount{{Kind: "status", Count: n}}
	for d, c := range perDay {
		v.Days = append(v.Days, view.Day{
			Date: fmt.Sprintf("2026-09-%02d", d+1), Events: c, Commits: c,
			Kinds: []view.KindCount{{Kind: "status", Count: c}},
		})
	}
	return v
}

// T6: the day table stands from three days, and the glance folds its lines
// above fifty, apart from the run fold of alike rows.
func TestHistoryDayFold(t *testing.T) {
	const dayTable, lineTable = "| Day | Events | Commits | Kinds |", "| Date | By | Event | Task | Gate and state, or pin |"
	cases := []struct {
		name         string
		n, days      int
		o            Options
		day, lines   bool
		fold         string
		items, trail int
	}{
		{"two days", 6, 2, Options{Level: view.Glance}, false, true, "", 0, 0},
		{"three days", 6, 3, Options{Level: view.Glance}, true, true, "", 0, 0},
		{"fifty lines", 50, 4, Options{Level: view.Glance}, true, true, "", 0, 0},
		{"fifty-one lines", 51, 4, Options{Level: view.Glance}, true, false, "**51 lines fold into the days above.**", 0, 0},
		{"fifty-one lines in two days", 51, 2, Options{Level: view.Glance}, true, false, "**51 lines fold into the days above.**", 0, 0},
		{"fifty-one with Unfold", 51, 4, Options{Level: view.Glance, Unfold: true}, true, true, "", 0, 0},
		{"fifty-one at detail", 51, 4, Options{Level: view.Detail}, false, false, "", 51, 0},
		{"fifty-one at provenance", 51, 4, Options{Level: view.Provenance}, false, false, "", 51, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := renderString(t, historyOf(c.n, c.days), c.o)
			if got := strings.Contains(out, dayTable); got != c.day {
				t.Errorf("day table: got %v, want %v", got, c.day)
			}
			if got := strings.Contains(out, lineTable); got != c.lines {
				t.Errorf("line table: got %v, want %v", got, c.lines)
			}
			if got := strings.Contains(out, "fold into the days above"); got != (c.fold != "") {
				t.Errorf("fold sentence: got %v", got)
			}
			if c.fold != "" && !strings.Contains(out, "\n"+c.fold+"\n") {
				t.Errorf("the fold sentence is not %q", c.fold)
			}
			if c.items > 0 {
				if got := strings.Count(out, "\n- **2026-09-"); got != c.items {
					t.Errorf("%d items, want %d", got, c.items)
				}
			}
			if strings.Contains(out, "more:") {
				t.Errorf("the run fold applies to the history:\n%s", out)
			}
		})
	}
}

// T7: the words of each effect key and each model reading.
func TestHistoryWords(t *testing.T) {
	render := func(l view.Line) string {
		v := historyOf(1, 1)
		v.Lines = []view.Line{l}
		return renderString(t, v, Options{Level: view.Detail})
	}
	effects := map[string]string{
		"authorised_commit":  "an authority commits the task file on the trunk, so the task is authorised",
		"authorised_merge":   "an authority merges the change, so the task is authorised",
		"authorised_trailer": "an authority's trailer authorises the task",
		"proposed":           "no authority accepts the change, so the task stands proposed",
		"handoff":            "the contributor hands the work to the reviewer",
		"accepts":            "the reviewer accepts the work at the gate",
		"stands":             "none; the authorisation stands as the review",
		"none":               "none",
		"unheard-of":         "unheard-of",
	}
	for key, words := range effects {
		l := view.Line{Events: []view.Event{statusEvent("2026-09-01", "")}, Effect: key}
		if out := render(l); !strings.Contains(out, "\n  - Effect: "+words+"\n") {
			t.Errorf("effect %q does not read %q:\n%s", key, words, out)
		}
	}
	if out := render(view.Line{Events: []view.Event{statusEvent("2026-09-01", "")}}); strings.Contains(out, "Effect:") || strings.Contains(out, "Model:") {
		t.Errorf("a line with no effect and no model prints them:\n%s", out)
	}
	readings := []struct {
		name string
		l    view.Line
		want string
	}{
		{"agrees", view.Line{Check: &view.ModelCheck{Reading: "agrees", Stated: "m", Gate: "defined"}, Model: "m-1"}, "`m-1`; the 📝 defined junction states `m`"},
		{"differs", view.Line{Check: &view.ModelCheck{Reading: "differs", Stated: "m", Gate: "defined"}, Model: "x-1"}, "`x-1`; the 📝 defined junction states `m`, which differs"},
		{"missing", view.Line{Check: &view.ModelCheck{Reading: "missing", Stated: "m", Gate: "defined"}}, "no trailer; the 📝 defined junction states `m`"},
		{"exempt", view.Line{Check: &view.ModelCheck{Reading: "exempt"}}, "none recorded; exempt"},
		{"unstated", view.Line{Check: &view.ModelCheck{Reading: "unstated", Gate: "defined"}, Model: "m-1"}, "`m-1`; the junction states no model"},
		{"a trailer and no check", view.Line{Model: "m-1"}, "`m-1`"},
		{"a stated model and no gate", view.Line{Check: &view.ModelCheck{Reading: "agrees", Stated: "m"}, Model: "m-1"}, "`m-1`; the junction states `m`"},
		{"an unknown reading", view.Line{Check: &view.ModelCheck{Reading: "puzzling"}, Model: "m-1"}, "`m-1`"},
	}
	for _, c := range readings {
		t.Run(c.name, func(t *testing.T) {
			c.l.Events = []view.Event{statusEvent("2026-09-01", "")}
			if out := render(c.l); !strings.Contains(out, "\n  - Model: "+c.want+"\n") {
				t.Errorf("want %q:\n%s", c.want, out)
			}
		})
	}
}

// TestHistoryCells covers the cells of a line no fixture holds: the last cell
// takes the first of its candidates, a person is bold, and a time prints as
// Git prints it.
func TestHistoryCells(t *testing.T) {
	v := historyOf(1, 1)
	v.Params.Person = "a@example.org"
	ev := func(by, kind string, mod func(*view.Event)) view.Event {
		e := view.Event{Date: "2026-09-01", By: by, Kind: kind}
		if mod != nil {
			mod(&e)
		}
		return e
	}
	status := func(e *view.Event) { e.Recorded = &view.EventStatus{Gate: "defined"} }
	review := func(e *view.Event) { e.Gate = "defined" }
	pin := func(e *view.Event) { e.Pin = &view.PinMove{URL: "u", New: "0123456789"} }
	v.Lines = []view.Line{
		{Task: view.TaskRef{ID: "aaaa"}, Events: []view.Event{
			ev("a@example.org", "status", status), ev("a@example.org", "reviewed", review), ev("a@example.org", "pin", pin)}},
		{Task: view.TaskRef{ID: "bbbb"}, Events: []view.Event{ev("b@example.org", "reviewed", review), ev("b@example.org", "pin", pin)}},
		{Task: view.TaskRef{ID: "cccc"}, Events: []view.Event{ev("b@example.org", "pin", func(e *view.Event) {
			e.Pin = &view.PinMove{URL: "u", Old: "abcdefabcdef", New: "0123456789"}
		})}},
		{Task: view.TaskRef{ID: "dddd"}, Events: []view.Event{ev("b@example.org", "authorised", nil)}},
	}
	out := renderString(t, v, Options{Level: view.Glance})
	for _, want := range []string{
		"| 2026-09-01 | **a@example.org** | status, reviewed, pin | `aaaa` | 📝 defined |\n",
		"| 2026-09-01 | b@example.org | reviewed, pin | `bbbb` | 📝 defined |\n",
		"| 2026-09-01 | b@example.org | pin | `cccc` | `u` `abcdefa` → `0123456` |\n",
		"| 2026-09-01 | b@example.org | authorised | `dddd` |  |\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the table lacks %q:\n%s", want, out)
		}
	}
	for in, want := range map[string]string{
		"2026-09-29T16:38:16-04:00": "2026-09-29 16:38:16 -0400",
		"2026-09-29T16:38:16Z":      "2026-09-29 16:38:16 +0000",
		"last Tuesday":              "last Tuesday",
	} {
		if got := gitTime(in); got != want {
			t.Errorf("gitTime(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestHistoryTaken: the subject titles the history, the root by default
// (8ed1 decision 5); a line taken from a subproject nests beneath the pin of
// that subproject that follows it, which counts its events, and stands in
// its place with no such pin (8ed1 decision 2).
func TestHistoryTaken(t *testing.T) {
	v := historyOf(1, 1)
	from := &view.SubEvent{URL: "firmware", Task: "f1a0"}
	took := func(commit string) view.Event {
		e := statusEvent("2026-09-01", "b@example.org")
		e.Commit, e.Subproject = commit, from
		return e
	}
	pin := view.Event{Date: "2026-09-02", By: "b@example.org", Kind: "pin", Pin: &view.PinMove{URL: "firmware", New: "0123456789"}}
	v.Lines = []view.Line{
		{Task: view.TaskRef{ID: "c07d"}, Events: []view.Event{took("c1"), took("c1")}},
		{Task: view.TaskRef{ID: "c07d"}, Events: []view.Event{took("c2")}},
		{Task: view.TaskRef{ID: "c07d"}, Events: []view.Event{pin}},
		{Task: view.TaskRef{ID: "c07d"}, Events: []view.Event{took("c3")}},
	}
	out := renderString(t, v, Options{Level: view.Detail})
	for _, want := range []string{
		"# History: `437e` Tableaux tooling\n",
		"- **2026-09-02** · b@example.org · pin · `c07d` · `firmware` new `0123456`\n" +
			"  - Subproject events: 3 events in 2 commits of `firmware`\n" +
			"    - 2026-09-01 · b@example.org · status, status · `f1a0` · 📝 defined 🟢 nominal\n" +
			"    - 2026-09-01 · b@example.org · status · `f1a0` · 📝 defined 🟢 nominal\n",
		"- **2026-09-01** · b@example.org · status · `c07d`, taken from `firmware` `f1a0` · 📝 defined 🟢 nominal\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output lacks %q:\n%s", want, out)
		}
	}
	if n := strings.Count(out, "taken from"); n != 1 {
		t.Errorf("%d taken lines stand in their place, want 1:\n%s", n, out)
	}
}

// TestHistoryLevel: a history the data holds at glance only refuses detail.
func TestHistoryLevel(t *testing.T) {
	v := readIn(t, "history", "weather-all")
	if err := Render(io.Discard, v, Options{Level: view.Detail}); !errors.Is(err, ErrLevel) {
		t.Errorf("got %v, want ErrLevel", err)
	}
}
