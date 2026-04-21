// Package main demonstrates custom emitter error handling.
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

// ErrSimulatedListener is returned by the example listener.
var ErrSimulatedListener = errors.New("simulated error in listener")

// CustomErrorHandler logs listener errors and marks them handled.
func CustomErrorHandler(event emitter.Event, err error) error {
	log.Printf("Error processing event: %s with payload: %v - error: %s\n", event.Topic(), event.Payload(), err.Error())
	return nil
}

func main() {
	e := emitter.NewMemoryEmitter(emitter.WithErrorHandler(CustomErrorHandler))

	listener := func(evt emitter.Event) error {
		return fmt.Errorf("simulated error in listener for event: %w", ErrSimulatedListener)
	}

	_, err := e.On("user.created", listener)
	if err != nil {
		log.Fatalf("Failed to subscribe listener: %v", err)
	}

	errChan := e.Emit("user.created", "Jane Doe")
	for err := range errChan {
		if err != nil {
			log.Printf("Error received from error channel: %v", err)
		}
	}
}
