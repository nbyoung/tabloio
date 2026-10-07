package cli

import (
	"context"
	"fmt"
)

// runVersion is the version command: one line, "tabloio <version>".
//
// The --json form waits for tablo's release: it names the version of tablo
// the binary links, which only tablo's constants know.
func runVersion(_ context.Context, env *Env, a *Args) int {
	if len(a.Operands) > 0 {
		return usageError(env, lookup("version"), fmt.Errorf("unexpected argument %q", a.Operands[0]))
	}
	_, _ = fmt.Fprintf(env.Stdout, "tabloio %s\n", env.Version)
	return ExitOK
}
