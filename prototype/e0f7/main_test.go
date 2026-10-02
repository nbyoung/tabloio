package main

import (
	"os"
	"strings"
	"testing"
)

func TestWidthSymbols(t *testing.T) {
	cases := []struct {
		s    string
		m    Model
		want int
	}{
		{"⚡", VS16Wide, 2},   // emoji presentation by default
		{"⚓", VS16Narrow, 2}, // likewise
		{"🟢", VS16Wide, 2},
		{"🧑", VS16Wide, 2},
		{"⚙", VS16Wide, 1},    // text form, no selector
		{"⚙️", VS16Wide, 2},   // selector widens
		{"⚙️", VS16Narrow, 1}, // legacy: selector is zero width
		{"⚙️", VS16Strip, 1},  // selector removed
		{"🛠️", VS16Wide, 2},
		{"🛠️", VS16Narrow, 1},
		{"⚡️", VS16Wide, 2}, // selector on a wide base adds nothing
		{"🧑👀", VS16Wide, 4},
		{"—", VS16Wide, 1},
		{"│─", VS16Wide, 2},
	}
	for _, c := range cases {
		if got := Width(c.s, c.m); got != c.want {
			t.Errorf("Width(%q, %d) = %d, want %d", c.s, c.m, got, c.want)
		}
	}
}

func load(t *testing.T) *Tableau {
	data, err := os.ReadFile("testdata/global-tableau.json")
	if err != nil {
		t.Fatal(err)
	}
	tb, err := ParseTableau(data)
	if err != nil {
		t.Fatal(err)
	}
	return tb
}

// Every line of the table has the same display width, under each model,
// at each terminal width the columns allow.
func TestRowsAlign(t *testing.T) {
	tb := load(t)
	for _, m := range []Model{VS16Wide, VS16Narrow, VS16Strip} {
		for _, width := range []int{120, 100, 80} {
			out := RenderTableau(tb, width, m)
			var want int
			for i, ln := range strings.Split(strings.TrimSpace(out), "\n") {
				if strings.HasPrefix(ln, "folded") {
					break
				}
				w := Width(ln, m)
				if i == 0 {
					want = w
				}
				if w != want {
					t.Errorf("model %d width %d line %d: width %d, want %d\n%s", m, width, i, w, want, out)
				}
				if w > width {
					t.Errorf("model %d width %d: line is %d wide", m, width, w)
				}
			}
		}
	}
}

func TestShrinksTaskColumn(t *testing.T) {
	out := RenderTableau(load(t), 60, VS16Wide)
	if !strings.Contains(out, "…") {
		t.Errorf("no ellipsis at width 60:\n%s", out)
	}
}

func TestIndentsTree(t *testing.T) {
	out := RenderTableau(load(t), 120, VS16Wide)
	if !strings.Contains(out, "│   Sensor node") || !strings.Contains(out, "│     Sensor board") {
		t.Errorf("tree indentation missing:\n%s", out)
	}
}

func TestStripRemovesSelector(t *testing.T) {
	if strings.ContainsRune(RenderTableau(load(t), 120, VS16Strip), vs16) {
		t.Error("selector survives under strip")
	}
}
