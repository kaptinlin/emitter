package emitter

import "crypto/rand"

// EmitterOption configures a [MemoryEmitter].
type EmitterOption func(*MemoryEmitter)

// DefaultErrorHandler returns listener errors unchanged.
var DefaultErrorHandler = func(event Event, err error) error {
	return err
}

// DefaultIDGenerator returns listener IDs.
var DefaultIDGenerator = rand.Text

// WithErrorHandler sets the error handler for a [MemoryEmitter].
func WithErrorHandler(errHandler func(Event, error) error) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetErrorHandler(errHandler)
	}
}

// WithIDGenerator sets the listener ID generator for a [MemoryEmitter].
func WithIDGenerator(idGen func() string) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetIDGenerator(idGen)
	}
}

// WithPool sets the pool used by Emit on a [MemoryEmitter].
func WithPool(pool Pool) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetPool(pool)
	}
}

// WithErrChanBufferSize sets the buffer size used by Emit error channels.
func WithErrChanBufferSize(size int) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetErrChanBufferSize(size)
	}
}
