package emitter

import (
	"errors"
	"fmt"
)

// Sentinel errors for the emitter package.
var (
	ErrNilListener      = errors.New("listener cannot be nil")
	ErrInvalidTopicName = errors.New("invalid topic name")
	ErrInvalidPriority  = errors.New("invalid priority")

	ErrTopicNotFound    = errors.New("topic not found")
	ErrListenerNotFound = errors.New("listener not found")

	ErrEmitterClosed        = errors.New("emitter is closed")
	ErrEmitterAlreadyClosed = errors.New("emitter is already closed")
	ErrListenerPanic        = errors.New("listener panicked")
)

// PanicError wraps a recovered listener panic so callers can inspect it with errors.Is.
type PanicError struct {
	Value any
	Cause error
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("emitter: listener panicked: %v", e.Value)
}

func (e *PanicError) Unwrap() error {
	if e.Cause != nil {
		return errors.Join(ErrListenerPanic, e.Cause)
	}
	return ErrListenerPanic
}

func newPanicError(recovered any) *PanicError {
	panicErr := &PanicError{Value: recovered}
	if err, ok := recovered.(error); ok {
		panicErr.Cause = err
	}
	return panicErr
}
