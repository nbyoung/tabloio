package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	taskPattern   = regexp.MustCompile(`^[0-9a-f]{4}$`)
	emailPattern  = regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)
	numberPattern = regexp.MustCompile(`^[0-9]+$`)
	gatePattern   = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
)

// invocation is what a read command line builds: the query, the format,
// where the bytes go, and whether the person comes from the viewer.
type invocation struct {
	Query  Query
	Text   bool   // --text; false is Markdown
	Output string // the --output path as given; "" is standard output
	Viewer bool   // the person defaults to the viewer, and none was given
}

// readInvocation checks a view command's operands and options, in the
// order of the vocabulary table, and builds its invocation. It checks
// spelling and shape only: whether a ref, a task or a gate exists is for
// tablo to say. The options of the line already stand in the command's row
// (checkApplies).
func readInvocation(c *Command, a *Args) (*invocation, error) {
	inv := &invocation{Query: Query{View: c.Name, Level: "glance"}}
	q := &inv.Query
	value := func(name string) (string, bool) {
		v, ok := a.Values[name]
		if !ok || len(v) == 0 {
			return "", ok
		}
		return v[0], true
	}
	has := func(name string) bool { _, ok := a.Values[name]; return ok }

	if c.Operands == "" {
		if len(a.Operands) > 0 {
			return nil, fmt.Errorf("unexpected argument %q", a.Operands[0])
		}
	} else {
		switch {
		case len(a.Operands) == 0:
			return nil, errors.New("the task id is missing")
		case len(a.Operands) > 1:
			return nil, fmt.Errorf("unexpected argument %q", a.Operands[1])
		case !taskPattern.MatchString(a.Operands[0]):
			return nil, fmt.Errorf("the task id takes four lowercase hexadecimal digits, not %q", a.Operands[0])
		}
		q.Task = a.Operands[0]
	}

	if v, ok := value("C"); ok {
		if v == "" {
			return nil, errors.New("-C takes a directory")
		}
		q.Dir = v
	}
	if v, ok := value("ref"); ok {
		if err := checkRef(c.Name, v); err != nil {
			return nil, err
		}
		q.Ref = v
	}
	if v, ok := value("level"); ok {
		if v != "glance" && v != "detail" && v != "provenance" {
			return nil, fmt.Errorf("--level takes glance, detail or provenance, not %q", v)
		}
		q.Level = v
	}
	if has("markdown") && has("text") {
		return nil, errors.New("--markdown and --text exclude each other")
	}
	inv.Text = has("text")
	if v, ok := value("output"); ok {
		if v == "" {
			return nil, errors.New("--output takes a path, or - for standard output")
		}
		if err := guardOutput(v); err != nil {
			return nil, err
		}
		if v != "-" {
			inv.Output = v
		}
	}
	if v, ok := value("task"); ok {
		if !taskPattern.MatchString(v) {
			return nil, fmt.Errorf("--task takes four lowercase hexadecimal digits, not %q", v)
		}
		q.Task = v
	}
	if v, ok := value("person"); ok {
		if !emailPattern.MatchString(v) {
			return nil, fmt.Errorf("--person takes an email, not %q", v)
		}
		if c.Name == "context" && has("task") {
			return nil, errors.New("--task and --person exclude each other")
		}
		q.Person = v
	}
	if v, ok := value("window"); ok {
		n, err := wholeNumber("--window", v)
		if err != nil {
			return nil, err
		}
		if has("columns") {
			return nil, errors.New("--window and --columns exclude each other")
		}
		q.Window = &n
	}
	if v, ok := value("columns"); ok {
		for _, g := range strings.Split(v, ",") {
			if !gatePattern.MatchString(g) {
				return nil, fmt.Errorf("--columns takes gate keys joined by commas, not %q", v)
			}
			q.Columns = append(q.Columns, g)
		}
	}
	q.Historical = has("historical")
	q.Proposed = has("proposed")
	if v, ok := value("stale"); ok {
		n, err := wholeNumber("--stale", v)
		if err != nil {
			return nil, err
		}
		q.Stale = &n
	}
	if v, ok := a.Values["brief"]; ok {
		if !taskPattern.MatchString(v[0]) || !gatePattern.MatchString(v[1]) {
			return nil, fmt.Errorf("--brief takes a task id and a gate key, not %q %q", v[0], v[1])
		}
		if has("level") || has("task") {
			return nil, errors.New("--brief excludes --level and --task")
		}
		q.Brief = &Brief{Task: v[0], Gate: v[1]}
		q.Level = "provenance"
	}

	inv.Viewer = !has("person") && (c.Name == "queue" || (c.Name == "context" && !has("task")))
	return inv, nil
}

// checkRef checks the shape of a --ref value: not empty, and a range
// <from>..<to>, with both sides present and no third dot, on history only.
func checkRef(command, ref string) error {
	if ref == "" {
		return errors.New("--ref takes a ref")
	}
	from, to, isRange := strings.Cut(ref, "..")
	switch {
	case !isRange:
		return nil
	case command != "history":
		return fmt.Errorf("a range applies to history only, not to %s", command)
	case from == "" || to == "" || strings.HasPrefix(to, ".") || strings.Contains(to, ".."):
		return fmt.Errorf("a range reads FROM..TO, with both sides present, not %q", ref)
	}
	return nil
}

// wholeNumber reads a value of --window or --stale: digits only.
func wholeNumber(option, v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil || !numberPattern.MatchString(v) {
		return 0, fmt.Errorf("%s takes a whole number, zero or more, not %q", option, v)
	}
	return n, nil
}

// runRead is the flow of a read command, as design 5ca9 states it: the
// checks, the renderer, the output guard, the viewer, the view, the
// rendering and the write.
func runRead(ctx context.Context, env *Env, a *Args) int {
	c := lookup(a.Command)
	inv, err := readInvocation(c, a)
	if err != nil {
		return usageError(env, c, err)
	}
	q := &inv.Query

	format, table := "markdown", markdown
	if inv.Text {
		format, table = "text", text
	}
	key := q.View
	if q.Brief != nil {
		key = "brief"
	}
	render, ok := table[key]
	if !ok {
		return usageError(env, c, fmt.Errorf("the %s format does not cover %s yet", format, key))
	}

	out := inv.Output
	if out != "" {
		out = resolveOutput(q.Dir, out)
		if err := guardOutput(out); err != nil {
			return usageError(env, c, err)
		}
	}

	// Git's one habitual write on a read is the index refresh.
	if err := os.Setenv("GIT_OPTIONAL_LOCKS", "0"); err != nil {
		return failure(env, c, err)
	}

	if inv.Viewer {
		viewer, err := env.Source.Viewer(ctx, q.Dir)
		if err != nil {
			return failure(env, c, err)
		}
		if viewer == "" {
			return usageError(env, c, errors.New("no person: give --person, or set user.email in Git"))
		}
		q.Person = viewer
	}

	res, err := env.Source.View(ctx, *q)
	if err != nil {
		return failure(env, c, err)
	}
	if res.Data == nil {
		return noView(env, c, res.Diagnostics)
	}

	base := ""
	if out != "" {
		base = linkBase(q.Dir, out)
	}
	b, err := render(res.Data, q.Level, base)
	if err != nil {
		return failure(env, c, err)
	}

	if out == "" {
		_, err = env.Stdout.Write(b)
	} else {
		err = writeFile(out, b)
	}
	if err != nil {
		return failure(env, c, err)
	}
	return ExitOK
}

// failure prints a failed read, rendering or write and returns ExitFailure.
func failure(env *Env, c *Command, err error) int {
	_, _ = fmt.Fprintf(env.Stderr, "tabloio %s: %v\n", c.Name, err)
	return ExitFailure
}

// noView reports a view that an error diagnostic prevents: one line per
// error diagnostic, and ExitInvalid. A source that returns neither data nor
// an error diagnostic fails.
func noView(env *Env, c *Command, ds []Diagnostic) int {
	n := 0
	for _, d := range ds {
		if d.Severity == "error" {
			_, _ = fmt.Fprintln(env.Stderr, diagnosticLine(d))
			n++
		}
	}
	if n == 0 {
		return failure(env, c, errors.New("the source returned no data and no error diagnostic"))
	}
	return ExitInvalid
}

// diagnosticLine writes a diagnostic as
// <path>:<line>:<col>: error <code>: <message>, leaving out a zero position
// and an empty path.
func diagnosticLine(d Diagnostic) string {
	var b strings.Builder
	if d.Path != "" {
		b.WriteString(d.Path)
		if d.Line > 0 {
			b.WriteString(":" + strconv.Itoa(d.Line))
			if d.Col > 0 {
				b.WriteString(":" + strconv.Itoa(d.Col))
			}
		}
		b.WriteString(": ")
	}
	b.WriteString("error " + d.Code + ": " + d.Message)
	return b.String()
}
