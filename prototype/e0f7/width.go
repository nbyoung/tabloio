package main

import "strings"

// Model says how a terminal treats U+FE0F (variation selector 16), the
// selector that asks for the emoji form of a symbol such as ⚙️ or 🛠️.
type Model int

const (
	// VS16Wide: the selector widens a narrow base to two cells. Modern
	// terminals (kitty, WezTerm, iTerm2, Windows Terminal, recent VTE) do this.
	VS16Wide Model = iota
	// VS16Narrow: the selector takes no width; the cluster keeps the width of
	// its base. Older wcwidth-based terminals do this, and draw the glyph wide.
	VS16Narrow
	// VS16Strip: the renderer removes the selector before it draws, so the
	// symbol shows in its text form and every terminal agrees on one cell.
	VS16Strip
)

const vs16 = '️'

// narrowEmoji lists emoji-block code points whose East Asian Width is Neutral,
// so that a base without a selector takes one cell.
var narrowEmoji = [][2]rune{
	{0x1F321, 0x1F32C}, {0x1F336, 0x1F336}, {0x1F37D, 0x1F37D},
	{0x1F394, 0x1F39F}, {0x1F3CB, 0x1F3CE}, {0x1F3D4, 0x1F3DF},
	{0x1F3F1, 0x1F3F3}, {0x1F3F5, 0x1F3F7}, {0x1F43F, 0x1F43F},
	{0x1F441, 0x1F441}, {0x1F4FD, 0x1F4FE}, {0x1F53E, 0x1F54A},
	{0x1F56F, 0x1F570}, {0x1F573, 0x1F579}, {0x1F587, 0x1F587},
	{0x1F58A, 0x1F58D}, {0x1F590, 0x1F590}, {0x1F5A5, 0x1F5A8},
	{0x1F5B1, 0x1F5B2}, {0x1F5BC, 0x1F5BC}, {0x1F5C2, 0x1F5C4},
	{0x1F5D1, 0x1F5D3}, {0x1F5DC, 0x1F5DE}, {0x1F5E1, 0x1F5E1},
	{0x1F5E3, 0x1F5E3}, {0x1F5E8, 0x1F5E8}, {0x1F5EF, 0x1F5EF},
	{0x1F5F3, 0x1F5F3}, {0x1F5FA, 0x1F5FA}, {0x1F6CB, 0x1F6CF},
	{0x1F6E0, 0x1F6E5}, {0x1F6E9, 0x1F6E9}, {0x1F6F0, 0x1F6F0},
	{0x1F6F3, 0x1F6F3},
}

// wideSymbols lists BMP symbols with default emoji presentation, which take
// two cells without a selector.
var wideSymbols = map[rune]bool{
	0x231A: true, 0x231B: true, 0x23E9: true, 0x23EA: true, 0x23EB: true,
	0x23EC: true, 0x23F0: true, 0x23F3: true, 0x25FD: true, 0x25FE: true,
	0x2614: true, 0x2615: true, 0x2648: true, 0x2649: true, 0x264A: true,
	0x264B: true, 0x264C: true, 0x264D: true, 0x264E: true, 0x264F: true,
	0x2650: true, 0x2651: true, 0x2652: true, 0x2653: true, 0x267F: true,
	0x2693: true, 0x26A1: true, 0x26AA: true, 0x26AB: true, 0x26BD: true,
	0x26BE: true, 0x26C4: true, 0x26C5: true, 0x26CE: true, 0x26D4: true,
	0x26EA: true, 0x26F2: true, 0x26F3: true, 0x26F5: true, 0x26FA: true,
	0x26FD: true, 0x2705: true, 0x270A: true, 0x270B: true, 0x2728: true,
	0x274C: true, 0x274E: true, 0x2753: true, 0x2754: true, 0x2755: true,
	0x2757: true, 0x2795: true, 0x2796: true, 0x2797: true, 0x27B0: true,
	0x27BF: true, 0x2B1B: true, 0x2B1C: true, 0x2B50: true, 0x2B55: true,
}

func inRanges(r rune, rs [][2]rune) bool {
	for _, p := range rs {
		if r >= p[0] && r <= p[1] {
			return true
		}
	}
	return false
}

// baseWidth is the width of one code point without any selector.
func baseWidth(r rune) int {
	switch {
	case r == 0 || r == '‍' || r == vs16 || r == '︎':
		return 0
	case r >= 0x0300 && r <= 0x036F:
		return 0
	case r < 0x20 || (r >= 0x7F && r < 0xA0):
		return 0
	case wideSymbols[r]:
		return 2
	case r >= 0x1F300 && r <= 0x1FAFF:
		if inRanges(r, narrowEmoji) {
			return 1
		}
		return 2
	case r >= 0x1100 && r <= 0x115F, r >= 0x2E80 && r <= 0xA4CF,
		r >= 0xAC00 && r <= 0xD7A3, r >= 0xF900 && r <= 0xFAFF,
		r >= 0xFE30 && r <= 0xFE6F, r >= 0xFF00 && r <= 0xFF60,
		r >= 0xFFE0 && r <= 0xFFE6, r >= 0x20000 && r <= 0x3FFFD:
		return 2
	}
	return 1
}

// Width is the number of terminal cells s takes under model m.
func Width(s string, m Model) int {
	w := 0
	var prev rune
	for _, r := range s {
		switch {
		case r == vs16 && m == VS16Wide && prev != 0 && baseWidth(prev) == 1:
			w++ // the selector widens a narrow base to two cells
		case r == vs16 && m == VS16Strip:
		default:
			w += baseWidth(r)
		}
		prev = r
	}
	return w
}

// Prepare returns s as the renderer writes it: under VS16Strip the selectors
// leave the text, under the other models it stays as it is.
func Prepare(s string, m Model) string {
	if m == VS16Strip {
		return strings.ReplaceAll(s, string(vs16), "")
	}
	return s
}

// Pad centres (align 'c') or left-aligns (align 'l') s in w cells.
func Pad(s string, w int, align byte, m Model) string {
	gap := w - Width(s, m)
	if gap <= 0 {
		return s
	}
	if align == 'c' {
		l := gap / 2
		return strings.Repeat(" ", l) + s + strings.Repeat(" ", gap-l)
	}
	return s + strings.Repeat(" ", gap)
}

// Truncate cuts s to at most w cells, ending in an ellipsis when it cuts.
func Truncate(s string, w int, m Model) string {
	if Width(s, m) <= w {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		if Width(b.String()+string(r), m)+1 > w {
			break
		}
		b.WriteRune(r)
	}
	return b.String() + "…"
}
