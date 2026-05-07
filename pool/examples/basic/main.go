// Bounded asynchronous emit via the optional pool sibling module.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/kaptinlin/emitter"
	"github.com/kaptinlin/emitter/pool"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	e := emitter.New()
	defer e.Close()

	var wg sync.WaitGroup
	if _, err := e.On("metric.cpu", func(_ context.Context, ev emitter.Event) error {
		defer wg.Done()
		fmt.Printf("metric.cpu: %v\n", ev.Payload())
		return nil
	}); err != nil {
		return err
	}

	p := pool.New(8, 256) // 8 workers, queue cap 256
	defer p.Close()

	for i := range 4 {
		wg.Add(1)
		if err := p.Submit(context.Background(), e, "metric.cpu", i); err != nil {
			wg.Done()
			if errors.Is(err, pool.ErrPoolFull) {
				continue // shed load — pool saturated
			}
			return err
		}
	}
	wg.Wait()
	return nil
}
