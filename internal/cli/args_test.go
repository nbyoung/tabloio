package cli

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		argv []string
		want Args
	}{
		{"empty", nil, Args{Values: map[string][]string{}}},
		{"command alone", []string{"tableau"}, Args{Command: "tableau", Values: map[string][]string{}}},
		{"options before and after", []string{"-C", "d", "task", "--level", "detail", "e9c6", "--historical"},
			Args{Command: "task", Operands: []string{"e9c6"}, Values: map[string][]string{
				"C": {"d"}, "level": {"detail"}, "historical": {}}}},
		{"equals", []string{"tableau", "--ref=main", "--window=0"},
			Args{Command: "tableau", Values: map[string][]string{"ref": {"main"}, "window": {"0"}}}},
		{"equals keeps later equals", []string{"tableau", "--ref=a=b"},
			Args{Command: "tableau", Values: map[string][]string{"ref": {"a=b"}}}},
		{"short options", []string{"tableau", "-o", "S.md", "-h"},
			Args{Command: "tableau", Values: map[string][]string{"output": {"S.md"}, "help": {}}}},
		{"two values", []string{"queue", "--brief", "c2ad", "validate", "--ref", "main"},
			Args{Command: "queue", Values: map[string][]string{"brief": {"c2ad", "validate"}, "ref": {"main"}}}},
		{"a value may start with a hyphen", []string{"tableau", "--window", "-1"},
			Args{Command: "tableau", Values: map[string][]string{"window": {"-1"}}}},
		{"a lone hyphen is a value", []string{"tableau", "--output", "-"},
			Args{Command: "tableau", Values: map[string][]string{"output": {"-"}}}},
		{"an empty value", []string{"history", "--ref", ""},
			Args{Command: "history", Values: map[string][]string{"ref": {""}}}},
		{"double hyphen ends the options", []string{"task", "--", "--level", "x"},
			Args{Command: "task", Operands: []string{"--level", "x"}, Values: map[string][]string{}}},
		{"double hyphen may precede the command", []string{"--", "tableau"},
			Args{Command: "tableau", Values: map[string][]string{}}},
	}
	for _, c := range cases {
		got, err := Parse(c.argv, vocabulary)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if !reflect.DeepEqual(*got, c.want) {
			t.Errorf("%s:\n got %#v\nwant %#v", c.name, *got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		argv []string
		want string
	}{
		{[]string{"tableau", "--bogus"}, "unknown option --bogus"},
		{[]string{"tableau", "--bogus=1"}, "unknown option --bogus"},
		{[]string{"tableau", "-x"}, "unknown option -x"},
		{[]string{"tableau", "-ref", "main"}, "takes two hyphens, as in --ref"},
		{[]string{"tableau", "--C", "x"}, "unknown option --C"},
		{[]string{"tableau", "-C"}, "-C needs a value"},
		{[]string{"tableau", "--level"}, "--level needs a value"},
		{[]string{"queue", "--brief", "c2ad"}, "--brief needs two values"},
		{[]string{"queue", "--brief"}, "--brief needs two values"},
		{[]string{"tableau", "--level", "glance", "--level", "detail"}, "--level is given twice"},
		{[]string{"tableau", "-o", "a", "--output", "b"}, "--output is given twice"},
		{[]string{"tableau", "-C", "a", "-C", "b"}, "-C is given twice"},
		{[]string{"tableau", "--historical=yes"}, "--historical takes no =value"},
		{[]string{"queue", "--brief=c2ad"}, "--brief takes two values"},
		{[]string{"tableau", "-C=d"}, "takes two hyphens"},
	}
	for _, c := range cases {
		a, err := Parse(c.argv, vocabulary)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%v: error %v, want one with %q", c.argv, err, c.want)
		}
		if a == nil || a.Command != c.argv[0] {
			t.Errorf("%v: the partial result lost the command: %#v", c.argv, a)
		}
	}
}

// TestParseKeepsWhatItRead checks that a --help after an error still reaches
// Run, which lets it win over every other check.
func TestParseKeepsWhatItRead(t *testing.T) {
	a, err := Parse([]string{"tableau", "--bogus", "-h", "--level"}, vocabulary)
	if err == nil || !strings.Contains(err.Error(), "--bogus") {
		t.Errorf("error %v, want the first one, about --bogus", err)
	}
	if _, ok := a.Values["help"]; !ok {
		t.Errorf("Parse dropped --help: %#v", a.Values)
	}
}

func TestVocabularyIsWellFormed(t *testing.T) {
	names := map[string]bool{}
	shorts := map[byte]bool{}
	for _, o := range vocabulary {
		if names[o.Name] {
			t.Errorf("option %s twice", o.Name)
		}
		names[o.Name] = true
		if o.Short != 0 {
			if shorts[o.Short] {
				t.Errorf("short option -%c twice", o.Short)
			}
			shorts[o.Short] = true
		}
		if o.NoLong && o.Short == 0 {
			t.Errorf("%s has no spelling", o.Name)
		}
		if (o.Arity == 0) != (o.Values == "") {
			t.Errorf("%s: arity %d and operand %q disagree", o.Name, o.Arity, o.Values)
		}
		if o.Help == "" {
			t.Errorf("%s has no help line", o.Name)
		}
	}
}
