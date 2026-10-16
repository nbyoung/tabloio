package view

import "testing"

// TestDecodeHistory: Decode reads the history, ignores an unknown key and leaves an absent field zero.
func TestDecodeHistory(t *testing.T) {
	v, err := Decode("history", []byte(`{"command": "history", "params": {"level": "detail"}, "data": {
		"extra": 1,
		"subject": {"id": "e9c6", "title": "Abstract views"},
		"lines": [{"task": {"id": "e9c6"}, "events": [
			{"date": "2026-09-29", "commit": "abc", "by": "a@b.c", "kind": "task"},
			{"date": "2026-09-29", "commit": "abc", "by": "a@b.c", "kind": "pin", "pin": {"url": "u", "new": "abc"},
				"subproject": {"url": "u", "task": "f1a0"}}],
			"check": {"reading": "agrees", "stated": "m"}, "effect": "authorised_commit"}],
		"events": 2, "commits": 1, "kinds": [{"kind": "task", "count": 1}],
		"days": [{"date": "2026-09-29", "events": 2, "commits": 1}]}}`))
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
	if len(l.Events) != 2 || l.Events[1].Pin.Old != "" || l.Events[1].Pin.New != "abc" || l.Events[1].Subproject.Task != "f1a0" ||
		l.Events[0].Recorded != nil || l.Proposal || l.Check.Reading != "agrees" || l.Effect != "authorised_commit" || l.After != nil {
		t.Errorf("decoded line %+v", l)
	}

	if _, err := Decode("history", []byte(`{"command": "history", "params": {"level": "deep"}, "data": {}}`)); err == nil {
		t.Error("an unknown level decodes")
	}
}
