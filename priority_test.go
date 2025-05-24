package emitter

import (
	"sync"
	"testing"
)

// TestPriorityOrdering checks if the Emitter calls listeners in the correct order of their priorities.
func TestPriorityOrdering(t *testing.T) {
	em := NewMemoryEmitter()

	var mu sync.Mutex // Mutex to protect access to callOrder slice.
	topic := "test_priority_topic"
	callOrder := make([]Priority, 0)
	var wg sync.WaitGroup // WaitGroup to wait for listeners to finish.

	// Helper function to subscribe to the emitter with synchronization.
	subscribeWithPriority := func(priority Priority) {
		wg.Add(1) // Increment the WaitGroup counter.
		_, err := em.On(topic, func(e Event) error {
			defer wg.Done() // Decrement the counter when the function completes.
			mu.Lock()       // Lock the mutex to safely append to callOrder.
			callOrder = append(callOrder, priority)
			mu.Unlock() // Unlock the mutex after appending.
			return nil
		}, WithPriority(priority))
		if err != nil {
			t.Fatalf("Error adding listener with priority %v: %v", priority, err)
		}
	}

	// Set up listeners with different priorities.
	subscribeWithPriority(High)
	subscribeWithPriority(Low)
	subscribeWithPriority(Normal)
	subscribeWithPriority(Lowest)
	subscribeWithPriority(Highest)

	// Emit an event to the topic
	em.Emit(topic, "test_payload")

	wg.Wait() // Wait for all listeners to process the event

	// Verify the call order of listeners matches the expected priority order.
	expectedOrder := []Priority{Highest, High, Normal, Low, Lowest}
	mu.Lock() // Lock the mutex to safely read callOrder.
	defer mu.Unlock()
	if len(callOrder) != len(expectedOrder) {
		t.Fatalf("Expected call order length to be %d, got %d", len(expectedOrder), len(callOrder))
	}

	for i, priority := range expectedOrder {
		if callOrder[i] != priority {
			t.Errorf("Expected priority %v at index %d, got %v", priority, i, callOrder[i])
		}
	}
}

// TestEmitSyncWithAbort tests the synchronous EmitSync method with a listener that aborts the event.
func TestEmitSyncWithAbort(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Create three listeners with different priorities.
	highPriorityListener := func(e Event) error {
		// This listener has the lowest priority and should be called first.
		return nil
	}

	abortingListener := func(e Event) error {
		// This listener aborts the event processing.
		e.SetAborted(true)
		return nil
	}

	lowPriorityListener := func(e Event) error {
		t.Error("The low priority listener should not be called after the event is aborted")
		return nil
	}

	// Subscribe the listeners to the "testTopic".
	if _, err := emitter.On("testTopic", lowPriorityListener, WithPriority(Low)); err != nil {
		t.Fatalf("Error adding low priority listener: %v", err)
	}
	if _, err := emitter.On("testTopic", abortingListener, WithPriority(Normal)); err != nil {
		t.Fatalf("Error adding aborting listener: %v", err)
	}
	if _, err := emitter.On("testTopic", highPriorityListener, WithPriority(High)); err != nil {
		t.Fatalf("Error adding high priority listener: %v", err)
	}

	// Emit the event synchronously.
	emitter.EmitSync("testTopic", "testPayload")
}

// TestEmitWithAbort tests the asynchronous Emit method with a listener that aborts the event.
func TestEmitWithAbort(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Create three listeners with different priorities.
	highPriorityListener := func(e Event) error {
		// This listener has the highest priority and should be called first.
		return nil
	}

	abortingListener := func(e Event) error {
		// This listener aborts the event processing.
		e.SetAborted(true)
		return nil
	}

	lowPriorityListenerCalled := false
	lowPriorityListener := func(e Event) error {
		// This flag should remain false if the event processing is correctly aborted.
		lowPriorityListenerCalled = true
		return nil
	}

	// Subscribe the listeners to the "testTopic".
	if _, err := emitter.On("testTopic", lowPriorityListener, WithPriority(Low)); err != nil {
		t.Fatalf("Error adding low priority listener: %v", err)
	}
	if _, err := emitter.On("testTopic", abortingListener, WithPriority(Normal)); err != nil {
		t.Fatalf("Error adding aborting listener: %v", err)
	}
	if _, err := emitter.On("testTopic", highPriorityListener, WithPriority(High)); err != nil {
		t.Fatalf("Error adding high priority listener: %v", err)
	}

	// Emit the event asynchronously.
	errChan := emitter.Emit("testTopic", "testPayload")

	// Wait for all errors to be collected.
	var emitErrors []error
	for err := range errChan {
		if err != nil {
			emitErrors = append(emitErrors, err)
		}
	}

	// Check that the low priority listener was not called.
	if lowPriorityListenerCalled {
		t.Error("The low priority listener should not have been called")
	}

	// Check that there were no errors during emission.
	if len(emitErrors) != 0 {
		t.Errorf("Emit() resulted in errors: %v", emitErrors)
	}
}
