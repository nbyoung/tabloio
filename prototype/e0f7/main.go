// Command e0f7 is the prototype of tabloio's --text renderer. It draws the
// global tableau from tablo's view data as a Unicode box table sized to a
// terminal width, and answers how wide emoji gate symbols and variation
// selectors are in a box-drawn column.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	in := flag.String("in", "testdata/global-tableau.json", "global tableau view data")
	width := flag.Int("width", 0, "terminal width in cells; 0 reads $COLUMNS, else 80")
	vs := flag.String("vs16", "wide", "terminal model for U+FE0F: wide, narrow or strip")
	probe := flag.Bool("probe", false, "print a per-symbol alignment probe instead of the table")
	flag.Parse()

	var m Model
	switch *vs {
	case "wide":
		m = VS16Wide
	case "narrow":
		m = VS16Narrow
	case "strip":
		m = VS16Strip
	default:
		fmt.Fprintln(os.Stderr, "e0f7: -vs16 takes wide, narrow or strip")
		os.Exit(2)
	}
	if *width == 0 {
		*width = 80
		if c, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && c > 0 {
			*width = c
		}
	}
	data, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "e0f7:", err)
		os.Exit(1)
	}
	t, err := ParseTableau(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "e0f7:", err)
		os.Exit(1)
	}
	if *probe {
		fmt.Print(Probe(t, m))
		return
	}
	fmt.Print(RenderTableau(t, *width, m))
}
