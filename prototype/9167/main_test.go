package main

import (
	"strings"
	"testing"
)

func hist(t *testing.T, file string, o Options) string {
	t.Helper()
	var evs []Event
	if err := load("testdata/"+file, &evs); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	History(&b, evs, o)
	return b.String()
}

func audit(t *testing.T, file string, o Options) string {
	t.Helper()
	var fs []Finding
	if err := load("testdata/"+file, &fs); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	Audit(&b, fs, o)
	return b.String()
}

func TestHistoryGroupsCommitEvents(t *testing.T) {
	out := hist(t, "history-w9-w13.json", Options{})
	if !strings.Contains(out, "| 2026-09-26 | ben@example.org | `c07d` | reviewed, status, pin |") {
		t.Errorf("same-commit events not merged:\n%s", out)
	}
	if !strings.Contains(out, "function stalled (blocked)") {
		t.Errorf("status missing:\n%s", out)
	}
}

func TestHistoryLevels(t *testing.T) {
	g := hist(t, "history-w9-w13.json", Options{Level: Glance})
	d := hist(t, "history-w9-w13.json", Options{Level: Detail})
	p := hist(t, "history-w9-w13.json", Options{Level: Provenance, Title: "W9..W13"})
	if strings.Contains(g, "Barometer") || !strings.Contains(d, "Barometer") {
		t.Error("note belongs at detail only")
	}
	if strings.Contains(d, "`efa5c1e`") || !strings.Contains(p, "`a2393fb`") || !strings.Contains(p, "git log") {
		t.Errorf("commit belongs at provenance:\n%s", p)
	}
}

func TestHistoryFilters(t *testing.T) {
	out := hist(t, "history-all.json", Options{Task: "9f31"})
	if strings.Contains(out, "`c07d`") || !strings.Contains(out, "`9f31`") {
		t.Errorf("task filter:\n%s", out)
	}
	out = hist(t, "history-all.json", Options{Person: "nobody@example.org"})
	if !strings.Contains(out, "No events.") {
		t.Errorf("empty:\n%s", out)
	}
}

func TestAuditLevels(t *testing.T) {
	g := audit(t, "audit-w3-w13-stale3.json", Options{})
	if !strings.Contains(g, "2 warnings (1 introduced, 1 resolved)") || !strings.Contains(g, "| Count |") {
		t.Errorf("glance:\n%s", g)
	}
	d := audit(t, "audit-w3-w13-stale3.json", Options{Level: Detail})
	if !strings.Contains(d, "more than 3 days") {
		t.Errorf("detail:\n%s", d)
	}
	p := audit(t, "audit-w3-w13-stale3.json", Options{Level: Provenance, Stale: 3})
	if !strings.Contains(p, "`55f57c1`") || !strings.Contains(p, "--stale 3") {
		t.Errorf("provenance:\n%s", p)
	}
}

func TestAuditEmpty(t *testing.T) {
	if out := audit(t, "audit-w13-w13.json", Options{}); !strings.Contains(out, "No findings.") {
		t.Error(out)
	}
}

func TestEscapesPipes(t *testing.T) {
	var b strings.Builder
	table(&b, []string{"a"}, [][]string{{"x|y"}})
	if !strings.Contains(b.String(), `x\|y`) {
		t.Error(b.String())
	}
}
