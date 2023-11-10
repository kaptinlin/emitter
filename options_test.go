package emitter

import (
	"errors"
	"sync"
	"testing"
)

// TestWithErrorHandler tests that the custom error handler is called on error.
func TestWithErrorHandler(t *testing.T) {
	// Define a variable to determine if the custom error handler was called.
	var handlerCalled bool

	// Define a custom error to be returned by a listener.
	customError := errors.New("custom error")

	// Define a custom error handler that sets handlerCalled to true.
	customErrorHandler := func(err error) error {
		if err == customError {
			handlerCalled = true
		}
		return err
	}

	// Create a new MemoryEmitter with the custom error handler.
	emitter := NewMemoryEmitter(WithErrorHandler(customErrorHandler))

	// Define a listener that returns the custom error.
	listener := func(e Event) error {
		return customError
	}

	// Subscribe the listener to a topic.
	_, err := emitter.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Emit the event synchronously to trigger the error.
	emitter.EmitSync("testTopic", "testPayload")

	// Check if the custom error handler was called.
	if !handlerCalled {
		t.Fatalf("Custom error handler was not called on listener error")
	}
}

func TestWithErrorHandlerAsync(t *testing.T) {
	// Define a variable to determine if the custom error handler was called.
	var handlerCalled bool
	var handlerMutex sync.Mutex // To safely update handlerCalled from different goroutines

	// Define a custom error to be returned by a listener.
	customError := errors.New("custom error")

	// Define a custom error handler that sets handlerCalled to true.
	customErrorHandler := func(err error) error {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()
		if err == customError {
			handlerCalled = true
		}
		return err
	}

	// Create a new MemoryEmitter with the custom error handler.
	emitter := NewMemoryEmitter(WithErrorHandler(customErrorHandler))

	// Define a listener that returns the custom error.
	listener := func(e Event) error {
		return customError
	}

	// Subscribe the listener to a topic.
	_, err := emitter.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Emit the event asynchronously to trigger the error.
	errChan := emitter.Emit("testTopic", "testPayload")

	// Wait for all errors to be processed.
	for err := range errChan {
		if err != customError {
			t.Errorf("Expected custom error, got: %v", err)
		}
	}

	// Check if the custom error handler was called.
	handlerMutex.Lock()
	wasHandlerCalled := handlerCalled
	handlerMutex.Unlock()

	if !wasHandlerCalled {
		t.Fatalf("Custom error handler was not called on listener error")
	}
}

func TestWithPanicHandlerSync(t *testing.T) {
	// Flag to indicate panic handler invocation
	var panicHandlerInvoked bool

	// Define a custom panic handler
	customPanicHandler := func(p interface{}) {
		if p == "test panic" {
			panicHandlerInvoked = true
		}
	}

	// Create a new MemoryEmitter with the custom panic handler.
	emitter := NewMemoryEmitter(WithPanicHandler(customPanicHandler))

	// Define a listener that panics
	listener := func(e Event) error {
		panic("test panic")
	}

	// Subscribe the listener to a topic.
	_, err := emitter.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Recover from panic to prevent test failure
	defer func() {
		if r := recover(); r != nil {
			// This is expected
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	// Emit the event synchronously to trigger the panic.
	emitter.EmitSync("testTopic", "testPayload")

	// Verify that the custom panic handler was invoked
	if !panicHandlerInvoked {
		t.Fatalf("Custom panic handler was not called on listener panic")
	}
}

func TestWithPanicHandlerAsync(t *testing.T) {
	// Flag to indicate panic handler invocation
	var panicHandlerInvoked bool
	var panicHandlerMutex sync.Mutex // To safely update panicHandlerInvoked from different goroutines

	// Define a custom panic handler
	customPanicHandler := func(p interface{}) {
		panicHandlerMutex.Lock()
		defer panicHandlerMutex.Unlock()
		if p == "test panic" {
			panicHandlerInvoked = true
		}
	}

	// Create a new MemoryEmitter with the custom panic handler.
	emitter := NewMemoryEmitter(WithPanicHandler(customPanicHandler))

	// Define a listener that panics
	listener := func(e Event) error {
		panic("test panic")
	}

	// Subscribe the listener to a topic.
	_, err := emitter.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Emit the event asynchronously to trigger the panic.
	errChan := emitter.Emit("testTopic", "testPayload")

	// Wait for all events to be processed (which includes recovering from panic).
	for range errChan {
		// Normally, you'd check for errors here, but in this case, we expect a panic, not an error
	}

	// Verify that the custom panic handler was invoked
	panicHandlerMutex.Lock()
	wasPanicHandlerInvoked := panicHandlerInvoked
	panicHandlerMutex.Unlock()

	if !wasPanicHandlerInvoked {
		t.Fatalf("Custom panic handler was not called on listener panic")
	}
}

func TestWithIDGenerator(t *testing.T) {
	// Custom ID to be returned by the custom ID generator
	customID := "customID"

	// Define a custom ID generator that returns the custom ID
	customIDGenerator := func() string {
		return customID
	}

	// Create a new MemoryEmitter with the custom ID generator.
	emitter := NewMemoryEmitter(WithIDGenerator(customIDGenerator))

	// Define a no-op listener.
	listener := func(e Event) error {
		return nil
	}

	// Subscribe the listener to a topic and capture the returned ID.
	returnedID, err := emitter.On("testTopic", listener)
	if err != nil {
		t.Fatalf("On() failed with error: %v", err)
	}

	// Check if the returned ID matches the custom ID.
	if returnedID != customID {
		t.Fatalf("Expected ID to be '%s', but got '%s'", customID, returnedID)
	}
}
