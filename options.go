package emitter

import (
	"crypto/rand"
	"fmt"
)

// EmitterOption defines a function type for Emitter configuration options.
type EmitterOption func(Emitter)

// DefaultErrorHandler is the default error handler that returns the error as-is.
var DefaultErrorHandler = func(event Event, err error) error {
	return err
}

// DefaultIDGenerator generates a unique identifier using cryptographically secure random text.
// Uses crypto/rand.Text (Go 1.24+) which returns a base32-encoded string
// with at least 128 bits of randomness.
var DefaultIDGenerator = func() string {
	return rand.Text()
}

// DefaultPanicHandler is the default panic handler that prints the panic value.
var DefaultPanicHandler = func(p any) {
	fmt.Printf("Panic occurred: %v\n", p)
}

// WithErrorHandler sets a custom error handler for an Emitter.
func WithErrorHandler(errHandler func(Event, error) error) EmitterOption {
	return func(m Emitter) {
		m.SetErrorHandler(errHandler)
	}
}

// WithIDGenerator sets a custom ID generator for an Emitter.
func WithIDGenerator(idGen func() string) EmitterOption {
	return func(m Emitter) {
		m.SetIDGenerator(idGen)
	}
}

// WithPool sets a custom pool for an Emitter.
func WithPool(pool Pool) EmitterOption {
	return func(m Emitter) {
		m.SetPool(pool)
	}
}

// PanicHandler is a function type that handles panics during event processing.
type PanicHandler func(any)

// WithPanicHandler sets a custom panic handler for an Emitter.
func WithPanicHandler(panicHandler PanicHandler) EmitterOption {
	return func(m Emitter) {
		m.SetPanicHandler(panicHandler)
	}
}

// WithErrChanBufferSize sets the buffer size for the error channel
// returned by [Emitter.Emit].
func WithErrChanBufferSize(size int) EmitterOption {
	return func(m Emitter) {
		m.SetErrChanBufferSize(size)
	}
}
