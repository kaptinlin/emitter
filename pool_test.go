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
	emitter := NewMemoryEmitter(WithPool(NewPondPool(10, 100)))

	numConcurrentEvents := 10

	var wg sync.WaitGroup
	wg.Add(numConcurrentEvents)

	var processingError error

	_, err := emitter.On("testEvent", func(event Event) error {
		time.Sleep(100 * time.Millisecond)
		wg.Done()
		return nil
	})
	require.NoError(t, err, "Error adding listener")

	for range numConcurrentEvents {
		go func() {
			errChan := emitter.Emit("testEvent", nil)
			for err := range errChan {
				if err != nil {
					processingError = err
					break
				}
			}
		}()
	}

	wg.Wait()

	assert.NoError(t, processingError, "Error processing event")
}
