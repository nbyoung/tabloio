package action

import (
	"errors"
	"strings"
	"testing"
)

// TestApplyStale is the first half of T23: a commit between plan and apply
// is stale, and Apply then makes nothing.
func TestApplyStale(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo(agentNm, agent)
	p, err := r.Reaffirm(ctx, []string{"b618"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	tr.run("commit", "-q", "--allow-empty", "-m", "Meanwhile")
	head := tr.head()
	_, err = r.Apply(ctx, p)
	var refusal Refusal
	if !errors.As(err, &refusal) || refusal.Code != "stale" {
		t.Fatalf("error %v, want the refusal stale", err)
	}
	if tr.head() != head {
		t.Error("a stale plan moved HEAD")
	}
}

// TestApplyAltered is the second half of T23: a commit-msg hook that
// appends a line is altered, with the commit's hash.
func TestApplyAltered(t *testing.T) {
	tr := newRepo(t)
	r := tr.repo(agentNm, agent)
	p, err := r.Reaffirm(ctx, []string{"b618"}, Options{Agent: Agent{Model: "claude-sonnet-5-5"}})
	if err != nil {
		t.Fatal(err)
	}
	tr.hook("commit-msg", `echo "Signed-off-by: Hook <hook@example.org>" >> "$1"`+"\n")
	res, err := r.Apply(ctx, p)
	var refusal Refusal
	if !errors.As(err, &refusal) || refusal.Code != "altered" {
		t.Fatalf("error %v, want the refusal altered", err)
	}
	if res.Commit == "" || res.Commit != tr.head() || !strings.Contains(refusal.Text, short(res.Commit)) {
		t.Errorf("result %+v, HEAD %s, refusal %q", res, tr.head(), refusal.Text)
	}
	if n := tr.run("rev-list", "--count", p.Base+"..HEAD"); n != "1" {
		t.Errorf("%s commits after the base; Apply never amends", n)
	}
}
