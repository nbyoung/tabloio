package brief

import (
	"fmt"
	"strings"
)

// The ways a gate passes, as way names them.
const (
	wayDefined = "defined"
	wayHandoff = "handoff"
	waySelf    = "self"
	wayNone    = "none"
)

// way returns how the gate of d passes: "defined", "handoff" (another
// reviewer accepts it), "self" (the contributor reviews it) or "none".
func way(d Data) string {
	contributor, reviewer := d.Junction.Contributor.Value, d.Junction.Reviewer.Value
	switch {
	case d.Gate.Key == "defined":
		return wayDefined
	case reviewer != "" && reviewer != contributor:
		return wayHandoff
	case reviewer != "":
		return waySelf
	}
	return wayNone
}

// lines returns the lines of the brief of d, without their line feeds.
func lines(d Data) []string {
	var l []string
	l = append(l, title(d)...)
	l = append(l, partItem(d)...)
	l = append(l, partTask(d)...)
	l = append(l, partRequirements(d)...)
	l = append(l, partStatus(d)...)
	l = append(l, partCommit(d)...)
	l = append(l, partCommands(d)...)
	return l
}

// base returns the name of the ref and whether it names more than the
// commit: the name HEAD and a prefix of the commit name nothing.
func base(d Data) (name string, named bool) {
	r := d.Ref
	named = r.Name != "HEAD" && !strings.HasPrefix(r.Commit, r.Name)
	if named {
		return r.Name, true
	}
	return short(r.Commit), false
}

// title returns the title line, the lead paragraph and a blank line.
func title(d Data) []string {
	name, named := base(d)
	at := "at " + short(d.Ref.Commit)
	if named {
		at = fmt.Sprintf("on %s at %s", name, short(d.Ref.Commit))
	}
	l := []string{fmt.Sprintf("Brief: %s %s at %s", d.Task.ID, d.Task.Title, d.Gate.Key), ""}
	l = append(l, wrap(fmt.Sprintf("This brief stands alone. It is one item of the work queue of %s %s, %s: %s.",
		d.Person, at, d.Ref.Date, d.Kind), 0)...)
	return append(l, "")
}

// junctionRow is one row of the junction values in part 1.
type junctionRow struct{ label, value, source string }

// partItem returns part 1, the item.
func partItem(d Data) []string {
	j, g, tid := d.Junction, d.Gate, d.Task.ID
	l := []string{
		"1. The item",
		"",
		fmt.Sprintf("  Task         %s %s", tid, d.Task.Title),
		fmt.Sprintf("  Gate         %s %s, %s: %s", g.Symbol, g.Key, g.Name, g.Criteria),
	}
	rows := []junctionRow{{"Contributor", j.Contributor.Value, sourceLabel(j.Contributor.Source, tid)}}
	if j.Model.Value != "" {
		rows = append(rows, junctionRow{"Model", j.Model.Value, sourceLabel(j.Model.Source, tid)})
	}
	if j.Reviewer.Value != "" {
		rows = append(rows, junctionRow{"Reviewer", j.Reviewer.Value, sourceLabel(j.Reviewer.Source, tid)})
	} else {
		rows = append(rows, junctionRow{"Reviewer", "none", ""})
	}
	pad := 0
	for _, r := range rows {
		pad = max(pad, len([]rune(r.value)))
	}
	for _, r := range rows {
		l = append(l, trimRight(fmt.Sprintf("  %s%s   %s", padRight(r.label, 13), padRight(r.value, pad), r.source)))
	}
	if j.Model.Value != "" {
		l = append(l, "")
		l = append(l, wrap("The model is the plan's statement. Run on a model that matches it, as an "+
			"identifier or by prefix, and record the model you in fact run in the "+
			"Model: trailer.", 2)...)
	}
	return append(l, "")
}

// reference returns a reference as its URL and its text.
func reference(r Reference) string {
	if r.Text == "" {
		return r.URL
	}
	return r.URL + ", " + r.Text
}

// partTask returns part 2, the task.
func partTask(d Data) []string {
	t, j := d.Task, d.Junction
	l := []string{"2. The task", ""}
	l = append(l, wrap(t.Description, 2)...)
	l = append(l, "")
	if len(t.References) > 0 {
		for _, r := range t.References {
			l = append(l, "  Reference: "+reference(r))
		}
	} else {
		l = append(l, "  References: none")
	}
	if len(j.References) > 0 {
		for _, r := range j.References {
			l = append(l, trimRight(fmt.Sprintf("  Junction reference: %s   %s", reference(r), sourceLabel(r.Source, t.ID))))
		}
		l = append(l, wrap("A junction reference expands the gate's criteria for this task. "+
			"Read each one before you start.", 2)...)
	} else {
		l = append(l, "  Junction references: none")
	}
	if len(t.Path) > 0 {
		titles := make([]string, len(t.Path))
		for i, p := range t.Path {
			titles[i] = p.Title
		}
		l = append(l, wrap("Under: "+strings.Join(titles, " › "), 2)...)
	}
	if s := d.Supplier; s != nil {
		l = append(l, "")
		l = append(l, wrap(fmt.Sprintf("%s (%s), whose entry sends this work to you: %s",
			s.Title, s.ID, strings.Join(strings.Fields(s.Description), " ")), 2)...)
	}
	return append(l, "")
}

// who names the tasks of a run: one task by id and title, a folded run by
// its count and its first and last id.
func who(run []Edge) string {
	if len(run) == 1 {
		return run[0].ID + " " + run[0].Title
	}
	return fmt.Sprintf("%d tasks, %s to %s,", len(run), run[0].ID, run[len(run)-1].ID)
}

// unblocks returns what a run of dependents the gate unblocks says.
func unblocks(run []Edge) string {
	return fmt.Sprintf("%s at %s, %s", who(run), run[0].To, run[0].Text)
}

// also returns what a run of the other dependents says.
func also(run []Edge) string {
	e := run[0]
	return fmt.Sprintf("%s from %s to %s, %s: %s", who(run), e.From, e.To, e.Text, e.Condition)
}

// requires returns the Requires line of one requirement.
func requires(e Edge) string {
	where := ""
	if e.URL != "" {
		where = fmt.Sprintf(" in %s at %s", e.URL, short(e.Commit))
	}
	return fmt.Sprintf("Requires: %s %s%s from %s to %s, %s: %s",
		e.ID, e.Title, where, e.From, e.To, e.Text, e.Condition)
}

// partRequirements returns part 3, the requirements.
func partRequirements(d Data) []string {
	tid, gkey := d.Task.ID, d.Gate.Key
	l := []string{"3. The requirements", ""}
	if len(d.Requires) > 0 {
		for _, e := range d.Requires {
			l = append(l, wrap(requires(e), 2)...)
		}
	} else {
		l = append(l, "  Requires: nothing")
	}
	var now, rest []Edge
	for _, e := range d.Dependents {
		if e.From == gkey {
			now = append(now, e)
		} else {
			rest = append(rest, e)
		}
	}
	switch {
	case len(d.Dependents) == 0:
		l = append(l, wrap("Unblocks: nothing; no task requires "+tid, 2)...)
	case len(now) == 0:
		l = append(l, "  Unblocks: nothing at this gate")
	}
	for _, run := range fold(now) {
		l = append(l, wrap("Unblocks: "+unblocks(run), 2)...)
	}
	for _, run := range fold(rest) {
		l = append(l, wrap("Also required by: "+also(run), 2)...)
	}
	if p := d.Parent; p != nil {
		var pnow []Edge
		for _, e := range p.Dependents {
			if e.From == gkey {
				pnow = append(pnow, e)
			}
		}
		if len(pnow) > 0 {
			l = append(l, wrap(fmt.Sprintf("Its parent %s %s at %s is required by:", p.ID, p.Title, gkey), 2)...)
			for _, run := range fold(pnow) {
				l = append(l, wrap(unblocks(run), 4)...)
			}
		}
	}
	return append(l, "")
}

// partStatus returns part 4, the status.
func partStatus(d Data) []string {
	st := d.Status
	head := fmt.Sprintf("%s %s %s %s", st.GateSymbol, st.Gate, st.StateSymbol, st.State)
	if st.Reason != "" {
		head += fmt.Sprintf(" %s %s", st.ReasonSymbol, st.Reason)
	}
	head += fmt.Sprintf(", %s, %s, %s", st.Date, st.Recorder, short(st.Commit))
	l := []string{"4. The status", ""}
	if st.Note != "" {
		l = append(l, wrap(head+":", 2)...)
		l = append(l, wrap(st.Note, 2)...)
	} else {
		l = append(l, wrap(head, 2)...)
	}
	return append(l, "")
}

// partCommit returns part 5, the commit the agent makes when done.
func partCommit(d Data) []string {
	st, g, tid := d.Status, d.Gate, d.Task.ID
	contributor, model, reviewer := d.Junction.Contributor.Value, d.Junction.Model.Value, d.Junction.Reviewer.Value
	w := way(d)
	name, _ := base(d)
	l := []string{"5. The commit you make when done", ""}

	lead := fmt.Sprintf("Commit your work and %s, as %s, on a branch off %s. Do not merge. ", st.Path, contributor, name)
	switch w {
	case wayHandoff:
		lead += fmt.Sprintf("The reviewer %s accepts this gate, so hand the work off: the status stays at "+
			"the gate it stands at and takes the reason review.", reviewer)
	case waySelf:
		lead += "You are the reviewer of this gate as well as its contributor, so the commit " +
			"that records the status passes the gate."
	case wayDefined:
		lead += "Authorisation stands as the review of defined, so the commit that records the " +
			"status passes the gate and carries no Reviewed: trailer."
	default:
		lead += "No one reviews this gate, so the commit that records the status passes it."
	}
	l = append(l, wrap(lead, 2)...)
	l = append(l, "")

	switch {
	case w == wayHandoff:
		l = append(l,
			"    gate: "+st.Gate,
			"    state: nominal",
			"    reason: review",
			"    note: <one line that says what waits on the branch>")
	case g.Follows == "recursive":
		l = append(l, "    gate: "+g.Key)
	default:
		state := "complete"
		if g.Follows != "" {
			state = "nominal"
		}
		l = append(l,
			"    gate: "+g.Key,
			"    state: "+state,
			"    note: <one line that says what the gate now holds>")
	}
	l = append(l, "")
	if w != wayHandoff && g.Follows == "recursive" {
		l = append(l, wrap("The next junction is recursive, so the file holds the gate alone. Edit "+
			"it by hand, and change no other file under .tableaux.", 2)...)
	} else {
		l = append(l, wrap("Edit the file by hand so that it reads as above, with the note on one "+
			"line and no colon followed by a space inside it. Change no other file "+
			"under .tableaux.", 2)...)
	}
	l = append(l, "")

	switch {
	case model != "":
		l = append(l, wrap(fmt.Sprintf("Author and committer are both %s. Set the identity in the environment, "+
			"which no configuration overrides, and check it on each commit:", contributor), 2)...)
		l = append(l, "",
			"    export GIT_AUTHOR_EMAIL="+contributor,
			"    export GIT_COMMITTER_EMAIL="+contributor,
			`    export GIT_AUTHOR_NAME="<your model's display name>"`,
			`    export GIT_COMMITTER_NAME="<your model's display name>"`,
			"    git log -1 --format='%an <%ae> / %cn <%ce>%n%B'",
			"")
		tail := "End each commit message with one final paragraph of trailers, with no blank " +
			"line between them and nothing after them."
		if w == waySelf {
			tail += " The commit that changes the status file carries all three; an earlier " +
				"commit carries the last two."
		}
		l = append(l, wrap(tail, 2)...)
		l = append(l, "")
		if w == waySelf {
			l = append(l, fmt.Sprintf("    Reviewed: %s %s", tid, g.Key))
		}
		l = append(l,
			fmt.Sprintf("    Model: <the identifier your harness reports, matching %s>", model),
			fmt.Sprintf("    Co-Authored-By: <your model's display name> <%s>", contributor))
	case w == waySelf:
		l = append(l, wrap("End the message of the commit that changes the status file with one "+
			"final paragraph of one trailer:", 2)...)
		l = append(l, "", fmt.Sprintf("    Reviewed: %s %s", tid, g.Key))
	default:
		l = append(l, wrap("A person contributes here, so the commit carries no trailer.", 2)...)
	}
	l = append(l, "")

	if w == wayHandoff {
		l = append(l, wrap(fmt.Sprintf("The reviewer merges the branch and passes the gate with a commit that "+
			"carries the trailer Reviewed: %s %s", tid, g.Key), 2)...)
	} else {
		l = append(l, wrap("Whoever dispatches you merges the branch.", 2)...)
	}
	return append(l, "")
}

// partCommands returns part 6, the commands that reproduce the brief and show
// the task. The last line ends the text.
func partCommands(d Data) []string {
	at := short(d.Ref.Commit)
	return []string{
		"6. The commands that reproduce this brief and show the task",
		"",
		fmt.Sprintf("  tabloio queue --person %s --ref %s --brief %s %s", d.Person, at, d.Task.ID, d.Gate.Key),
		fmt.Sprintf("  tabloio task %s --ref %s", d.Task.ID, at),
	}
}
