// Prototype e4c7 writes the brief of one work queue item as Markdown.
//
// It reads the view data of the tablo prototypes and renders one
// self-contained document for an agent session.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "e4c7:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("e4c7", flag.ContinueOnError)
	queue := fs.String("queue", "", "work queue JSON (tablo prototype 886d)")
	task := fs.String("task", "", "the task of the item")
	gate := fs.String("gate", "", "the gate of the item")
	taskView := fs.String("taskview", "", "task view JSON (tablo prototype 493e -view task)")
	gateView := fs.String("gateview", "", "gate view JSON (tablo prototype 493e -view gate -level detail)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *queue == "" || *task == "" || *taskView == "" {
		return fmt.Errorf("-queue, -task and -taskview are required")
	}
	var in inputs
	if err := load(*queue, &in.Queue); err != nil {
		return err
	}
	if err := load(*taskView, &in.Task); err != nil {
		return err
	}
	if *gateView != "" {
		if err := load(*gateView, &in.Gate); err != nil {
			return err
		}
	}
	it, err := selectItem(in.Queue, *task, *gate)
	if err != nil {
		return err
	}
	render(os.Stdout, in, it)
	return nil
}
