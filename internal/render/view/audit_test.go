package view

import "testing"

// TestDecodeAudit: Decode reads the audit, ignores an unknown key and leaves
// an absent field zero.
func TestDecodeAudit(t *testing.T) {
	v, err := Decode("audit", []byte(`{"command": "audit", "params": {"level": "provenance"}, "extra": 1, "data": {
		"extra": 1, "errors": 1,
		"most": {"email": "a@example.org", "count": 3},
		"findings": [{"key": "S11", "rule": "S11", "severity": "error", "act": "review", "tasks": [{"id": "e3cb"}], "files": ["status/e3cb.yaml"],
			"source": {"text": "README.md", "url": "README.md#status"}, "facts": [{"task": "e3cb", "junction": {"gate": "implementation"}, "status": null}],
			"commits": [{"hash": "abc"}], "commands": [{"text": "git log", "comment": "c"}]}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	a, ok := v.(*Audit)
	if !ok {
		t.Fatalf("got %T", v)
	}
	f := a.Findings[0]
	if a.Errors != 1 || a.Most.Count != 3 || f.Rule != "S11" || f.Source.URL != "README.md#status" || f.Facts[0].Junction.Gate != "implementation" || f.Facts[0].Status != nil || f.Act != "review" ||
		f.Commits[0].Hash != "abc" || f.Commands[0].Comment != "c" || f.Gate != "" {
		t.Errorf("decoded %+v", a)
	}

	if _, err := Decode("audit", []byte(`{"command": "audit", "params": {"level": "deep"}, "data": {}}`)); err == nil {
		t.Error("an unknown level decodes")
	}
}
