package view

import (
	"bytes"
	"testing"
)

// The structural views decode from their fixtures, ignore an unknown key and
// reach their data through the embedded head.
func TestDecodeStructural(t *testing.T) {
	for _, c := range []struct{ view, name string }{
		{"authority", "tooling"},
		{"authority", "weather-proposed"},
		{"assignment", "tooling"},
		{"assignment", "weather"},
	} {
		t.Run(c.view+"/"+c.name, func(t *testing.T) {
			data := fixture(t, c.view, c.name)
			v, err := Decode(c.view, data)
			if err != nil {
				t.Fatal(err)
			}
			if h := v.Header(); h.Level != Provenance || h.Project.ID == "" || len(h.Legend.Gates) == 0 {
				t.Errorf("head not decoded: %+v", h.Project)
			}
			extra := bytes.Replace(data, []byte("{"), []byte(`{"unknown_key": [1, {"a": 2}], `), 1)
			if _, err := Decode(c.view, extra); err != nil {
				t.Errorf("an unknown key fails: %v", err)
			}
			deep := bytes.Replace(data, []byte(`"level": "`), []byte(`"level": "deep`), 1)
			if _, err := Decode(c.view, deep); err == nil {
				t.Errorf("an unknown level decodes")
			}
		})
	}

	d, err := Decode("authority", fixture(t, "authority", "tooling"))
	if err != nil {
		t.Fatal(err)
	}
	v := d.(*Delegation)
	if len(v.Rows) != 37 || len(v.Parents) != 5 || len(v.Deciding) != 4 || v.Deciding[0].Merged != "320cf2b" {
		t.Errorf("authority: %d rows, %d parents, %d deciding", len(v.Rows), len(v.Parents), len(v.Deciding))
	}
	if m := v.Parents[1].Defaults[0]; m.Gate != "design" || len(m.Marks) != 2 || m.ReviewerFrom != "437e" {
		t.Errorf("a default: %+v", m)
	}

	a, err := Decode("assignment", fixture(t, "assignment", "tooling"))
	if err != nil {
		t.Fatal(err)
	}
	w := a.(*Assignment)
	if len(w.People) != 2 || len(w.Sections) != 2 || len(w.Recursive) != 4 || len(w.Sections[1].Contributes) != 79 {
		t.Errorf("assignment: %d people, %d sections", len(w.People), len(w.Sections))
	}
	if r := w.Sections[0].Assigned[3]; r.ID != "ac33" || r.Under != "bc63" || r.Next == nil || r.Gate != "implementation" {
		t.Errorf("an assigned row: %+v", r)
	}
}
