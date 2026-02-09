package emitter

import "errors"

// Sentinel errors for the emitter package.
var (
	// Initialization errors relate to the setup of listeners and topics.
	ErrNilListener      = errors.New("listener cannot be nil")
	ErrInvalidTopicName = errors.New("invalid topic name")
	ErrInvalidPriority  = errors.New("invalid priority")

	// Runtime errors occur during event emission and listener execution.
	ErrTopicNotFound          = errors.New("topic not found")
	ErrListenerNotFound       = errors.New("listener not found")
	ErrEventProcessingAborted = errors.New("event processing aborted")

	// Manager errors are related to the emitter lifecycle.
	ErrEmitterClosed        = errors.New("emitter is closed")
	ErrEmitterAlreadyClosed = errors.New("emitter is already closed")
)
