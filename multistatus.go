/*
Package multistatus will print a continuously updating block of text to the
stdout to show the current status of a set of concurrent goroutines.

Usage:
		import (
			"context"
			"fmt"
			"math/rand"
			"time"

			ms "github.com/zikes/multistatus"
		)

		func main() {
			// Create a new WorkerSet
			ws := ms.New()

			// Populate the WorkerSet with Workers
			for i := 0; i < 10; i++ {
				w := ws.Add(fmt.Sprintf("Task #%d", i))
				go func(w *ms.Worker) {
					// Sleep for 0-8 seconds, then tell the Worker it failed or completed
					time.Sleep(time.Millisecond * time.Duration(rand.Intn(8000)))
					if rand.Intn(5) == 1 {
						w.Fail()
					} else {
						w.Done()
					}
				}(w)
			}

			// Print the WorkerSet's status until all Workers have completed
			ws.Print(context.Background())
		}
*/
package multistatus

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	spin "github.com/tj/go-spin"
	"golang.org/x/crypto/ssh/terminal"
)

// WorkerState represent the current state of a Worker
type WorkerState int

// Available Worker states
const (
	Completed WorkerState = iota
	Failed
	Pending
)

// Worker is used to track the status of a worker task
type Worker struct {
	State  WorkerState
	Name   string
	parent *WorkerSet
}

// Done will set the Worker.State to Completed and decrement the parent
// WorkerSet's sync.WaitGroup
func (w *Worker) Done() {
	w.State = Completed
	w.parent.wg.Done()
}

// Fail will set the Worker.State to Fail and decrement the parent
// WorkerSet's sync.WaitGroup
func (w *Worker) Fail() {
	w.State = Failed
	w.parent.wg.Done()
}

// Active will return `true` if the Worker.State is Pending
func (w *Worker) Active() bool {
	return w.State == Pending
}

// A WorkerSet is a collection of Workers
type WorkerSet struct {
	Workers []*Worker
	wg      sync.WaitGroup
	spinner *spin.Spinner
}

// New returns an empty WorkerSet
func New() *WorkerSet {
	return &WorkerSet{spinner: spin.New()}
}

// Add creates and returns a new Worker, and increments the WorkerSet's
// sync.WaitGroup
func (w *WorkerSet) Add(s string) *Worker {
	w.wg.Add(1)
	worker := &Worker{Pending, s, w}
	w.Workers = append(w.Workers, worker)
	return worker
}

// Print initiates the WorkerSet's sync.WaitGroup.Wait() and continuously
// prints the status of all the Workers in its collection, cancelable via
// context cancelation.
//
// If the stdout is determined to not be a terminal then it will not print
// until the WaitGroup has finished, and its output will be free of terminal
// escapes.
func (w *WorkerSet) Print(ctx context.Context) {
	done := make(chan bool)
	go func() {
		w.wg.Wait()
		done <- true
	}()

	if terminal.IsTerminal(int(os.Stdout.Fd())) {
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
	} else {
		<-done
		w.print(true)
	}
}

func (w *WorkerSet) print(end bool) {
	failed := "✗"
	completed := "✔"
	inProgress := "-"

	isTerm := terminal.IsTerminal(int(os.Stdout.Fd()))

	if isTerm {
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
		failed = "\033[0;31m✗\033[0m"
		completed = "\033[0;32m✔\033[0m"
		inProgress = w.spinner.Next()
	}

	for _, v := range w.Workers {
		p := inProgress
		if v.State == Completed {
			p = completed
		} else if v.State == Failed {
			p = failed
		}
		fmt.Printf("  %s %s\n", p, v.Name)
	}
	if isTerm {
		// Hide the cursor
		fmt.Print("\033[?25l")
		if end {
			// Show the cursor
			fmt.Print("\033[?25h")
		} else {
			fmt.Printf("%s", strings.Repeat("\033[A", len(w.Workers)))
		}
	}
}
