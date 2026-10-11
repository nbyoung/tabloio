package text

import (
	"strings"
	"unicode"

	"golang.org/x/text/width"
)

// selector is U+FE0F, the variation selector that asks for the emoji form.
const selector = '\uFE0F'

// runeWidth is the number of cells one code point takes: 0, 1 or 2.
func runeWidth(r rune) int {
	switch {
	case r < 0x20, r >= 0x7F && r < 0xA0:
		return 0
	case unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf):
		return 0
	}
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	}
	return 1
}

// cells is the number of cells s takes: the sum over its code points.
func cells(s string) int {
	n := 0
	for _, r := range s {
		n += runeWidth(r)
	}
	return n
}

// clean replaces each control character with U+FFFD and removes each U+FE0F
// that follows a code point one cell wide, so that the symbol prints as the
// plain character and takes the one cell the measure counts.
func clean(s string) string {
	var b strings.Builder
	prev := 0
	for _, r := range s {
		if unicode.IsControl(r) {
			r = unicode.ReplacementChar
		}
		if r == selector && prev == 1 {
			prev = 0
			continue
		}
		prev = runeWidth(r)
		b.WriteRune(r)
	}
	return b.String()
}

// units splits a word into the pieces a hard break may fall between: a code
// point with a width, with the zero-width code points after it, and with
// whatever a zero-width joiner ties to it.
func units(s string) []string {
	var out []string
	joined := false
	for _, r := range s {
		if len(out) > 0 && (runeWidth(r) == 0 || joined) {
			out[len(out)-1] += string(r)
		} else {
			out = append(out, string(r))
		}
		joined = r == '\u200D'
	}
	return out
}

// wrap fills lines of at most w cells with the words of s, which single
// spaces separate. A word wider than w breaks between units.
func wrap(s string, w int) []string {
	if s == "" {
		return nil
	}
	w = max(w, 1)
	var lines []string
	line, used := "", 0
	flush := func() {
		lines = append(lines, line)
		line, used = "", 0
	}
	for _, word := range strings.Split(s, " ") {
		ww := cells(word)
		if used > 0 && used+1+ww <= w {
			line += " " + word
			used += 1 + ww
			continue
		}
		if used > 0 {
			flush()
		}
		if ww <= w {
			line, used = word, ww
			continue
		}
		for _, u := range units(word) {
			uw := cells(u)
			if used > 0 && used+uw > w {
				flush()
			}
			line += u
			used += uw
		}
	}
	if used > 0 || line != "" {
		flush()
	}
	return lines
}

// widestWord is the width of the widest word of s.
func widestWord(s string) int {
	n := 0
	for _, word := range strings.Split(s, " ") {
		n = max(n, cells(word))
	}
	return n
}
