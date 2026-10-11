package text

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// parseCodePoints reads "U+2699 U+FE0F" as a string.
func parseCodePoints(t *testing.T, s string) string {
	t.Helper()
	var b strings.Builder
	for _, f := range strings.Fields(s) {
		n, err := strconv.ParseInt(strings.TrimPrefix(f, "U+"), 16, 32)
		if err != nil {
			t.Fatalf("code point %q: %v", f, err)
		}
		b.WriteRune(rune(n))
	}
	return b.String()
}

// T1: every row of widths.tsv, what the writer writes and its cells.
func TestWidthsTable(t *testing.T) {
	f, err := os.Open("testdata/widths.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Scan() // header
	rows := 0
	for sc.Scan() {
		cols := strings.Split(sc.Text(), "\t")
		if len(cols) != 5 {
			t.Fatalf("row %q has %d columns", sc.Text(), len(cols))
		}
		in, want := parseCodePoints(t, cols[2]), parseCodePoints(t, cols[3])
		wantCells, err := strconv.Atoi(cols[4])
		if err != nil {
			t.Fatal(err)
		}
		if in != cols[1] {
			t.Errorf("%s: the symbol column differs from its code points", cols[1])
		}
		rows++
		t.Run(cols[1], func(t *testing.T) {
			got := clean(in)
			if got != want {
				t.Errorf("clean writes % X, want % X", []rune(got), []rune(want))
			}
			if n := cells(got); n != wantCells {
				t.Errorf("cells = %d, want %d", n, wantCells)
			}
		})
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 60 {
		t.Errorf("widths.tsv holds %d rows, want 60", rows)
	}
}

func TestRuneWidth(t *testing.T) {
	for _, c := range []struct {
		r    rune
		want int
	}{
		{'a', 1}, {'—', 1}, {'─', 1}, {'日', 2}, {'🟢', 2}, {0x0301, 0}, {0x200D, 0},
		{0x1B, 0}, {0x7F, 0}, {0x9F, 0}, {0xFE0F, 0}, {0xFFFD, 1},
	} {
		if got := runeWidth(c.r); got != c.want {
			t.Errorf("runeWidth(%U) = %d, want %d", c.r, got, c.want)
		}
	}
}
