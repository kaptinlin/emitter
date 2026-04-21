// Package emitter provides an in-memory, thread-safe event emitter with
// wildcard topics and priority-ordered listeners.
//
// Emit and EmitSync recover listener panics and return them as errors matching
// ErrListenerPanic. The package has no Must-style APIs and its public runtime
// operations are intended to return errors instead of panicking.
package emitter
