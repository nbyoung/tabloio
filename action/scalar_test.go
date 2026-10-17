package action

import (
	"os"
	"strings"
	"testing"
)

// noteOf20261005 is the note whose colon and space the scalar rule quotes.
const noteOf20261005 = "prototype/171b/ shows live refresh: git-polling watcher"

// TestNoteWithColonAndSpace is T4: the note of 2026-10-05 lands in double
// quotes, and every row of scalars.tsv gives its block and flow form.
func TestNoteWithColonAndSpace(t *testing.T) {
	t.Run("the note of 2026-10-05", func(t *testing.T) {
		got, err := editStatus("gate: function\nstate: nominal\n", map[string]*string{"note": ptr(noteOf20261005)})
		if want := "gate: function\nstate: nominal\nnote: \"" + noteOf20261005 + "\"\n"; err != nil || got != want {
			t.Errorf("got %q, %v; want %q", got, err, want)
		}
	})
	t.Run("through Record", testRecordNote)
	t.Run("scalars.tsv", func(t *testing.T) {
		b, err := os.ReadFile("testdata/scalars.tsv")
		if err != nil {
			t.Fatal(err)
		}
		rows := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
		if rows[0] != "value\tblock\tflow" || len(rows) != 35 {
			t.Fatalf("scalars.tsv holds %d lines under %q, want 34 rows", len(rows)-1, rows[0])
		}
		for _, row := range rows[1:] {
			f := strings.Split(row, "\t")
			if len(f) != 3 {
				t.Fatalf("malformed row %q", row)
			}
			if got := scalar(f[0]); got != f[1] {
				t.Errorf("block %q: got %s, want %s", f[0], got, f[1])
			}
			if got := flowScalar(f[0]); got != f[2] {
				t.Errorf("flow %q: got %s, want %s", f[0], got, f[2])
			}
		}
	})
}

func ptr(s string) *string { return &s }
