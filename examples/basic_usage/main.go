package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

func main() {
	// Create a new memory-based event emitter
	e := emitter.NewMemoryEmitter()

	// Define a simple event listener
	userCreatedListener := func(evt emitter.Event) error {
		username := evt.Payload().(string)
		fmt.Printf("User created: %s\n", username)
		return nil
	}

	// Subscribe the listener to the "user.created" topic
	listenerID, err := e.On("user.created", userCreatedListener)
	if err != nil {
		log.Fatalf("Failed to subscribe listener: %v", err)
	}
	fmt.Printf("Listener registered with ID: %s\n", listenerID)

	// Emit events synchronously
	fmt.Println("\n--- Synchronous Emission ---")
	errs := e.EmitSync("user.created", "alice@example.com")
	if len(errs) > 0 {
		log.Printf("Errors during emission: %v", errs)
	}

	errs = e.EmitSync("user.created", "bob@example.com")
	if len(errs) > 0 {
		log.Printf("Errors during emission: %v", errs)
	}

	// Emit events asynchronously
	fmt.Println("\n--- Asynchronous Emission ---")
	errChan := e.Emit("user.created", "charlie@example.com")

	// Wait for async completion and handle any errors
	for err := range errChan {
		if err != nil {
			log.Printf("Error from async emission: %v", err)
		}
	}

	fmt.Println("\n--- Multiple Listeners ---")

	// Add another listener to the same topic
	notificationListener := func(evt emitter.Event) error {
		username := evt.Payload().(string)
		fmt.Printf("Sending welcome email to: %s\n", username)
		return nil
	}

	_, err = e.On("user.created", notificationListener)
	if err != nil {
		log.Fatalf("Failed to subscribe notification listener: %v", err)
	}

	// This will trigger both listeners
	e.EmitSync("user.created", "david@example.com")

	// Clean up
	err = e.Close()
	if err != nil {
		log.Printf("Error closing emitter: %v", err)
	}

	fmt.Println("\n--- Emitter Closed Successfully ---")
}
