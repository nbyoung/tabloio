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

// kinds words the kind of a work item, as the work queue names it.
var kinds = map[string]string{"ready": "work ready", "waiting": "work waiting"}

// causes words why an item waits, one phrase per kind of cause, as the
// work queue words them, joined by "; ".
func causes(d Data) string {
	var parts []string
	for _, w := range d.Causes {
		switch w.Kind {
		case "requirement":
			if r := w.Requires; r != nil {
				parts = append(parts, fmt.Sprintf("waits for %s at %s", strings.TrimSpace(r.Task.ID+" "+r.Task.Title), r.From))
			}
		case "reason":
			p := w.Reason
			if d.Status.Note != "" {
				p += ": " + d.Status.Note
			}
			parts = append(parts, p)
		case "authorisation":
			parts = append(parts, "proposed; waits for its authorisation by "+w.By)
		case "review":
			parts = append(parts, "handed off; waits for the review by "+w.By)
		}
	}
	return strings.Join(parts, "; ")
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
		d.Person, at, d.Ref.Date, kinds[d.Kind]), 0)...)
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
	switch {
	case j.Reviewer.Value != "":
		rows = append(rows, junctionRow{"Reviewer", j.Reviewer.Value, sourceLabel(j.Reviewer.Source, tid)})
	case j.Reviewer.Source.Kind == "task":
		// The entry waives review with reviewer: false.
		rows = append(rows, junctionRow{"Reviewer", "none", sourceLabel(j.Reviewer.Source, tid)})
	default:
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

// partStatus returns part 4, the status. The head line leaves out an empty
// recorder and an empty commit.
func partStatus(d Data) []string {
	st := d.Status
	head := fmt.Sprintf("%s %s %s %s", st.GateSymbol, st.Gate, st.StateSymbol, st.State)
	if st.Reason != "" {
		head += fmt.Sprintf(" %s %s", st.ReasonSymbol, st.Reason)
	}
	head += ", " + st.Date
	for _, f := range []string{st.Recorder, short(st.Commit)} {
		if f != "" {
			head += ", " + f
		}
	}
	l := []string{"4. The status", ""}
	if st.Note != "" {
		l = append(l, wrap(head+":", 2)...)
		l = append(l, wrap(st.Note, 2)...)
	} else {
		l = append(l, wrap(head, 2)...)
	}
	return append(l, "")
}

// partCommit returns part 5, the commit the agent makes when done. The
// status file is the write commands' to change: the brief quotes the
// `tabloio record` line that writes it and makes the commit (design b618,
// A10 and decision 14 d), and, for a hand-off, the `tabloio review` line of
// the reviewer.
func partCommit(d Data) []string {
	st, g, tid := d.Status, d.Gate, d.Task.ID
	contributor, model, reviewer := d.Junction.Contributor.Value, d.Junction.Model.Value, d.Junction.Reviewer.Value
	w := way(d)
	name, _ := base(d)
	l := []string{"5. The commit you make when done", ""}

	lead := fmt.Sprintf("Commit your work, as %s, on a branch off %s, and record %s there "+
		"with tabloio record. Do not merge. ", contributor, name, st.Path)
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

	// The file as the command writes it, and the command's own options.
	var file, opts []string
	switch {
	case w == wayHandoff:
		file = []string{
			"    gate: " + st.Gate,
			"    state: nominal",
			"    reason: review",
			"    note: <one line that says what waits on the branch>"}
		opts = []string{"--state nominal --reason review", `--note "<one line that says what waits on the branch>"`}
	case g.Follows == "recursive":
		file = []string{"    gate: " + g.Key}
		opts = []string{"--gate " + g.Key}
	default:
		state := "complete"
		if g.Follows != "" {
			state = "nominal"
		}
		file = []string{
			"    gate: " + g.Key,
			"    state: " + state,
			"    note: <one line that says what the gate now holds>"}
		opts = []string{"--gate " + g.Key + " --state " + state, `--note "<one line that says what the gate now holds>"`}
	}
	if model != "" {
		opts = append(opts,
			fmt.Sprintf(`--model "<the identifier your harness reports, matching %s>"`, model),
			fmt.Sprintf(`--co-author "<your model's display name> <%s>"`, contributor))
	}
	opts[len(opts)-1] += " --yes"
	if w != wayHandoff && g.Follows == "recursive" {
		l = append(l, wrap("The next junction is recursive, so the file holds the gate alone. tabloio "+
			"record writes it so, and makes the commit that records it:", 2)...)
	} else {
		l = append(l, wrap("tabloio record writes the file so, with the note on one line, and makes "+
			"the commit that records it:", 2)...)
	}
	l = append(l, "")
	l = append(l, file...)
	l = append(l, "")
	cmd := "    tabloio record " + tid + " " + opts[0]
	for _, o := range opts[1:] {
		l = append(l, cmd+" \\")
		cmd = "      " + o
	}
	l = append(l, cmd, "")
	l = append(l, wrap("Change no file under .tableaux by hand: the command refuses a status the "+
		"method cannot read, and names the rule.", 2)...)
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
			"line between them and nothing after them. tabloio record composes those of the " +
			"commit that records the status from --model and --co-author"
		if w == waySelf {
			tail += ", and adds the first, since you review this gate; an earlier commit " +
				"carries the last two."
		} else {
			tail += "."
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
		l = append(l, wrap("tabloio record ends the message of the commit that records the status "+
			"with one final paragraph of one trailer, since you review this gate:", 2)...)
		l = append(l, "", fmt.Sprintf("    Reviewed: %s %s", tid, g.Key))
	default:
		// A person, or an agent whose junction states no model.
		l = append(l, wrap("The junction states no model: an agent records the model it runs in a Model: trailer, and a person's commit carries none.", 2)...)
	}
	l = append(l, "")

	if w == wayHandoff {
		l = append(l, wrap(fmt.Sprintf("The reviewer merges the branch and accepts the gate with tabloio "+
			"review %s %s, whose commit carries the trailer Reviewed: %s %s", tid, g.Key, tid, g.Key), 2)...)
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
