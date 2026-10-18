// Package cli runs the tabloio commands.
//
// Run parses a command line, checks it against the vocabulary of options,
// asks a Source for the data of a view, hands the data to a renderer and
// writes the bytes to standard output or to a file. A read command changes
// no project state. Only source.go depends on how tablo reaches the read
// commands, so that a test replaces the Source and the renderers with fakes.
//
// A write command (write.go) turns its line into a request of package
// action, prints the plan, asks, and has action make the commit.
package cli
