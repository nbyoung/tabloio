// Command 5ca9 is the prototype of tabloio's read commands.
//
// One subcommand per view takes the view's focusing parameters as flags,
// builds the argument vector of the plumbing command `tablo view`, runs it (or
// a stand-in that reads testdata), and renders the envelope's data as Markdown
// or Unicode text to stdout or to a file. It never changes project state.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Exit codes follow the plumbing command (prototype 4ed9).
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
	exitRead  = 3
)

// view describes one subcommand: the `tablo view` name it calls, the focusing
// flags it accepts and its positional argument.
type view struct {
	name   string   // tabloio subcommand
	tablo  string   // `tablo view` name
	flags  []string // focusing flags beyond the common ones
	pos    string   // positional argument naming a flag, or ""
	answer string   // the question, for usage
}

var views = []view{
	{"gates", "gate-definition", []string{"task", "window", "columns"}, "", "What does each gate demand?"},
	{"task", "task-definition", []string{"for", "window", "columns"}, "task", "What does one task say, gate by gate?"},
	{"authority", "authority-delegation", []string{"task", "for", "proposed"}, "", "Who authorises what?"},
	{"assignment", "task-assignment", []string{"task", "for", "window", "columns"}, "", "Who holds which junction?"},
	{"queue", "work-queue", []string{"task", "for", "window", "columns", "brief"}, "", "What do I do next?"},
	{"blockage", "work-blockage-tree", []string{"task", "for", "window", "columns"}, "", "What holds the work up?"},
	{"tableau", "global-tableau", []string{"for", "window", "columns", "historical"}, "", "Where does every task stand?"},
	{"context", "contextual-tableau", []string{"task", "for", "window", "columns", "historical"}, "", "Where does my corner stand?"},
	{"history", "history", []string{"task", "for", "window", "columns"}, "", "What happened, when, and who did it?"},
	{"audit", "audit", []string{"task", "for", "stale"}, "", "Where do files and history disagree?"},
}

// validRoles and validLevels come from README.md#roles and VIEWS.md#levels.
var (
	validRoles  = []string{"owner", "authority", "assignee", "contributor", "agent", "reviewer", "observer"}
	validLevels = []string{"glance", "detail", "provenance"}
)

// opts holds every flag value a read command takes.
type opts struct {
	markdown, text bool
	output         string
	dir, ref       string
	role, level    string
	tabloBin       string
	testdata       string

	task, person, columns, brief string
	window                       int // -1 when unset
	historical, proposed         bool
	stale                        int // 0 when unset
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		usage(stderr)
		if len(args) == 0 {
			return exitUsage
		}
		return exitOK
	}
	v := findView(args[0])
	if v == nil {
		pf(stderr, "tabloio: unknown command %q\n", args[0])
		usage(stderr)
		return exitUsage
	}
	o, err := parse(v, args[1:], stderr)
	if err != nil {
		if err != flag.ErrHelp {
			pf(stderr, "tabloio %s: %v\n", v.name, err)
		}
		return exitUsage
	}

	argv := tabloArgv(v, o)
	var r runner = execRunner{}
	if o.testdata != "" {
		r = standIn{dir: o.testdata}
	}
	env, code, err := r.run(argv)
	if err != nil {
		pf(stderr, "tabloio %s: %v\n", v.name, err)
		return code
	}
	for _, d := range env.Diagnostics {
		pf(stderr, "%s %s: %s\n", d.Severity, d.Code, d.Message)
	}
	format := "markdown"
	if o.text {
		format = "text"
	}
	out := Render(v, env, format)
	if o.output == "" || o.output == "-" {
		pf(stdout, "%s", out)
	} else if err := os.WriteFile(o.output, []byte(out), 0o644); err != nil {
		pf(stderr, "tabloio %s: %v\n", v.name, err)
		return exitRead
	}
	if hasError(env) {
		return exitError
	}
	return exitOK
}

func findView(name string) *view {
	for i := range views {
		if views[i].name == name {
			return &views[i]
		}
	}
	return nil
}

func (v *view) accepts(f string) bool {
	for _, x := range v.flags {
		if x == f {
			return true
		}
	}
	return false
}

// parse reads the flags of one subcommand. It accepts flags before and after
// the positional argument, and registers only the flags the view takes, so a
// parameter a view does not name is a usage error here (VIEWS.md says it has
// no effect; a front end says so early).
func parse(v *view, args []string, stderr io.Writer) (*opts, error) {
	o := &opts{window: -1}
	fs := flag.NewFlagSet("tabloio "+v.name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.BoolVar(&o.markdown, "markdown", false, "Markdown, the default")
	fs.BoolVar(&o.text, "text", false, "Unicode text")
	fs.StringVar(&o.output, "o", "", "write to `file` instead of stdout")
	fs.StringVar(&o.output, "output", "", "same as -o")
	fs.StringVar(&o.dir, "C", "", "run as if in `dir`")
	fs.StringVar(&o.ref, "ref", "", "commit in view; history also takes a range `a..b`")
	fs.StringVar(&o.role, "role", "", "role whose level the view opens at")
	fs.StringVar(&o.level, "level", "", "level: glance, detail or provenance")
	fs.StringVar(&o.tabloBin, "tablo", "tablo", "the plumbing command")
	fs.StringVar(&o.testdata, "testdata", "", "stand-in: read view data from `dir` instead of running tablo")
	if v.accepts("task") || v.pos == "task" {
		fs.StringVar(&o.task, "task", "", "one task, or the subtree under it")
	}
	if v.accepts("for") {
		fs.StringVar(&o.person, "for", "", "one person's email")
	}
	if v.accepts("window") {
		fs.IntVar(&o.window, "window", -1, "gate columns `n` either side of the next gates")
	}
	if v.accepts("columns") {
		fs.StringVar(&o.columns, "columns", "", "explicit gate columns, comma-separated")
	}
	if v.accepts("historical") {
		fs.BoolVar(&o.historical, "historical", false, "show the marks of historical junctions")
	}
	if v.accepts("proposed") {
		fs.BoolVar(&o.proposed, "proposed", false, "proposed tasks only")
	}
	if v.accepts("stale") {
		fs.IntVar(&o.stale, "stale", 0, "age in `days` beyond which a status is stale")
	}
	if v.accepts("brief") {
		fs.StringVar(&o.brief, "brief", "", "write one item as a brief: `TASK:GATE`")
	}

	var pos []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		pos = append(pos, rest[0])
		args = rest[1:]
	}
	if v.pos == "" && len(pos) > 0 {
		return nil, fmt.Errorf("unexpected argument %q", pos[0])
	}
	if v.pos == "task" {
		if len(pos) > 1 {
			return nil, fmt.Errorf("unexpected argument %q", pos[1])
		}
		if len(pos) == 1 {
			if o.task != "" {
				return nil, fmt.Errorf("the task is given twice")
			}
			o.task = pos[0]
		}
		if o.task == "" {
			return nil, fmt.Errorf("the task id is required")
		}
	}
	return o, o.validate(v)
}

func (o *opts) validate(v *view) error {
	if o.markdown && o.text {
		return fmt.Errorf("--markdown and --text exclude each other")
	}
	if o.window >= 0 && o.columns != "" {
		return fmt.Errorf("--window and --columns exclude each other")
	}
	if v.name == "context" && o.task != "" && o.person != "" {
		return fmt.Errorf("--task and --for exclude each other: the contextual tableau takes one")
	}
	if o.task != "" && !isID(o.task) {
		return fmt.Errorf("task %q is not four lowercase hex digits", o.task)
	}
	if o.level != "" && !contains(validLevels, o.level) {
		return fmt.Errorf("level %q is not one of %s", o.level, strings.Join(validLevels, ", "))
	}
	if o.role != "" && !contains(validRoles, o.role) {
		return fmt.Errorf("role %q is not one of %s", o.role, strings.Join(validRoles, ", "))
	}
	if o.person != "" && !strings.Contains(o.person, "@") {
		return fmt.Errorf("--for wants an email, not %q", o.person)
	}
	if o.columns != "" {
		for _, c := range strings.Split(o.columns, ",") {
			if c == "" {
				return fmt.Errorf("--columns has an empty gate")
			}
		}
	}
	if strings.Contains(o.ref, "..") && v.name != "history" {
		return fmt.Errorf("a range is a parameter of history only")
	}
	if o.brief != "" {
		t, g, ok := strings.Cut(o.brief, ":")
		if !ok || !isID(t) || g == "" {
			return fmt.Errorf("--brief wants TASK:GATE, for example 9f31:function")
		}
	}
	return nil
}

// tabloArgv composes the argument vector of the plumbing command. Global
// options come first, as the envelope README (4ed9) has them; the focusing
// parameters follow the view name, in a fixed order so a test can compare.
func tabloArgv(v *view, o *opts) []string {
	a := []string{o.tabloBin}
	if o.dir != "" {
		a = append(a, "-C", o.dir)
	}
	if o.ref != "" {
		a = append(a, "--ref", o.ref)
	}
	a = append(a, "--json", "view", v.tablo)
	add := func(flag, val string) {
		if val != "" {
			a = append(a, flag, val)
		}
	}
	add("--task", o.task)
	add("--person", o.person)
	if o.window >= 0 {
		add("--window", strconv.Itoa(o.window))
	}
	add("--columns", o.columns)
	if o.historical {
		a = append(a, "--historical-junctions")
	}
	if o.proposed {
		a = append(a, "--proposed")
	}
	if o.stale > 0 {
		add("--stale", strconv.Itoa(o.stale))
	}
	add("--brief", o.brief)
	add("--role", o.role)
	add("--level", o.level)
	return a
}

func isID(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func usage(w io.Writer) {
	pl(w, "usage: tabloio <command> [flags]")
	pl(w, "\ncommands:")
	names := make([]string, 0, len(views))
	byName := map[string]view{}
	for _, v := range views {
		names = append(names, v.name)
		byName[v.name] = v
	}
	sort.Strings(names)
	for _, n := range names {
		v := byName[n]
		fl := make([]string, len(v.flags))
		for i, f := range v.flags {
			fl[i] = "--" + f
		}
		arg := ""
		if v.pos != "" {
			arg = " <" + v.pos + ">"
		}
		pf(w, "  %-12s%s  %s\n", v.name+arg, v.answer, strings.Join(fl, " "))
	}
	pl(w, "\ncommon flags: --markdown --text -o FILE -C DIR --ref REF --role ROLE --level LEVEL")
}

// pf and pl write and drop the error: a failed write to a terminal has no
// recourse.
func pf(w io.Writer, f string, a ...any) { _, _ = fmt.Fprintf(w, f, a...) }
func pl(w io.Writer, a ...any)           { _, _ = fmt.Fprintln(w, a...) }
