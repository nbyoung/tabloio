// Command tabloio is the Tableaux command-line front end.
//
// The read commands render a view as Markdown or Unicode text; the write
// commands operate the method through Git commits.
package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/nbyoung/tabloio/internal/cli"
)

// version is the release version. GoReleaser sets it through -ldflags at
// build time; a development build keeps the default.
var version = "dev"

func main() {
	// An interrupt cancels the context, so that a read in progress stops.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := cli.Run(ctx, os.Args[1:], &cli.Env{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Version: versionString(version),
		Source:  cli.NewSource(),
	})
	stop()
	os.Exit(code)
}

// versionString returns v, or the module version that go install recorded
// in the build information when v is the default and that version exists.
func versionString(v string) string {
	if v != "dev" {
		return v
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if mv := info.Main.Version; mv != "" && mv != "(devel)" {
			return mv
		}
	}
	return v
}
