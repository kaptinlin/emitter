package emitter

import (
	"crypto/rand"
	"fmt"
)

// EmitterOption defines a function type for Emitter configuration options.
type EmitterOption func(Emitter)

var DefaultErrorHandler = func(event Event, err error) error {
	return err
}

// DefaultIDGenerator generates a unique identifier using cryptographically secure random text.
// This is more efficient than hex encoding and timestamp concatenation.
// Uses crypto/rand.Text (Go 1.24+) which returns a base32-encoded string with at least 128 bits of randomness.
var DefaultIDGenerator = func() string {
	// Generate cryptographically secure random text (base32-encoded with at least 128 bits of randomness)
	return rand.Text()
}

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

type PanicHandler func(any)

func WithPanicHandler(panicHandler PanicHandler) EmitterOption {
	return func(m Emitter) {
		m.SetPanicHandler(panicHandler)
	}
}

func WithErrChanBufferSize(size int) EmitterOption {
	return func(m Emitter) {
		m.SetErrChanBufferSize(size)
	}
}
