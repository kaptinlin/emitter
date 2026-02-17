package emitter

import "sync/atomic"

// Event represents the structure of an event with topic, payload, and abort state.
type Event interface {
	Topic() string
	Payload() any
	SetPayload(any)
	SetAborted(bool)
	IsAborted() bool
}

// BaseEvent provides a thread-safe implementation of the [Event] interface
// using atomic operations for payload and abort state.
type BaseEvent struct {
	topic   string
	payload atomic.Pointer[any]
	aborted atomic.Bool
}

// NewBaseEvent creates a new BaseEvent with the given topic and payload.
func NewBaseEvent(topic string, payload any) *BaseEvent {
	e := &BaseEvent{
		topic: topic,
	}
	e.payload.Store(&payload)
	return e
}

// Topic returns the event's topic name.
func (e *BaseEvent) Topic() string {
	return e.topic
}

// Payload returns the event's payload data.
func (e *BaseEvent) Payload() any {
	if p := e.payload.Load(); p != nil {
		return *p
	}
	return nil
}

// SetPayload updates the event's payload.
func (e *BaseEvent) SetPayload(payload any) {
	e.payload.Store(&payload)
}

// SetAborted sets the abort state of the event.
func (e *BaseEvent) SetAborted(abort bool) {
	e.aborted.Store(abort)
}

// IsAborted reports whether the event has been aborted.
func (e *BaseEvent) IsAborted() bool {
	return e.aborted.Load()
}
