package render

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/view"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// golden is one fixture and the levels, with the options, at which it is
// compared with its golden files.
type golden struct {
	view   string
	name   string
	levels []view.Level
	opts   func(*Options) // adjusts the options; nil for none
}

var (
	glanceDetail = []view.Level{view.Glance, view.Detail}
	allLevels    = []view.Level{view.Glance, view.Detail, view.Provenance}
)

// goldens lists the golden cases. The files are testdata/<view>/<name>.json
// and testdata/<view>/<name>.<level>.md.
var goldens = []golden{
	// T2
	{view: "gates", name: "draft", levels: glanceDetail},
	// T3 and T4
	{view: "task", name: "e9c6", levels: allLevels},
	{view: "gates", name: "tooling", levels: []view.Level{view.Provenance}},
	{view: "gates", name: "tooling-e9c6", levels: []view.Level{view.Detail}},
	{view: "task", name: "5fe3", levels: allLevels},
	{view: "task", name: "595e", levels: allLevels},
	// T5
	{view: "gates", name: "weather-9f31", levels: []view.Level{view.Detail}},
	{view: "task", name: "weather-9f31", levels: allLevels},
	{view: "task", name: "weather-c07d", levels: allLevels},
	{view: "task", name: "weather-3c5d", levels: glanceDetail},
}

func readFixture(t testing.TB, viewName, name string) view.View {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", viewName, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := view.Decode(viewName, data)
	if err != nil {
		t.Fatalf("%s/%s: %v", viewName, name, err)
	}
	return v
}

func renderString(t testing.TB, v view.View, o Options) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Render(&buf, v, o); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// T2 to T5: each fixture at each level against its golden file, byte for byte.
func TestGolden(t *testing.T) {
	for _, g := range goldens {
		for _, level := range g.levels {
			t.Run(g.view+"/"+g.name+"."+level.String(), func(t *testing.T) {
				v := readFixture(t, g.view, g.name)
				o := Options{Level: level}
				if g.opts != nil {
					g.opts(&o)
				}
				got := renderString(t, v, o)
				file := filepath.Join("testdata", g.view, g.name+"."+level.String()+".md")
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
			})
		}
	}
}
