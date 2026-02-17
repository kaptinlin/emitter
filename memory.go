package emitter

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// MemoryEmitter is an in-memory, thread-safe implementation of the [Emitter] interface.
type MemoryEmitter struct {
	topics            sync.Map
	errorHandler      atomic.Pointer[func(Event, error) error]
	idGenerator       atomic.Pointer[func() string]
	panicHandler      atomic.Pointer[func(any)]
	pool              Pool
	closed            atomic.Bool
	errChanBufferSize atomic.Int32
}

// NewMemoryEmitter initializes a new MemoryEmitter with optional configuration options.
func NewMemoryEmitter(opts ...EmitterOption) *MemoryEmitter {
	m := &MemoryEmitter{}

	errorHandler := DefaultErrorHandler
	idGenerator := DefaultIDGenerator
	panicHandler := DefaultPanicHandler
	m.errorHandler.Store(&errorHandler)
	m.idGenerator.Store(&idGenerator)
	m.panicHandler.Store(&panicHandler)
	m.errChanBufferSize.Store(10)

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// On subscribes a listener to a topic with the given name.
// It returns a unique ID for the listener and an error, if any.
func (m *MemoryEmitter) On(topicName string, listener Listener, opts ...ListenerOption) (string, error) {
	if listener == nil {
		return "", ErrNilListener
	}

	if !isValidTopicName(topicName) {
		return "", ErrInvalidTopicName
	}

	topic := m.EnsureTopic(topicName)
	listenerID := (*m.idGenerator.Load())()
	topic.AddListener(listenerID, listener, opts...)
	return listenerID, nil
}

// Off unsubscribes a listener from a topic using the listener's unique ID.
func (m *MemoryEmitter) Off(topicName string, listenerID string) error {
	topic, err := m.GetTopic(topicName)
	if err != nil {
		return err
	}

	return topic.RemoveListener(listenerID)
}

// Emit asynchronously dispatches an event to all subscribers of the topic.
// It returns a channel that will receive any errors encountered during event handling.
func (m *MemoryEmitter) Emit(topicName string, payload any) <-chan error {
	errChan := make(chan error, m.errChanBufferSize.Load())

	if m.closed.Load() {
		errChan <- ErrEmitterClosed
		close(errChan)
		return errChan
	}

	task := func() {
		defer close(errChan)
		m.handleEvents(topicName, payload, func(err error) {
			errChan <- err
		})
	}

	if m.pool != nil {
		m.pool.Submit(task)
	} else {
		go task()
	}

	return errChan
}

// EmitSync dispatches an event synchronously to all subscribers of the topic.
// This method blocks until all listeners have been notified.
func (m *MemoryEmitter) EmitSync(topicName string, payload any) []error {
	if m.closed.Load() {
		return []error{ErrEmitterClosed}
	}

	var errs []error
	m.handleEvents(topicName, payload, func(err error) {
		errs = append(errs, err)
	})
	return errs
}

// handleEvents processes an event and notifies all matching listeners,
// with error handling and panic recovery.
func (m *MemoryEmitter) handleEvents(topicName string, payload any, onError func(error)) {
	defer func() {
		if r := recover(); r != nil {
			if handler := m.panicHandler.Load(); handler != nil {
				(*handler)(r)
			}
		}
	}()

	event := NewBaseEvent(topicName, payload)
	errorHandler := m.errorHandler.Load()
	m.topics.Range(func(key, value any) bool {
		topicPattern := key.(string)
		if !matchTopicPattern(topicPattern, topicName) {
			return true
		}

		topic := value.(*Topic)
		for _, err := range topic.Trigger(event) {
			if errorHandler != nil {
				err = (*errorHandler)(event, err)
			}
			if err != nil {
				onError(err)
			}
		}
		return true
	})
}

// GetTopic retrieves a topic by name.
func (m *MemoryEmitter) GetTopic(topicName string) (*Topic, error) {
	topic, ok := m.topics.Load(topicName)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrTopicNotFound, topicName)
	}
	return topic.(*Topic), nil
}

// EnsureTopic retrieves an existing topic or creates a new one.
func (m *MemoryEmitter) EnsureTopic(topicName string) *Topic {
	topic, _ := m.topics.LoadOrStore(topicName, NewTopic())
	return topic.(*Topic)
}

// SetErrorHandler assigns a custom error handler for the Emitter.
// A nil handler is ignored; the previous handler remains active.
func (m *MemoryEmitter) SetErrorHandler(handler func(Event, error) error) {
	if handler != nil {
		m.errorHandler.Store(&handler)
	}
}

// SetIDGenerator assigns a custom ID generator for new listeners.
// A nil generator is ignored; the previous generator remains active.
func (m *MemoryEmitter) SetIDGenerator(generator func() string) {
	if generator != nil {
		m.idGenerator.Store(&generator)
	}
}

// SetPool sets a custom goroutine pool for managing concurrent event handling.
func (m *MemoryEmitter) SetPool(p Pool) {
	m.pool = p
}

// SetPanicHandler sets a function that will be called when a panic occurs during event handling.
// A nil handler is ignored; the previous handler remains active.
func (m *MemoryEmitter) SetPanicHandler(handler PanicHandler) {
	if handler != nil {
		fn := func(v any) { handler(v) }
		m.panicHandler.Store(&fn)
	}
}

// SetErrChanBufferSize sets the size of the buffered channel for errors returned by [MemoryEmitter.Emit].
func (m *MemoryEmitter) SetErrChanBufferSize(size int) {
	m.errChanBufferSize.Store(int32(size))
}

// Close terminates the emitter and releases resources.
// Calling Close on an already closed emitter returns an error.
func (m *MemoryEmitter) Close() error {
	if !m.closed.CompareAndSwap(false, true) {
		return ErrEmitterAlreadyClosed
	}

	m.topics.Clear()

	if m.pool != nil {
		m.pool.Release()
	}

	return nil
}
