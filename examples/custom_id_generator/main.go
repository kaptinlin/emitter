package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kaptinlin/emitter"
)

func main() {
	// Custom ID generator using UUID v4
	uuidGenerator := uuid.NewString

	// Create a new emitter instance with the custom ID generator
	e := emitter.NewMemoryEmitter(emitter.WithIDGenerator(uuidGenerator))

	// Define an event listener
	listener := func(evt emitter.Event) error {
		// The listener does something with the event
		fmt.Printf("Received event: %s with payload: %+v\n", evt.Topic(), evt.Payload())
		return nil
	}

	// Subscribe the listener to a topic and retrieve the listener's unique ID
	listenerID, err := e.On("user.created", listener)
	if err != nil {
		fmt.Printf("Error subscribing listener: %v\n", err)
		return
	}

	// The listenerID returned from the subscription is the unique UUID generated for the listener
	fmt.Printf("Listener with ID %s subscribed to topic 'user.created'\n", listenerID)

	// Emit an event
	e.Emit("user.created", "Jane Doe")
}
