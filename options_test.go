package emitter

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTestCustomError = errors.New("custom error")

// TestWithErrorHandler tests that the custom error handler is called on error.
func TestWithErrorHandler(t *testing.T) {
	// Define a variable to determine if the custom error handler was called.
	var handlerCalled bool

	// Define a custom error to be returned by a listener.
	customError := errTestCustomError

	// Define a custom error handler that sets handlerCalled to true.
	customErrorHandler := func(event Event, err error) error {
		if errors.Is(err, customError) {
			handlerCalled = true
			t.Logf("Custom error handler called with event: %s and error: %s", event.Topic(), err.Error())
		}
		return nil // Returning nil to indicate the error is handled.
	}

	// Create a new MemoryEmitter with the custom error handler.
	emitter := NewMemoryEmitter(WithErrorHandler(customErrorHandler))

	// Define a listener that returns the custom error.
	listener := func(e Event) error {
		return customError
	}

	// Subscribe the listener to a topic.
	_, err := emitter.On("testTopic", listener)
	require.NoError(t, err, "On() failed with error")

	// Emit the event synchronously to trigger the error.
	emitter.EmitSync("testTopic", NewBaseEvent("testTopic", "testPayload"))

	// Check if the custom error handler was called.
	assert.True(t, handlerCalled, "Custom error handler was not called on listener error")
}

func TestWithErrorHandlerAsync(t *testing.T) {
	// Define a variable to determine if the custom error handler was called.
	var handlerCalled bool
	var handlerMutex sync.Mutex // To safely update handlerCalled from different goroutines

	// Define a custom error to be returned by a listener.
	customError := errTestCustomError

	// Define a custom error handler that sets handlerCalled to true.
	customErrorHandler := func(event Event, err error) error {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()
		if errors.Is(err, customError) {
			handlerCalled = true
		}
		return nil // Assume the error is handled and return nil.
	}

	// Create a new MemoryEmitter with the custom error handler.
	emitter := NewMemoryEmitter(WithErrorHandler(customErrorHandler))

	// Define a listener that returns the custom error.
	listener := func(e Event) error {
		return customError
	}

	// Subscribe the listener to a topic.
	_, err := emitter.On("testTopic", listener)
	require.NoError(t, err, "On() failed with error")

	// Emit the event asynchronously to trigger the error.
	errChan := emitter.Emit("testTopic", NewBaseEvent("testTopic", "testPayload"))

	// Wait for all errors to be processed.
	for err := range errChan {
		assert.NoError(t, err, "Expected nil error due to custom handler")
	}

	// Check if the custom error handler was called.
	handlerMutex.Lock()
	defer handlerMutex.Unlock()
	assert.True(t, handlerCalled, "Custom error handler was not called on listener error")
}

func TestEmitSyncReturnsRecoveredPanicError(t *testing.T) {
	emitter := NewMemoryEmitter()

	_, err := emitter.On("testTopic", func(Event) error {
		panic(errTestCustomError)
	})
	require.NoError(t, err, "On() failed with error")

	errs := emitter.EmitSync("testTopic", "testPayload")
	require.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], ErrListenerPanic)
	assert.ErrorIs(t, errs[0], errTestCustomError)

	var panicErr *PanicError
	require.ErrorAs(t, errs[0], &panicErr)
	assert.Equal(t, errTestCustomError, panicErr.Value)
}

func TestEmitReturnsRecoveredPanicError(t *testing.T) {
	emitter := NewMemoryEmitter()

	_, err := emitter.On("testTopic", func(Event) error {
		panic("test panic")
	})
	require.NoError(t, err, "On() failed with error")

	errChan := emitter.Emit("testTopic", "testPayload")
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	require.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], ErrListenerPanic)

	var panicErr *PanicError
	require.ErrorAs(t, errs[0], &panicErr)
	assert.Equal(t, "test panic", panicErr.Value)
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
	require.NoError(t, err, "On() failed with error")

	// Check if the returned ID matches the custom ID.
	assert.Equal(t, customID, returnedID, "Expected ID to match custom ID")
}

func TestWithErrChanBufferSize(t *testing.T) {
	t.Parallel()

	emitter := NewMemoryEmitter(WithErrChanBufferSize(2))
	listenerErr := errors.New("listener error")

	for i := range 2 {
		_, err := emitter.On("testTopic", func(Event) error {
			return listenerErr
		})
		require.NoError(t, err, "On() failed for listener %d", i)
	}

	errChan := emitter.Emit("testTopic", "testPayload")
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	require.Len(t, errs, 2)
	for _, err := range errs {
		assert.ErrorIs(t, err, listenerErr)
	}
}

func TestSetErrChanBufferSizeClampsNegativeValues(t *testing.T) {
	t.Parallel()

	emitter := NewMemoryEmitter()
	emitter.SetErrChanBufferSize(-1)

	listenerErr := errors.New("listener error")
	_, err := emitter.On("testTopic", func(Event) error {
		return listenerErr
	})
	require.NoError(t, err)

	errChan := emitter.Emit("testTopic", "testPayload")
	assert.Zero(t, cap(errChan))

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}
	if assert.Len(t, errs, 1) {
		assert.ErrorIs(t, errs[0], listenerErr)
	}
}

func TestSetErrorHandlerNilKeepsExistingHandler(t *testing.T) {
	t.Parallel()

	emitter := NewMemoryEmitter()
	handledErr := errors.New("handled error")
	emitter.SetErrorHandler(func(Event, error) error {
		return handledErr
	})
	emitter.SetErrorHandler(nil)

	_, err := emitter.On("testTopic", func(Event) error {
		return errTestCustomError
	})
	require.NoError(t, err)

	errChan := emitter.Emit("testTopic", "testPayload")
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}
	if assert.Len(t, errs, 1) {
		assert.ErrorIs(t, errs[0], handledErr)
	}
}

func TestSetIDGeneratorNilKeepsExistingGenerator(t *testing.T) {
	t.Parallel()

	emitter := NewMemoryEmitter()
	emitter.SetIDGenerator(func() string { return "custom-id" })
	emitter.SetIDGenerator(nil)

	returnedID, err := emitter.On("testTopic", func(Event) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, "custom-id", returnedID)
}

func TestPanicErrorFormatsRecoveredValue(t *testing.T) {
	t.Parallel()

	err := &PanicError{Value: "boom"}
	assert.Equal(t, "emitter: listener panicked: boom", err.Error())
}

func TestDefaultHandlers(t *testing.T) {
	t.Parallel()

	err := errors.New("listener error")
	assert.Same(t, err, DefaultErrorHandler(NewBaseEvent("testTopic", nil), err))
	assert.NotEmpty(t, DefaultIDGenerator())
}
