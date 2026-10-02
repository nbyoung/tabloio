package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

// Envelope is the plumbing command's output (prototype 4ed9): five keys.
type Envelope struct {
	Schema      string         `json:"schema"`
	Command     string         `json:"command"`
	Ref         map[string]any `json:"ref,omitempty"`
	Data        map[string]any `json:"data"`
	Diagnostics []Diagnostic   `json:"diagnostics"`
}

// Diagnostic carries a severity, a stable code, a path and a message.
type Diagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Path     string `json:"path"`
	Message  string `json:"message"`
}

func hasError(e *Envelope) bool {
	for _, d := range e.Diagnostics {
		if d.Severity == "error" {
			return true
		}
	}
	return false
}

// runner runs an argument vector and returns the envelope, or an exit code
// and an error.
type runner interface {
	run(argv []string) (*Envelope, int, error)
}

// execRunner runs the real plumbing command.
type execRunner struct{}

func (execRunner) run(argv []string) (*Envelope, int, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			// tablo's own codes pass through; its stderr is the message.
			return nil, ee.ExitCode(), fmt.Errorf("%s: %s", argv[0], bytes.TrimSpace(errb.Bytes()))
		}
		return nil, exitRead, err
	}
	var env Envelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		return nil, exitRead, fmt.Errorf("%s: not an envelope: %v", argv[0], err)
	}
	if env.Schema != "tablo/1" {
		return nil, exitRead, fmt.Errorf("%s: schema %q, want tablo/1", argv[0], env.Schema)
	}
	return &env, exitOK, nil
}
