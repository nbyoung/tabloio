// Package brief writes the brief of one work item of the contributor work
// queue: the queue's provenance level for that item, as plain text.
//
// The text is the same under --markdown and --text. It carries six parts: the
// item, the task, the requirements, the status, the commit the agent makes
// when done and the commands that reproduce the brief. Decode reads the
// brief object that tablo emits, and Render writes it, byte for byte the same
// for the same data. The package reads no file and calls no program.
package brief
