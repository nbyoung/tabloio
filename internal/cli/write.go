package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nbyoung/tabloio/action"
)

// writeOptions are the options every write command takes beyond -C and
// --help (design b618, "The commands").
var writeOptions = []string{"dry-run", "yes", "message", "body", "model", "co-author"}

// writeCommands returns the five write commands, in the order of the usage
// text.
func writeCommands() []Command {
	write := func(name, operands, summary string, more ...string) Command {
		return Command{Name: name, Group: "write", Operands: operands, Summary: summary,
			Options: append(append([]string(nil), writeOptions...), more...), Run: runWrite}
	}
	return []Command{
		write("propose", "", "Propose a task under a parent",
			"title", "description", "parent", "assignee", "order", "reference", "id"),
		write("record", "<id>", "Record a task's status, or move its pin",
			"gate", "state", "reason", "no-reason", "note", "no-note", "pin"),
		write("review", "<id>... <gate>", "Accept the work of tasks at a gate"),
		write("reaffirm", "<id>...", "Confirm that the statuses of tasks stand"),
		write("authorise", "<id>...", "Accept proposed tasks"),
	}
}

// planFunc is one plan function of package action, bound to its request.
type planFunc func(ctx context.Context, r *action.Repo) (*action.Plan, error)

// runWrite is the flow of a write command: the request, the plan on
// standard output, the prompt, the commit.
func runWrite(ctx context.Context, env *Env, a *Args) int {
	c := lookup(a.Command)
	plan, err := writeRequest(c, a, env.getenv)
	if err != nil {
		return usageError(env, c, err)
	}
	_, dry := a.Values["dry-run"]
	_, yes := a.Values["yes"]
	if !dry && !yes && !env.terminal() {
		return usageError(env, c, errors.New("standard input is no terminal; pass --yes to commit, or --dry-run to show the plan alone"))
	}
	repo := &action.Repo{Dir: first(a, "C")}
	p, err := plan(ctx, repo)
	if err != nil {
		return writeFailure(env, c, err)
	}
	_, _ = fmt.Fprint(env.Stdout, p.String())
	if dry {
		return ExitOK
	}
	if !yes {
		_, _ = fmt.Fprint(env.Stderr, "Commit? [y/N] ")
		line, _ := bufio.NewReader(env.stdin()).ReadString('\n')
		if answer := strings.ToLower(strings.TrimSpace(line)); answer != "y" && answer != "yes" {
			_, _ = fmt.Fprintln(env.Stderr, "not committed")
			return ExitInvalid
		}
	}
	res, err := repo.Apply(ctx, p)
	if res.Commit != "" {
		_, _ = fmt.Fprintf(env.Stdout, "committed %s %s\n", res.Commit[:min(7, len(res.Commit))], p.Subject)
	}
	if err != nil {
		return writeFailure(env, c, err)
	}
	return ExitOK
}

// writeFailure reports a plan or a commit that does not happen: each
// refusal on a line of its own with exit 1, any other error with exit 3.
func writeFailure(env *Env, c *Command, err error) int {
	var refusals action.Refusals
	var refusal action.Refusal
	switch {
	case errors.As(err, &refusals):
		for _, r := range refusals {
			_, _ = fmt.Fprintf(env.Stderr, "tabloio %s: %v\n", c.Name, r)
		}
		return ExitInvalid
	case errors.As(err, &refusal):
		_, _ = fmt.Fprintf(env.Stderr, "tabloio %s: %v\n", c.Name, refusal)
		return ExitInvalid
	}
	return failure(env, c, err)
}

// writeRequest turns a write command's line into its plan function, or a
// usage error.
func writeRequest(c *Command, a *Args, getenv func(string) string) (planFunc, error) {
	o := action.Options{Subject: first(a, "message"), Body: strings.Join(a.Values["body"], "\n\n")}
	o.Agent.Model, o.Agent.CoAuthor = first(a, "model"), first(a, "co-author")
	if _, ok := a.Values["model"]; !ok {
		o.Agent.Model = getenv("TABLOIO_MODEL")
	}
	if _, ok := a.Values["co-author"]; !ok {
		o.Agent.CoAuthor = getenv("TABLOIO_CO_AUTHOR")
	}
	ids := a.Operands
	switch c.Name {
	case "propose":
		if len(ids) > 0 {
			return nil, fmt.Errorf("unexpected argument %q", ids[0])
		}
		p := action.Proposal{Title: first(a, "title"), Description: first(a, "description"), Parent: first(a, "parent"),
			Assignee: first(a, "assignee"), ID: first(a, "id")}
		for _, name := range []string{"title", "description", "parent"} {
			if _, ok := a.Values[name]; !ok {
				return nil, fmt.Errorf("propose needs --%s", name)
			}
		}
		if v, ok := a.Values["order"]; ok {
			n, err := wholeNumber("--order", v[0])
			if err != nil {
				return nil, err
			}
			p.Order = n
		}
		for _, ref := range a.Values["reference"] {
			url, text, _ := strings.Cut(ref, "=")
			p.References = append(p.References, action.Reference{URL: url, Text: text})
		}
		return func(ctx context.Context, r *action.Repo) (*action.Plan, error) { return r.Propose(ctx, p, o) }, nil
	case "record":
		if len(ids) != 1 {
			return nil, errors.New("record takes one task id")
		}
		_, noReason := a.Values["no-reason"]
		_, noNote := a.Values["no-note"]
		_, pin := a.Values["pin"]
		if _, ok := a.Values["reason"]; ok && noReason {
			return nil, errors.New("--reason excludes --no-reason")
		}
		if _, ok := a.Values["note"]; ok && noNote {
			return nil, errors.New("--note excludes --no-note")
		}
		rec := action.Recording{ID: ids[0], Gate: first(a, "gate"), State: first(a, "state"), Reason: first(a, "reason"),
			Note: first(a, "note"), NoReason: noReason, NoNote: noNote, Pin: pin}
		return func(ctx context.Context, r *action.Repo) (*action.Plan, error) { return r.Record(ctx, rec, o) }, nil
	case "review":
		if len(ids) < 2 {
			return nil, errors.New("review takes one or more task ids and then the gate")
		}
		gate, ids := ids[len(ids)-1], ids[:len(ids)-1]
		return func(ctx context.Context, r *action.Repo) (*action.Plan, error) { return r.Review(ctx, ids, gate, o) }, nil
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("%s takes one or more task ids", c.Name)
	}
	if c.Name == "reaffirm" {
		return func(ctx context.Context, r *action.Repo) (*action.Plan, error) { return r.Reaffirm(ctx, ids, o) }, nil
	}
	return func(ctx context.Context, r *action.Repo) (*action.Plan, error) { return r.Authorise(ctx, ids, o) }, nil
}

// first returns the first value of an option, or "".
func first(a *Args, name string) string {
	if v := a.Values[name]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// stdin returns the reader the prompt reads.
func (env *Env) stdin() io.Reader {
	if env.Stdin != nil {
		return env.Stdin
	}
	return os.Stdin
}

// terminal reports whether the prompt's input is a terminal.
func (env *Env) terminal() bool {
	if env.Terminal != nil {
		return env.Terminal()
	}
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// getenv reads the process environment, or the test's.
func (env *Env) getenv(key string) string {
	if env.Getenv != nil {
		return env.Getenv(key)
	}
	return os.Getenv(key)
}
