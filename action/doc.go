// Package action plans and makes the commits that operate the Tableaux
// method. A plan function changes neither a ref, the index nor the working
// tree; Apply makes the planned commit or leaves the checkout as it found it.
//
// Five plan functions serve the five write commands of tabloio: Propose adds
// a task file, Record writes a status file or moves a pin, and Review,
// Reaffirm and Authorise make an empty commit with one trailer per task. Each
// returns a Plan, the commit that does not exist yet, or Refusals, the rules
// the commit would break. The facts come from a Source, by default Tablo,
// which reads them through tablo's library; the bytes come from Git, which
// the package runs itself. The package edits a status file as lines of text
// and never parses or dumps YAML (design b618).
//
// The package stands outside internal/ so that tablotui makes the same
// commits: it imports action.
package action
