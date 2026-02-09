package emitter

// Listener is a function type that can handle events of any type.
type Listener func(Event) error

// listenerItem stores a listener along with its unique identifier and priority.
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
		// Validate priority and use boundary values if out of range
		item.priority = max(Lowest, min(priority, Highest))
	}
}
