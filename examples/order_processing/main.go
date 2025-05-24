package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kaptinlin/emitter"
)

func main() {
	// Initialize the emitter
	e := emitter.NewMemoryEmitter()

	// High-priority listener for order validation
	validateOrderListener := func(evt emitter.Event) error {
		orderID := evt.Payload().(string)
		// Perform validation logic...
		fmt.Printf("Validating order: %s\n", orderID)
		// Simulate order validation failure
		if orderID == "order123" {
			fmt.Println("Validation failed. Aborting event propagation.")
			evt.SetAborted(true)
		}
		return nil
	}

	// Listener for processing the payment
	processPaymentListener := func(evt emitter.Event) error {
		if evt.IsAborted() {
			fmt.Println("Payment processing skipped due to previous validation failure.")
			return nil
		}
		orderID := evt.Payload().(string)
		// Process payment logic...
		fmt.Printf("Processing payment for order: %s\n", orderID)
		return nil
	}

	// Listener for sending confirmation email
	sendConfirmationEmailListener := func(evt emitter.Event) error {
		if evt.IsAborted() {
			fmt.Println("Confirmation email not sent due to event abort.")
			return nil
		}
		orderID := evt.Payload().(string)
		// Send email logic...
		fmt.Printf("Sending confirmation email for order: %s\n", orderID)
		return nil
	}

	// Subscribe listeners with specified priorities
	_, err := e.On("order.created", validateOrderListener, emitter.WithPriority(emitter.Highest))
	if err != nil {
		log.Fatalf("Failed to subscribe validate order listener: %v", err)
	}
	_, err = e.On("order.created", processPaymentListener, emitter.WithPriority(emitter.Normal))
	if err != nil {
		log.Fatalf("Failed to subscribe process payment listener: %v", err)
	}
	_, err = e.On("order.created", sendConfirmationEmailListener, emitter.WithPriority(emitter.Low))
	if err != nil {
		log.Fatalf("Failed to subscribe send confirmation email listener: %v", err)
	}

	// Emit events for order creation
	fmt.Println("Emitting event for order creation...")
	e.Emit("order.created", "order123") // This order will fail validation
	e.Emit("order.created", "order456") // This order will pass validation

	// Allow time for events to be processed
	time.Sleep(1 * time.Second) // Replace with proper synchronization in production
}
