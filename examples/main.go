package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	ms "git.zikes.me/multistatus"
)

func main() {
	ws := ms.New()

	for i := 0; i < 10; i++ {
		w := ws.Add(fmt.Sprintf("Task #%d", i))
		go func(w *ms.Worker, i int) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(8000)))
			if rand.Intn(5) == 1 {
				w.Fail()
			} else {
				w.Done()
			}
		}(w, i)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ws.Print(ctx)
}
