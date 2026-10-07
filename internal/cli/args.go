package cli

import (
	"fmt"
	"strings"
)

// Option is one entry of the vocabulary.
type Option struct {
	Name   string // the key in Args, and the long spelling unless NoLong
	Short  byte   // 'C', 'o', 'h', or 0
	NoLong bool   // true for -C alone
	Arity  int    // the number of values: 0, 1 or 2
	Values string // "<dir>", "<id> <gate>": the help text's operand
	Help   string
}

// Args is a parsed command line: the command, its operands, and the values
// of each option given.
type Args struct {
	Command  string
	Operands []string
	Values   map[string][]string // an option given with no value holds an empty slice
}

// vocabulary is the one set of options every command draws from, in the
// order of design 5ca9's vocabulary table. A command's help lists its own
// options in this order.
var vocabulary = []Option{
	{Name: "C", Short: 'C', NoLong: true, Arity: 1, Values: "<dir>",
		Help: "Run as if started in <dir>; the project is the nearest .tableaux at or above it"},
	{Name: "ref", Arity: 1, Values: "<ref>",
		Help: "Read the project at this commit; history also takes <from>..<to>"},
	{Name: "level", Arity: 1, Values: "<level>",
		Help: "Write at glance (the default), detail or provenance"},
	{Name: "markdown", Help: "Write Markdown (the default)"},
	{Name: "text", Help: "Write Unicode text"},
	{Name: "output", Short: 'o', Arity: 1, Values: "<file>",
		Help: "Write to <file>; - is standard output (the default)"},
	{Name: "task", Arity: 1, Values: "<id>",
		Help: "Focus on one task, or on the subtree under it"},
	{Name: "person", Arity: 1, Values: "<email>",
		Help: "Focus on one person"},
	{Name: "window", Arity: 1, Values: "<n>",
		Help: "Show <n> gate columns either side of the next gates in view"},
	{Name: "columns", Arity: 1, Values: "<gates>",
		Help: "Show these gate columns: gate keys joined by commas"},
	{Name: "historical",
		Help: "Show the marks of historical junctions"},
	{Name: "proposed",
		Help: "Show proposed tasks only"},
	{Name: "stale", Arity: 1, Values: "<days>",
		Help: "Take a status older than <days> as stale"},
	{Name: "brief", Arity: 2, Values: "<id> <gate>",
		Help: "Write one item as a brief"},
	{Name: "help", Short: 'h',
		Help: "Print this text"},
}

// label returns the spelling an error message uses for the option: the long
// form, or -C.
func (o *Option) label() string {
	if o.NoLong {
		return "-" + string(o.Short)
	}
	return "--" + o.Name
}

// Parse splits argv by the rules of "The command line" of design 5ca9. It
// knows the vocabulary and no command; Run checks the options against the
// command.
//
// Parse never returns nil. It reads the whole line and returns the first
// error it met, together with every option it did read, so that Run can let
// --help win over any other check and name the command in a usage error.
func Parse(argv []string, vocabulary []Option) (*Args, error) {
	a := &Args{Values: map[string][]string{}}
	var first error
	fail := func(format string, args ...any) {
		if first == nil {
			first = fmt.Errorf(format, args...)
		}
	}
	find := func(match func(*Option) bool) *Option {
		for i := range vocabulary {
			if match(&vocabulary[i]) {
				return &vocabulary[i]
			}
		}
		return nil
	}

	ended := false // after --, every word is the command or an operand
	for i := 0; i < len(argv); i++ {
		w := argv[i]
		if ended || w == "-" || !strings.HasPrefix(w, "-") {
			if a.Command == "" {
				a.Command = w
			} else {
				a.Operands = append(a.Operands, w)
			}
			continue
		}
		if w == "--" {
			ended = true
			continue
		}

		var o *Option
		var inline string
		var hasInline bool
		if strings.HasPrefix(w, "--") {
			var name string
			name, inline, hasInline = strings.Cut(w[2:], "=")
			o = find(func(o *Option) bool { return !o.NoLong && o.Name == name })
			if o == nil {
				fail("unknown option --%s", name)
				continue
			}
		} else {
			if len(w) > 2 {
				fail("%s: a long option takes two hyphens, as in -%s", w, w)
				continue
			}
			o = find(func(o *Option) bool { return o.Short != 0 && o.Short == w[1] })
			if o == nil {
				fail("unknown option %s", w)
				continue
			}
		}

		var values []string
		switch {
		case hasInline && o.Arity == 0:
			fail("%s takes no =value", o.label())
		case hasInline && o.Arity == 2:
			fail("%s takes two values as separate words, not =value", o.label())
		case hasInline:
			values = []string{inline}
		case i+o.Arity >= len(argv):
			if o.Arity == 1 {
				fail("%s needs a value", o.label())
			} else {
				fail("%s needs two values", o.label())
			}
			i = len(argv)
			continue
		default:
			values = append(values, argv[i+1:i+1+o.Arity]...)
			i += o.Arity
		}
		if _, dup := a.Values[o.Name]; dup {
			fail("%s is given twice", o.label())
			continue
		}
		if values == nil {
			values = []string{}
		}
		a.Values[o.Name] = values
	}
	return a, first
}
