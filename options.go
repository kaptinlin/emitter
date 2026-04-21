package emitter

import "crypto/rand"

// EmitterOption defines a function type for MemoryEmitter configuration options.
type EmitterOption func(*MemoryEmitter)

// DefaultErrorHandler returns the error as-is.
var DefaultErrorHandler = func(event Event, err error) error {
	return err
}

// DefaultIDGenerator generates a unique identifier using crypto/rand.Text (Go 1.24+),
// returning a base32-encoded string with at least 128 bits of randomness.
var DefaultIDGenerator = rand.Text

// WithErrorHandler sets a custom error handler for a MemoryEmitter.
func WithErrorHandler(errHandler func(Event, error) error) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetErrorHandler(errHandler)
	}
}

// WithIDGenerator sets a custom ID generator for a MemoryEmitter.
func WithIDGenerator(idGen func() string) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetIDGenerator(idGen)
	}
}

// WithPool sets a custom pool for a MemoryEmitter.
func WithPool(pool Pool) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetPool(pool)
	}
}

// WithErrChanBufferSize sets the buffer size for the error channel
// returned by [MemoryEmitter.Emit].
func WithErrChanBufferSize(size int) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetErrChanBufferSize(size)
	}
}
