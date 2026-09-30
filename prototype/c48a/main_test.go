package main

import (
	"os"
	"strings"
	"testing"
)

func render(t *testing.T, file string, level int) string {
	t.Helper()
	f, err := os.Open("testdata/" + file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	out, err := Render(f, level)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func has(t *testing.T, out string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
}

func TestAuthorityGlance(t *testing.T) {
	out := render(t, "authority.json", glance)
	has(t, out,
		"a1c0 Weather station    ada@example.org\n",
		"    c07d Node firmware  ·\n",
		"3c5d Dashboard        dan@example.org  proposed\n")
	if strings.Contains(out, "## ") {
		t.Error("glance shows a lower level")
	}
}

func TestAuthorityLevelsNest(t *testing.T) {
	d := render(t, "authority.json", detail)
	p := render(t, "authority.json", provenance)
	has(t, d, "mockup reviewer: ben@example.org", "ben@example.org, ada@example.org")
	if strings.Contains(d, "Deciding commits") {
		t.Error("detail shows provenance")
	}
	if !strings.HasPrefix(p, strings.TrimSuffix(d, "\n")) {
		t.Error("provenance does not contain detail")
	}
	has(t, p, "`f9e8746` | 2026-09-19 | merge")
}

func TestAuthorityBranch(t *testing.T) {
	out := render(t, "authority-branch.json", glance)
	has(t, out, "off the trunk")
	if strings.Contains(out, "authorised") {
		t.Error("an off-trunk task reads as authorised")
	}
}

func TestAuthorityPerson(t *testing.T) {
	has(t, render(t, "authority-person.json", glance), "person ada@example.org")
}

func TestAssignment(t *testing.T) {
	g := render(t, "assignment.json", glance)
	has(t, g, "| ada@example.org  |        3 |                2 |            0 |", "`claude-opus-5-5`")
	p := render(t, "assignment.json", provenance)
	has(t, p, "## ben@example.org", "Reviews", "reviewer `4e2b`", "function, stalled, blocked")
}

func TestAssignmentPerson(t *testing.T) {
	out := render(t, "assignment-person.json", detail)
	has(t, out, "person ada@example.org", "## ada@example.org")
	if strings.Contains(out, "## ben@example.org") {
		t.Error("another person appears")
	}
}

func TestOtherView(t *testing.T) {
	if _, err := Render(strings.NewReader(`{"view":"gate"}`), glance); err == nil {
		t.Error("a gate view renders")
	}
}
