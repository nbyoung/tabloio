package brief

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// width is the longest line, in runes, the text wraps to, indent included.
const width = 76

// foldAbove is the length a run of dependents alike but for the task must
// exceed to fold into one line.
const foldAbove = 8

// wrap fills text greedily to width, each line indented by indent spaces. A
// blank line in text ends a paragraph and shows as an empty line.
func wrap(text string, indent int) []string {
	var out []string
	pad := strings.Repeat(" ", indent)
	for n, para := range strings.Split(strings.TrimSpace(text), "\n\n") {
		if n > 0 {
			out = append(out, "")
		}
		line := ""
		for _, word := range strings.Fields(para) {
			switch {
			case line == "":
				line = pad + word
			case utf8.RuneCountInString(line)+1+utf8.RuneCountInString(word) <= width:
				line += " " + word
			default:
				out = append(out, line)
				line = pad + word
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// short returns the first seven digits of a commit.
func short(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

// sourceLabel returns the parenthesis that names where a field resolves
// from, or "" when the source has no known kind.
func sourceLabel(s Source, taskID string) string {
	switch s.Kind {
	case "task":
		if s.ID == taskID {
			return "(this task)"
		}
		return fmt.Sprintf("(%s, %s)", s.Title, s.ID)
	case "assignee":
		return "(the assignee; the junction states none)"
	case "default":
		return "(the assignee, by default)"
	}
	return ""
}

// fold groups consecutive edges alike but for the task. A run of more than
// foldAbove edges stays one group; a shorter run splits into groups of one.
func fold(edges []Edge) [][]Edge {
	var runs [][]Edge
	for _, e := range edges {
		if n := len(runs); n > 0 && alike(runs[n-1][0], e) {
			runs[n-1] = append(runs[n-1], e)
		} else {
			runs = append(runs, []Edge{e})
		}
	}
	var out [][]Edge
	for _, run := range runs {
		if len(run) > foldAbove {
			out = append(out, run)
			continue
		}
		for _, e := range run {
			out = append(out, []Edge{e})
		}
	}
	return out
}

// alike reports whether two edges differ at most in the task.
func alike(a, b Edge) bool {
	return a.From == b.From && a.To == b.To && a.Text == b.Text && a.Condition == b.Condition
}

// padRight pads s with spaces to n runes.
func padRight(s string, n int) string {
	if k := utf8.RuneCountInString(s); k < n {
		return s + strings.Repeat(" ", n-k)
	}
	return s
}

// trimRight removes trailing white space.
func trimRight(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
