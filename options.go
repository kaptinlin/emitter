package emitter

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(randomBytes) + fmt.Sprintf("%x", timestamp)
}

var DefaultPanicHandler = func(p interface{}) {
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

type PanicHandler func(interface{})

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
