// Package emitter is an in-memory pub/sub primitive.
//
// # Topic grammar (EBNF)
//
//	topic    := segment ('.' segment)*
//	segment  := name | wildcard
//	name     := [a-zA-Z0-9_-]+
//	wildcard := '*' | '**'
//
// '*' matches exactly one segment. '**' matches zero or more segments;
// for example, "event.**" matches "event", "event.x", and "event.x.y".
//
// Listeners run synchronously in priority order (high to low) within a topic.
// Emit returns errors.Join of all listener errors. Listener panics are
// wrapped as PanicError; check with errors.Is(err, ErrListenerPanic).
//
// Context cancellation stops further listener invocation within an emit.
package emitter
