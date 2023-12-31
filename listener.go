package emitter

// Listener is a function type that can handle events of any type.
type Listener func(Event) error

// listenerItem stores a listener along with its unique identifier and priority.
type listenerItem struct {
	listener Listener
	priority Priority
}

type ListenerOption func(*listenerItem)

func WithPriority(priority Priority) ListenerOption {
	return func(item *listenerItem) {
		item.priority = priority
	}
}
