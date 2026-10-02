package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func loadInputs(t *testing.T, queue, task string) inputs {
	t.Helper()
	var in inputs
	if err := load("testdata/"+queue+".json", &in.Queue); err != nil {
		t.Fatal(err)
	}
	if err := load("testdata/task-"+task+".json", &in.Task); err != nil {
		t.Fatal(err)
	}
	if err := load("testdata/gate-"+task+".json", &in.Gate); err != nil {
		t.Fatal(err)
	}
	return in
}

func brief(t *testing.T, in inputs, task, gate string) string {
	t.Helper()
	it, err := selectItem(in.Queue, task, gate)
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	render(&b, in, it)
	return b.String()
}

func TestGolden(t *testing.T) {
	for _, c := range []struct{ queue, task, gate string }{
		{"queue-opus", "c07d", "unit"},
		{"queue-ada", "7b2e", "defined"},
		{"queue-ada", "9f31", "performance"},
	} {
		got := brief(t, loadInputs(t, c.queue, c.task), c.task, c.gate)
		want, err := os.ReadFile("testdata/brief-" + c.task + "-" + c.gate + ".md")
		if err != nil {
			t.Fatal(err)
		}
		if got != string(want) {
			t.Errorf("%s at %s differs from its golden file:\n%s", c.task, c.gate, got)
		}
	}
}

func TestModelAndReviewerReachTheBrief(t *testing.T) {
	got := brief(t, loadInputs(t, "queue-opus", "c07d"), "c07d", "unit")
	for _, want := range []string{
		"model `claude-opus-5-5`",
		"Reviewer: ben@example.org",
		"matching claude-opus-5-5",
		"reason: review",
		"Reviewed: c07d unit",
		"Co-Authored-By:",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("brief lacks %q", want)
		}
	}
}

// An agent that is its own reviewer passes the gate and carries the trailer,
// in one paragraph with Model and Co-Authored-By.
func TestSelfReview(t *testing.T) {
	in := loadInputs(t, "queue-opus", "c07d")
	for i := range in.Task.Detail.Junctions {
		if j := &in.Task.Detail.Junctions[i]; j.Gate == "unit" {
			j.Reviewer = "opus@example.org"
		}
	}
	got := brief(t, in, "c07d", "unit")
	want := "Reviewed: c07d unit\nModel: <the identifier the harness reports, matching claude-opus-5-5>\nCo-Authored-By:"
	if !strings.Contains(got, want) {
		t.Errorf("trailers not one paragraph:\n%s", got)
	}
	if !strings.Contains(got, "gate: unit\nstate: nominal\nnote:") || strings.Contains(got, "reason: review") {
		t.Errorf("self-review status wrong:\n%s", got)
	}
}

func TestPersonHasNoModelTrailer(t *testing.T) {
	got := brief(t, loadInputs(t, "queue-ada", "7b2e"), "7b2e", "defined")
	if strings.Contains(got, "Model:") {
		t.Errorf("a person's brief carries a Model trailer:\n%s", got)
	}
}

func TestOtherKinds(t *testing.T) {
	in := loadInputs(t, "queue-ada", "9f31")
	if got := brief(t, in, "9f31", "performance"); !strings.Contains(got, "Do not start") {
		t.Errorf("work waiting must not start:\n%s", got)
	}
	if got := brief(t, in, "9f31", "function"); !strings.Contains(got, "Reaffirmed: 9f31") {
		t.Errorf("reaffirmation lacks its trailer:\n%s", got)
	}
}

func TestNoItem(t *testing.T) {
	in := loadInputs(t, "queue-ada", "9f31")
	if _, err := selectItem(in.Queue, "9f31", "release"); err == nil {
		t.Error("expected an error for a missing item")
	}
}

func TestRunFlags(t *testing.T) {
	if err := run(nil); err == nil {
		t.Error("expected an error without flags")
	}
}
