# emitter

[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/emitter)](https://goreportcard.com/report/github.com/kaptinlin/emitter)
[![Go Version](https://img.shields.io/badge/Go-1.26.2+-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A thread-safe in-memory event emitter for Go with wildcard topics, listener priorities, and optional goroutine pooling

## Features

- **In-memory core**: Keep event routing local with no broker or storage dependency.
- **Wildcard topics**: Match one segment with `*` and zero or more segments with `**`.
- **Priority ordering**: Run higher-priority listeners before lower-priority listeners.
- **Sync and async delivery**: Use `EmitSync` for immediate results or `Emit` for channel-based delivery.
- **Panic recovery**: Recover listener panics and surface them as errors matching `ErrListenerPanic`.
- **Pool integration**: Route async work through `PondPool` or any custom `Pool` implementation.

## Installation

Requires Go 1.26.2 or newer.

```bash
go get github.com/kaptinlin/emitter
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

func main() {
	bus := emitter.NewMemoryEmitter()

	_, err := bus.On("user.created", func(evt emitter.Event) error {
		fmt.Printf("%s %v\n", evt.Topic(), evt.Payload())
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	errs := bus.EmitSync("user.created", "alice@example.com")
	if len(errs) > 0 {
		log.Fatalf("listener errors: %v", errs)
	}
}
```

## API Overview

| API | Purpose |
| --- | --- |
| `NewMemoryEmitter(opts...)` | Create an emitter with optional configuration. |
| `On(topic, listener, opts...)` | Register a listener and get back its listener ID. |
| `Off(topic, listenerID)` | Remove a listener by topic and ID. |
| `Emit(topic, payload)` | Emit asynchronously and receive listener errors on a channel. |
| `EmitSync(topic, payload)` | Emit synchronously and receive listener errors as a slice. |
| `EnsureTopic(name)` / `GetTopic(name)` | Create or fetch a concrete topic registry. |
| `Close()` | Clear topics and release the configured pool. |

Full API docs: [pkg.go.dev/github.com/kaptinlin/emitter](https://pkg.go.dev/github.com/kaptinlin/emitter)

## Topic Patterns

- `user.created` matches only `user.created`
- `user.*` matches `user.created` and `user.deleted`
- `order.**` matches `order.created` and `order.item.added`
- `event.**` does not match bare `event`

## Listener Priorities

Use `WithPriority` when listener order matters.
Higher priorities run first: `Highest`, `High`, `Normal`, `Low`, `Lowest`.

```go
_, _ = bus.On("order.created", validateOrder, emitter.WithPriority(emitter.Highest))
_, _ = bus.On("order.created", persistOrder, emitter.WithPriority(emitter.Normal))
_, _ = bus.On("order.created", auditOrder, emitter.WithPriority(emitter.Low))
```

## Configuration

| Option | Purpose |
| --- | --- |
| `WithErrorHandler(func(Event, error) error)` | Rewrite, wrap, or suppress listener errors. |
| `WithIDGenerator(func() string)` | Control listener ID generation. |
| `WithPool(Pool)` | Run `Emit` through a pool. |
| `WithErrChanBufferSize(int)` | Set the async error channel buffer size. |

## Error Handling

Listener-returned errors flow through the configured error handler.
Recovered listener panics are returned as `PanicError` values and match `ErrListenerPanic`.

```go
for err := range bus.Emit("user.created", "alice@example.com") {
	if err != nil {
		fmt.Println(err)
	}
}
```

## Examples

See [examples/README.md](examples/README.md) for runnable examples covering:

- basic usage
- wildcard subscriptions
- listener management
- custom error handling
- custom ID generation
- listener panic recovery
- goroutine pool integration

## Development

```bash
task test
task lint
task test-fuzz
```

For development guidelines, see [AGENTS.md](AGENTS.md).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is licensed under the MIT License. See [LICENSE.md](LICENSE.md).
