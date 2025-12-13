package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

func main() {
	e := emitter.NewMemoryEmitter()

	fmt.Println("=== Wildcard Pattern Matching Examples ===")
	fmt.Println()

	// Single-level wildcard (*)
	// Matches exactly one segment between dots
	fmt.Println("--- Single-level Wildcard (*) ---")

	singleWildcardListener := func(evt emitter.Event) error {
		fmt.Printf("Single wildcard matched: %s with payload: %v\n", evt.Topic(), evt.Payload())
		return nil
	}

	_, err := e.On("order.*", singleWildcardListener)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// These will match "order.*"
	e.EmitSync("order.created", "Order #123")
	e.EmitSync("order.updated", "Order #124")
	e.EmitSync("order.cancelled", "Order #125")

	// This will NOT match "order.*" (too many segments)
	e.EmitSync("order.item.added", "Order #126")

	fmt.Println()

	// Multi-level wildcard (**)
	// Matches zero or more segments
	fmt.Println("--- Multi-level Wildcard (**) ---")

	multiWildcardListener := func(evt emitter.Event) error {
		fmt.Printf("Multi wildcard matched: %s with payload: %v\n", evt.Topic(), evt.Payload())
		return nil
	}

	_, err = e.On("system.**", multiWildcardListener)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// All of these will match "system.**"
	e.EmitSync("system.cpu.high", "CPU usage 95%")
	e.EmitSync("system.memory.low", "Memory < 100MB")
	e.EmitSync("system.disk.full.warning", "Disk usage 98%")
	e.EmitSync("system.network.latency.spike.detected", "Latency > 500ms")

	// This will NOT match "system.**"
	e.EmitSync("application.error", "App crashed")

	fmt.Println()

	// Combined patterns
	fmt.Println("--- Combined Patterns ---")

	combinedListener := func(evt emitter.Event) error {
		fmt.Printf("Combined pattern matched: %s\n", evt.Topic())
		return nil
	}

	// Match any event that has "error" as the last segment
	_, err = e.On("**.error", combinedListener)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	e.EmitSync("database.connection.error", nil)
	e.EmitSync("api.timeout.error", nil)
	e.EmitSync("validation.error", nil)

	fmt.Println()

	// Multiple wildcard listeners
	fmt.Println("--- Multiple Wildcard Listeners ---")

	allEventsListener := func(evt emitter.Event) error {
		fmt.Printf("All events listener: %s\n", evt.Topic())
		return nil
	}

	_, err = e.On("**", allEventsListener)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// This will trigger multiple listeners
	fmt.Println("\nEmitting 'system.cpu.error' - will match multiple patterns:")
	e.EmitSync("system.cpu.error", "CPU error detected")

	// Clean up
	err = e.Close()
	if err != nil {
		log.Printf("Error closing emitter: %v", err)
	}
}
