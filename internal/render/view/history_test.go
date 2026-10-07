package view

import "testing"

// TestDecodeHistory: Decode reads the history, ignores an unknown key and leaves an absent field zero.
func TestDecodeHistory(t *testing.T) {
	v, err := Decode("history", []byte(`{
		"level": "detail", "extra": 1,
		"subject": {"id": "e9c6", "title": "Abstract views"},
		"lines": [{"date": "2026-09-29", "kinds": ["task", "authorised"], "task": {"id": "e9c6"},
			"pin": {"url": "u", "new": "abc"}, "subproject": {"url": "u", "events": 2, "lines": [{"model": "m"}]}}],
		"events": 2, "commits": 1, "kinds": [{"kind": "task", "count": 1}],
		"days": [{"date": "2026-09-29", "events": 2, "commits": 1}]}`))
	if err != nil {
		t.Fatal(err)
	}
	h, ok := v.(*History)
	if !ok {
		t.Fatalf("got %T", v)
	}
	if h.Header().Level != Detail || h.Subject.ID != "e9c6" || h.Events != 2 || h.Kinds[0].Count != 1 || h.Days[0].Commits != 1 {
		t.Errorf("decoded %+v", h)
	}
	l := h.Lines[0]
	if len(l.Kinds) != 2 || l.Pin.Old != "" || l.Pin.New != "abc" || l.Sub.Events != 2 || l.Sub.Lines[0].Model != "m" || l.Status != nil || l.Proposal {
		t.Errorf("decoded line %+v", l)
	}

	if _, err := Decode("history", []byte(`{"level": "deep"}`)); err == nil {
		t.Error("an unknown level decodes")
	}
}
