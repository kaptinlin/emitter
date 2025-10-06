package emitter

import (
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errTestListenerError = errors.New("listener error")
)

// TestNewMemoryEmitter tests the creation of a new MemoryEmitter.
func TestNewMemoryEmitter(t *testing.T) {
	emitter := NewMemoryEmitter()
	assert.NotNil(t, emitter, "NewMemoryEmitter() should not return nil")
}

// TestOnOff tests subscribing to and unsubscribing from a topic.
func TestOnOff(t *testing.T) {
	emitter := NewMemoryEmitter()

	listener := func(e Event) error {
		return nil
	}

	// On to a topic.
	id, err := emitter.On("testTopic", listener)
	require.NoError(t, err, "On() failed with error")
	assert.NotEmpty(t, id, "On returned an empty ID")

	// Now unsubscribe and ensure the listener is removed.
	err = emitter.Off("testTopic", id)
	require.NoError(t, err, "Off() failed with error")
}

// TestEmitAsyncSuccess tests the asynchronous Emit method for successful event handling.
func TestEmitAsyncSuccess(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Create a listener that does not return an error.
	listener := func(e Event) error {
		// Simulate some work.
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	// Subscribe the listener to the "testTopic".
	_, err := emitter.On("testTopic", listener)
	require.NoError(t, err, "On() failed with error")

	// Emit the event asynchronously.
	errChan := emitter.Emit("testTopic", "testPayload")

	// Collect errors from the error channel.
	var emitErrors []error
	for err := range errChan {
		if err != nil {
			emitErrors = append(emitErrors, err)
		}
	}

	// Check that there were no errors during emission.
	assert.Empty(t, emitErrors, "Emit() resulted in errors")
}

// TestEmitAsyncFailure tests the asynchronous Emit method for event handling that returns an error.
func TestEmitAsyncFailure(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Create a listener that returns an error.
	listener := func(e Event) error {
		// Simulate some work.
		time.Sleep(10 * time.Millisecond)
		return errTestListenerError
	}

	// Subscribe the listener to the "testTopic".
	_, err := emitter.On("testTopic", listener)
	require.NoError(t, err, "On() failed with error")

	// Emit the event asynchronously.
	errChan := emitter.Emit("testTopic", "testPayload")

	// Collect errors from the error channel.
	var emitErrors []error
	for err := range errChan {
		if err != nil {
			emitErrors = append(emitErrors, err)
		}
	}

	// Check that the errors slice is not empty, indicating that an error was returned by the listener.
	assert.NotEmpty(t, emitErrors, "Emit() should have resulted in errors, but none were returned")
}

// TestEmitSyncSuccess tests emitting to a topic.
func TestEmitSyncSuccess(t *testing.T) {
	emitter := NewMemoryEmitter()
	received := make(chan string, 1) // Buffered channel to receive one message.

	// Prepare the listener.
	listener := createTestListener(received)

	// On to the topic.
	_, err := emitter.On("testTopic", listener)
	require.NoError(t, err, "On() failed with error")

	// Emit the event and ignore the error channel for this test
	go func() {
		errChan := emitter.Emit("testTopic", "testPayload")
		for range errChan {
			// Consume errors to prevent goroutine leak
		}
	}()

	// Wait for the listener to handle the event or timeout after a specific duration.
	select {
	case topic := <-received:
		assert.Equal(t, "testTopic", topic, "Expected to receive event on 'testTopic'")
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for the event to be received")
	}
}

// TestEmitSyncFailure tests the synchronous EmitSync method for event handling that returns an error.
func TestEmitSyncFailure(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Create a listener that returns an error.
	listener := func(e Event) error {
		return errTestListenerError
	}

	// Subscribe the listener to the "testTopic".
	_, err := emitter.On("testTopic", listener)
	require.NoError(t, err, "On() failed with error")

	// Emit the event synchronously and collect errors.
	errors := emitter.EmitSync("testTopic", "testPayload")

	// Check that the errors slice is not empty, indicating that an error was returned by the listener.
	assert.NotEmpty(t, errors, "EmitSync() should have resulted in errors, but none were returned")
}

// TestGetTopic tests getting a topic.
func TestGetTopic(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Creating a topic by subscribing to it.
	_, err := emitter.On("testTopic", func(e Event) error { return nil })
	require.NoError(t, err, "On() failed with error")

	topic, err := emitter.GetTopic("testTopic")
	require.NoError(t, err, "GetTopic() failed with error")
	assert.NotNil(t, topic, "GetTopic() returned nil")
}

// TestEnsureTopic tests getting or creating a topic.
func TestEnsureTopic(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Get or create a new topic.
	topic := emitter.EnsureTopic("newTopic")
	assert.NotNil(t, topic, "EnsureTopic() should not return nil")

	// Try to retrieve the same topic and check if it's the same instance.
	sameTopic, err := emitter.GetTopic("newTopic")
	require.NoError(t, err, "GetTopic() failed with error")
	assert.Same(t, topic, sameTopic, "EnsureTopic() did not return the same instance of the topic")
}

func TestWildcardSubscriptionAndEmiting(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		event           string
		expectedMatches []string
	}{
		{
			name:            "full_match",
			event:           "event.some.thing.run",
			expectedMatches: []string{"event.some.*.*", "event.some.*.run", "event.some.**", "**.thing.run"},
		},
		{
			name:            "partial_match",
			event:           "event.some.thing.do",
			expectedMatches: []string{"event.some.*.*", "event.some.**"},
		},
		{
			name:            "simple_match",
			event:           "event.some.thing",
			expectedMatches: []string{"event.some.**"},
		},
	}

	topics := []string{
		"event.some.*.*",
		"event.some.*.run",
		"event.some.**",
		"**.thing.run",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			emitter := NewMemoryEmitter()
			receivedEvents := make(map[string]bool)
			var mu sync.Mutex
			var wg sync.WaitGroup

			// Subscribe to all topics with listeners that track received events
			for _, topicName := range topics {
				_, err := emitter.On(topicName, func(e Event) error {
					mu.Lock()
					receivedEvents[topicName] = true
					mu.Unlock()
					wg.Done()
					return nil
				})
				require.NoError(t, err, "Failed to subscribe to topic %s", topicName)
			}

			// Set up wait group to wait for expected listeners
			wg.Add(len(tc.expectedMatches))

			// Emit the event
			errChan := emitter.Emit(tc.event, tc.event)
			go func() {
				for range errChan {
					// Consume errors to prevent goroutine leak
				}
			}()

			// Wait for all expected listeners to be called or timeout
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// All expected listeners were called
			case <-time.After(5 * time.Second):
				t.Fatal("Test timed out waiting for listeners to be called")
			}

			// Verify that the correct topics were notified
			mu.Lock()
			defer mu.Unlock()

			for _, expectedTopic := range tc.expectedMatches {
				assert.True(t, receivedEvents[expectedTopic], "Expected topic %s to be notified for event %s", expectedTopic, tc.event)
			}

			// Verify no unexpected topics were notified
			for topic := range receivedEvents {
				assert.True(t, slices.Contains(tc.expectedMatches, topic), "Topic %s was unexpectedly notified for event %s", topic, tc.event)
			}
		})
	}
}

func TestMemoryEmitterClose(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Set up topics and listeners
	topic1 := "topic1"
	listener1 := func(e Event) error { return nil }
	_, err := emitter.On(topic1, listener1)
	require.NoError(t, err, "On() failed with error")

	topic2 := "topic2"
	listener2 := func(e Event) error { return nil }
	_, err = emitter.On(topic2, listener2)
	require.NoError(t, err, "On() failed with error")

	// Use a Pool
	pool := NewPondPool(10, 1000)
	emitter.SetPool(pool)

	// Close the emitter
	err = emitter.Close()
	require.NoError(t, err, "Close() should not return an error")

	// Verify topics have been removed
	_, err = emitter.GetTopic(topic1)
	assert.Error(t, err, "GetTopic() should return an error after Close()")

	_, err = emitter.GetTopic(topic2)
	assert.Error(t, err, "GetTopic() should return an error after Close()")

	// Verify the pool has been released
	assert.Equal(t, 0, pool.Running(), "Pool should be released and have no running workers after Close()")

	// Verify that no new events can be emitted
	errChan := emitter.Emit(topic1, "payload")
	select {
	case err := <-errChan:
		assert.Error(t, err, "Emit() should return an error after Close()")
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for the error to be received")
	}
}

func createTestListener(received chan<- string) func(Event) error {
	return func(e Event) error {
		// Send the topic to the received channel.
		received <- e.Topic()
		return nil
	}
}
