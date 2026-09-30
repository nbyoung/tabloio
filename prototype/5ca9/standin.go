package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// standIn plays `tablo view`: it parses the argument vector tabloio composed
// and answers from a testdata file, wrapped in the envelope. It applies no
// focusing: the file stands for one fixed answer. It does check that the
// vector has the shape the plumbing command takes.
type standIn struct{ dir string }

// level maps a role to the level a view opens at (VIEWS.md#levels); tablo
// resolves this, so the stand-in only mimics it.
func roleLevel(role string) string {
	switch role {
	case "", "observer":
		return "glance"
	case "owner":
		return "provenance"
	}
	return "detail"
}

func (s standIn) run(argv []string) (*Envelope, int, error) {
	// argv: bin [-C d] [--ref r] --json view NAME [flags]
	i := 1
	f := map[string]string{}
	for i < len(argv) && argv[i] != "view" {
		if argv[i] == "--json" {
			i++
			continue
		}
		if i+1 >= len(argv) {
			return nil, exitUsage, fmt.Errorf("stand-in: dangling %s", argv[i])
		}
		f[argv[i]] = argv[i+1]
		i += 2
	}
	if i+1 >= len(argv) {
		return nil, exitUsage, fmt.Errorf("stand-in: no view name")
	}
	name := argv[i+1]
	i += 2
	for i < len(argv) {
		if argv[i] == "--historical-junctions" || argv[i] == "--proposed" {
			f[argv[i]] = "true"
			i++
			continue
		}
		if i+1 >= len(argv) {
			return nil, exitUsage, fmt.Errorf("stand-in: dangling %s", argv[i])
		}
		f[argv[i]] = argv[i+1]
		i += 2
	}

	level := f["--level"]
	if level == "" {
		level = roleLevel(f["--role"])
	}
	file, kind := "", "json"
	switch name {
	case "gate-definition":
		file = "gate-detail.json"
	case "task-definition":
		file = "task-detail.json" // 9f31 only
		if f["--task"] != "9f31" {
			return nil, exitRead, fmt.Errorf("stand-in: no test data for task %s (only 9f31)", f["--task"])
		}
	case "authority-delegation":
		file = "authority-detail.json"
	case "task-assignment":
		file = "assignment-detail.json"
	case "work-queue":
		file = "queue-ada.json"
	case "work-blockage-tree":
		file = "work-blockage-tree.json"
	case "global-tableau":
		file = "global-tableau.json"
	case "contextual-tableau":
		file = "contextual-person-ben.json"
		if f["--task"] == "4e2b" {
			file = "contextual-task-4e2b.json"
		}
	case "history":
		file, kind = "history-w9-w13.json", "events"
	case "audit":
		file, kind = "audit-w3-w13.json", "findings"
	default:
		return nil, exitUsage, fmt.Errorf("stand-in: unknown view %q", name)
	}
	raw, err := os.ReadFile(filepath.Join(s.dir, file))
	if err != nil {
		return nil, exitRead, err
	}
	var data map[string]any
	if kind == "json" {
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, exitRead, err
		}
		data = atLevel(data, level)
	} else {
		var list []any
		if err := json.Unmarshal(raw, &list); err != nil {
			return nil, exitRead, err
		}
		data = map[string]any{"view": name, kind: list}
	}
	data["view"] = name
	data["level"] = level
	ref := f["--ref"]
	if ref == "" {
		ref = "HEAD"
	}
	return &Envelope{
		Schema:      "tablo/1",
		Command:     "view",
		Ref:         map[string]any{"name": ref, "commit": "(stand-in)"},
		Data:        data,
		Diagnostics: []Diagnostic{},
	}, exitOK, nil
}

// atLevel flattens a view that nests glance, detail and provenance (the 493e
// shape) to the keys up to level; the levels nest, so deeper keys override.
func atLevel(d map[string]any, level string) map[string]any {
	if _, ok := d["glance"]; !ok {
		return d
	}
	out := map[string]any{}
	for k, v := range d {
		if k != "glance" && k != "detail" && k != "provenance" {
			out[k] = v
		}
	}
	for _, l := range []string{"glance", "detail", "provenance"} {
		if m, ok := d[l].(map[string]any); ok {
			for k, v := range m {
				out[k] = v
			}
		}
		if l == level {
			break
		}
	}
	return out
}
