package main

import (
	"bytes"
	"strings"
	"testing"
)

func render(t *testing.T, args ...string) string {
	t.Helper()
	var b bytes.Buffer
	if err := run(args, &b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func mustContain(t *testing.T, out string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
}

func TestGateGlance(t *testing.T) {
	out := render(t, "testdata/gate.json")
	mustContain(t, out,
		"| ❔ undefined | 📝 defined | 📌 mockup | ⚙️ function | ⚡ performance |",
		"States ⚪ undefined · 🟢 nominal · 🟡 at_risk · 🔴 stalled · ✅ complete.",
		"Reasons 🪫 overloaded · ⛔ blocked.",
		"🧑 a person contributes")
	if strings.Contains(out, "## Gate criteria") {
		t.Error("glance shows criteria")
	}
}

func TestGateLevelsNest(t *testing.T) {
	g := render(t, "testdata/gate.json")
	d := render(t, "-level", "detail", "testdata/gate.json")
	p := render(t, "-level", "provenance", "testdata/gate.json")
	if !strings.HasPrefix(d, g) || !strings.HasPrefix(p, d) {
		t.Error("a level does not include the one before")
	}
	mustContain(t, d, "A technical demonstration of function exists", "| 🔴 stalled | 3 |", "👓 review")
	mustContain(t, p, "Language version 0.3.1, trunk `main`", "`.tableaux/gates.yaml` last changed by `68d9021` 2026-09-15")
}

func TestGateForTask(t *testing.T) {
	out := render(t, "-level", "detail", "testdata/gate-c07d.json")
	mustContain(t, out, "# Legend for task `c07d`", "| ⚓ reliability | Reliability prototype | Target reliability shown in the fundamental technology | — |")
}

func TestTaskGlance(t *testing.T) {
	out := render(t, "-gates", "testdata/gate.json", "testdata/task-c07d.json")
	mustContain(t, out,
		"> `c07d` **Node firmware** — ben@example.org — under `4e2b` Sensor node, order 2 — 📐 design 🟢 nominal, 2026-09-17, ben@example.org: Sleep scheduler in progress",
		"Snapshot of `f1a0` in `firmware` at `cf5b169`")
}

func TestTaskReasonAndRollUp(t *testing.T) {
	out := render(t, "testdata/task-4e2b.json")
	mustContain(t, out, "⚙️ function 🔴 stalled ⛔ blocked, 2026-09-17, rolled up from `9f31`: Barometer ICs on 14-week backorder")
	root := render(t, "testdata/task-a1c0.json")
	mustContain(t, root, "— the root —")
}

func TestTaskDetail(t *testing.T) {
	out := render(t, "-level", "detail", "testdata/task-c07d.json")
	mustContain(t, out,
		"| `9f31` Sensor board | design | implementation | Pin map and sensor bus | unmet, due |",
		"| 🧩 unit | 🤖👀 | opus@example.org | claude-opus-5-5 | ben@example.org |",
		"| 🛠️ implementation | 🪆 |",
		"Authorisation: authorised.")
	dep := render(t, "-level", "detail", "testdata/task-9f31.json")
	mustContain(t, dep, "## Dependents", "[Barometer datasheet](https://example.org/datasheets/bmp390.pdf)", "<docs/sensor-board.md>")
	par := render(t, "-level", "detail", "testdata/task-4e2b.json")
	mustContain(t, par, "`9f31`, `c07d`, in display order.")
}

func TestTaskProvenance(t *testing.T) {
	out := render(t, "-level", "provenance", "testdata/task-9f31.json")
	mustContain(t, out, "way: merge", "`f9e8746`", "| 2026-09-28 | reaffirmed |", "git log --format=")
}

func TestErrors(t *testing.T) {
	var b bytes.Buffer
	if run([]string{"-level", "deep", "testdata/gate.json"}, &b) == nil {
		t.Error("accepted an unknown level")
	}
	if run([]string{}, &b) == nil {
		t.Error("accepted no file")
	}
}
