package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/nbyoung/tablo"
)

// versionSchema names the shape of the version command's JSON line.
const versionSchema = "tabloio/1"

// runVersion is the version command: one line, "tabloio <version>", or with
// --json one line that names the tablo the binary links and the tableaux
// version it accepts.
func runVersion(_ context.Context, env *Env, a *Args) int {
	if len(a.Operands) > 0 {
		return usageError(env, lookup("version"), fmt.Errorf("unexpected argument %q", a.Operands[0]))
	}
	if _, ok := a.Values["json"]; !ok {
		_, _ = fmt.Fprintf(env.Stdout, "tabloio %s\n", env.Version)
		return ExitOK
	}
	// A map marshals with its keys sorted, which fixes the line's order.
	b, err := json.Marshal(map[string]any{
		"command": "version",
		"data": map[string]string{
			"tableaux": strconv.Itoa(tablo.AcceptedMajor) + "." + strconv.Itoa(tablo.AcceptedMinor),
			"tablo":    "v" + tablo.Version,
			"tabloio":  env.Version,
		},
		"diagnostics": []Diagnostic{},
		"schema":      versionSchema,
	})
	if err != nil {
		return failure(env, lookup("version"), err)
	}
	_, _ = fmt.Fprintf(env.Stdout, "%s\n", b)
	return ExitOK
}
