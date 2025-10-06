package emitter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitEventWithPool(t *testing.T) {
	emitter := NewMemoryEmitter(WithPool(NewPondPool(10, 100)))

	var processedEvents int32
	listenerID, err := emitter.On("testEvent", func(event Event) error {
		atomic.AddInt32(&processedEvents, 1)
		time.Sleep(10 * time.Millisecond) // Simulating work
		return nil
	})

	require.NoError(t, err, "Error adding listener")

	errChan := emitter.Emit("testEvent", nil)

	// Collect all errors from the channel.
	var errors []error
	go func() {
		for err := range errChan {
			if err != nil {
				errors = append(errors, err)
			}
		}
	}()

	// Wait for a short duration to ensure event processing has a chance to complete.
	time.Sleep(100 * time.Millisecond)

	// Check for errors reported by the listener.
	assert.Empty(t, errors, "Listener reported errors")

	// Unregister the listener as cleanup.
	err = emitter.Off("testEvent", listenerID)
	require.NoError(t, err, "Failed to unregister listener")

	// Final assertion after cleanup.
	assert.Equal(t, int32(1), atomic.LoadInt32(&processedEvents), "Expected 1 event to be processed")
}

func TestEmitMultipleEventsWithPool(t *testing.T) {
	// Create a MemoryEmitter instance with a PondPool.
	emitter := NewMemoryEmitter(WithPool(NewPondPool(10, 100)))

	// Define the number of concurrent events to emit.
	numConcurrentEvents := 10

	// Define a wait group to wait for all events to be processed.
	var wg sync.WaitGroup
	wg.Add(numConcurrentEvents)

	// Define a variable to keep track of any errors encountered during event processing.
	var processingError error

	// Add an event listener to handle "testEvent" and increment the processedEvents count.
	_, err := emitter.On("testEvent", func(event Event) error {
		// Simulate some processing.
		// For testing, we just sleep for a short time to simulate work.
		// In a real scenario, you should replace this with your actual event processing logic.
		// Sleep for 100 milliseconds to simulate processing.
		// You can adjust the sleep duration based on your test requirements.
		time.Sleep(100 * time.Millisecond)

		// Decrement the wait group to signal event processing completion.
		wg.Done()

		return nil
	})
	require.NoError(t, err, "Error adding listener")

	// Emit multiple events concurrently.
	for i := 0; i < numConcurrentEvents; i++ {
		go func() {
			// Emit an event using the emitter.
			errChan := emitter.Emit("testEvent", nil)

			// Wait for the event to be processed.
			for err := range errChan {
				if err != nil {
					// Capture the first error encountered during event processing.
					processingError = err
					break
				}
			}
		}()
	}

	// Wait for all events to be processed.
	wg.Wait()

	// Check if any errors occurred during event processing.
	assert.NoError(t, processingError, "Error processing event")
}
