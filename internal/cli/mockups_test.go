package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// renderTestdata is the renderers' testdata, whose fixtures tablo draws and
// whose golden files the renderer designs fix.
const renderTestdata = "../render/testdata"

// handMade names the fixtures no ref of a corpus entry holds: the draft of
// design e3ed (its T2), the blockage illustration of design a3cc (its T10)
// and two of the mockup's cases outside its pictures (design 9167, T4).
var handMade = map[string]bool{
	"gates/draft": true, "blockage/illustration": true,
	"history/proposal": true, "history/reaffirmed": true,
}

// auditEntries names the corpus entry of each audit fixture of the rules
// H2 to H6 (design 9167, T10).
var auditEntries = map[string]string{
	"corpus-h2": "review-by-non-reviewer", "corpus-h3": "model-mismatch",
	"corpus-h4": "handoff-inferred", "corpus-h5": "handoff-stale",
	"corpus-h6": "model-trailer-missing",
}

// mockupFixture is one fixture of the renderers and the head of its
// envelope: what the command line names.
type mockupFixture struct {
	view, name string
	env        struct {
		Command string `json:"command"`
		Source  struct {
			Ref      string `json:"ref"`
			Worktree bool   `json:"worktree"`
			From     *struct {
				Ref string `json:"ref"`
			} `json:"from"`
		} `json:"source"`
		Project struct {
			Root string `json:"root"`
		} `json:"project"`
		Params struct {
			Task       string   `json:"task"`
			Person     string   `json:"person"`
			Window     *int     `json:"window"`
			Columns    []string `json:"columns"`
			Historical bool     `json:"historical"`
			Proposed   bool     `json:"proposed"`
			Stale      int      `json:"stale"`
			Now        string   `json:"now"`
		} `json:"params"`
	}
}

// mockupFixtures reads every fixture of the renderers but the hand-made.
func mockupFixtures(t *testing.T) []mockupFixture {
	t.Helper()
	var paths []string
	for _, pattern := range []string{"*/*.json", "*/in/*.json"} {
		p, err := filepath.Glob(filepath.Join(renderTestdata, pattern))
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p...)
	}
	sort.Strings(paths)
	var out []mockupFixture
	for _, p := range paths {
		rel, _ := filepath.Rel(renderTestdata, p)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		f := mockupFixture{view: parts[0], name: strings.TrimSuffix(parts[len(parts)-1], ".json")}
		if handMade[f.view+"/"+f.name] {
			continue
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &f.env); err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		out = append(out, f)
	}
	return out
}

// argv is the command line that reaches the fixture's view: its command,
// the parameters its envelope names, a view's default task left out, and
// the ref as given unless the fixture read the working tree.
func (f mockupFixture) argv(dir string) []string {
	e := f.env
	a := []string{"-C", dir, e.Command}
	defaults := map[string]bool{"authority": true, "assignment": true, "queue": true, "blockage": true, "history": true, "audit": true}
	switch {
	case e.Command == "task":
		a = append(a, e.Params.Task)
	case e.Params.Task != "" && (!defaults[e.Command] || e.Params.Task != e.Project.Root):
		a = append(a, "--task", e.Params.Task)
	}
	if e.Params.Person != "" {
		a = append(a, "--person", e.Params.Person)
	}
	if !e.Source.Worktree {
		ref := e.Source.Ref
		if e.Source.From != nil {
			ref = e.Source.From.Ref + ".." + ref
		}
		a = append(a, "--ref", ref)
	}
	if e.Params.Window != nil {
		a = append(a, "--window", strconv.Itoa(*e.Params.Window))
	} else if len(e.Params.Columns) > 0 {
		a = append(a, "--columns", strings.Join(e.Params.Columns, ","))
	}
	if e.Params.Historical {
		a = append(a, "--historical")
	}
	if e.Params.Proposed {
		a = append(a, "--proposed")
	}
	if e.Params.Stale > 0 {
		a = append(a, "--stale", strconv.Itoa(e.Params.Stale))
	}
	return a
}

// goldens lists the golden files of the fixture, each with the options that
// select its level, format and width.
func (f mockupFixture) goldens(t *testing.T) map[string][]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(renderTestdata, f.view, f.name+".*"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
	shape := regexp.MustCompile(`^\.(glance|detail|provenance)(\.w(\d+))?\.(md|txt)$`)
	for _, file := range files {
		m := shape.FindStringSubmatch(strings.TrimPrefix(filepath.Base(file), f.name))
		if m == nil {
			continue
		}
		opts := []string{"--level", m[1]}
		if m[4] == "txt" {
			opts = append(opts, "--text")
			if m[3] != "" {
				opts = append(opts, "--width", m[3])
			}
		}
		out[file] = opts
	}
	return out
}

// corpusLock is the lock of the CI job corpus, whose tooling names the
// commit of the umbrella that the tooling plan's fixtures describe.
const corpusLock = "../../.github/corpus.lock.yaml"

// toolingCopy copies the repository of the corpus entry tableaux-tooling,
// as tablo's TestCorpus reaches it through TABLO_CORPUS, into a temporary
// directory with its submodules. The entry reads the umbrella at the ref
// its expected.yaml names, corpus/tooling. The umbrella's owner moves that
// tag whenever the plan's facts change, and the fixtures follow only when a
// person moves the lock's tooling and draws them again; so the copy's own
// tag names the lock's tooling, as corpus.sh fetch sets it in CI. So that
// every caller reads that tree whatever clone TABLO_CORPUS names and
// wherever it stands, the copy is checked out at the tag, detached, and
// each submodule's working tree at its pin there, from the module stores
// the copied .git holds: nothing is fetched. So that no fact of a view
// depends on where a clone's branches stand today, the copy's main, local
// and origin's, stands at the tag, and each submodule's at its pin. It
// skips without TABLO_CORPUS. It returns the copy and the ref.
func toolingCopy(t *testing.T) (dir, ref string) {
	t.Helper()
	corpus := os.Getenv("TABLO_CORPUS")
	if corpus == "" {
		t.Skip("TABLO_CORPUS names no corpus beside a clone of the umbrella")
	}
	for _, tool := range []string{"git", "cp"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	entry := filepath.Join(corpus, "entries", "tableaux-tooling")
	expected, err := os.ReadFile(filepath.Join(entry, "expected.yaml"))
	if err != nil {
		t.Skipf("the corpus holds no tableaux-tooling: %v", err)
	}
	m := regexp.MustCompile(`(?m)^source:\s*\{\s*repository:\s*([^,\s]+),\s*ref:\s*([^\s}]+)\s*\}`).FindSubmatch(expected)
	if m == nil {
		t.Fatalf("tableaux-tooling's expected.yaml names no source")
	}
	src, ref := filepath.Join(entry, string(m[1])), string(m[2])
	dir = filepath.Join(t.TempDir(), "tableaux")
	if out, err := exec.Command("cp", "-a", src, dir).CombinedOutput(); err != nil {
		t.Fatalf("copy the umbrella: %v\n%s", err, out)
	}
	lock, err := os.ReadFile(corpusLock)
	if err != nil {
		t.Fatal(err)
	}
	tooling := regexp.MustCompile(`(?m)^tooling:\s*([0-9a-f]{40})\b`).FindSubmatch(lock)
	if tooling == nil {
		t.Fatalf("%s names no tooling commit", corpusLock)
	}
	if err := exec.Command("git", "-C", dir, "rev-parse", "-q", "--verify", string(tooling[1])+"^{commit}").Run(); err != nil {
		t.Fatalf("the clone TABLO_CORPUS names holds no commit %s, the lock's tooling; fetch it", tooling[1])
	}
	gitOut(t, dir, "tag", "-f", ref, string(tooling[1]))
	commit := strings.TrimSpace(gitOut(t, dir, "rev-parse", ref+"^{commit}"))
	// at checks a repository out at a commit, detached, then sets its main
	// there; detached first, so that moving main moves no checkout.
	at := func(repo, commit string) {
		gitOut(t, repo, "checkout", "-q", "-f", "--detach", commit)
		for _, r := range []string{"refs/heads/main", "refs/remotes/origin/main"} {
			gitOut(t, repo, "update-ref", r, commit)
		}
	}
	at(dir, commit)
	for _, line := range strings.Split(strings.TrimSpace(gitOut(t, dir, "ls-tree", commit, "subprojects/")), "\n") {
		f := strings.Fields(line)
		if len(f) != 4 || f[1] != "commit" {
			continue
		}
		sub := filepath.Join(dir, f[3])
		if _, err := os.Stat(filepath.Join(sub, ".git")); err != nil {
			t.Fatalf("the clone TABLO_CORPUS names has no checkout of the submodule %s; run git submodule update --init in it", f[3])
		}
		at(sub, f[2])
	}
	return dir, ref
}

// entryCopy copies a built corpus entry, with its .git and its submodule,
// into a temporary directory, as corpusCopy does for weather-station.
func entryCopy(t *testing.T, name string) string {
	t.Helper()
	corpus := os.Getenv("TABLEAUX_CORPUS")
	if corpus == "" {
		t.Skip("TABLEAUX_CORPUS names no built corpus")
	}
	src := filepath.Join(corpus, name)
	if _, err := os.Stat(src); err != nil {
		t.Skipf("the corpus holds no %s: %v", name, err)
	}
	dst := filepath.Join(t.TempDir(), name)
	if out, err := exec.Command("cp", "-a", src, dst).CombinedOutput(); err != nil {
		t.Fatalf("copy the corpus entry: %v\n%s", err, out)
	}
	return dst
}

// TestRealMockups is T18: the mockups' command lines give the mockups'
// renderings. Each golden file of the renderer designs whose fixture tablo
// draws, Markdown and text alike (e0f7 A2), is reached through Run over
// tablo's library: the tooling plan's on the corpus entry tableaux-tooling,
// weather-station's and those of the audit's rule entries on copies of the
// built corpus. The command line is the one the fixture's envelope names,
// with the level, the format and the width of the golden file; the audit's
// clock is the date its envelope states.
func TestRealMockups(t *testing.T) {
	fixtures := mockupFixtures(t)
	byRoot := map[string][]mockupFixture{}
	for _, f := range fixtures {
		byRoot[f.env.Project.Root] = append(byRoot[f.env.Project.Root], f)
	}
	run := func(t *testing.T, dir string, fs []mockupFixture) int {
		n := 0
		for _, f := range fs {
			for file, opts := range f.goldens(t) {
				want, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				argv := append(f.argv(dir), opts...)
				old := today
				if now := f.env.Params.Now; now != "" {
					today = func() string { return now }
				}
				code, out, errs := runReal(t, argv...)
				today = old
				out = unlinked.ReplaceAllString(out, "$1")
				name, _ := filepath.Rel(renderTestdata, file)
				if code != ExitOK || out != string(want) {
					t.Errorf("%s: tabloio %s: exit %d, stderr %q, output differs:\n%s", name, strings.Join(argv[2:], " "), code, errs, firstDifference(out, string(want)))
				}
				n++
			}
		}
		return n
	}
	t.Run("tableaux-tooling", func(t *testing.T) {
		dir, ref := toolingCopy(t)
		if ref != "corpus/tooling" {
			t.Errorf("tableaux-tooling reads %s, the fixtures corpus/tooling", ref)
		}
		n := run(t, dir, byRoot["437e"])
		if n < 100 {
			t.Errorf("only %d golden files of the tooling plan", n)
		}
		t.Logf("%d golden files of the tooling plan through Run", n)
	})
	t.Run("weather-station", func(t *testing.T) {
		dir := entryCopy(t, realCorpusEntry)
		n := run(t, dir, byRoot["a1c0"])
		if n < 50 {
			t.Errorf("only %d golden files of weather-station", n)
		}
		t.Logf("%d golden files of weather-station through Run", n)
	})
	t.Run("the audit's rules", func(t *testing.T) {
		for _, f := range byRoot["e4a1"] {
			entry, ok := auditEntries[f.name]
			if !ok {
				t.Errorf("%s/%s names no corpus entry", f.view, f.name)
				continue
			}
			t.Logf("%s: %d golden files through Run", f.name, run(t, entryCopy(t, entry), []mockupFixture{f}))
		}
	})
	if len(byRoot) != 3 {
		t.Errorf("the fixtures name %d projects, want 3", len(byRoot))
	}
}

// unlinked matches a task-file link, which Run writes at the link base of
// standard output, "." (R13), and the golden files, rendered with no link
// base, do not.
var unlinked = regexp.MustCompile(`\[([^\]]*)\]\(\.tableaux/tasks/[0-9a-f]{4}\.yaml\)`)

// firstDifference shows the first line at which two renderings part.
func firstDifference(got, want string) string {
	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		var gl, wl string
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			return "line " + strconv.Itoa(i+1) + "\n got  " + gl + "\n want " + wl
		}
	}
	return "no line differs"
}
