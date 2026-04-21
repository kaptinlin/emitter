// Package main demonstrates emitter goroutine-pool integration.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kaptinlin/emitter"
)

func main() {
	pool := emitter.NewPondPool(5, 1000)
	e := emitter.NewMemoryEmitter(emitter.WithPool(pool))

	timeConsumingListener := func(evt emitter.Event) error {
		fmt.Printf("Processing event: %s with payload: %v\n", evt.Topic(), evt.Payload())
		time.Sleep(2 * time.Second)
		fmt.Printf("Finished processing event: %s\n", evt.Topic())
		return nil
	}

	_, err := e.On("user.signup", timeConsumingListener)
	if err != nil {
		log.Fatalf("Failed to subscribe listener: %v", err)
	}

	for i := range 10 {
		go func() {
			payload := fmt.Sprintf("User #%d", i)
			e.Emit("user.signup", payload)
		}()
	}

	time.Sleep(10 * time.Second)
	pool.Release()
	fmt.Println("All events have been processed and the pool has been released.")
}
