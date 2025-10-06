package emitter

import "sync/atomic"

// Event is an interface representing the structure of an event.
type Event interface {
	Topic() string
	Payload() any
	SetPayload(any)
	SetAborted(bool)
	IsAborted() bool
}

// BaseEvent provides a basic implementation of the Event interface.
type BaseEvent struct {
	topic   string
	payload atomic.Pointer[any]
	aborted atomic.Bool
}

// NewBaseEvent creates a new instance of BaseEvent with a payload.
func NewBaseEvent(topic string, payload any) *BaseEvent {
	e := &BaseEvent{
		topic: topic,
	}
	e.payload.Store(&payload)
	return e
}

// Topic returns the event's topic.
func (e *BaseEvent) Topic() string {
	return e.topic
}

// Payload returns the event's payload.
func (e *BaseEvent) Payload() any {
	if p := e.payload.Load(); p != nil {
		return *p
	}
	return nil
}

// SetPayload sets the event's payload.
func (e *BaseEvent) SetPayload(payload any) {
	e.payload.Store(&payload)
}

// SetAborted sets the event's aborted status.
func (e *BaseEvent) SetAborted(abort bool) {
	e.aborted.Store(abort)
}

// IsAborted checks the event's aborted status.
func (e *BaseEvent) IsAborted() bool {
	return e.aborted.Load()
}
