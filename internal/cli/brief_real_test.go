package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/brief"
)

// TestRealBrief is the integrate gate of design e4c7: the brief of a ready
// item of this repository, from tablo's library, confirms A1 to A7. The
// repository is tabloio at the pin the tooling plan's tag corpus/tooling
// holds, the submodule subprojects/tabloio of the corpus entry
// tableaux-tooling, where the agent's queue holds e3ed at integrate as work
// ready, 5ca9 at implementation as work waiting and 0a2f at implementation
// as a reaffirmation.
func TestRealBrief(t *testing.T) {
	dir, _ := toolingCopy(t)
	proj := filepath.Join(dir, "subprojects", "tabloio")
	const agent = "noreply@anthropic.com"
	ask := func(task, gate string) (Result, error) {
		return tabloSource{}.View(context.Background(), Query{Dir: proj, View: "queue", Person: agent,
			Brief: &Brief{Task: task, Gate: gate}, Level: "provenance"})
	}

	// A1: tablo answers every brief with the envelope, and data.brief
	// stands exactly for work ready or waiting.
	for _, c := range []struct {
		task, gate string
		brief      bool
	}{{"e3ed", "integrate", true}, {"5ca9", "implementation", true}, {"0a2f", "implementation", false}} {
		res, err := ask(c.task, c.gate)
		if err != nil {
			t.Fatalf("%s %s: %v", c.task, c.gate, err)
		}
		var env struct {
			Data struct {
				Brief json.RawMessage `json:"brief"`
			} `json:"data"`
		}
		raw, ok := res.Data.([]byte)
		if !ok || json.Unmarshal(raw, &env) != nil {
			t.Fatalf("%s %s: Result.Data is no envelope", c.task, c.gate)
		}
		if has := len(env.Data.Brief) > 0 && string(env.Data.Brief) != "null"; has != c.brief {
			t.Errorf("%s %s: data.brief stands %v, want %v", c.task, c.gate, has, c.brief)
		}
	}

	res, err := ask("e3ed", "integrate")
	if err != nil {
		t.Fatal(err)
	}
	d, err := brief.Decode(res.Data.([]byte), brief.Request{Person: agent, Task: "e3ed", Gate: "integrate"})
	if err != nil {
		t.Fatal(err)
	}
	full := regexp.MustCompile(`^[0-9a-f]{40}$`)
	// A2: the brief stands alone, with the person and the ref.
	if d.Person != agent || d.Ref.Name != "HEAD" || !full.MatchString(d.Ref.Commit) || d.Ref.Date != "2026-10-07" {
		t.Errorf("A2: person %q, ref %+v", d.Person, d.Ref)
	}
	// A3: each junction field with its value and source; an agent's
	// reviewer that the junction states none of is the assignee.
	tabloio := brief.Source{Kind: "task", ID: "07e0", Title: "tabloio"}
	if j := d.Junction; j.Contributor != (brief.Field{Value: agent, Source: tabloio}) ||
		j.Model != (brief.Field{Value: "claude-sonnet", Source: tabloio}) ||
		j.Reviewer != (brief.Field{Value: agent, Source: brief.Source{Kind: "assignee"}}) {
		t.Errorf("A3: %+v", j)
	}
	// A4: the gate's legend and the status with its symbols and commit.
	if g := d.Gate; g != (brief.Gate{Key: "integrate", Symbol: "🔗", Name: "Integration", Criteria: "Assembled with neighbouring components", Follows: "plain"}) {
		t.Errorf("A4: gate %+v", g)
	}
	if s := d.Status; s.Path != ".tableaux/status/e3ed.yaml" || s.Gate != "unit" || s.GateSymbol != "📏" || s.State != "nominal" ||
		s.StateSymbol != "🟢" || s.Date != "2026-10-07" || s.Recorder != agent || !full.MatchString(s.Commit) {
		t.Errorf("A4: status %+v", s)
	}
	// A5: the task whose entry supplies the contributor, with its
	// description.
	if s := d.Supplier; s == nil || s.ID != "07e0" || !strings.HasPrefix(s.Description, "The Tableaux command-line front end.") {
		t.Errorf("A5: supplier %+v", s)
	}
	// A6: every requirement, dependent and dependent of the parent, with
	// its gates, text and condition.
	if !reflect.DeepEqual(d.Requires, []brief.Edge{{ID: "0a2f", Title: "Bootstrap", From: "implementation", To: "implementation", Text: "The module", Condition: "met"}}) {
		t.Errorf("A6: requires %+v", d.Requires)
	}
	var dependents []string
	for _, e := range d.Dependents {
		dependents = append(dependents, e.ID+" "+e.From+" "+e.To+" "+e.Condition)
	}
	if want := []string{"c48a implementation implementation met", "a3cc implementation implementation met", "9167 implementation implementation met"}; !reflect.DeepEqual(dependents, want) {
		t.Errorf("A6: dependents %v", dependents)
	}
	if p := d.Parent; p == nil || p.ID != "1093" || len(p.Dependents) != 1 || p.Dependents[0].ID != "5ca9" || p.Dependents[0].Condition != "unmet" {
		t.Errorf("A6: parent %+v", p)
	}
	// A7: the path root first, without the task; the kind.
	if !reflect.DeepEqual(d.Task.Path, []brief.Link{{ID: "07e0", Title: "tabloio"}, {ID: "1093", Title: "Renderers"}}) || d.Kind != "ready" {
		t.Errorf("A7: path %+v, kind %q", d.Task.Path, d.Kind)
	}

	// The command: the brief on standard output, in the write commands'
	// vocabulary (b618 A10); a waiting item and an item that is no work
	// exit 1 with the table's messages.
	code, out, errs := runReal(t, "-C", proj, "queue", "--person", agent, "--brief", "e3ed", "integrate")
	if code != ExitOK || errs != "" || !strings.HasPrefix(out, "Brief: e3ed Legend and task renderers at integrate\n") {
		t.Fatalf("exit %d, stderr %q:\n%s", code, errs, out)
	}
	for _, want := range []string{
		"noreply@anthropic.com at 4bb4300, 2026-10-07: work ready.\n",
		"  Reviewer     noreply@anthropic.com   (the assignee; the junction states none)\n",
		"\n    tabloio record e3ed --gate integrate --state nominal \\\n",
		"\n    Reviewed: e3ed integrate\n",
		"\n  tabloio queue --person noreply@anthropic.com --ref 4bb4300 --brief e3ed integrate\n  tabloio task e3ed --ref 4bb4300\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the brief lacks %q", want)
		}
	}
	for _, c := range []struct{ task, gate, msg string }{
		{"5ca9", "implementation", "tabloio queue: 5ca9 at implementation waits and no brief starts it: waits for 1093 Renderers at implementation\n"},
		{"0a2f", "implementation", "tabloio queue: no work for noreply@anthropic.com at 0a2f implementation; tabloio queue --person noreply@anthropic.com lists the items\n"},
	} {
		if code, out, errs := runReal(t, "-C", proj, "queue", "--person", agent, "--brief", c.task, c.gate); code != ExitInvalid || out != "" || errs != c.msg {
			t.Errorf("%s %s: exit %d, stdout %q, stderr %q", c.task, c.gate, code, out, errs)
		}
	}
}
