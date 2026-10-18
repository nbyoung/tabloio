package action

import (
	"fmt"
	"strings"
)

// statusKeys are the four fields of a status file, in the order the edit
// takes them and inserts a new one.
var statusKeys = []string{"gate", "state", "reason", "note"}

// statusFormError is a status file the line edit cannot read: a line in
// column one that is neither a comment, blank nor one of the four keys, or a
// key that appears twice. Record refuses it as status-form.
type statusFormError struct{ text string }

func (e *statusFormError) Error() string { return e.text }

// editStatus changes a status file line by line and returns the new bytes
// (design b618, "The edit"). edits maps a key to its new value, or to nil to
// remove the field; a key edits lacks keeps its lines byte for byte. A
// changed field's lines become the one line "<key>: <scalar>"; a new field
// enters after the lines of the nearest earlier field in the order of
// statusKeys, else before the nearest later one, else at the end. The line
// ending is the file's first one.
func editStatus(text string, edits map[string]*string) (string, error) {
	lines := strings.SplitAfter(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	eol := "\n"
	if len(lines) > 0 && strings.HasSuffix(lines[0], "\r\n") {
		eol = "\r\n"
	}
	seen := map[string]bool{}
	for _, l := range lines {
		if strings.TrimSpace(l) == "" || strings.HasPrefix(l, "#") || indented(l) {
			continue
		}
		key := keyOf(l)
		if key == "" {
			return "", &statusFormError{fmt.Sprintf("the line %q is no field of a status file; edit the file by hand", strings.TrimRight(l, "\r\n"))}
		}
		if seen[key] {
			return "", &statusFormError{fmt.Sprintf("the key %s appears twice; edit the file by hand", key)}
		}
		seen[key] = true
	}
	for k, key := range statusKeys {
		v, ok := edits[key]
		if !ok {
			continue
		}
		start, end, found := extent(lines, key)
		if v == nil {
			if found {
				lines = append(lines[:start], lines[end:]...)
			}
			continue
		}
		line := key + ": " + scalar(*v) + eol
		if found {
			lines = append(lines[:start], append([]string{line}, lines[end:]...)...)
			continue
		}
		pos := -1
		for j := k - 1; j >= 0 && pos < 0; j-- {
			if _, e, ok := extent(lines, statusKeys[j]); ok {
				pos = e
			}
		}
		for j := k + 1; j < len(statusKeys) && pos < 0; j++ {
			if s, _, ok := extent(lines, statusKeys[j]); ok {
				pos = s
			}
		}
		if pos < 0 {
			pos = len(lines)
		}
		if pos > 0 && !strings.HasSuffix(lines[pos-1], "\n") {
			lines[pos-1] += eol
		}
		lines = append(lines[:pos], append([]string{line}, lines[pos:]...)...)
	}
	return strings.Join(lines, ""), nil
}

// keyOf returns the status key a line begins with in column one, followed by
// a space, a line ending or nothing; "" for none.
func keyOf(l string) string {
	for _, key := range statusKeys {
		if rest, ok := strings.CutPrefix(l, key+":"); ok &&
			(rest == "" || rest[0] == ' ' || rest == "\n" || rest == "\r\n") {
			return key
		}
	}
	return ""
}

// indented reports whether a line begins with a space or a tab.
func indented(l string) bool { return l != "" && (l[0] == ' ' || l[0] == '\t') }

// extent returns the lines of a field: the first line that begins with the
// key in column one, and the indented lines after it, blank lines between
// them included, so that a folded value goes whole.
func extent(lines []string, key string) (start, end int, found bool) {
	for i, l := range lines {
		if keyOf(l) != key {
			continue
		}
		e := i + 1
		for e < len(lines) {
			if indented(lines[e]) {
				e++
				continue
			}
			if strings.TrimSpace(lines[e]) == "" {
				j := e
				for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
					j++
				}
				if j < len(lines) && indented(lines[j]) {
					e = j
					continue
				}
			}
			break
		}
		return i, e, true
	}
	return 0, 0, false
}
