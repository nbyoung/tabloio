package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/view"
)

// tabloGolden is one golden file of tablo's views, copied unchanged into
// testdata/<view>/tablo/; the README there names the commit. The file holds
// the data of one envelope, and params are the parameters of tablo's golden
// table for it, without the level, which the name ends in.
type tabloGolden struct {
	view, file string
	params     map[string]any
}

// tabloGoldens lists tablo's golden files of the views tabloio renders.
var tabloGoldens = []tabloGolden{
	{"gates", "gates-9f31.provenance.json", map[string]any{"task": "9f31"}},
	{"task", "task-3c5d.detail.json", map[string]any{"task": "3c5d", "person": nil}},
	{"task", "task-c07d.provenance.json", map[string]any{"task": "c07d", "person": "ben@example.org"}},
}

// path is the file's path under testdata.
func (g tabloGolden) path() string { return filepath.Join("testdata", g.view, "tablo", g.file) }

// level is the level the file's name ends in.
func (g tabloGolden) level(t testing.TB) view.Level {
	t.Helper()
	name := strings.TrimSuffix(g.file, ".json")
	l, err := view.ParseLevel(name[strings.LastIndex(name, ".")+1:])
	if err != nil {
		t.Fatalf("%s: %v", g.file, err)
	}
	return l
}

// envelope wraps the file's data as tablo's schema test wraps its golden
// data: the source main with no commit and an undetermined trunk, no
// project, an observer, and the parameters with the role and the level.
func (g tabloGolden) envelope(t testing.TB) (raw, env []byte) {
	t.Helper()
	raw, err := os.ReadFile(g.path())
	if err != nil {
		t.Fatal(err)
	}
	params := map[string]any{"role": "observer", "level": g.level(t).String()}
	for k, v := range g.params {
		params[k] = v
	}
	env, err = json.Marshal(map[string]any{
		"schema":  "tablo/1",
		"command": g.view,
		"source": map[string]any{
			"ref": "main", "worktree": false, "commit": nil, "date": nil, "from": nil,
			"trunk": map[string]any{"name": nil, "resolved": "undetermined"}, "on_trunk": false,
		},
		"project":     nil,
		"viewer":      map[string]any{"email": nil, "roles": []string{"observer"}},
		"params":      params,
		"data":        json.RawMessage(raw),
		"diagnostics": []any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw, env
}

// TestTabloGoldens decodes each of tablo's golden files and renders it at
// every level it holds, as Markdown and as text, under the rules every
// fixture keeps: the levels nest (T6), no rune appears that the data does
// not hold (T7 and the text T11), two renderings agree (T10), and the text
// keeps its width and its boxes (the text T10).
func TestTabloGoldens(t *testing.T) {
	const own = "·…→›—×═─│┌┬┐├┼┤└┴┘┼�"
	seen := map[string]bool{}
	for _, g := range tabloGoldens {
		seen[g.path()] = true
		t.Run(g.view+"/"+g.file, func(t *testing.T) {
			raw, env := g.envelope(t)
			v, err := view.Decode(g.view, env)
			if err != nil {
				t.Fatal(err)
			}
			if v.Header().Level != g.level(t) {
				t.Fatalf("the data holds %s", v.Header().Level)
			}
			checkLevelsNest(t, v)
			for _, l := range levelsOf(v) {
				for _, o := range []Options{
					{Level: l, Links: Links{Project: "."}},
					{Format: Text, Level: l, Links: Links{Project: "."}},
				} {
					out := renderString(t, v, o)
					if again := renderString(t, v, o); again != out {
						t.Errorf("%s: two renderings differ", l)
					}
					for _, r := range out {
						if r > 0x7F && !strings.ContainsRune(own, r) && !strings.ContainsRune(string(raw), r) {
							t.Errorf("%s: the output holds %q (%U), which the data does not", l, r, r)
						}
					}
					if o.Format == Text {
						checkTextInvariants(t, out, 0, buildFor(t, v, o))
					}
				}
			}
		})
	}
	files, err := filepath.Glob(filepath.Join("testdata", "*", "tablo", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if !seen[f] {
			t.Errorf("%s is in no case", f)
		}
	}
}
