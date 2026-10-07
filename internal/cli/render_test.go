package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/view"
)

// views lists the ten view commands in the order of VIEWS.md, each with a
// view data file of package render's testdata.
var views = []struct{ name, data string }{
	{"gates", "gates/tooling.json"},
	{"task", "task/e9c6.json"},
	{"authority", "authority/tooling.json"},
	{"assignment", "assignment/tooling.json"},
	{"queue", "queue/tooling-owner.json"},
	{"blockage", "blockage/tooling.json"},
	{"tableau", "tableau/tooling.json"},
	{"context", "context/tooling-2034.json"},
	{"history", "history/in/tooling-e9c6.json"},
	{"audit", "audit/in/tooling.json"},
}

func decode(t *testing.T, name, file string) view.View {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "render", "testdata", filepath.FromSlash(file)))
	if err != nil {
		t.Fatal(err)
	}
	v, err := view.Decode(name, b)
	if err != nil {
		t.Fatalf("decode %s: %v", file, err)
	}
	return v
}

// TestMarkdownTableIsComplete is T15: the markdown table holds the ten views.
// The brief entry and the text table wait for the tasks that own them, and
// their cases skip with that reason until the entries stand.
func TestMarkdownTableIsComplete(t *testing.T) {
	for _, v := range views {
		t.Run(v.name, func(t *testing.T) {
			render, ok := markdown[v.name]
			if !ok {
				t.Fatalf("the markdown table has no entry for %s", v.name)
			}
			data := decode(t, v.name, v.data)
			out, err := render(data, "glance", "")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(out), "# ") || !strings.HasSuffix(string(out), "\n") {
				t.Errorf("not a document ending in a newline:\n%s", out)
			}
		})
	}
	t.Run("only the ten and brief", func(t *testing.T) {
		known := map[string]bool{"brief": true}
		for _, v := range views {
			known[v.name] = true
		}
		for k := range markdown {
			if !known[k] {
				t.Errorf("the markdown table holds %q, which is no view", k)
			}
		}
	})
	t.Run("brief", func(t *testing.T) {
		if markdown["brief"] == nil {
			t.Skip("the brief entry waits for the Agent briefs task, e4c7")
		}
	})
	t.Run("text", func(t *testing.T) {
		if len(text) == 0 {
			t.Skip("the text table waits for the Text renderers task, e0f7")
		}
		for _, v := range views {
			if text[v.name] == nil {
				t.Errorf("the text table has no entry for %s", v.name)
			}
		}
	})
}

func TestMarkdownEntryChecksItsData(t *testing.T) {
	tableau := decode(t, "tableau", "tableau/tooling.json")
	context := decode(t, "context", "context/tooling-2034.json")
	gates := decode(t, "gates", "gates/tooling.json")
	for _, c := range []struct {
		entry string
		data  any
	}{
		{"gates", tableau},
		{"task", gates},
		{"tableau", context},
		{"context", tableau},
		{"queue", nil},
		{"audit", "not a view"},
	} {
		if _, err := markdown[c.entry](c.data, "glance", ""); err == nil || !strings.Contains(err.Error(), "the "+c.entry+" view cannot render") {
			t.Errorf("%s given %T: error %v", c.entry, c.data, err)
		}
	}
}

func TestMarkdownEntryErrors(t *testing.T) {
	gates := decode(t, "gates", "gates/tooling.json")
	if _, err := markdown["gates"](gates, "summary", ""); err == nil || !strings.Contains(err.Error(), "unknown level") {
		t.Errorf("unknown level: error %v", err)
	}
	// The data holds glance only; a deeper level is the renderer's error.
	h := gates.Header()
	h.Level = view.Glance
	if _, err := markdown["gates"](gates, "provenance", ""); err == nil {
		t.Error("data that holds less than the level asked for rendered")
	}
}

// TestMarkdownEntryTurnsAPanicIntoAnError feeds the tableau renderer a grid
// whose row holds the wrong number of cells; package render panics with the
// error there, and the entry returns it.
func TestMarkdownEntryTurnsAPanicIntoAnError(t *testing.T) {
	v := decode(t, "tableau", "tableau/tooling.json").(*view.Tableau)
	if len(v.Rows) == 0 || len(v.Rows[0].Cells) == 0 {
		t.Skip("the fixture has no grid row")
	}
	v.Rows[0].Cells = v.Rows[0].Cells[1:]
	if _, err := markdown["tableau"](v, "detail", ""); err == nil {
		t.Error("a malformed grid rendered")
	}
}

// TestMarkdownLinkBase checks that the link base reaches the links: "" is the
// project directory, anything else is joined in front of the path.
func TestMarkdownLinkBase(t *testing.T) {
	data := decode(t, "task", "task/e9c6.json")
	for base, want := range map[string]string{
		"":        "(.tableaux/tasks/e9c6.yaml)",
		"..":      "(../.tableaux/tasks/e9c6.yaml)",
		"../proj": "(../proj/.tableaux/tasks/e9c6.yaml)",
	} {
		out, err := markdown["task"](data, "glance", base)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), want) {
			t.Errorf("link base %q: no %s in\n%s", base, want, out)
		}
	}
}
