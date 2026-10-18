package action

import "strings"

// reserved lists the plain words YAML 1.1 reads as a boolean or a null, in
// lower case; a value that is one of them in any case goes in quotes.
var reserved = map[string]bool{
	"true": true, "false": true, "yes": true, "no": true, "on": true, "off": true,
	"null": true, "y": true, "n": true,
}

// scalar returns v as a YAML scalar on one line: plain when the plain form
// reads back as v in a block mapping, and in double quotes otherwise, with \
// and " escaped (design b618, "The scalar"). The caller refuses a value with
// a line break or a control character before it gets here.
func scalar(v string) string {
	if plainBlock(v) {
		return v
	}
	return quote(v)
}

// flowScalar returns v as a scalar inside a flow mapping, where the plain
// form also holds only the characters of flowPlain.
func flowScalar(v string) string {
	if plainBlock(v) && flowPlain(v) {
		return v
	}
	return quote(v)
}

// plainBlock reports whether v stays plain in a block mapping: it begins with
// an ASCII letter, is no reserved word, holds neither ": " nor " #", and ends
// with neither a colon nor a space.
func plainBlock(v string) bool {
	if v == "" {
		return false
	}
	if c := v[0] | 0x20; c < 'a' || c > 'z' {
		return false
	}
	return !reserved[strings.ToLower(v)] &&
		!strings.Contains(v, ": ") && !strings.Contains(v, " #") &&
		!strings.HasSuffix(v, ":") && !strings.HasSuffix(v, " ")
}

// flowPlain reports whether every character of v may stand plain inside a
// flow mapping: A-Z a-z 0-9, the space and ._/~()'+=@%:#-.
func flowPlain(v string) bool {
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9':
		case strings.IndexByte(" ._/~()'+=@%:#-", c) >= 0:
		default:
			return false
		}
	}
	return true
}

// quote returns v in double quotes with \ and " escaped.
func quote(v string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v) + `"`
}
