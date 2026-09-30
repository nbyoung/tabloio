// Command tabloio is the Tableaux command-line front end.
//
// It prints its version until the subcommands land: the read commands render
// a view as Markdown or Unicode text, and the write commands operate the
// method through Git commits.
package main

import (
	"fmt"
	"runtime/debug"
)

// version is the release version. GoReleaser sets it through -ldflags at
// build time; a development build keeps the default.
var version = "dev"

func main() {
	fmt.Println("tabloio", versionString(version))
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
