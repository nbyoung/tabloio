package render

import "testing"

// T12: the count helper.
func TestCount(t *testing.T) {
	for _, c := range []struct {
		n         int
		one, many string
		want      string
	}{
		{0, "cause", "causes", "0 causes"},
		{1, "cause", "causes", "1 cause"},
		{2, "cause", "causes", "2 causes"},
		{0, "status off nominal", "statuses off nominal", "0 statuses off nominal"},
		{1, "status off nominal", "statuses off nominal", "1 status off nominal"},
		{2, "status off nominal", "statuses off nominal", "2 statuses off nominal"},
		{0, "review owed", "reviews owed", "0 reviews owed"},
		{1, "review owed", "reviews owed", "1 review owed"},
		{2, "review owed", "reviews owed", "2 reviews owed"},
		{0, "work ready", "work ready", "0 work ready"},
		{1, "work ready", "work ready", "1 work ready"},
		{2, "work ready", "work ready", "2 work ready"},
		{1, "day", "days", "1 day"},
		{28, "day", "days", "28 days"},
	} {
		if got := count(c.n, c.one, c.many); got != c.want {
			t.Errorf("count(%d, %q, %q) = %q, want %q", c.n, c.one, c.many, got, c.want)
		}
	}
}
