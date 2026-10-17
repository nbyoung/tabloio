package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/brief"
	rtext "github.com/nbyoung/tabloio/internal/render/text"
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

// TestMarkdownTableIsComplete is T15: the markdown table holds the ten views
// and the brief.
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
		render, ok := markdown["brief"]
		if !ok {
			t.Fatal("the markdown table has no entry for the brief")
		}
		out, err := render(briefData(t, "e4c7-design.json"), "provenance", "")
		if err != nil {
			t.Fatal(err)
		}
		if want := briefData(t, "e4c7-design.txt"); string(out) != string(want) {
			t.Errorf("the brief differs from e4c7-design.txt:\n%s", out)
		}
	})
	t.Run("text", func(t *testing.T) {
		for _, v := range views {
			entry, ok := text[v.name]
			if !ok {
				t.Errorf("the text table has no entry for %s", v.name)
				continue
			}
			out, err := entry(decode(t, v.name, v.data), "glance", "", rtext.Options{})
			if err != nil {
				t.Errorf("%s: %v", v.name, err)
				continue
			}
			if !strings.HasSuffix(string(out), "\n") || strings.HasPrefix(string(out), "# ") {
				t.Errorf("%s: not a text document ending in a newline:\n%s", v.name, out)
			}
		}
		if _, ok := text["brief"]; ok {
			t.Error("the text table holds a brief entry")
		}
		if len(text) != len(views) {
			t.Errorf("the text table holds %d entries, want %d", len(text), len(views))
		}
	})
}

// briefData reads a file of package brief's testdata: an envelope or the
// golden brief it gives.
func briefData(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "brief", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestBriefEntryChecksItsData: the brief entry takes the envelope's bytes
// and nothing else, and returns the composer's errors as they stand.
func TestBriefEntryChecksItsData(t *testing.T) {
	for _, data := range []any{decode(t, "queue", "queue/tooling-owner.json"), "not bytes", nil} {
		if _, err := markdown["brief"](data, "provenance", ""); err == nil || !strings.Contains(err.Error(), "the brief cannot render") {
			t.Errorf("given %T: error %v", data, err)
		}
	}
	if _, err := markdown["brief"]([]byte("{"), "provenance", ""); err == nil || !strings.HasPrefix(err.Error(), "brief: envelope:") {
		t.Errorf("a broken envelope: error %v", err)
	}
	var waits *brief.WaitingError
	if _, err := markdown["brief"](briefData(t, "1422-implementation.json"), "provenance", ""); !errors.As(err, &waits) {
		t.Errorf("a waiting item: error %v", err)
	}
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

// TestTextLinkBase is the text counterpart of TestMarkdownLinkBase (T15): the
// link base reaches the address of a reference, and the text adds no
// task-file link.
func TestTextLinkBase(t *testing.T) {
	data := decode(t, "task", "task/e9c6.json")
	for base, want := range map[string]string{
		"":        "Proposed views <PLAN.md#views>",
		"..":      "Proposed views <../PLAN.md#views>",
		"../proj": "Proposed views <../proj/PLAN.md#views>",
	} {
		out, err := text["task"](data, "detail", base, rtext.Options{Width: 200})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), want) {
			t.Errorf("link base %q: no %s in\n%s", base, want, out)
		}
		if strings.Contains(string(out), ".tableaux/tasks/") {
			t.Errorf("link base %q: the text holds a task-file link", base)
		}
	}
}

func TestTextEntryChecksItsData(t *testing.T) {
	tableau := decode(t, "tableau", "tableau/tooling.json")
	if _, err := text["context"](tableau, "glance", "", rtext.Options{}); err == nil || !strings.Contains(err.Error(), "the context view cannot render") {
		t.Errorf("error %v", err)
	}
	if _, err := text["tableau"](tableau, "glance", "", rtext.Options{Width: 19}); !errors.Is(err, rtext.ErrWidth) {
		t.Errorf("width 19: error %v", err)
	}
}
