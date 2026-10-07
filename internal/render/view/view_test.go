package view

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture reads testdata/<view>/<name>.json of the render package.
func fixture(t *testing.T, view, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "testdata", view, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// T1: Decode reads each view, ignores an unknown key, rejects an unknown
// level and an unknown view name.
func TestDecode(t *testing.T) {
	cases := []struct{ view, name string }{
		{"gates", "draft"},
		{"task", "e9c6"},
	}
	for _, c := range cases {
		t.Run(c.view+"/"+c.name, func(t *testing.T) {
			data := fixture(t, c.view, c.name)
			v, err := Decode(c.view, data)
			if err != nil {
				t.Fatal(err)
			}
			h := v.Header()
			if h.Project.ID == "" || h.Ref.Commit == "" || len(h.Legend.Gates) == 0 {
				t.Errorf("head not decoded: %+v", h.Project)
			}

			extra := bytes.Replace(data, []byte("{"), []byte(`{"unknown_key": {"a": [1, 2]}, `), 1)
			w, err := Decode(c.view, extra)
			if err != nil {
				t.Fatalf("an unknown key fails: %v", err)
			}
			if w.Header().Project != h.Project || w.Header().Level != h.Level {
				t.Errorf("an unknown key changes the head")
			}

			deep := bytes.Replace(data, []byte(`"level": "`), []byte(`"level": "deep`), 1)
			if bytes.Equal(deep, data) {
				t.Fatal("the fixture has no level to corrupt")
			}
			if _, err := Decode(c.view, deep); err == nil || !strings.Contains(err.Error(), "deep") {
				t.Errorf("an unknown level gives %v", err)
			}
		})
	}
	if _, err := Decode("nope", []byte("{}")); !errors.Is(err, ErrUnknown) {
		t.Errorf("an unknown view gives %v", err)
	}
	if _, err := Decode("gates", []byte("[")); err == nil {
		t.Error("malformed JSON passes")
	}
}

func TestGatesFields(t *testing.T) {
	v, err := Decode("gates", fixture(t, "gates", "draft"))
	if err != nil {
		t.Fatal(err)
	}
	g := v.(*Gates)
	if g.Level != Detail || g.Project.Title != "Tableaux tooling" || g.Ref.Name != "main" || !g.Ref.OnTrunk {
		t.Errorf("head: %+v %+v", g.Level, g.Ref)
	}
	if len(g.Legend.Gates) != 10 || g.Legend.Gates[3].Symbol != "🔧" || g.Legend.Gates[5].Criteria == "" {
		t.Errorf("gates: %+v", g.Legend.Gates)
	}
	if s := g.Legend.States; len(s) != 5 || s[3].Severity != 3 || s[3].Key != "stalled" {
		t.Errorf("states: %+v", s)
	}
	if r := g.Legend.Reasons; len(r) != 3 || !r[2].Reserved || r[0].Reserved {
		t.Errorf("reasons: %+v", r)
	}
	if m := g.Legend.Marks; len(m) != 5 || m[4].Key != "exempt" || m[4].Symbol != "—" {
		t.Errorf("marks: %+v", m)
	}
}

func TestLevel(t *testing.T) {
	for i, w := range []string{"glance", "detail", "provenance"} {
		l, err := ParseLevel(w)
		if err != nil || l != Level(i) || l.String() != w {
			t.Errorf("%s: %v %v", w, l, err)
		}
		b, err := l.MarshalText()
		if err != nil || string(b) != w {
			t.Errorf("marshal %s: %s %v", w, b, err)
		}
	}
	if _, err := ParseLevel("deep"); err == nil {
		t.Error("deep parses")
	}
	if got := Level(7).String(); got != "level(7)" {
		t.Errorf("out of range: %s", got)
	}
	if _, err := Level(7).MarshalText(); err == nil {
		t.Error("out of range marshals")
	}
}
