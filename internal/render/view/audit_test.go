package view

import "testing"

// TestDecodeAudit: Decode reads the audit, ignores an unknown key and leaves
// an absent field zero.
func TestDecodeAudit(t *testing.T) {
	v, err := Decode("audit", []byte(`{
		"level": "provenance", "extra": 1, "errors": 1,
		"most": {"email": "a@example.org", "count": 3},
		"findings": [{"rule": "S11", "severity": "error", "tasks": [{"id": "e3cb"}], "files": ["status/e3cb.yaml"],
			"source": {"text": "README.md", "url": "README.md#status"}, "facts": [{"name": "Junction", "value": "v"}],
			"commits": [{"hash": "abc"}], "commands": [{"text": "git log", "comment": "c"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	a, ok := v.(*Audit)
	if !ok {
		t.Fatalf("got %T", v)
	}
	f := a.Findings[0]
	if a.Errors != 1 || a.Most.Count != 3 || f.Rule != "S11" || f.Source.URL != "README.md#status" || f.Facts[0].Name != "Junction" ||
		f.Commits[0].Hash != "abc" || f.Commands[0].Comment != "c" || f.Gate != "" {
		t.Errorf("decoded %+v", a)
	}

	if _, err := Decode("audit", []byte(`{"level": "deep"}`)); err == nil {
		t.Error("an unknown level decodes")
	}
}
