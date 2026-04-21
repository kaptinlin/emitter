package emitter

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

// MemoryEmitter is an in-memory [Emitter].
type MemoryEmitter struct {
	topics            sync.Map
	errorHandler      atomic.Pointer[func(Event, error) error]
	idGenerator       atomic.Pointer[func() string]
	pool              Pool
	closed            atomic.Bool
	errChanBufferSize atomic.Int32
}

// NewMemoryEmitter returns a [MemoryEmitter] configured with opts.
func NewMemoryEmitter(opts ...EmitterOption) *MemoryEmitter {
	m := &MemoryEmitter{}

	errorHandler := DefaultErrorHandler
	idGenerator := DefaultIDGenerator
	m.errorHandler.Store(&errorHandler)
	m.idGenerator.Store(&idGenerator)
	m.errChanBufferSize.Store(10)

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// On registers listener for topicName and returns its listener ID.
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

// Off removes the listener identified by listenerID from topicName.
func (m *MemoryEmitter) Off(topicName string, listenerID string) error {
	topic, err := m.GetTopic(topicName)
	if err != nil {
		return err
	}

	return topic.RemoveListener(listenerID)
}

// Emit emits payload to matching listeners and returns their errors.
func (m *MemoryEmitter) Emit(topicName string, payload any) <-chan error {
	errChan := make(chan error, m.errChanBufferSize.Load())

	if m.closed.Load() {
		closedErrChan := make(chan error, 1)
		closedErrChan <- ErrEmitterClosed
		close(closedErrChan)
		return closedErrChan
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

// EmitSync emits payload to matching listeners and waits for completion.
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

func (m *MemoryEmitter) handleEvents(topicName string, payload any, onError func(error)) {
	event := NewBaseEvent(topicName, payload)
	errorHandler := m.errorHandler.Load()
	m.topics.Range(func(key, value any) bool {
		topicPattern := key.(string)
		if !matchTopicPattern(topicPattern, topicName) {
			return true
		}

		topic := value.(*Topic)
		for _, err := range topic.Trigger(event) {
			err = (*errorHandler)(event, err)
			if err != nil {
				onError(err)
			}
		}
		return true
	})
}

// GetTopic returns the topic registered as topicName.
func (m *MemoryEmitter) GetTopic(topicName string) (*Topic, error) {
	topic, ok := m.topics.Load(topicName)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrTopicNotFound, topicName)
	}
	return topic.(*Topic), nil
}

// EnsureTopic returns the topic registered as topicName, creating it if needed.
func (m *MemoryEmitter) EnsureTopic(topicName string) *Topic {
	topic, _ := m.topics.LoadOrStore(topicName, NewTopic())
	return topic.(*Topic)
}

// SetErrorHandler sets the handler used to rewrite listener errors.
// A nil handler is ignored.
func (m *MemoryEmitter) SetErrorHandler(handler func(Event, error) error) {
	if handler == nil {
		return
	}
	m.errorHandler.Store(&handler)
}

// SetIDGenerator sets the generator used for listener IDs.
// A nil generator is ignored.
func (m *MemoryEmitter) SetIDGenerator(generator func() string) {
	if generator == nil {
		return
	}
	m.idGenerator.Store(&generator)
}

// SetPool sets the pool used by Emit.
func (m *MemoryEmitter) SetPool(p Pool) {
	m.pool = p
}

// SetErrChanBufferSize sets the buffer size used by Emit error channels.
func (m *MemoryEmitter) SetErrChanBufferSize(size int) {
	if size < 0 {
		size = 0
	}
	if size > math.MaxInt32 {
		size = math.MaxInt32
	}
	m.errChanBufferSize.Store(int32(size))
}

// Close releases the emitter's resources.
// Close returns ErrEmitterAlreadyClosed after the first call.
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
