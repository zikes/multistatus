package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	ms "github.com/zikes/multistatus"
)

func main() {
	ws := ms.New()

	for i := 0; i < 10; i++ {
		w := ws.Add(fmt.Sprintf("Task #%d", i))
		go func(w *ms.Worker) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(8000)))
			if rand.Intn(5) == 1 {
				w.Fail()
			} else {
				w.Done()
			}
		}(w)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			cancel()
			time.Sleep(10 * time.Millisecond)
			os.Exit(0)
		}
	}()

	ws.Print(ctx)
}
