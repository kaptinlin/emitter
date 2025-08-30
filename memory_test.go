package emitter

import (
	"errors"
	"sync"
	"testing"
	"time"
)

var (
	errTestListenerError = errors.New("listener error")
)

// TestNewMemoryEmitter tests the creation of a new MemoryEmitter.
func TestNewMemoryEmitter(t *testing.T) {
	emitter := NewMemoryEmitter()
	if emitter == nil {
		t.Fatal("NewMemoryEmitter() should not return nil")
	}
}

// TestOnOff tests subscribing to and unsubscribing from a topic.
func TestOnOff(t *testing.T) {
	emitter := NewMemoryEmitter()

	listener := func(e Event) error {
		return nil
	}

	// On to a topic.
	id, err := emitter.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}
	if id == "" {
		t.Fatal("Onrned an empty ID")
	}

	// Now unsubscribe and ensure the listener is removed.
	if err := emitter.Off("testTopic", id); err != nil {
		t.Fatalf("Off() failed with error: %v", err)
	}
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
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

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
	if len(emitErrors) != 0 {
		t.Errorf("Emit() resulted in errors: %v", emitErrors)
	}
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
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

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
	if len(emitErrors) == 0 {
		t.Error("Emit() should have resulted in errors, but none were returned")
	}
}

// TestEmitSyncSuccess tests emitting to a topic.
func TestEmitSyncSuccess(t *testing.T) {
	emitter := NewMemoryEmitter()
	received := make(chan string, 1) // Buffered channel to receive one message.

	// Prepare the listener.
	listener := createTestListener(received)

	// On to the topic.
	_, err := emitter.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

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
		if topic != "testTopic" {
			t.Fatalf("Expected to receive event on 'testTopic', got '%v'", topic)
		}
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
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Emit the event synchronously and collect errors.
	errors := emitter.EmitSync("testTopic", "testPayload")

	// Check that the errors slice is not empty, indicating that an error was returned by the listener.
	if len(errors) == 0 {
		t.Error("EmitSync() should have resulted in errors, but none were returned")
	}
}

// TestGetTopic tests getting a topic.
func TestGetTopic(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Creating a topic by subscribing to it.
	_, err := emitter.On("testTopic", func(e Event) error { return nil })
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	topic, err := emitter.GetTopic("testTopic")
	if err != nil {
		t.Fatalf("GetTopic() failed with error: %v", err)
	}
	if topic == nil {
		t.Fatal("GetTopic() returned nil")
	}
}

// TestEnsureTopic tests getting or creating a topic.
func TestEnsureTopic(t *testing.T) {
	emitter := NewMemoryEmitter()

	// Get or create a new topic.
	topic := emitter.EnsureTopic("newTopic")
	if topic == nil {
		t.Fatal("EnsureTopic() should not return nil")
	}

	// Try to retrieve the same topic and check if it's the same instance.
	sameTopic, err := emitter.GetTopic("newTopic")
	if err != nil {
		t.Fatalf("GetTopic() failed with error: %v", err)
	}
	if sameTopic != topic {
		t.Fatal("EnsureTopic() did not return the same instance of the topic")
	}
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
				name := topicName // Capture the variable for the closure
				_, err := emitter.On(name, func(e Event) error {
					mu.Lock()
					receivedEvents[name] = true
					mu.Unlock()
					wg.Done()
					return nil
				})
				if err != nil {
					t.Fatalf("Failed to subscribe to topic %s: %v", name, err)
				}
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
				if !receivedEvents[expectedTopic] {
					t.Errorf("Expected topic %s to be notified for event %s, but it was not", expectedTopic, tc.event)
				}
			}

			// Verify no unexpected topics were notified
			for topic := range receivedEvents {
				if !contains(tc.expectedMatches, topic) {
					t.Errorf("Topic %s was unexpectedly notified for event %s", topic, tc.event)
				}
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
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	topic2 := "topic2"
	listener2 := func(e Event) error { return nil }
	if _, err := emitter.On(topic2, listener2); err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Use a Pool
	pool := NewPondPool(10, 1000)
	emitter.SetPool(pool)

	// Close the emitter
	if err := emitter.Close(); err != nil {
		t.Errorf("Close() should not return an error: %v", err)
	}

	// Verify topics have been removed
	_, err = emitter.GetTopic(topic1)
	if err == nil {
		t.Errorf("GetTopic() should return an error after Close()")
	}

	_, err = emitter.GetTopic(topic2)
	if err == nil {
		t.Errorf("GetTopic() should return an error after Close()")
	}

	// Verify the pool has been released
	if pool.Running() > 0 {
		t.Errorf("Pool should be released and have no running workers after Close()")
	}

	// Verify that no new events can be emitted
	errChan := emitter.Emit(topic1, "payload")
	select {
	case err := <-errChan:
		if err == nil {
			t.Errorf("Emit() should return an error after Close()")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for the error to be received")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func createTestListener(received chan<- string) func(Event) error {
	return func(e Event) error {
		// Send the topic to the received channel.
		received <- e.Topic()
		return nil
	}
}
