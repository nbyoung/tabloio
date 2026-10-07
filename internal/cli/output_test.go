package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardOutput(t *testing.T) {
	for path, want := range map[string]string{
		"STATUS.md":                  "",
		"docs/STATUS.md":             "",
		"-":                          "",
		".tableaux":                  ".tableaux",
		".tableaux/status/5ca9.yaml": ".tableaux",
		"a/.git/index":               ".git",
		"/abs/.git":                  ".git",
		".tableaux/../x":             ".tableaux",
		"x/.git/../../y":             ".git",
		".github/ci.yml":             "",
		"a.tableaux/x":               "",
		".gitignore":                 "",
	} {
		err := guardOutput(path)
		switch {
		case want == "" && err != nil:
			t.Errorf("%s: refused: %v", path, err)
		case want != "" && (err == nil || !strings.Contains(err.Error(), "does not write inside "+want)):
			t.Errorf("%s: error %v, want one naming %s", path, err, want)
		}
	}
}

func TestResolveOutput(t *testing.T) {
	for _, c := range []struct{ dir, out, want string }{
		{"", "S.md", "S.md"},
		{"p", "S.md", filepath.Join("p", "S.md")},
		{"p", "docs/S.md", filepath.Join("p", "docs", "S.md")},
		{"p", "/abs/S.md", "/abs/S.md"},
		{"../p", "S.md", filepath.Join("..", "p", "S.md")},
	} {
		if got := resolveOutput(c.dir, c.out); got != c.want {
			t.Errorf("resolveOutput(%q, %q) = %q, want %q", c.dir, c.out, got, c.want)
		}
	}
}

// TestLinkBase is T10.
func TestLinkBase(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	for _, c := range []struct{ name, dir, out, want string }{
		{"the same directory", project, filepath.Join(project, "STATUS.md"), ""},
		{"a subdirectory", project, filepath.Join(project, "docs", "S.md"), ".."},
		{"two levels down", project, filepath.Join(project, "a", "b", "S.md"), "../.."},
		{"a sibling directory", project, filepath.Join(root, "out", "S.md"), "../project"},
		{"a parent directory", project, filepath.Join(root, "S.md"), "project"},
		{"a nested project", filepath.Join(project, "sub", "proj"), filepath.Join(project, "S.md"), "sub/proj"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := linkBase(c.dir, c.out); got != c.want {
				t.Errorf("linkBase(%q, %q) = %q, want %q", c.dir, c.out, got, c.want)
			}
		})
	}
	t.Run("relative paths read from the working directory", func(t *testing.T) {
		if got := linkBase("", "S.md"); got != "" {
			t.Errorf("no -C, output in the working directory: %q", got)
		}
		if got := linkBase("", filepath.Join("docs", "S.md")); got != ".." {
			t.Errorf("no -C, output in docs: %q", got)
		}
		if got := linkBase("project", filepath.Join("project", "S.md")); got != "" {
			t.Errorf("-C project, output inside it: %q", got)
		}
	})
}

func entries(t *testing.T, dir string) []string {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range es {
		names = append(names, e.Name())
	}
	return names
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "S.md")
	if err := writeFile(path, []byte("one\n")); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(path, []byte("two\n")); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "two\n" {
		t.Errorf("read %q, %v", b, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Errorf("mode %v, want 0644", info.Mode().Perm())
	}
	if got := entries(t, dir); len(got) != 1 {
		t.Errorf("the directory holds %v, want the file alone", got)
	}
}

func TestWriteFileFailures(t *testing.T) {
	t.Run("a missing directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeFile(filepath.Join(dir, "no", "S.md"), []byte("x")); err == nil {
			t.Error("wrote into a directory that does not exist")
		}
	})
	t.Run("a read-only directory leaves the file as it stood", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "S.md")
		if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dir, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		probe, err := os.CreateTemp(dir, "probe-*")
		if err == nil {
			_ = probe.Close()
			t.Skip("the directory stays writable, as it does for root")
		}
		if err := writeFile(path, []byte("new\n")); err == nil {
			t.Error("wrote into a read-only directory")
		}
		if b, _ := os.ReadFile(path); string(b) != "old\n" {
			t.Errorf("the file reads %q, want it as it stood", b)
		}
		if got := entries(t, dir); len(got) != 1 {
			t.Errorf("the directory holds %v, want the file alone", got)
		}
	})
	t.Run("a rename onto a directory removes the temporary file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "S.md")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := writeFile(path, []byte("x")); err == nil {
			t.Error("replaced a directory")
		}
		if got := entries(t, dir); len(got) != 1 {
			t.Errorf("the directory holds %v, want the directory alone", got)
		}
	})
}
