package emitter

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// EmitterOption defines a function type for Emitter configuration options.
type EmitterOption func(Emitter)

var DefaultErrorHandler = func(event Event, err error) error {
	return err
}

// DefaultIDGenerator generates a unique identifier using a combination of the current time
// and random bytes, encoded in hexadecimal.
var DefaultIDGenerator = func() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err)
	}

	var b strings.Builder
	b.Grow(32 + 16) // Pre-allocate: 32 hex chars + 16 timestamp chars
	b.WriteString(hex.EncodeToString(randomBytes))
	b.WriteString(fmt.Sprintf("%x", timestamp))
	return b.String()
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
