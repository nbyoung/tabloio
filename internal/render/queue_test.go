package render

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// queueGoldens lists the golden cases of the work queue (T8). The glance
// golden files of the first two are copies of the drafts of design a3cc,
// with the flag the owner ruled.
var queueGoldens = []statusGolden{
	{"queue", "tooling-agent", glanceDetail},
	{"queue", "tooling-owner", glanceDetail},
	{"queue", "review-owed", glanceDetail},
	{"queue", "waiting", glanceDetail},
	{"queue", "weather-ada", glanceDetail},
}

func TestQueueGolden(t *testing.T) { runStatusGoldens(t, queueGoldens) }

// T8, the decoding half: an unknown key is ignored and an unknown level is
// rejected.
func TestDecodeQueue(t *testing.T) {
	if _, err := view.Decode("queue", []byte(`{"command": "queue", "params": {"level": "glance"}, "data": {"unknown": 1}}`)); err != nil {
		t.Error(err)
	}
	if _, err := view.Decode("queue", []byte(`{"command": "queue", "params": {"level": "deep"}, "data": {}}`)); err == nil {
		t.Error("an unknown level decodes")
	}
}

// T8: the queue's folds at glance and at detail.
func TestQueueFold(t *testing.T) {
	mk := func(n int) *view.Queue {
		v := readFixture(t, "queue", "tooling-owner").(*view.Queue)
		cp := *v
		for i := 0; i < n; i++ {
			cp.Items = append(cp.Items, view.QueueItem{
				Kind: "ready", Task: view.TaskRef{ID: fmt.Sprintf("d%03d", i), Title: "Mockup"},
				Gate: "mockup", Model: "claude-opus", Since: "2026-09-29",
			})
		}
		return &cp
	}
	rows := func(out string) []string { return section(out, "") }
	_ = rows

	for _, c := range []struct {
		n      int
		unfold bool
		fold   bool
	}{{8, false, false}, {9, false, true}, {9, true, false}} {
		out := renderString(t, mk(c.n), Options{Level: view.Detail, Unfold: c.unfold})
		if got := strings.Contains(out, "| … and 6 more: "); got != c.fold {
			t.Errorf("%d alike, unfold %v: a fold row stands %v, want %v", c.n, c.unfold, got, c.fold)
		}
		if got := strings.Contains(out, "**… and 6 more** at "); got != c.fold {
			t.Errorf("%d alike, unfold %v: a fold paragraph stands %v, want %v", c.n, c.unfold, got, c.fold)
		}
		if c.fold && !strings.Contains(out, "**… and 6 more** at 📌 mockup, model `claude-opus`: `d003` `d004` `d005` `d006` `d007` `d008`.") {
			t.Errorf("the fold paragraph:\n%s", out)
		}
		if got := strings.Count(out, "\n**`d"); got != map[bool]int{true: 3, false: c.n}[c.fold] {
			t.Errorf("%d alike: %d item paragraphs", c.n, got)
		}
	}

	// An item with a cause never folds: the cause belongs to its task.
	q := mk(12)
	for i := range q.Items {
		q.Items[i].Kind = "waiting"
		q.Items[i].Causes = []view.Wait{{Kind: "requirement", Requires: &view.Requirement{Task: view.TaskRef{ID: "x001"}, From: "design"}}}
	}
	out := renderString(t, q, Options{Level: view.Detail})
	if strings.Contains(out, "more") || strings.Count(out, "; waits for `x001` at design |") != 12 {
		t.Errorf("items with a cause fold:\n%s", out)
	}
}

// T8: the glance table's rows of the owner's empty queue and of a kind with
// no item, and the sections of detail.
func TestQueueSections(t *testing.T) {
	out := renderString(t, readFixture(t, "queue", "tooling-owner"), Options{Level: view.Detail})
	for _, w := range []string{
		"0 items: 0 reviews owed, 0 authorisations owed, 0 work ready, 0 reaffirmations, 0 work waiting.\n\n**The queue is empty.**\n",
		"## 1. Reviews owed: none\n\n## 2. Authorisations owed: none\n\n## 3. Work ready: none\n\n## 4. Reaffirmations: none\n\n## 5. Work waiting: none\n",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("the empty queue lacks %q:\n%s", w, out)
		}
	}
	if strings.Contains(out, "| Kind |") {
		t.Error("an empty queue draws a table")
	}
	// A kind with no item gives no row: the count line names it.
	out = renderString(t, readFixture(t, "queue", "review-owed"), Options{Level: view.Glance})
	if strings.Count(out, "\n|") != 3 {
		t.Errorf("one item should give a header, a delimiter and one row:\n%s", out)
	}
	// The queue lines come from the data and each kind's command.
	out = renderString(t, readFixture(t, "queue", "tooling-agent"), Options{Level: view.Detail})
	for _, w := range []string{
		"- Brief: `tabloio queue --person noreply@anthropic.com --ref corpus/tooling --brief c2ad validate`\n",
		"- Contributor: 🤖 **noreply@anthropic.com**, model `claude-opus`, from `bc63`\n",
		"- Reviewer: 👀 nbyoung@nbyoung.com, from `437e`\n",
		"- Reviewer: 👀 **noreply@anthropic.com**, the assignee\n",
		"- References: [The mockup](docs/mockups/gates.md)\n",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("the queue lacks %q", w)
		}
	}
	// At 1b6916e twenty mockups wait for e9c6 at design, which unblocks
	// them, and each one's parent unblocks the subprojects.
	wait := renderString(t, readFixture(t, "queue", "waiting"), Options{Level: view.Detail})
	for _, w := range []string{
		"- Waits for: waits for `e9c6` Abstract views at design\n",
		"- Its parent unblocks:\n  - `bc86` Markdown views at 📌 mockup is required by `6103` tabloio: command line, output and input, mockup → design, The Markdown mockups it reproduces: unmet, not yet due\n",
		"- Unblocks:\n  - `99f0` Gate definition view in Markdown, design → mockup, The abstract definition of the view: unmet\n",
		"- Also required by: `77b2` tablo: backend library and plumbing, implementation → design, The abstract views it derives: unmet, not yet due\n",
	} {
		if !strings.Contains(wait, w) {
			t.Errorf("the waiting queue lacks %q", w)
		}
	}
	if !strings.HasSuffix(strings.SplitN(wait, "\n\nCommand:", 2)[0], "- Do not start: the cause stands.") {
		t.Errorf("a waiting item does not close with its instruction:\n%s", wait)
	}
}

// T9: the queue's provenance is the detail but for the frame's level and
// ref. The brief of one item is package brief's, and a brief in the data
// changes nothing here.
func TestQueueProvenance(t *testing.T) {
	v := readFixture(t, "queue", "weather-ada").(*view.Queue)
	cp := *v
	cp.Level = view.Provenance
	o := Options{Level: view.Provenance}
	withBrief := cp
	withBrief.Brief = json.RawMessage(`{"task": "9f31", "gate": "performance"}`)
	if a, b := renderString(t, &withBrief, o), renderString(t, &cp, o); a != b {
		t.Errorf("a brief in the data changes the queue:\n%s\n---\n%s", a, b)
	}

	detail := renderString(t, &cp, Options{Level: view.Detail})
	prov := renderString(t, &cp, o)
	body := func(s string) string {
		lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
		return strings.Join(lines[3:len(lines)-1], "\n")
	}
	if body(detail) != body(prov) {
		t.Errorf("provenance differs from detail:\n%s\n---\n%s", body(prov), body(detail))
	}
	if !strings.Contains(prov, "level provenance") || !strings.Contains(prov, "--level provenance") {
		t.Errorf("the frame lacks the level:\n%s", prov)
	}
}

// T8: the words of each cause of a waiting item, and of an owed
// authorisation, by its typed facts. A cause reads the item's own facts at
// every level, a reason's note and an authorisation's why (886d decision
// 15): a status elsewhere in the item gives no note.
func TestQueueCauses(t *testing.T) {
	v := readFixture(t, "queue", "weather-ada").(*view.Queue)
	p := newPage(&v.Head, Options{Level: view.Detail})
	plain := func(xs []doc.Inline) string { return plainInline(xs, map[string]bool{}) }
	req := &view.Requirement{Task: view.TaskRef{ID: "e9c6"}, From: "design"}
	for _, c := range []struct {
		it   view.QueueItem
		want string
	}{
		{view.QueueItem{Kind: "waiting", Causes: []view.Wait{{Kind: "requirement", Requires: req}}}, "waits for e9c6 at design"},
		{view.QueueItem{Kind: "waiting", Causes: []view.Wait{{Kind: "reason", Reason: "blocked"}}}, "blocked"},
		{view.QueueItem{Kind: "waiting", Causes: []view.Wait{{Kind: "reason", Reason: "blocked", Note: "No parts"}}},
			"blocked: No parts"},
		{view.QueueItem{Kind: "waiting", Causes: []view.Wait{{Kind: "reason", Reason: "blocked"}}, Status: &view.Status{Note: "No parts"}},
			"blocked"},
		{view.QueueItem{Kind: "waiting", Causes: []view.Wait{{Kind: "authorisation", By: "a@b.c"}}}, "proposed; waits for its authorisation by a@b.c"},
		{view.QueueItem{Kind: "waiting", Causes: []view.Wait{{Kind: "review", By: "r@b.c"}}}, "handed off; waits for the review by r@b.c"},
		{view.QueueItem{Kind: "waiting", Causes: []view.Wait{{Kind: "requirement", Requires: req}, {Kind: "review", By: "r@b.c"}}},
			"waits for e9c6 at design; handed off; waits for the review by r@b.c"},
		{view.QueueItem{Kind: "authorisation"}, "proposed; the deciding commit is not by an authority"},
		{view.QueueItem{Kind: "authorisation", Why: "no_trunk"}, "proposed; the trunk is undetermined"},
		{view.QueueItem{Kind: "authorisation", Authorisation: &view.Authorisation{Why: "no_trunk"}}, "proposed; the deciding commit is not by an authority"},
		{view.QueueItem{Kind: "ready"}, ""},
	} {
		if got := plain(p.queueCause(c.it)); got != c.want {
			t.Errorf("%s %+v: %q, want %q", c.it.Kind, c.it.Causes, got, c.want)
		}
	}
}
