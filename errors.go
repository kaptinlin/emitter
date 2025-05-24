package emitter

import (
	"errors"
)

// Initialization Errors relate to the setup of listeners and topics.
var (
	ErrNilListener      = errors.New("listener cannot be nil")
	ErrInvalidTopicName = errors.New("invalid topic name")
	ErrInvalidPriority  = errors.New("invalid priority")
)

// Runtime Errors occur during the event emission and listener execution.
var (
	ErrTopicNotFound          = errors.New("topic not found")
	ErrListenerNotFound       = errors.New("listener not found")
	ErrEventProcessingAborted = errors.New("event processing aborted")
)

// Manager Errors are related to the emitter.
var (
	ErrEmitterClosed        = errors.New("emitter is closed")
	ErrEmitterAlreadyClosed = errors.New("emitter is already closed")
)
