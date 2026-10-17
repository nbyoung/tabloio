package action

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// rainGauge is the proposal of task.yaml.
var rainGauge = Proposal{
	Title:       "Rain gauge: tipping bucket",
	Description: "A tipping-bucket gauge on the mast that counts 0.2 mm tips, debounces them in firmware and reports the hourly total to the gateway.",
	Parent:      "07e0",
	Assignee:    ada,
	References: []Reference{
		{URL: "docs/rain-gauge.md"},
		{URL: "https://example.org/datasheets/rg11.pdf", Text: "Gauge datasheet, revision 2"},
	},
}

func propose(p Proposal, o Options) func(*Repo) (*Plan, error) {
	return func(r *Repo) (*Plan, error) { return r.Propose(ctx, p, o) }
}

// TestNeverRedumps is T3: a proposal under 40e8 and one under 07e0 write
// id: "40e8" and id: "07e0", and no existing file under .tableaux changes a
// byte.
func TestNeverRedumps(t *testing.T) {
	tr := newRepo(t)
	base := tr.head()
	r := tr.repo("Owner", owner)
	for _, parent := range []string{"40e8", "07e0"} {
		p := rainGauge
		p.Parent = parent
		plan, _ := tr.commit(r, propose(p, Options{}))
		if want := `parent: { id: "` + parent + `", order: `; !strings.Contains(string(plan.Files[0].New), want) {
			t.Errorf("under %s:\n%s", parent, plan.Files[0].New)
		}
	}
	if got := tr.run("diff", "--name-status", "--diff-filter=MDRT", base, "HEAD"); got != "" {
		t.Errorf("existing files change: %q", got)
	}
	if got := tr.run("diff", "--name-status", base, "HEAD"); strings.Count(got, "A\t.tableaux/tasks/") != 2 {
		t.Errorf("the commits change %q", got)
	}
}

// TestProposeFile is T15: the proposal of task.yaml gives that file byte for
// byte; a Rand that repeats a taken id draws again; an id in use is
// id-taken; the order is one above the largest.
func TestProposeFile(t *testing.T) {
	want, err := os.ReadFile("testdata/task.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tr := newRepo(t)
	tr.write(".tableaux/status/abcd.yaml", "gate: defined\nstate: nominal\n") // in the working tree alone
	r := tr.repo("Owner", owner)
	r.Rand = bytes.NewReader([]byte{0xb6, 0x18, 0xab, 0xcd, 0x9f, 0x31, 0x12, 0x34})
	p, err := r.Propose(ctx, rainGauge, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if p.NewID != "1234" || p.Tasks[0] != "1234" || p.Files[0].Path != ".tableaux/tasks/1234.yaml" || p.Files[0].Old != nil {
		t.Errorf("id %s, tasks %v, file %s", p.NewID, p.Tasks, p.Files[0].Path)
	}
	if got := string(p.Files[0].New); got != string(want) {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if p.Subject != "Propose rain gauge: tipping bucket (1234)" {
		t.Errorf("subject %q", p.Subject)
	}
	for _, id := range []string{"c07d", "abcd"} {
		stated := rainGauge
		stated.ID = id
		_, err := r.Propose(ctx, stated, Options{})
		wantRefusal(t, err, "id-taken")
	}
	t.Run("the order stated and the assignee by default", func(t *testing.T) {
		stated := rainGauge
		stated.Order, stated.Assignee, stated.ID = 7, "", "5678"
		p, err := r.Propose(ctx, stated, Options{})
		if err != nil || !strings.Contains(string(p.Files[0].New), "assignee: "+owner+"\n") ||
			!strings.HasSuffix(string(p.Files[0].New), `parent: { id: "07e0", order: 7 }`+"\n") {
			t.Errorf("%v\n%s", err, p.Files[0].New)
		}
	})
	t.Run("bad-field", func(t *testing.T) {
		for name, p := range map[string]Proposal{
			"empty title":       {Description: "d", Parent: "07e0"},
			"title with break":  {Title: "a\nb", Description: "d", Parent: "07e0"},
			"empty description": {Title: "t", Parent: "07e0"},
			"assignee":          {Title: "t", Description: "d", Parent: "07e0", Assignee: "ada"},
			"reference":         {Title: "t", Description: "d", Parent: "07e0", References: []Reference{{Text: "no url"}}},
			"order":             {Title: "t", Description: "d", Parent: "07e0", Order: -1},
			"no parent":         {Title: "t", Description: "d"},
			"bad id":            {Title: "t", Description: "d", Parent: "07e0", ID: "XYZ1"},
		} {
			_, err := r.Propose(ctx, p, Options{})
			if codes := refusalCodes(t, err); len(codes) != 1 || codes[0] != "bad-field" {
				t.Errorf("%s: %v", name, codes)
			}
		}
		_, err := r.Propose(ctx, Proposal{Title: "t", Description: "d", Parent: "beef"}, Options{})
		wantRefusal(t, err, "no-task")
	})
	t.Run("a thousand draws", func(t *testing.T) {
		r := tr.repo("Owner", owner)
		r.Rand = bytes.NewReader(bytes.Repeat([]byte{0xb6, 0x18}, freshDraws))
		if _, err := r.Propose(ctx, rainGauge, Options{}); err == nil || !strings.Contains(err.Error(), "no free task id") {
			t.Errorf("error %v", err)
		}
	})
}

// TestProposeNeverAuthorises is T16: no plan of propose holds an
// Authorised: trailer.
func TestProposeNeverAuthorises(t *testing.T) {
	tr := newRepo(t)
	for _, o := range []Options{{}, sonnet} {
		for _, email := range []string{owner, agent} {
			p, err := tr.repo("Hand", email).Propose(ctx, rainGauge, o)
			if err != nil {
				t.Fatal(err)
			}
			for _, tl := range p.Trailers {
				if tl.Key == "Authorised" {
					t.Errorf("as %s: %v", email, p.Trailers)
				}
			}
			want := 0
			if o.Agent.Model != "" {
				want = 2 // Model: and Co-Authored-By:
			}
			if len(p.Trailers) != want {
				t.Errorf("as %s with %+v: trailers %v", email, o.Agent, p.Trailers)
			}
		}
	}
}

// testProposeRestored is T22 for a proposal: after a failing pre-commit
// hook, the proposed file is gone and unstaged.
func testProposeRestored(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo("Owner", owner)
	r.Rand = bytes.NewReader([]byte{0x12, 0x34})
	p, err := r.Propose(ctx, rainGauge, Options{})
	if err != nil {
		t.Fatal(err)
	}
	tr.hook("pre-commit", "exit 1\n")
	head := tr.head()
	if _, err := r.Apply(ctx, p); err == nil {
		t.Fatal("the commit passed a failing hook")
	}
	if tr.head() != head || tr.read(".tableaux/tasks/1234.yaml") != "<absent>" || tr.run("status", "--porcelain") != "" {
		t.Errorf("HEAD %s, file %q, status %q", tr.head(), tr.read(".tableaux/tasks/1234.yaml"), tr.run("status", "--porcelain"))
	}
}
