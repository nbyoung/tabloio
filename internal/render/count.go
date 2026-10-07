package render

import "strconv"

// count prints a number with the noun that agrees with it: `1 cause`,
// `2 causes`, `0 causes`.
func count(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}
