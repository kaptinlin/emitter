package emitter

import (
	"context"
	"fmt"
	"reflect"
)

// Subscribe registers a typed listener on e. The callback receives the event
// payload pre-asserted to T. If a future emit on a matching topic carries a
// payload whose dynamic type is not T, the callback is not invoked and the
// listener returns an error wrapping [ErrPayloadType] which is joined into the
// Emit result. For interface payload types, a nil payload is delivered as T's
// zero value.
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
		payload, ok := payloadAs[T](ev.Payload())
		if !ok {
			return payloadTypeError[T](ev)
		}
		return fn(ctx, ev, payload)
	}, append(opts, withPayloadTypeFilter[T]())...)
}

func withPayloadTypeFilter[T any]() ListenerOption {
	return func(o *listenerOpts) {
		o.filter = func(ev Event) error {
			if _, ok := payloadAs[T](ev.Payload()); !ok {
				return payloadTypeError[T](ev)
			}
			return nil
		}
	}
}

func payloadAs[T any](payload any) (T, bool) {
	if value, ok := payload.(T); ok {
		return value, true
	}

	var zero T
	if payload == nil && reflect.TypeFor[T]().Kind() == reflect.Interface {
		return zero, true
	}
	return zero, false
}

func payloadTypeError[T any](ev Event) error {
	return fmt.Errorf("%w: topic %q expected %v, got %T", ErrPayloadType, ev.Topic(), reflect.TypeFor[T](), ev.Payload())
}

// Publish is sugar over [Emitter.Emit] with a typed payload.
// The payload is delivered as any to listeners; use [Subscribe] for the typed
// receive side.
func Publish[T any](ctx context.Context, e *Emitter, topic string, payload T) error {
	return e.Emit(ctx, topic, payload)
}
