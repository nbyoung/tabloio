package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/view"
)

// statusGolden is one fixture of the status renderers and the levels at
// which it meets its golden files. The files are testdata/<view>/<name>.json
// and testdata/<view>/<name>.<level>.md. The tests of design e3ed that run
// for every fixture (T6, T7, T8, T10) pick the fixtures up from testdata with
// no entry in any list.
type statusGolden struct {
	view   string
	name   string
	levels []view.Level
}

// runStatusGoldens compares each fixture at each level with its golden file,
// byte for byte; -update rewrites the file.
func runStatusGoldens(t *testing.T, list []statusGolden) {
	t.Helper()
	for _, g := range list {
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
