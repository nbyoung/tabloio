package action

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// plans is one plan function of each command, each valid on newRepo's
// checkout as the agent once the submodule has moved.
var plans = map[string]func(*Repo) (*Plan, error){
	"propose": propose(rainGauge, sonnet),
	"record":  record(Recording{ID: "c07d", Gate: "design", Pin: true}, sonnet),
	"review": func(r *Repo) (*Plan, error) {
		return r.Review(ctx, []string{"b618"}, "function", sonnet)
	},
	"reaffirm": func(r *Repo) (*Plan, error) { return r.Reaffirm(ctx, []string{"b618"}, sonnet) },
	"authorise": func(r *Repo) (*Plan, error) {
		return r.Authorise(ctx, []string{"b618"}, sonnet)
	},
}

// agentRepo returns newRepo's checkout with the submodule moved, and a Repo
// that commits as the agent, which the fake makes an authority.
func agentRepo(t *testing.T) (*testRepo, *Repo) {
	tr := newRepo(t)
	tr.runIn(filepath.Join(tr.dir, "sub"), "checkout", "-q", "main")
	tr.src.tasks["b618"].Authorities = []string{agent}
	tr.src.tasks["c07d"].Junctions[3].Contributor = agent // design
	return tr, tr.repo(agentNm, agent)
}

// TestPlanChangesNothing is T19: after each plan function HEAD, git status
// --porcelain and the index file's bytes are as before, though a touched
// file would make git status rewrite the index.
func TestPlanChangesNothing(t *testing.T) {
	tr, r := agentRepo(t)
	index := filepath.Join(tr.dir, ".git", "index")
	noLocks := append(hand("Builder", "builder@example.org"), "GIT_OPTIONAL_LOCKS=0")
	for name, plan := range plans {
		t.Run(name, func(t *testing.T) {
			later := time.Now().Add(time.Duration(len(name)) * time.Second)
			if err := os.Chtimes(filepath.Join(tr.dir, ".tableaux/status/b618.yaml"), later, later); err != nil {
				t.Fatal(err)
			}
			head, status := tr.head(), tr.runAs(noLocks, "status", "--porcelain")
			before, err := os.ReadFile(index)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := plan(r); err != nil {
				t.Fatalf("plan: %v", err)
			}
			after, _ := os.ReadFile(index)
			if tr.head() != head || tr.runAs(noLocks, "status", "--porcelain") != status || !bytes.Equal(before, after) {
				t.Errorf("HEAD %s, status %q, index changed %t", tr.head(), tr.runAs(noLocks, "status", "--porcelain"), !bytes.Equal(before, after))
			}
		})
	}
}

// TestNeverRuns is T26: a git wrapper first on the PATH logs every Git run
// of a plan and an apply of each of the five commands; none is push, fetch,
// pull, merge, rebase, tag, switch, checkout or config, and no commit
// carries --amend or --no-verify.
func TestNeverRuns(t *testing.T) {
	real, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not on the PATH")
	}
	forbidden := []string{"push", "fetch", "pull", "merge", "rebase", "tag", "switch", "checkout", "config"}
	for _, name := range []string{"propose", "record", "review", "reaffirm", "authorise"} {
		t.Run(name, func(t *testing.T) {
			tr, r := agentRepo(t)
			bin, log := t.TempDir(), filepath.Join(t.TempDir(), "git.log")
			wrapper := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$*\" >> %q\nexec %q \"$@\"\n", log, real)
			if err := os.WriteFile(filepath.Join(bin, "git"), []byte(wrapper), 0o755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			tr.commit(r, plans[name])
			b, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			runs := strings.Split(strings.TrimSpace(string(b)), "\n")
			commits := 0
			for _, run := range runs {
				words := strings.Fields(run)
				if len(words) == 0 || slices.Contains(forbidden, words[0]) {
					t.Errorf("git %s", run)
				}
				if len(words) > 0 && words[0] == "commit" {
					commits++
					if slices.Contains(words, "--amend") || slices.Contains(words, "--no-verify") {
						t.Errorf("git %s", run)
					}
				}
			}
			if commits != 1 || len(runs) < 10 {
				t.Errorf("%d runs, %d commits:\n%s", len(runs), commits, b)
			}
		})
	}
}

// TestDirtyRefused checks step 7: a file the plan changes must be clean, and
// a file it creates must not exist in the working tree.
func TestDirtyRefused(t *testing.T) {
	tr := newRepo(t)
	tr.write(".tableaux/status/b618.yaml", b618Status+"# an edit\n")
	_, err := tr.repo(agentNm, agent).Record(ctx, b618Function, sonnet)
	wantRefusal(t, err, "dirty")
	tr.write(".tableaux/status/9f31.yaml", "")
	tr.run("rm", "-q", "--cached", ".tableaux/status/9f31.yaml")
	tr.run("commit", "-q", "-m", "Untrack the barometer's status")
	tr.src.tasks["9f31"].Status = nil
	_, err = tr.repo("Ada", ada).Record(ctx, Recording{ID: "9f31", Gate: "defined", State: "nominal"}, Options{})
	wantRefusal(t, err, "dirty")
}
