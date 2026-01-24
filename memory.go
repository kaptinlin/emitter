package emitter

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// MemoryEmitter is an in-memory implementation of the Emitter interface. It provides
// facilities for adding and removing listeners, emitting events, and configuring
// the behavior of event handling within the application.
type MemoryEmitter struct {
	topics            sync.Map                                 // Stores topics with concurrent access support.
	errorHandler      atomic.Pointer[func(Event, error) error] // Handles errors that occur during event handling.
	idGenerator       atomic.Pointer[func() string]            // Generates unique IDs for listeners.
	panicHandler      atomic.Pointer[func(any)]                // Handles panics that occur during event handling.
	Pool              Pool                                     // Manages concurrent execution of event handlers.
	closed            atomic.Bool                              // Indicates whether the emitter is closed.
	errChanBufferSize int                                      // Size of the buffer for the error channel in Emit.
}

// NewMemoryEmitter initializes a new MemoryEmitter with optional configuration options.
// Default configurations are applied, which can be overridden by the provided options.
func NewMemoryEmitter(opts ...EmitterOption) *MemoryEmitter {
	m := &MemoryEmitter{
		topics:            sync.Map{},
		errChanBufferSize: 10,
	}

	// Set default handlers using atomic pointers
	errorHandler := DefaultErrorHandler
	idGenerator := DefaultIDGenerator
	panicHandler := DefaultPanicHandler
	m.errorHandler.Store(&errorHandler)
	m.idGenerator.Store(&idGenerator)
	m.panicHandler.Store(&panicHandler)
	m.closed.Store(false)

	// Apply each provided option to the emitter to configure it.
	for _, opt := range opts {
		opt(m)
	}

	return m
}

// On subscribes a listener to a topic with the given name. Listener options can be specified
// to configure the listener's behavior. It returns a unique ID for the listener and an error, if any.
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

// Off unsubscribes a listener from a topic using the listener's unique ID. It returns
// an error if the listener cannot be found or if there is a problem with unsubscribing.
func (m *MemoryEmitter) Off(topicName string, listenerID string) error {
	topic, err := m.GetTopic(topicName)
	if err != nil {
		return err
	}

	return topic.RemoveListener(listenerID)
}

// Emit asynchronously dispatches an event to all the subscribers of the event's topic.
// It returns a channel that will receive any errors encountered during event handling.
// Emit asynchronously dispatches an event to all the subscribers of the event's topic.
// It returns a channel that will receive any errors encountered during event handling.
func (m *MemoryEmitter) Emit(eventName string, payload any) <-chan error {
	errChan := make(chan error, m.errChanBufferSize)

	// Before starting new goroutine, check if Emitter is closed
	if m.closed.Load() {
		errChan <- ErrEmitterClosed
		close(errChan)
		return errChan
	}

	task := func() {
		defer close(errChan)
		m.handleEvents(eventName, payload, func(err error) {
			errChan <- err
		})
	}

	if m.Pool != nil {
		m.Pool.Submit(task)
	} else {
		go task()
	}

	return errChan
}

// EmitSync dispatches an event synchronously to all subscribers of the event's topic and
// collects any errors that occurred. This method will block until all notifications are completed.
func (m *MemoryEmitter) EmitSync(eventName string, payload any) []error {
	if m.closed.Load() {
		return []error{ErrEmitterClosed}
	}

	var errs []error
	m.handleEvents(eventName, payload, func(err error) {
		errs = append(errs, err)
	})
	return errs
}

// handleEvents is an internal method that processes an event and notifies all
// registered listeners. It takes care of error handling and panic recovery.
func (m *MemoryEmitter) handleEvents(topicName string, payload any, errorHandler func(error)) {
	defer func() {
		if r := recover(); r != nil {
			if handler := m.panicHandler.Load(); handler != nil {
				(*handler)(r)
			}
		}
	}()

	event := NewBaseEvent(topicName, payload)
	m.topics.Range(func(key, value any) bool {
		topicPattern := key.(string)
		if matchTopicPattern(topicPattern, topicName) {
			topic := value.(*Topic)
			topicErrors := topic.Trigger(event)
			for _, err := range topicErrors {
				if handler := m.errorHandler.Load(); handler != nil {
					err = (*handler)(event, err)
				}
				if err != nil {
					errorHandler(err)
				}
			}
		}
		return true
	})
}

// GetTopic retrieves a topic by its name. If the topic does not exist, it returns an error.
func (m *MemoryEmitter) GetTopic(topicName string) (*Topic, error) {
	topic, ok := m.topics.Load(topicName)
	if !ok {
		return nil, fmt.Errorf("%w: unable to find topic '%s'", ErrTopicNotFound, topicName)
	}
	return topic.(*Topic), nil
}

// EnsureTopic retrieves or creates a new topic by its name. If the topic does not
// exist, it is created and returned. This ensures that a topic is always available.
func (m *MemoryEmitter) EnsureTopic(topicName string) *Topic {
	topic, _ := m.topics.LoadOrStore(topicName, NewTopic())
	return topic.(*Topic)
}

func (m *MemoryEmitter) SetErrorHandler(handler func(Event, error) error) {
	if handler != nil {
		m.errorHandler.Store(&handler)
	}
}

func (m *MemoryEmitter) SetIDGenerator(generator func() string) {
	if generator != nil {
		m.idGenerator.Store(&generator)
	}
}

func (m *MemoryEmitter) SetPool(pool Pool) {
	m.Pool = pool
}

func (m *MemoryEmitter) SetPanicHandler(panicHandler PanicHandler) {
	if panicHandler != nil {
		handler := func(p any) { panicHandler(p) }
		m.panicHandler.Store(&handler)
	}
}

func (m *MemoryEmitter) SetErrChanBufferSize(size int) {
	m.errChanBufferSize = size
}

// Close terminates the emitter, ensuring all pending events are processed. It performs cleanup
// and releases resources. Calling Close on an already closed emitter will result in an error.
func (m *MemoryEmitter) Close() error {
	if m.closed.Load() {
		return ErrEmitterAlreadyClosed
	}

	m.closed.Store(true)

	// Perform cleanup operations
	m.topics.Range(func(key, value any) bool {
		m.topics.Delete(key)
		return true
	})

	if m.Pool != nil {
		m.Pool.Release()
	}

	return nil
}
