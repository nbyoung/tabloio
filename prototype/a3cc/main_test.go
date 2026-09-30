package main

import (
	"os"
	"strings"
	"testing"
)

func render(t *testing.T, file, level string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + file)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(data, level)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func wantAll(t *testing.T, out string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(out, s) {
			t.Errorf("output lacks %q:\n%s", s, out)
		}
	}
}

func TestGlobalDetail(t *testing.T) {
	out := render(t, "global-tableau.json", "detail")
	wantAll(t, out, "## Global tableau on `main`", "**Sensor node**", "🔴⛔", "🖼️…🚀 ×0",
		"rolls up from `9f31`", "subproject at `F2`", "&nbsp;&nbsp;&nbsp;&nbsp;Node firmware")
	// every table row has the same number of cells as the header
	var want int
	for _, l := range strings.Split(out, "\n") {
		if !strings.HasPrefix(l, "|") {
			continue
		}
		n := strings.Count(l, " | ")
		if want == 0 {
			want = n
		} else if strings.Contains(l, "---") {
			continue
		} else if n != want {
			t.Errorf("row has %d separators, header %d: %s", n, want, l)
		}
	}
}

func TestGlobalGlance(t *testing.T) {
	out := render(t, "global-tableau.json", "glance")
	wantAll(t, out, "Sensor node**", "(2)")
	if strings.Contains(out, "Node firmware") {
		t.Error("glance shows a depth-two row")
	}
}

func TestWindowFold(t *testing.T) {
	out := render(t, "global-tableau-window0.json", "detail")
	wantAll(t, out, "❔ ×1", "🧩…🚀 ×0")
}

func TestContextual(t *testing.T) {
	wantAll(t, render(t, "contextual-task-4e2b.json", "detail"), "task `4e2b`")
	wantAll(t, render(t, "contextual-person-ben.json", "detail"), "ben@example.org", "Node firmware ◀")
}

func TestBlockage(t *testing.T) {
	out := render(t, "work-blockage-tree.json", "detail")
	wantAll(t, out, "holds 2", "`c07d` Node firmware at 🛠️ implementation", "not yet due", "requires `7b2e`")
	g := render(t, "work-blockage-tree.json", "glance")
	if strings.Contains(g, "action:") {
		t.Error("glance shows the tree")
	}
}

func TestQueue(t *testing.T) {
	wantAll(t, render(t, "queue-ada.json", "glance"), "authorisation owed", "work waiting", "blocked: Barometer")
	wantAll(t, render(t, "queue-ben.json", "detail"), "queue is empty")
	wantAll(t, render(t, "queue-dan.json", "detail"), "`3c5d` Dashboard", "reaffirmation")
}

func TestUnknownView(t *testing.T) {
	if _, err := Render([]byte(`{"view":"x"}`), "detail"); err == nil {
		t.Error("want an error")
	}
}
