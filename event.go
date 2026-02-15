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

func (e *BaseEvent) Topic() string {
	return e.topic
}

func (e *BaseEvent) Payload() any {
	if p := e.payload.Load(); p != nil {
		return *p
	}
	return nil
}

func (e *BaseEvent) SetPayload(payload any) {
	e.payload.Store(&payload)
}

func (e *BaseEvent) SetAborted(abort bool) {
	e.aborted.Store(abort)
}

func (e *BaseEvent) IsAborted() bool {
	return e.aborted.Load()
}
