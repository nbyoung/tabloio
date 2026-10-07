package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// The exit codes of every tabloio command. They keep the meanings tablo's
// plumbing command gives them (prototype 4ed9).
const (
	ExitOK      = 0 // the command wrote what it was asked for
	ExitInvalid = 1 // an error diagnostic prevents the view; nothing is written
	ExitUsage   = 2 // the command line is wrong; nothing is read
	ExitFailure = 3 // a read, a rendering or a write fails
)

// Env carries what a command takes from the process, so that a test
// replaces each part.
type Env struct {
	Stdout, Stderr io.Writer
	Version        string // main.version, as GoReleaser sets it
	Source         Source
}

// Command is one subcommand. The read commands fill ten; the Write commands
// task adds its own to the same table.
type Command struct {
	Name     string
	Group    string   // "read", "write" or "other": the section of the usage text
	Operands string   // "<id>" for task, "" otherwise
	Summary  string   // the view's question, verbatim from VIEWS.md
	Options  []string // the option names it takes beyond -C and --help
	Run      func(ctx context.Context, env *Env, a *Args) int
}

// groups lists the sections of the usage text in order, each with its
// heading. A group with no command leaves its section out. The Write
// commands task adds its own entry.
var groups = []struct{ name, heading string }{
	{"read", "Read commands write one view and change nothing in the project."},
	{"other", "Other commands."},
}

// commands is the command table, in the order of the usage text. init fills
// it, since the help command reads the table it stands in.
var commands []Command

func init() {
	common := []string{"ref", "level", "markdown", "text", "output"}
	with := func(more ...string) []string { return append(append([]string(nil), common...), more...) }
	read := func(name, operands, summary string, more ...string) Command {
		return Command{Name: name, Group: "read", Operands: operands, Summary: summary, Options: with(more...), Run: runRead}
	}
	commands = []Command{
		read("gates", "", "What do the columns and symbols mean?", "task"),
		read("task", "<id>", "What is this task and where does it stand?", "person"),
		read("authority", "", "Who may accept what?", "task", "person", "proposed"),
		read("assignment", "", "What does each person carry?", "task", "person"),
		read("queue", "", "What do I do next?", "task", "person", "brief"),
		read("blockage", "", "What waits on what?", "task", "person"),
		read("tableau", "", "How does the whole project stand?", "person", "window", "columns", "historical"),
		read("context", "", "How does my corner stand?", "task", "person", "window", "columns", "historical"),
		read("history", "", "What happened, when, and who did it?", "task", "person"),
		read("audit", "", "Where do files and history disagree?", "task", "person", "stale"),
		{Name: "version", Group: "other", Summary: "Print the version", Run: runVersion},
		{Name: "help", Group: "other", Summary: "Print this text, or the options of one command", Run: runHelp},
	}
}

// lookup returns the command of that name, or nil.
func lookup(name string) *Command {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}

// Run parses argv, the arguments after the program name, runs the command
// they name and returns its exit code. It never calls os.Exit.
func Run(ctx context.Context, argv []string, env *Env) int {
	a, parseErr := Parse(argv, vocabulary)
	c := lookup(a.Command)
	if _, ok := a.Values["help"]; ok {
		// --help wins over every other check.
		switch {
		case a.Command == "":
			printUsage(env.Stdout)
		case c == nil:
			return usageError(env, nil, fmt.Errorf("unknown command %q", a.Command))
		default:
			printHelp(env.Stdout, c)
		}
		return ExitOK
	}
	switch {
	case parseErr != nil:
		return usageError(env, c, parseErr)
	case a.Command == "":
		printUsage(env.Stderr)
		return ExitUsage
	case c == nil:
		return usageError(env, nil, fmt.Errorf("unknown command %q", a.Command))
	}
	if err := checkApplies(c, a); err != nil {
		return usageError(env, c, err)
	}
	return c.Run(ctx, env, a)
}

// usageError prints a usage error as the message and a pointer to the
// command's help, and returns ExitUsage. The command is nil when the line
// names none that exists.
func usageError(env *Env, c *Command, err error) int {
	if c == nil {
		_, _ = fmt.Fprintf(env.Stderr, "tabloio: %v\nRun 'tabloio help' for the commands.\n", err)
		return ExitUsage
	}
	_, _ = fmt.Fprintf(env.Stderr, "tabloio %s: %v\nRun 'tabloio %s --help' for its options.\n", c.Name, err, c.Name)
	return ExitUsage
}

// checkApplies refuses an option that stands outside the command's row of
// the vocabulary table. -C and --help belong to every command.
func checkApplies(c *Command, a *Args) error {
	for _, o := range vocabulary {
		if _, ok := a.Values[o.Name]; ok && !takes(c, o.Name) {
			return fmt.Errorf("%s does not apply to %s", o.label(), c.Name)
		}
	}
	return nil
}

// takes reports whether the command takes the option.
func takes(c *Command, name string) bool {
	if name == "C" || name == "help" {
		return true
	}
	for _, n := range c.Options {
		if n == name {
			return true
		}
	}
	return false
}

// usagePad is the width of the name column of the usage text; helpPad the
// width of the spelling column of a command's help.
const (
	usagePad = 14
	helpPad  = 23
)

// printUsage writes the text of "tabloio help": the commands by group, in
// table order.
func printUsage(w io.Writer) {
	var b strings.Builder
	b.WriteString("usage: tabloio [-C <dir>] <command> [<options>]\n")
	for _, g := range groups {
		var rows []string
		for _, c := range commands {
			if c.Group != g.name {
				continue
			}
			name := c.Name
			if c.Operands != "" {
				name += " " + c.Operands
			}
			rows = append(rows, "  "+pad(name, usagePad)+c.Summary+"\n")
		}
		if len(rows) == 0 {
			continue
		}
		b.WriteString("\n" + g.heading + "\n\n")
		for _, r := range rows {
			b.WriteString(r)
		}
	}
	b.WriteString("\nRun 'tabloio <command> --help' for the options of one command.\n")
	_, _ = io.WriteString(w, b.String())
}

// printHelp writes the help of one command: the usage line, the summary,
// then -C, the command's own options and --help in the order of the
// vocabulary table.
func printHelp(w io.Writer, c *Command) {
	var b strings.Builder
	b.WriteString("usage: tabloio [-C <dir>] " + c.Name)
	if c.Operands != "" {
		b.WriteString(" " + c.Operands)
	}
	b.WriteString(" [<options>]\n\n" + c.Summary + "\n\n")
	for _, o := range vocabulary {
		if takes(c, o.Name) {
			b.WriteString("  " + pad(spelling(o), helpPad) + o.Help + "\n")
		}
	}
	_, _ = io.WriteString(w, b.String())
}

// spelling is how a command's help writes an option: "-o, --output <file>".
func spelling(o Option) string {
	var s string
	switch {
	case o.NoLong:
		s = "-" + string(o.Short)
	case o.Short != 0:
		s = "-" + string(o.Short) + ", --" + o.Name
	default:
		s = "--" + o.Name
	}
	if o.Values != "" {
		s += " " + o.Values
	}
	return s
}

// pad returns s followed by spaces up to width, and at least one space.
func pad(s string, width int) string {
	n := width - len(s)
	if n < 1 {
		n = 1
	}
	return s + strings.Repeat(" ", n)
}

// runHelp is the help command: the usage text, or the help of the command
// its operand names.
func runHelp(_ context.Context, env *Env, a *Args) int {
	switch len(a.Operands) {
	case 0:
		printUsage(env.Stdout)
		return ExitOK
	case 1:
		c := lookup(a.Operands[0])
		if c == nil {
			return usageError(env, lookup("help"), fmt.Errorf("unknown command %q", a.Operands[0]))
		}
		printHelp(env.Stdout, c)
		return ExitOK
	}
	return usageError(env, lookup("help"), fmt.Errorf("unexpected argument %q", a.Operands[1]))
}
