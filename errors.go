package emitter

import (
	"errors"
	"fmt"
)

var (
	// ErrNilListener is returned when On receives a nil listener.
	ErrNilListener = errors.New("listener cannot be nil")
	// ErrInvalidTopicName is returned when a topic name is invalid.
	ErrInvalidTopicName = errors.New("invalid topic name")
	// ErrInvalidPriority is returned when a priority is invalid.
	ErrInvalidPriority = errors.New("invalid priority")

	// ErrTopicNotFound is returned when a topic does not exist.
	ErrTopicNotFound = errors.New("topic not found")
	// ErrListenerNotFound is returned when a listener ID does not exist.
	ErrListenerNotFound = errors.New("listener not found")

	// ErrEmitterClosed is returned when an operation uses a closed emitter.
	ErrEmitterClosed = errors.New("emitter is closed")
	// ErrEmitterAlreadyClosed is returned when Close is called more than once.
	ErrEmitterAlreadyClosed = errors.New("emitter is already closed")
	// ErrListenerPanic marks an error that came from a recovered listener panic.
	ErrListenerPanic = errors.New("listener panicked")
)

// PanicError wraps a recovered listener panic.
type PanicError struct {
	// Value is the recovered panic value.
	Value any
	// Cause is the recovered value when it implements error.
	Cause error
}

// Error returns the panic error message.
func (e *PanicError) Error() string {
	return fmt.Sprintf("emitter: listener panicked: %v", e.Value)
}

// Unwrap returns ErrListenerPanic and the recovered error, when available.
func (e *PanicError) Unwrap() error {
	if e.Cause != nil {
		return errors.Join(ErrListenerPanic, e.Cause)
	}
	return ErrListenerPanic
}
