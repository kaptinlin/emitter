// Package main demonstrates custom listener ID generation.
package main

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/kaptinlin/emitter"
)

func main() {
	uuidGenerator := uuid.NewString
	e := emitter.NewMemoryEmitter(emitter.WithIDGenerator(uuidGenerator))

	listener := func(evt emitter.Event) error {
		fmt.Printf("Received event: %s with payload: %+v\n", evt.Topic(), evt.Payload())
		return nil
	}

	listenerID, err := e.On("user.created", listener)
	if err != nil {
		fmt.Printf("Error subscribing listener: %v\n", err)
		return
	}

	fmt.Printf("Listener with ID %s subscribed to topic 'user.created'\n", listenerID)
	e.Emit("user.created", "Jane Doe")
}
