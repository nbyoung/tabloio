package render

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/view"
)

// tableauGoldens lists the golden cases of the two tableaux (T1 to T5).
var tableauGoldens = []statusGolden{
	// T1
	{"tableau", "tooling-glance", []view.Level{view.Glance}},
	// T2
	{"context", "tooling-2034", allLevels},
	// T3
	{"tableau", "tooling", []view.Level{view.Detail, view.Provenance}},
	{"context", "tooling-2034-spine", []view.Level{view.Detail}},
	{"context", "tooling-nbyoung", allLevels},
	// T4
	{"tableau", "tooling-historical", []view.Level{view.Detail}},
	{"tableau", "tooling-person", []view.Level{view.Detail}},
	{"tableau", "tooling-window0", []view.Level{view.Detail}},
	{"tableau", "tooling-reason", []view.Level{view.Detail}},
	// T5
	{"tableau", "weather", glanceDetail},
	{"tableau", "weather-window0", []view.Level{view.Detail}},
	{"context", "weather-4e2b", glanceDetail},
	{"context", "weather-ben", glanceDetail},
}

func TestTableauGolden(t *testing.T) { runStatusGoldens(t, tableauGoldens) }

// recoverFailure runs f and returns the error of a builder that failed, as
// registerStatus panics with it.
func recoverFailure(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			bf, ok := r.(buildFailure)
			if !ok {
				panic(r)
			}
			err = bf
		}
	}()
	f()
	return nil
}

func rawFixture(t testing.TB, viewName, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", viewName, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// T1 and T2, the decoding half: both names return one type, and the
// contextual flag follows the name.
func TestDecodeTableau(t *testing.T) {
	data := rawFixture(t, "tableau", "tooling-glance")
	for name, want := range map[string]bool{"tableau": false, "context": true} {
		v, err := view.Decode(name, data)
		if err != nil {
			t.Fatal(err)
		}
		tb, ok := v.(*view.Tableau)
		if !ok {
			t.Fatalf("%s decodes to %T", name, v)
		}
		if tb.Contextual != want {
			t.Errorf("%s: Contextual %v, want %v", name, tb.Contextual, want)
		}
		if len(tb.Columns) != 10 || len(tb.Rows) != 6 {
			t.Errorf("%s: %d columns, %d rows", name, len(tb.Columns), len(tb.Rows))
		}
	}
}

// T7: a row with the wrong number of cells is an error that names the row.
func TestGridRowCells(t *testing.T) {
	// A fixture with a cell removed, through Decode.
	var raw map[string]any
	if err := json.Unmarshal(rawFixture(t, "tableau", "tooling-glance"), &raw); err != nil {
		t.Fatal(err)
	}
	rows := raw["rows"].([]any)
	row := rows[2].(map[string]any)
	cells := row["cells"].([]any)
	row["cells"] = cells[:len(cells)-1]
	bad, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tableau", "context"} {
		_, err := view.Decode(name, bad)
		if err == nil || !strings.Contains(err.Error(), `"77b2"`) {
			t.Errorf("Decode(%s): %v, want an error that names 77b2", name, err)
		}
	}

	// A value built by hand reaches the builder, which reports the row too,
	// and Render passes it on as the panic of registerStatus.
	v := readFixture(t, "tableau", "tooling-glance").(*view.Tableau)
	cp := *v
	cp.Rows = slices.Clone(v.Rows)
	cp.Rows[3].Cells = append(slices.Clone(cp.Rows[3].Cells), view.GridCell{Kind: "empty"})
	if _, err := buildTableau(&cp, Options{Level: view.Glance}); err == nil || !strings.Contains(err.Error(), `"6103"`) {
		t.Errorf("buildTableau: %v, want an error that names 6103", err)
	}
	err = recoverFailure(func() { _ = Render(&bytes.Buffer{}, &cp, Options{Level: view.Glance}) })
	if err == nil || !strings.Contains(err.Error(), `"6103"`) {
		t.Errorf("Render: %v, want an error that names 6103", err)
	}
}

// notesOf returns the lines of the notes table of a rendering.
func notesOf(t *testing.T, v *view.Tableau, level view.Level) []string {
	t.Helper()
	return section(renderString(t, v, Options{Level: level}), "## Notes")
}

// T6: the notes group by date, note and origin, in display order of each
// group's first row, and the section vanishes when empty.
func TestGridNotes(t *testing.T) {
	v := readFixture(t, "tableau", "tooling").(*view.Tableau)

	got := notesOf(t, v, view.Detail)
	if len(got) != 2+15 {
		t.Fatalf("%d lines in the notes table, want 17:\n%s", len(got), strings.Join(got, "\n"))
	}
	// The twenty leaves share one line, placed where 99f0 stands, after the
	// roll-up of 437e and the notes of c2ad and e9c6, and before 5fe3's.
	first := []string{"| `437e` `bc63` `2034` `bc86` |", "| `c2ad` |", "| `e9c6` |", "| `99f0` `05a9` `a9ce`", "| `5fe3` |"}
	for i, want := range first {
		if !strings.HasPrefix(got[2+i], want) {
			t.Errorf("line %d reads %q, want it to start with %q", i, got[2+i], want)
		}
	}
	if !strings.Contains(got[2+3], "`a8b4`") {
		t.Errorf("the twenty leaves do not share a line: %s", got[2+3])
	}
	if !strings.HasSuffix(got[2], "| rolled up from `99f0` Gate definition view in Markdown |") {
		t.Errorf("the roll-up line reads %q", got[2])
	}
	if !strings.HasSuffix(got[2+3], "|  |") {
		t.Errorf("a status file leaves From empty: %q", got[2+3])
	}

	// A row with neither a date nor a note gives no line.
	cp := *v
	cp.Rows = slices.Clone(v.Rows)
	cp.Rows[2].Status.Date, cp.Rows[2].Status.Note = "", "" // c2ad
	if got := notesOf(t, &cp, view.Detail); len(got) != 2+14 {
		t.Errorf("a row with neither gives %d lines, want 16", len(got))
	}

	// Rows with one date and note and another origin stay apart; the same
	// origin joins them, at the place of the first.
	cp.Rows = slices.Clone(v.Rows)
	cp.Rows[0].Status.From = nil // 437e: the leaves' note, date and origin
	got = notesOf(t, &cp, view.Detail)
	if len(got) != 2+15 || !strings.HasPrefix(got[2], "| `437e` `99f0` `05a9`") {
		t.Errorf("a row with the leaves' origin joins them at its own place:\n%s", strings.Join(got, "\n"))
	}
	cp.Rows = slices.Clone(v.Rows)
	cp.Rows[1].Status.From = &view.Origin{Kind: "rollup", Task: view.TaskRef{ID: "zzzz"}} // bc63
	if got := notesOf(t, &cp, view.Detail); len(got) != 2+16 {
		t.Errorf("a differing origin gives %d lines, want 18", len(got))
	}

	// A fixture with no note at all prints no section.
	none := readFixture(t, "tableau", "tooling-reason").(*view.Tableau)
	out := renderString(t, none, Options{Level: view.Detail})
	if strings.Contains(out, "## Notes") {
		t.Errorf("a tableau with no note prints a Notes section:\n%s", out)
	}
}

// The grid's cells by kind, the header of a folded run, the person's
// brackets, the file links and the folded columns section.
func TestGridCells(t *testing.T) {
	v := readFixture(t, "tableau", "tooling-reason").(*view.Tableau)
	out := renderString(t, v, Options{Level: view.Detail})
	if !strings.Contains(out, "| 🟢👓 | 🤖👀 | — | — | — | — | — | — | — |") {
		t.Errorf("the reason beside the state:\n%s", out)
	}
	if !strings.Contains(out, "🟢 nominal · 👓 review · ") {
		t.Errorf("the key line lacks the reason:\n%s", out)
	}
	if strings.Contains(out, "## Folded columns") && !strings.Contains(out, "| ❔ | ❔ undefined | 0 |") {
		t.Errorf("the folded columns:\n%s", out)
	}

	w0 := readFixture(t, "tableau", "tooling-window0").(*view.Tableau)
	out = renderString(t, w0, Options{Level: view.Detail})
	if !strings.Contains(out, "| ❔…📝 ×29 | 📌 |") || !strings.Contains(out, "| 🚀 ×0 |") {
		t.Errorf("the folded headers:\n%s", out)
	}
	if got := section(out, "## Folded columns"); len(got) != 2+2+1 {
		t.Errorf("%d lines in the folded columns table, want 5:\n%s", len(got), strings.Join(got, "\n"))
	} else if got[2] != "| ❔…📝 | ❔ undefined | 0 |" || got[3] != "| ❔…📝 | 📝 defined | 29 |" {
		t.Errorf("the folded column lines:\n%s", strings.Join(got, "\n"))
	}

	p := readFixture(t, "tableau", "tooling-person").(*view.Tableau)
	out = renderString(t, p, Options{Level: view.Detail})
	if !strings.Contains(out, `\[🤖👀\]`) || !strings.Contains(out, `\[🧑\]`) {
		t.Errorf("the person's cells lack brackets:\n%s", out)
	}
	if !strings.Contains(out, "| `2034` | &nbsp;&nbsp;&nbsp;&nbsp;**Views** |") {
		t.Errorf("a parent is bold and indented by depth:\n%s", out)
	}

	h := readFixture(t, "tableau", "tooling-historical").(*view.Tableau)
	out = renderString(t, h, Options{Level: view.Detail, Links: Links{Project: "."}})
	if !strings.Contains(out, "| [`c2ad`](.tableaux/tasks/c2ad.yaml) |") || !strings.Contains(out, "| 🤖 | — | — | 🤖👀 | 🟢 |") {
		t.Errorf("the historical cells and the file links:\n%s", out)
	}

	// The person with no task, and a hidden count after a title.
	e := *v
	e.Contextual, e.Rows = true, nil
	out = renderString(t, &e, Options{Level: view.Glance})
	if !strings.Contains(out, "**No task stands in this corner.**\n") || strings.Contains(out, "| Id |") {
		t.Errorf("the empty form:\n%s", out)
	}
	g := readFixture(t, "tableau", "tooling-glance")
	if out := renderString(t, g, Options{Level: view.Glance}); !strings.Contains(out, "**Method** (31) |") {
		t.Errorf("the hidden count:\n%s", out)
	}
}

// The contextual title carries its subject.
func TestContextTitle(t *testing.T) {
	task := readFixture(t, "context", "tooling-2034")
	person := readFixture(t, "context", "tooling-nbyoung")
	for v, want := range map[view.View]string{
		task:   "# Contextual tableau: `2034` Views\n",
		person: "# Contextual tableau: nbyoung@nbyoung.com\n",
	} {
		if out := renderString(t, v, Options{Level: view.Glance}); !strings.HasPrefix(out, want) {
			t.Errorf("the title reads %q, want %q", strings.SplitN(out, "\n", 2)[0], want)
		}
	}
	if out := renderString(t, readFixture(t, "context", "tooling-2034-spine"), Options{Level: view.Detail}); !strings.Contains(out, "Roles · sibling |") || !strings.Contains(out, "**Tableaux tooling** · spine |") {
		t.Errorf("the roles:\n%s", out)
	}
}
