package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

func main() {
	e := emitter.NewMemoryEmitter()

	fmt.Println("=== Listener Management Examples ===")
	fmt.Println()

	// Adding listeners
	fmt.Println("--- Adding Listeners ---")

	listener1 := func(evt emitter.Event) error {
		fmt.Println("Listener 1 executed")
		return nil
	}

	listener2 := func(evt emitter.Event) error {
		fmt.Println("Listener 2 executed")
		return nil
	}

	listener3 := func(evt emitter.Event) error {
		fmt.Println("Listener 3 executed")
		return nil
	}

	// Subscribe multiple listeners to the same topic
	id1, err := e.On("notification", listener1)
	if err != nil {
		log.Fatalf("Failed to subscribe listener 1: %v", err)
	}
	fmt.Printf("Listener 1 registered with ID: %s\n", id1)

	id2, err := e.On("notification", listener2)
	if err != nil {
		log.Fatalf("Failed to subscribe listener 2: %v", err)
	}
	fmt.Printf("Listener 2 registered with ID: %s\n", id2)

	id3, err := e.On("notification", listener3)
	if err != nil {
		log.Fatalf("Failed to subscribe listener 3: %v", err)
	}
	fmt.Printf("Listener 3 registered with ID: %s\n", id3)

	// Emit to trigger all listeners
	fmt.Println("\nEmitting notification event (all 3 listeners active):")
	e.EmitSync("notification", nil)

	// Removing a specific listener
	fmt.Println("\n--- Removing Listener 2 ---")
	err = e.Off("notification", id2)
	if err != nil {
		log.Printf("Error removing listener: %v", err)
	} else {
		fmt.Println("Listener 2 removed successfully")
	}

	fmt.Println("\nEmitting notification event (only listeners 1 and 3):")
	e.EmitSync("notification", nil)

	// Removing another listener
	fmt.Println("\n--- Removing Listener 1 ---")
	err = e.Off("notification", id1)
	if err != nil {
		log.Printf("Error removing listener: %v", err)
	} else {
		fmt.Println("Listener 1 removed successfully")
	}

	fmt.Println("\nEmitting notification event (only listener 3):")
	e.EmitSync("notification", nil)

	// Attempting to remove non-existent listener
	fmt.Println("\n--- Attempting to Remove Non-existent Listener ---")
	err = e.Off("notification", "non-existent-id")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Priority-based listener execution
	fmt.Println("\n--- Priority-based Listener Management ---")

	highPriorityListener := func(evt emitter.Event) error {
		fmt.Println("High priority listener")
		return nil
	}

	normalPriorityListener := func(evt emitter.Event) error {
		fmt.Println("Normal priority listener")
		return nil
	}

	lowPriorityListener := func(evt emitter.Event) error {
		fmt.Println("Low priority listener")
		return nil
	}

	// Register listeners with different priorities
	_, err = e.On("task", highPriorityListener, emitter.WithPriority(emitter.High))
	if err != nil {
		log.Fatalf("Failed to subscribe high priority listener: %v", err)
	}

	_, err = e.On("task", normalPriorityListener, emitter.WithPriority(emitter.Normal))
	if err != nil {
		log.Fatalf("Failed to subscribe normal priority listener: %v", err)
	}

	_, err = e.On("task", lowPriorityListener, emitter.WithPriority(emitter.Low))
	if err != nil {
		log.Fatalf("Failed to subscribe low priority listener: %v", err)
	}

	fmt.Println("\nEmitting task event (listeners execute by priority):")
	e.EmitSync("task", nil)

	// Event abortion (stopping propagation)
	fmt.Println("\n--- Event Abortion (Stopping Propagation) ---")

	abortingListener := func(evt emitter.Event) error {
		fmt.Println("Aborting listener - stopping further propagation")
		evt.SetAborted(true)
		return nil
	}

	afterAbortListener := func(evt emitter.Event) error {
		fmt.Println("This should NOT execute after abort")
		return nil
	}

	_, err = e.On("abortable", abortingListener, emitter.WithPriority(emitter.Highest))
	if err != nil {
		log.Fatalf("Failed to subscribe aborting listener: %v", err)
	}

	_, err = e.On("abortable", afterAbortListener, emitter.WithPriority(emitter.Lowest))
	if err != nil {
		log.Fatalf("Failed to subscribe after-abort listener: %v", err)
	}

	fmt.Println("\nEmitting abortable event:")
	e.EmitSync("abortable", nil)

	// Clean up
	err = e.Close()
	if err != nil {
		log.Printf("Error closing emitter: %v", err)
	}

	fmt.Println("\n--- Emitter Closed Successfully ---")
}
