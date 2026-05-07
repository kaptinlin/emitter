package emitter

import "context"

// Listener handles an emitted event.
// Within a single emit, listeners run synchronously in priority order;
// implementations should be reasonably non-blocking and respect ctx.
type Listener func(ctx context.Context, ev Event) error

// ListenerOption configures a listener at registration time.
type ListenerOption func(*listenerOpts)

type listenerOpts struct {
	priority Priority
	once     bool
}

// WithPriority sets the listener's dispatch priority.
// Higher values run earlier; equal values run in registration order.
func WithPriority(p Priority) ListenerOption {
	return func(o *listenerOpts) { o.priority = p }
}

// Once registers the listener to fire at most once across all emits.
// After it fires, the subscription is automatically cancelled.
// Concurrent emits observe at-most-once semantics via atomic compare-and-swap.
func Once() ListenerOption {
	return func(o *listenerOpts) { o.once = true }
}
