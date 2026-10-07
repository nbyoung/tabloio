package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/view"
)

// structuralGoldens lists the golden cases of the structural views (design
// c48a), which golden_test.go does not hold: the files are
// testdata/<view>/<name>.json and testdata/<view>/<name>.<level>.md.
var structuralGoldens = []golden{
	// T1: the tooling plan, against the drafts of design c48a
	{view: "authority", name: "tooling", levels: allLevels},
	// T4: the weather station: a proposed task, a person's view, a ref off the trunk
	{view: "authority", name: "weather", levels: allLevels},
	{view: "authority", name: "weather-ada", levels: []view.Level{view.Detail}},
	{view: "authority", name: "weather-branch", levels: glanceDetail},
	// T7: --proposed and its empty form
	{view: "authority", name: "weather-proposed", levels: allLevels},
	{view: "authority", name: "tooling-proposed", levels: allLevels},
}

// T1, T4 and T7: each fixture at each level against its golden file,
// byte for byte.
func TestStructuralGolden(t *testing.T) {
	for _, g := range structuralGoldens {
		for _, level := range g.levels {
			t.Run(g.view+"/"+g.name+"."+level.String(), func(t *testing.T) {
				v := readFixture(t, g.view, g.name)
				got := renderString(t, v, Options{Level: level})
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
