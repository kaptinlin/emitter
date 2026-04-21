package emitter

// Emitter manages listeners and emits events by topic.
type Emitter interface {
	// On registers listener for topicName and returns its listener ID.
	On(topicName string, listener Listener, opts ...ListenerOption) (string, error)

	// Off removes the listener identified by listenerID from topicName.
	Off(topicName string, listenerID string) error

	// Emit emits payload to matching listeners and returns their errors.
	Emit(topicName string, payload any) <-chan error

	// EmitSync emits payload to matching listeners and waits for completion.
	EmitSync(topicName string, payload any) []error

	// GetTopic returns the topic registered as topicName.
	GetTopic(topicName string) (*Topic, error)

	// EnsureTopic returns the topic registered as topicName, creating it if needed.
	EnsureTopic(topicName string) *Topic

	// SetErrorHandler sets the handler used to rewrite listener errors.
	SetErrorHandler(func(Event, error) error)

	// SetIDGenerator sets the generator used for listener IDs.
	SetIDGenerator(func() string)

	// SetPool sets the pool used by Emit.
	SetPool(Pool)

	// SetErrChanBufferSize sets the buffer size used by Emit error channels.
	SetErrChanBufferSize(int)

	// Close releases the emitter's resources.
	Close() error
}
