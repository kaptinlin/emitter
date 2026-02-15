package emitter

import "errors"

// Sentinel errors for the emitter package.
var (
	ErrNilListener      = errors.New("listener cannot be nil")
	ErrInvalidTopicName = errors.New("invalid topic name")
	ErrInvalidPriority  = errors.New("invalid priority")

	ErrTopicNotFound    = errors.New("topic not found")
	ErrListenerNotFound = errors.New("listener not found")

	ErrEmitterClosed        = errors.New("emitter is closed")
	ErrEmitterAlreadyClosed = errors.New("emitter is already closed")
)
