// Package cli runs the tabloio commands.
//
// Run parses a command line, checks it against the vocabulary of options,
// asks a Source for the data of a view, hands the data to a renderer and
// writes the bytes to standard output or to a file. A read command changes
// no project state. Only source.go depends on how tablo is reached, so that
// a test replaces the Source and the renderers with fakes.
package cli
