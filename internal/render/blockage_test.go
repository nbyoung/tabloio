package render

import (
	"strings"
	"testing"

	"github.com/nbyoung/tabloio/internal/render/doc"
	"github.com/nbyoung/tabloio/internal/render/view"
)

// blockageGoldens lists the golden cases of the work-blockage tree (T10).
var blockageGoldens = []statusGolden{
	{"blockage", "illustration", allLevels},
	{"blockage", "tooling", allLevels},
	{"blockage", "weather", glanceDetail},
}

func TestBlockageGolden(t *testing.T) { runStatusGoldens(t, blockageGoldens) }

// T10, the decoding half: an unknown key is ignored and an unknown level is
// rejected.
func TestDecodeBlockage(t *testing.T) {
	if _, err := view.Decode("blockage", []byte(`{"level": "glance", "unknown": 1}`)); err != nil {
		t.Error(err)
	}
	if _, err := view.Decode("blockage", []byte(`{"level": "deep"}`)); err == nil {
		t.Error("an unknown level decodes")
	}
}

// T10: the blockage tree's forms.
func TestBlockageForms(t *testing.T) {
	// The empty form at each level: the count line, the bold sentence, and
	// the requirements not yet due.
	empty := readFixture(t, "blockage", "tooling")
	for _, l := range allLevels {
		out := renderString(t, empty, Options{Level: l})
		if !strings.Contains(out, "0 causes: 0 unmet requirements, 0 statuses off nominal, 0 reviews outstanding, 0 authorisations outstanding, 0 snapshots not advanced.\n\n**No cause holds any task.**\n") {
			t.Errorf("%s: the empty form:\n%s", l, out)
		}
		if l == view.Glance && !strings.Contains(out, "Next: 10 requirements are not yet due, 8 of them unmet.\n") {
			t.Errorf("glance: the line of requirements not yet due:\n%s", out)
		}
		if l >= view.Detail && !strings.Contains(out, "## Next: 10 requirements not yet due\n") {
			t.Errorf("%s: the section of requirements not yet due:\n%s", l, out)
		}
	}
	// Ten requirements are listed under their four tasks.
	out := renderString(t, empty, Options{Level: view.Detail})
	if got := strings.Count(out, ": requires `"); got != 10 {
		t.Errorf("%d requirements listed, want 10", got)
	}
	if !strings.Contains(out, "  - 📐 design: requires `e9c6` Abstract views at 🧱 implementation, The abstract views it derives: unmet, not yet due.\n") {
		t.Errorf("an entry not yet due:\n%s", out)
	}

	// With none not yet due, glance prints no line and detail says none.
	v := *empty.(*view.Blockage)
	v.NotDue, v.NotDueN, v.NotDueUnmet = nil, 0, 0
	if out := renderString(t, &v, Options{Level: view.Glance}); strings.Contains(out, "Next:") {
		t.Errorf("a Next line stands:\n%s", out)
	}
	if out := renderString(t, &v, Options{Level: view.Detail}); !strings.Contains(out, "## Next: none\n") {
		t.Errorf("no Next: none heading:\n%s", out)
	}

	// The glance table, and the tree at detail keeps it (R1).
	ill := readFixture(t, "blockage", "illustration")
	g := renderString(t, ill, Options{Level: view.Glance})
	if !strings.Contains(g, "| `5471` Work-blockage tree view in Html awaits review at 📌 mockup | 👀 nbyoung@nbyoung.com reviews | 4 |\n") {
		t.Errorf("the glance table:\n%s", g)
	}
	if d := renderString(t, ill, Options{Level: view.Detail}); !strings.Contains(d, strings.SplitN(strings.SplitN(g, "| Cause", 2)[1], "\n\n", 2)[0]) {
		t.Errorf("detail lacks the glance table")
	}
}

// T11: the words of each kind of cause, act and held task.
func TestBlockageWords(t *testing.T) {
	ill := readFixture(t, "blockage", "illustration").(*view.Blockage)
	p := newPage(&ill.Head, Options{Level: view.Detail})
	plain := func(xs []doc.Inline) string { return plainInline(xs, map[string]bool{}) }
	task := view.TaskRef{ID: "9f31", Title: "Sensor board"}

	causes := []struct {
		c    view.Cause
		want string
	}{
		{view.Cause{Kind: "requirement", Task: task, Gate: "design"}, "9f31 Sensor board has not passed 📐 design"},
		{view.Cause{Kind: "status", Task: task, Status: &view.Status{Gate: "implementation", State: "stalled", Reason: "blocked", Note: "No parts"}},
			"9f31 Sensor board stands at 🧱 implementation 🔴 stalled ⛔ blocked: No parts"},
		{view.Cause{Kind: "status", Task: task, Status: &view.Status{Gate: "design", State: "at_risk"}}, "9f31 Sensor board stands at 📐 design 🟡 at_risk"},
		{view.Cause{Kind: "review", Task: task, Gate: "mockup"}, "9f31 Sensor board awaits review at 📌 mockup"},
		{view.Cause{Kind: "authorisation", Task: task}, "9f31 Sensor board is proposed"},
		{view.Cause{Kind: "snapshot", Task: task, URL: "subprojects/tablo", Pin: "00f8f68"}, "9f31 Sensor board waits on subprojects/tablo at 00f8f68"},
	}
	for _, c := range causes {
		if got := plain(p.blockageCause(c.c)); got != c.want {
			t.Errorf("%s: %q, want %q", c.c.Kind, got, c.want)
		}
	}

	who := []struct {
		c    view.Cause
		want string
	}{
		{view.Cause{Resolver: "nbyoung@nbyoung.com", Mark: "reviewer", Act: "reviews"}, "👀 nbyoung@nbyoung.com reviews"},
		{view.Cause{Resolver: "noreply@anthropic.com", Mark: "agent", Act: "contributes"}, "🤖 noreply@anthropic.com contributes"},
		{view.Cause{Resolver: "a@b.c", Mark: "person", Act: "authorises"}, "🧑 a@b.c authorises"},
		{view.Cause{Resolver: "a@b.c", Act: "records"}, "a@b.c records"},
		{view.Cause{Resolver: "a@b.c", Mark: "subproject", Act: "advances"}, "🪆 a@b.c advances"},
	}
	for _, c := range who {
		if got := plain(p.blockageWho(c.c)); got != c.want {
			t.Errorf("%s: %q, want %q", c.c.Act, got, c.want)
		}
	}

	req := &view.Requirement{Task: view.TaskRef{ID: "5fe3"}, From: "mockup", To: "design", Text: "The HTML mockups"}
	held := []struct {
		h    view.Held
		want string
	}{
		{view.Held{Task: task, Gate: "mockup", Via: "self"}, "9f31 Sensor board at 📌 mockup: the cause holds the task itself."},
		{view.Held{Task: task, Gate: "mockup", Via: "parent", Parent: true, Child: "5471"}, "9f31 Sensor board at 📌 mockup: the parent of 5471."},
		{view.Held{Task: task, Gate: "design", Via: "requirement", Requires: req}, "9f31 Sensor board at 📐 design: requires 5fe3 at 📌 mockup, The HTML mockups."},
		{view.Held{Task: task, Gate: "design", Via: "requirement", Requires: req, Also: []int{1, 3}},
			"9f31 Sensor board at 📐 design: requires 5fe3 at 📌 mockup, The HTML mockups. Also under cause 1. Also under cause 3."},
		{view.Held{Task: view.TaskRef{ID: "9f31"}, Gate: "design", Via: "self"}, "9f31 at 📐 design: the cause holds the task itself."},
	}
	for _, c := range held {
		if got := plain(p.blockageHeld(c.h).Text); got != c.want {
			t.Errorf("%s: %q, want %q", c.h.Via, got, c.want)
		}
	}
	// A parent's title is bold; a plain task's is not.
	if it := p.blockageHeld(held[1].h); !hasStrong(it.Text, "Sensor board") {
		t.Error("a parent's title is not bold")
	}
	if it := p.blockageHeld(held[0].h); hasStrong(it.Text, "Sensor board") {
		t.Error("a plain task's title is bold")
	}

	// The count line names each kind by the helper.
	b := *ill
	b.Kinds = []view.KindCount{{Kind: "requirement", Count: 1}, {Kind: "status", Count: 2}, {Kind: "review", Count: 1}, {Kind: "authorisation"}, {Kind: "snapshot", Count: 2}}
	line := plain(p.blockageCounts(&b).Text)
	if want := "2 causes: 1 unmet requirement, 2 statuses off nominal, 1 review outstanding, 0 authorisations outstanding, 2 snapshots not advanced."; line != want {
		t.Errorf("the count line reads %q, want %q", line, want)
	}
}

func hasStrong(xs []doc.Inline, text string) bool {
	for _, x := range xs {
		if s, ok := x.(doc.Strong); ok && len(s) == 1 && s[0] == doc.Text(text) {
			return true
		}
	}
	return false
}
