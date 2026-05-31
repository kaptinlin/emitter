package emitter

import (
	"context"
	"fmt"
)

// Subscribe registers a typed listener on e. The callback receives the event
// payload pre-asserted to T. If a future emit on a matching topic carries a
// payload whose dynamic type is not T, the callback is not invoked and the
// listener returns an error wrapping [ErrPayloadType] which is joined into the
// Emit result.
//
// Subscribe is sugar over [Emitter.On]; it does not change dispatch semantics.
func Subscribe[T any](
	e *Emitter,
	pattern string,
	fn func(ctx context.Context, ev Event, payload T) error,
	opts ...ListenerOption,
) (Subscription, error) {
	if fn == nil {
		return nil, ErrNilListener
	}
	return e.On(pattern, func(ctx context.Context, ev Event) error {
		payload, ok := ev.Payload().(T)
		if !ok {
			return payloadTypeError[T](ev)
		}
		return fn(ctx, ev, payload)
	}, append(opts, withPayloadTypeFilter[T]())...)
}

func withPayloadTypeFilter[T any]() ListenerOption {
	return func(o *listenerOpts) {
		o.filter = func(ev Event) error {
			if _, ok := ev.Payload().(T); !ok {
				return payloadTypeError[T](ev)
			}
			return nil
		}
	}
}

func payloadTypeError[T any](ev Event) error {
	var zero T
	return fmt.Errorf("%w: topic %q expected %T, got %T", ErrPayloadType, ev.Topic(), zero, ev.Payload())
}

// Publish is sugar over [Emitter.Emit] with a typed payload.
// The payload is delivered as any to listeners; use [Subscribe] for the typed
// receive side.
func Publish[T any](ctx context.Context, e *Emitter, topic string, payload T) error {
	return e.Emit(ctx, topic, payload)
}
