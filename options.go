package emitter

import (
	"crypto/rand"
	"fmt"
)

// EmitterOption defines a function type for MemoryEmitter configuration options.
type EmitterOption func(*MemoryEmitter)

// DefaultErrorHandler returns the error as-is.
var DefaultErrorHandler = func(event Event, err error) error {
	return err
}

// DefaultIDGenerator generates a unique identifier using crypto/rand.Text (Go 1.24+),
// returning a base32-encoded string with at least 128 bits of randomness.
var DefaultIDGenerator = func() string {
	return rand.Text()
}

// DefaultPanicHandler prints the panic value to stdout.
var DefaultPanicHandler = func(p any) {
	fmt.Printf("Panic occurred: %v\n", p)
}

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

// PanicHandler is a function type that handles panics during event processing.
type PanicHandler func(any)

// WithPanicHandler sets a custom panic handler for a MemoryEmitter.
func WithPanicHandler(panicHandler PanicHandler) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetPanicHandler(panicHandler)
	}
}

// WithErrChanBufferSize sets the buffer size for the error channel
// returned by [MemoryEmitter.Emit].
func WithErrChanBufferSize(size int) EmitterOption {
	return func(m *MemoryEmitter) {
		m.SetErrChanBufferSize(size)
	}
}
