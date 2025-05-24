package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

// ErrSimulatedListener is a static error for demonstration purposes.
var ErrSimulatedListener = errors.New("simulated error in listener")

// CustomErrorHandler logs and handles errors that occur during event processing.
func CustomErrorHandler(event emitter.Event, err error) error {
	// Log the error with additional context or send it to an error tracking service
	log.Printf("Error processing event: %s with payload: %v - error: %s\n", event.Topic(), event.Payload(), err.Error())

	// Here you can decide whether to return the error or handle it so that
	// the emitter considers it resolved.
	// Returning nil will effectively 'swallow' the error, indicating it's been handled.
	return nil
}

func main() {
	// Create a new emitter instance with the custom error handler
	e := emitter.NewMemoryEmitter(emitter.WithErrorHandler(CustomErrorHandler))

	// Define an event listener that intentionally causes an error
	listener := func(evt emitter.Event) error {
		// Simulate an error
		return fmt.Errorf("simulated error in listener for event: %w", ErrSimulatedListener)
	}

	// Subscribe the listener to a topic
	_, err := e.On("user.created", listener)
	if err != nil {
		log.Fatalf("Failed to subscribe listener: %v", err)
	}

	// Emit an event which will cause the listener to error
	errChan := e.Emit("user.created", "Jane Doe")

	// Wait and collect errors from the error channel
	for err := range errChan {
		if err != nil {
			log.Printf("Error received from error channel: %v", err)
		}
	}
}
