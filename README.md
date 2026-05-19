# emitter

[![Go Reference](https://pkg.go.dev/badge/github.com/kaptinlin/emitter.svg)](https://pkg.go.dev/github.com/kaptinlin/emitter)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/emitter)](https://goreportcard.com/report/github.com/kaptinlin/emitter)
[![Go Module](https://img.shields.io/badge/Go-module-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An in-memory pub/sub primitive for Go. Synchronous, ordered, panic-safe — and small enough to read in a sitting.

## Why

Most "event bus" libraries grow into protocol shims, retry machinery, or transport layers. `emitter` does one thing: deliver events in-process, in order, with predictable failure modes.

- **One emitter, many topics.** No generic-typed bus that pins every event to a single payload type.
- **Synchronous dispatch.** Listeners run in priority order; `Emit` returns once they all finish.
- **Wildcard subscriptions.** `*` matches one segment, `**` matches zero or more.
- **Panic recovery.** Listener panics surface as errors that wrap `ErrListenerPanic`.
- **Zero third-party dependencies.** Stdlib only in the core package.

For high-throughput fire-and-forget dispatch, an optional sibling module ([`emitter/pool`](pool/)) adds a bounded asynchronous dispatcher.

### When not to use it

`emitter` is an in-process primitive. Reach for a broker (NATS, Kafka, Redis Streams) when you need persistence, cross-process delivery, retries, or replay. Reach for `chan` when one producer talks to one consumer with backpressure. Reach for `errgroup` / `sync.WaitGroup` when you're orchestrating goroutines, not events.

## Install

```bash
go get github.com/kaptinlin/emitter
```

Requires the Go version declared in `go.mod`.

## Quick start

```go
package main

import (
    "context"
    "fmt"

    "github.com/kaptinlin/emitter"
)

func main() {
    e := emitter.New()
    defer e.Close()

    _, _ = e.On("user.created", func(_ context.Context, ev emitter.Event) error {
        fmt.Printf("%s %v\n", ev.Topic(), ev.Payload())
        return nil
    })

    _ = e.Emit(context.Background(), "user.created", "alice@example.com")
}
```

More runnable examples in [`examples/`](examples/).

## Concepts

| Type | Purpose |
| --- | --- |
| `Emitter` | Owns topic routing and dispatch. Construct with `New`. |
| `Listener` | `func(ctx context.Context, ev Event) error`. |
| `Event` | Read-only view of the in-flight emission: `Topic()`, `Payload()`, `Stop()`. |
| `Subscription` | Handle returned by `On`; call `Cancel()` to remove the listener. |
| `Priority` | Plain `int`. Higher runs first. Sentinels: `Lowest`, `Low`, `Normal`, `High`, `Highest`. |

## Topic grammar

```ebnf
topic    := segment ('.' segment)*
segment  := name | wildcard
name     := [a-zA-Z0-9_-]+
wildcard := '*' | '**'
```

- `*` matches exactly one segment.
- `**` matches zero or more segments.
- Wildcards are valid in subscription patterns only — emit topics must be literal.

```go
_, _ = e.On("user.*",    handler) // user.created, user.deleted
_, _ = e.On("metric.**", handler) // metric, metric.cpu, metric.cpu.idle
_, _ = e.On("**",        handler) // every topic
```

## Listener options

```go
_, _ = e.On("evt", handler, emitter.WithPriority(emitter.High))
_, _ = e.On("evt", handler, emitter.Once())
```

- `WithPriority(p)` overrides the default `Normal`. Higher values run first; equal priorities run in registration order.
- `Once()` removes the listener after its first invocation. Concurrent emits fire it at most once.

## Stopping dispatch

`Event.Stop()` halts remaining listeners for the current emission. It takes effect after the current listener returns:

```go
_, _ = e.On("evt", func(_ context.Context, ev emitter.Event) error {
    if shouldShortCircuit() {
        ev.Stop()
    }
    return nil
}, emitter.WithPriority(emitter.High))
```

## Errors

```go
err := e.Emit(ctx, "evt", payload)
```

- All listener errors are joined via `errors.Join`. Use `errors.Is` / `errors.As` to inspect.
- A panicking listener returns a `*PanicError` wrapping `ErrListenerPanic`. The raw panic value is on `pe.Value`; if it was an `error`, it appears in the unwrap chain too.
- Cancelling `ctx` mid-dispatch surfaces `ctx.Err()` in the joined result and skips remaining listeners.
- Sentinels: `ErrEmitterClosed`, `ErrInvalidTopicName`, `ErrNilListener`, `ErrListenerPanic`, `ErrPayloadType`.

## Typed helpers

For payloads with a known type, the generic helpers add static checking at the boundary without coupling the emitter to a single type:

```go
type OrderShipped struct{ ID string }

_, _ = emitter.Subscribe(e, "order.shipped",
    func(_ context.Context, _ emitter.Event, p OrderShipped) error {
        fmt.Println(p.ID)
        return nil
    })

_ = emitter.Publish(context.Background(), e, "order.shipped", OrderShipped{ID: "ord-7"})
```

If a published payload's dynamic type does not match `T`, the listener is skipped and the emit returns an error wrapping `ErrPayloadType`.

## Asynchronous dispatch (optional)

The core dispatches synchronously. For bounded asynchronous emission with backpressure, import the sibling module:

```bash
go get github.com/kaptinlin/emitter/pool
```

```go
import (
    "context"

    "github.com/kaptinlin/emitter"
    "github.com/kaptinlin/emitter/pool"
)

p := pool.New(64, 1024) // 64 workers, queue cap 1024
defer p.Close()

if err := p.Submit(ctx, e, "user.created", payload); err != nil {
    // errors.Is(err, pool.ErrPoolFull) under saturation
}
```

The pool's underlying engine is an implementation detail; only `Pool`, `New`, `Submit`, `Close`, and `ErrPoolFull` are part of the API.

## Concurrency

- All public methods on `*Emitter` are safe for concurrent use.
- Each emission takes a snapshot of its listener set, so listeners may freely call `On` / `Cancel` on the same emitter without deadlocking.
- `Subscription.Cancel` and `Emitter.Close` are both idempotent.

## License

MIT — see [LICENSE.md](LICENSE.md).
