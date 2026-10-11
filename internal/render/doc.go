// Package render writes a view of a Tableaux project in a textual format.
//
// Each view has a builder that applies the level, the folds and the wording
// and returns a doc.Doc; a writer, the Markdown writer or the text writer, prints it.
// The rules R1 to R17 of design e3ed live in this package as helpers that
// every builder shares.
package render
