package emitter

import (
	"errors"
	"fmt"
)

var (
	// ErrEmitterClosed is returned when an operation is performed on a closed Emitter.
	ErrEmitterClosed = errors.New("emitter: closed")

	// ErrInvalidTopicName is returned when a topic name does not satisfy the topic grammar.
	ErrInvalidTopicName = errors.New("emitter: invalid topic name")

	// ErrNilListener is returned when On is called with a nil listener.
	ErrNilListener = errors.New("emitter: nil listener")

	// ErrListenerPanic marks an error that originates from a recovered listener panic.
	// Use errors.Is(err, ErrListenerPanic) to detect it.
	ErrListenerPanic = errors.New("emitter: listener panicked")

	// ErrPayloadType is returned by Subscribe[T] when the emitted payload does not match T.
	ErrPayloadType = errors.New("emitter: payload type mismatch")
)

// PanicError wraps a recovered listener panic.
// Inspect the recovered value via Value; the error chain includes ErrListenerPanic
// and the original error (if the panic value was an error).
type PanicError struct {
	// Value is the original recovered value (any).
	Value any
	cause error
}

// Error implements error.
func (e *PanicError) Error() string {
	return fmt.Sprintf("emitter: listener panicked: %v", e.Value)
}

// Unwrap returns the chained errors so errors.Is and errors.As traverse them.
func (e *PanicError) Unwrap() []error {
	if e.cause != nil {
		return []error{ErrListenerPanic, e.cause}
	}
	return []error{ErrListenerPanic}
}
