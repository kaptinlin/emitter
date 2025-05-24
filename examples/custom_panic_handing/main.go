package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

// CustomPanicHandler logs the panic information and performs necessary cleanup.
func CustomPanicHandler(recoveredPanic interface{}) {
	fmt.Printf("Recovered from panic: %v", recoveredPanic)
	// Additional panic recovery logic can go here.
	// For example, you might want to notify an administrator or restart the operation that caused the panic.
}

func main() {
	// Create a new emitter instance with the custom panic handler
	e := emitter.NewMemoryEmitter(emitter.WithPanicHandler(CustomPanicHandler))

	// Define an event listener that intentionally causes a panic
	listener := func(evt emitter.Event) error {
		// Simulating a panic situation
		panic(fmt.Sprintf("simulated panic in listener for event: %s", evt.Topic()))
	}

	// Subscribe the listener to a topic
	_, err := e.On("user.created", listener)
	if err != nil {
		log.Fatalf("Failed to subscribe listener: %v", err)
	}

	// Emit an event which will cause the listener to panic
	// Normally, you would check for errors and handle the error channel, but for the sake of this example, it's omitted.
	e.Emit("user.created", "Jane Doe")

	// Assuming there's additional application logic that continues after event emission,
	// it would carry on uninterrupted thanks to our panic handler.
	fmt.Println("Application continues running despite the panic.")
}
