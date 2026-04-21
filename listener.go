package emitter

// Listener handles an emitted event.
type Listener func(Event) error

type listenerItem struct {
	listener Listener
	priority Priority
}

// ListenerOption configures a [listenerItem] when registering a listener.
type ListenerOption func(*listenerItem)

// WithPriority sets the priority level for a listener.
// Invalid priorities are clamped to the valid range [Lowest, Highest].
func WithPriority(priority Priority) ListenerOption {
	return func(item *listenerItem) {
		item.priority = max(Lowest, min(priority, Highest))
	}
}
