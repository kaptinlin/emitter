package emitter

// Event is the view delivered to a [Listener] for a single emission.
// The emitter does not mutate Payload; listeners receive the value passed to Emit.
type Event interface {
	// Topic returns the topic the event was emitted on.
	Topic() string
	// Payload returns the value passed to Emit.
	Payload() any
	// Stop halts dispatch to remaining listeners within this emit.
	Stop()
}

// event is the internal Event implementation.
// It is created once per emit and accessed serially by listeners,
// so it does not need atomic fields.
type event struct {
	topic   string
	payload any
	stopped bool
}

func (e *event) Topic() string { return e.topic }
func (e *event) Payload() any  { return e.payload }
func (e *event) Stop()         { e.stopped = true }
