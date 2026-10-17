package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// briefEnvelope returns an envelope of package brief's testdata with edit
// applied to its decoded form, or as it stands for a nil edit.
func briefEnvelope(t *testing.T, name string, edit func(env map[string]any)) []byte {
	t.Helper()
	b := briefData(t, name)
	if edit == nil {
		return b
	}
	var env map[string]any
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatal(err)
	}
	edit(env)
	out, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// briefSource is a fake Source that returns the envelope as a brief's data.
func briefSource(envelope []byte) *fakeSource {
	src := newSource()
	src.result = Result{Data: envelope}
	return src
}

// TestBriefCommand is T15 of design e4c7: the brief through Run, with the
// fake Source of 5ca9's tests.
func TestBriefCommand(t *testing.T) {
	const person = "noreply@anthropic.com"

	t.Run("the query of A1", func(t *testing.T) {
		src := briefSource(briefEnvelope(t, "e4c7-design.json", nil))
		if code, _, errs := run(t, src, "queue", "--person", person, "--ref", "main", "--brief", "e4c7", "design"); code != ExitOK {
			t.Fatalf("exit %d, stderr %q", code, errs)
		}
		want := Query{View: "queue", Ref: "main", Person: person, Brief: &Brief{Task: "e4c7", Gate: "design"}, Level: "provenance"}
		if len(src.queries) != 1 || !reflect.DeepEqual(src.queries[0], want) {
			t.Fatalf("queries %+v, want %+v", src.queries, want)
		}
		// tablo takes the person, the ref and the brief, and no level.
		p, ref := tabloParams(src.queries[0])
		if p.Person != person || ref != "main" || p.Brief == nil || p.Brief.Task != "e4c7" || p.Brief.Gate != "design" || p.Level != "" {
			t.Errorf("tablo's parameters %+v at %q", p, ref)
		}
		if len(src.viewerDirs) != 0 {
			t.Error("a brief with --person asked for the viewer")
		}
	})

	t.Run("the person defaults to the viewer", func(t *testing.T) {
		src := briefSource(briefEnvelope(t, "e4c7-design.json", nil))
		if code, _, errs := run(t, src, "queue", "--brief", "e4c7", "design"); code != ExitOK {
			t.Fatalf("exit %d, stderr %q", code, errs)
		}
		if len(src.queries) != 1 || src.queries[0].Person != "viewer@example.org" {
			t.Errorf("queries %+v", src.queries)
		}
	})

	t.Run("the golden text under either format", func(t *testing.T) {
		for _, name := range []string{"e4c7-design", "e9c6-implementation", "c74a-mockup"} {
			want := string(briefData(t, name+".txt"))
			task, gate, _ := strings.Cut(name, "-")
			for _, format := range [][]string{nil, {"--markdown"}, {"--text"}} {
				argv := append([]string{"queue", "--person", person, "--brief", task, gate}, format...)
				code, out, errs := run(t, briefSource(briefEnvelope(t, name+".json", nil)), argv...)
				if code != ExitOK || errs != "" {
					t.Errorf("%v: exit %d, stderr %q", argv, code, errs)
				}
				if out != want {
					t.Errorf("%v: standard output differs from %s.txt:\n%s", argv, name, out)
				}
			}
		}
	})

	t.Run("-o writes the brief", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "brief.txt")
		code, out, errs := run(t, briefSource(briefEnvelope(t, "e4c7-design.json", nil)), "-C", dir, "queue", "--person", person, "--brief", "e4c7", "design", "-o", path)
		if code != ExitOK || out != "" || errs != "" {
			t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != string(briefData(t, "e4c7-design.txt")) {
			t.Errorf("the file holds %q, %v", got, err)
		}
	})

	// The rows of the design's table: standard error and the exit code, with
	// nothing on standard output and no file written.
	for _, c := range []struct {
		name     string
		envelope []byte
		task     string
		gate     string
		code     int
		stderr   string
	}{
		{
			"work waiting", briefEnvelope(t, "1422-implementation.json", nil), "1422", "implementation", ExitInvalid,
			"tabloio queue: 1422 at implementation waits and no brief starts it: waits for 5ca9 Read commands at implementation\n",
		},
		{
			"no work", briefEnvelope(t, "e4c7-design.json", func(env map[string]any) {
				env["data"].(map[string]any)["brief"] = nil
			}), "e4c7", "design", ExitInvalid,
			"tabloio queue: no work for noreply@anthropic.com at e4c7 design; tabloio queue --person noreply@anthropic.com lists the items\n",
		},
		{
			"a missing field", briefEnvelope(t, "e4c7-design.json", func(env map[string]any) {
				env["data"].(map[string]any)["brief"].(map[string]any)["task"].(map[string]any)["id"] = ""
			}), "e4c7", "design", ExitFailure,
			"tabloio queue: brief: task.id is missing in tablo's data\n",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "brief.txt")
			for _, extra := range [][]string{nil, {"--text"}, {"-o", path}} {
				argv := append([]string{"-C", dir, "queue", "--person", person, "--brief", c.task, c.gate}, extra...)
				code, out, errs := run(t, briefSource(c.envelope), argv...)
				if code != c.code || out != "" || errs != c.stderr {
					t.Errorf("%v: exit %d, stdout %q, stderr %q; want exit %d, stderr %q", argv, code, out, errs, c.code, c.stderr)
				}
			}
			if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("a refused brief wrote %s: %v", path, err)
			}
		})
	}

	t.Run("--brief with --level detail", func(t *testing.T) {
		src := briefSource(briefEnvelope(t, "e4c7-design.json", nil))
		code, out, errs := run(t, src, "queue", "--person", person, "--brief", "e4c7", "design", "--level", "detail")
		if code != ExitUsage || out != "" || !strings.HasPrefix(errs, "tabloio queue: --brief excludes --level and --task\n") || !src.untouched() {
			t.Errorf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})

	t.Run("tablo fails", func(t *testing.T) {
		src := newSource()
		src.result = Result{Diagnostics: []Diagnostic{{Severity: "error", Code: "E001", Path: ".tableaux/version.yaml", Message: "no version"}}}
		code, out, errs := run(t, src, "queue", "--person", person, "--brief", "e4c7", "design")
		if code != ExitInvalid || out != "" || errs != ".tableaux/version.yaml: error E001: no version\n" {
			t.Errorf("an error diagnostic: exit %d, stdout %q, stderr %q", code, out, errs)
		}
		src = newSource()
		src.viewErr = errors.New("the repository cannot be read")
		code, out, errs = run(t, src, "queue", "--person", person, "--brief", "e4c7", "design")
		if code != ExitFailure || out != "" || errs != "tabloio queue: the repository cannot be read\n" {
			t.Errorf("a failed read: exit %d, stdout %q, stderr %q", code, out, errs)
		}
	})
}
