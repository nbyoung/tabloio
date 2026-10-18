package action

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// statusCase is one case of status-edits.txt.
type statusCase struct {
	name, before, after string
	edits               map[string]*string
}

// statusCases reads testdata/status-edits.txt: "=== <name>", then the
// sections "--- before", "--- set" (one "key=value" or "key-" per line) and
// "--- after".
func statusCases(t *testing.T) []statusCase {
	t.Helper()
	b, err := os.ReadFile("testdata/status-edits.txt")
	if err != nil {
		t.Fatal(err)
	}
	var cases []statusCase
	for _, chunk := range strings.Split(string(b), "=== ")[1:] {
		name, rest, _ := strings.Cut(chunk, "\n")
		before, rest, ok1 := strings.Cut(strings.TrimPrefix(rest, "--- before\n"), "--- set\n")
		set, after, ok2 := strings.Cut(rest, "--- after\n")
		if !ok1 || !ok2 {
			t.Fatalf("malformed case %q", name)
		}
		c := statusCase{name: name, before: before, after: after, edits: map[string]*string{}}
		for _, l := range strings.Split(strings.TrimSuffix(set, "\n"), "\n") {
			if k, v, ok := strings.Cut(l, "="); ok {
				c.edits[k] = ptr(v)
			} else {
				c.edits[strings.TrimSuffix(l, "-")] = nil
			}
		}
		cases = append(cases, c)
	}
	if len(cases) != 10 {
		t.Fatalf("status-edits.txt holds %d cases, want 10", len(cases))
	}
	return cases
}

// TestRecordChangesLinesOnly is T2: each case of status-edits.txt gives its
// after; a CRLF file and a file with no final newline keep their endings,
// and through Record git diff --numstat counts no line beyond the changed
// fields.
func TestRecordChangesLinesOnly(t *testing.T) {
	for _, c := range statusCases(t) {
		got, err := editStatus(c.before, c.edits)
		if err != nil || got != c.after {
			t.Errorf("%s: got %q, %v; want %q", c.name, got, err, c.after)
		}
	}
	t.Run("CRLF", func(t *testing.T) {
		got, err := editStatus("gate: defined\r\nstate: nominal\r\n", map[string]*string{"gate": ptr("function"), "note": ptr("Shown")})
		if want := "gate: function\r\nstate: nominal\r\nnote: Shown\r\n"; err != nil || got != want {
			t.Errorf("got %q, %v; want %q", got, err, want)
		}
	})
	t.Run("no final newline", func(t *testing.T) {
		got, err := editStatus("gate: defined\nstate: nominal", map[string]*string{"note": ptr("Shown")})
		if want := "gate: defined\nstate: nominal\nnote: Shown\n"; err != nil || got != want {
			t.Errorf("insert: got %q, %v; want %q", got, err, want)
		}
		got, err = editStatus("gate: defined\nstate: nominal", map[string]*string{"gate": ptr("mockup")})
		if want := "gate: mockup\nstate: nominal"; err != nil || got != want {
			t.Errorf("replace: got %q, %v; want %q", got, err, want)
		}
	})
	t.Run("through Record", testRecordLines)
	t.Run("status-form", func(t *testing.T) {
		for _, bad := range []string{"{ gate: design }\n", "gate: a\ngate: b\n", "---\ngate: a\n", "gate: a\nextra: 1\n"} {
			var form *statusFormError
			if _, err := editStatus(bad, map[string]*string{"gate": ptr("x")}); !errors.As(err, &form) {
				t.Errorf("%q: error %v, want a status-form error", bad, err)
			}
		}
	})
}
