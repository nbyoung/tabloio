package action

import "testing"

// TestCorpusEntries is T29, which falls to the integrate gate: on the
// corpus entries weather-station, junction-kinds, submodule-subproject,
// status-gate-not-applicable, status-complete-early,
// status-state-with-recursive, status-unreviewed-gate and
// review-by-non-reviewer, each refusal fires from tablo's own facts, and
// after each commit tablo validate reports no new error and tablo history
// shows the event.
func TestCorpusEntries(t *testing.T) {
	t.Skip("T29 falls to the integrate gate: it needs tablo's 4ed9, 8118 and 493e at implementation")
}
