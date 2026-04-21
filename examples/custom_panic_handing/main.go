// Package main demonstrates listener panic errors returned by emitter.
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

func main() {
	e := emitter.NewMemoryEmitter()

	listener := func(evt emitter.Event) error {
		panic(fmt.Sprintf("simulated panic in listener for event: %s", evt.Topic()))
	}

	_, err := e.On("user.created", listener)
	if err != nil {
		log.Fatalf("Failed to subscribe listener: %v", err)
	}

	for err := range e.Emit("user.created", "Jane Doe") {
		if err != nil {
			log.Printf("Emit returned error: %v", err)
		}
	}

	fmt.Println("Application continues running despite the listener panic.")
}
