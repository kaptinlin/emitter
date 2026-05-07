package emitter

import "sync/atomic"

// Subscription is the handle returned by [Emitter.On].
// Cancel removes the listener; calling Cancel more than once is a no-op.
type Subscription interface {
	// Cancel removes the listener. Idempotent.
	Cancel()
	// Topic returns the pattern the listener was registered for.
	Topic() string
}

type subscription struct {
	emitter   *Emitter
	pattern   string
	id        uint64
	cancelled atomic.Bool
}

func (s *subscription) Topic() string { return s.pattern }

func (s *subscription) Cancel() {
	if s.cancelled.CompareAndSwap(false, true) {
		s.emitter.removeListener(s.pattern, s.id)
	}
}
