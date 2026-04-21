package emitter

import "sync/atomic"

// Event represents a value being emitted through a topic.
type Event interface {
	// Topic returns the event topic.
	Topic() string
	// Payload returns the event payload.
	Payload() any
	// SetPayload replaces the event payload.
	SetPayload(any)
	// SetAborted stops delivery to remaining listeners.
	SetAborted(bool)
	// IsAborted reports whether delivery has been stopped.
	IsAborted() bool
}

// BaseEvent is the default [Event] implementation.
type BaseEvent struct {
	topic   string
	payload atomic.Pointer[any]
	aborted atomic.Bool
}

// NewBaseEvent returns a [BaseEvent] for topic and payload.
func NewBaseEvent(topic string, payload any) *BaseEvent {
	e := &BaseEvent{
		topic: topic,
	}
	e.payload.Store(&payload)
	return e
}

// Topic returns the event topic.
func (e *BaseEvent) Topic() string {
	return e.topic
}

// Payload returns the event payload.
func (e *BaseEvent) Payload() any {
	if p := e.payload.Load(); p != nil {
		return *p
	}
	return nil
}

// SetPayload replaces the event payload.
func (e *BaseEvent) SetPayload(payload any) {
	e.payload.Store(&payload)
}

// SetAborted stops delivery to remaining listeners.
func (e *BaseEvent) SetAborted(abort bool) {
	e.aborted.Store(abort)
}

// IsAborted reports whether delivery has been stopped.
func (e *BaseEvent) IsAborted() bool {
	return e.aborted.Load()
}
