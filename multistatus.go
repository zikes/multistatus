package multistatus

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	spin "github.com/tj/go-spin"
)

type WorkerState int

const (
	Completed WorkerState = iota
	Failed
	Pending
)

type Worker struct {
	State  WorkerState
	Name   string
	parent *WorkerSet
}

func (w *Worker) Done() {
	w.State = Completed
	w.parent.wg.Done()
}

func (w *Worker) Fail() {
	w.State = Failed
	w.parent.wg.Done()
}

func (w *Worker) Active() bool {
	return w.State == Pending
}

type WorkerSet struct {
	Workers []*Worker
	wg      sync.WaitGroup
	spinner *spin.Spinner
}

func New() *WorkerSet {
	return &WorkerSet{spinner: spin.New()}
}

func (w *WorkerSet) Add(s string) *Worker {
	w.wg.Add(1)
	worker := &Worker{Pending, s, w}
	w.Workers = append(w.Workers, worker)
	return worker
}

func (w *WorkerSet) Print(ctx context.Context) {
	done := make(chan bool)
	go func() {
		w.wg.Wait()
		done <- true
	}()

	end := false
	for !end {
		select {
		case <-ctx.Done():
			w.print(true)
			return
		case <-time.After(100 * time.Millisecond):
			w.print(false)
		case end = <-done:
			w.print(true)
		}
	}
}

func (w *WorkerSet) print(end bool) {
	// wipe section
	fmt.Print(
		// Ensure the output area is at least N lines long
		strings.Repeat("\n", len(w.Workers)),

		// Move cursor up N lines
		strings.Repeat("\033[A", len(w.Workers)),

		// Move the cursor down N lines, erasing each line
		strings.Repeat("\033[B\033[2K", len(w.Workers)),

		// Move cursor up N lines
		strings.Repeat("\033[A", len(w.Workers)),
	)

	failed := "\033[0;31m✗\033[0m"
	completed := "\033[0;32m✔\033[0m"
	inProgress := w.spinner.Next()
	for _, v := range w.Workers {
		p := inProgress
		if v.State == Completed {
			p = completed
		} else if v.State == Failed {
			p = failed
		}
		fmt.Printf("  %s %s\n", p, v.Name)
	}
	// Hide the cursor
	fmt.Print("\033[?25l")
	if end {
		// Show the cursor
		fmt.Print("\033[?25h")
	} else {
		fmt.Printf("%s", strings.Repeat("\033[A", len(w.Workers)))
	}
}
