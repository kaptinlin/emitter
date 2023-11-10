# Emitter: A Go Event Emission Library

Emitter is a robust Go library that simplifies event management in applications. Offering a straightforward interface for event subscription and handling, it is designed for performance and thread safety.

## Features

- **In-Memory Management**: Host and manage events internally without external dependencies.
- **Listener Prioritization**: Specify invocation order for fine-grained control over event handling.
- **Concurrent Processing**: Utilize goroutines for handling events in parallel.
- **Wildcard Subscriptions**: Employ pattern matching for event subscriptions.
- **Customization**: Configure with custom handlers for errors, IDs, panics, and more.
- **Thread Safety**: Safely operate in concurrent environments.

## Installation

```sh
go get -u github.com/kaptinlin/emitter
```

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/kaptinlin/emitter"
)

func main() {
	e := emitter.NewMemoryEmitter()
	e.On("user.created", func(evt emitter.Event) error {
		fmt.Println("Event received:", evt.Topic())
		return nil
	})
	e.Emit("user.created", "Jane Doe")
}
```

## Configuration

Customize Emitter with a variety of options:

```go
e := emitter.NewMemoryEmitter(
	emitter.WithErrorHandler(customErrorHandler),
	emitter.WithIDGenerator(customIDGenerator),
	// More options...
)
```

### Options

| Option                                         | Description                                                  |
|------------------------------------------------|--------------------------------------------------------------|
| `WithPool(pool emitter.Pool)`                  | Assign a goroutine pool for concurrent event handling.       |
| `WithErrorHandler(handler func(emitter.Event, error) error)` | Set a custom error handler for the emitter that receives an event and an error. |
| `WithIDGenerator(generator func() string)`     | Define a function for generating unique listener IDs.        |
| `WithPanicHandler(handler func(interface{}))`  | Implement a panic recovery strategy.                         |

## Wildcard Event Subscription

Pattern-match event topics with wildcards:

- `*` - Matches a single segment.
- `**` - Matches multiple segments.

### Using Wildcards

```go
e := emitter.NewMemoryEmitter()
e.On("user.*", userEventListener)
e.On("order.**", orderEventListener)
e.On("**.completed", completionEventListener)
```

### Example

```go
e := emitter.NewMemoryEmitter()
e.On("user.*", func(evt emitter.Event) error {
	fmt.Printf("Event: %s, Payload: %+v\n", evt.Topic(), evt.Payload())
	return nil
})
e.Emit("user.signup", "John Doe")
// Use synchronization instead of sleep in production.
```

## Aborting Event Propagation

Stop event propagation using `SetAborted`:

```go
e := emitter.NewMemoryEmitter()
e.On("order.processed", func(evt emitter.Event) error {
	if /* condition fails */ false {
		evt.SetAborted(true)
	}
	return nil
}, emitter.WithPriority(emitter.High))
e.On("order.processed", func(evt emitter.Event) error {
	// This will not run if the event is aborted.
	return nil
}, emitter.WithPriority(emitter.Low))
e.Emit("order.processed", "Order data")
```

Abort event handling early based on custom logic.

## Examples

- [Managing Concurrency](#managing-concurrency-with-withpool)
- [Custom Error Handling](#custom-error-handling-with-witherrorhandler)
- [Listener Prioritization](#prioritizing-listeners-with-withpriority)
- [ID Generation](#generating-unique-ids-with-withidgenerator)
- [Panic Recovery](#handling-panics-gracefully-with-withpanichandler)

### Managing Concurrency with `WithPool`

Delegate concurrency management to a custom goroutine pool using the `WithPool` option:

```go
package main

import (
	"github.com/kaptinlin/emitter"
	"github.com/alitto/pond"
)

func main() {
	// Initialize a goroutine pool
	pool := emitter.NewPondPool(10, 1000) // 10 workers, queue size 1000

	// Set up the emitter with this pool
	e := emitter.NewMemoryEmitter(emitter.WithPool(pool))

	// Your emitter is now ready to handle events using the pool
}
```

This configuration employs 10 worker goroutines, optimizing task handling.

### Custom Error Handling with `WithErrorHandler`

Enhance error visibility by defining a custom error handler:

```go
package main

import (
	"log"
	"github.com/kaptinlin/emitter"
)

func main() {
	// Define a custom error handler that logs the event and the error
	customErrorHandler := func(event emitter.Event, err error) error {
		log.Printf("Error encountered during event '%s': %v, with payload: %v", event.Topic(), err, event.Payload())
		return nil  // Returning nil to indicate that the error has been handled
	}

	// Apply the custom error handler to the emitter
	e := emitter.NewMemoryEmitter(emitter.WithErrorHandler(customErrorHandler))

	// Your emitter will now log detailed errors encountered during event handling
}
```

With `logErrorHandler`, all errors are logged for review and action.

### Prioritizing Listeners with `WithPriority`

Control the invocation order of event listeners:

```go
package main

import (
	"fmt"
	"github.com/kaptinlin/emitter"
)

func main() {
	// Set up the emitter
	e := emitter.NewMemoryEmitter()

	// Define listeners with varying priorities
	normalPriorityListener := func(e emitter.Event) error {
		fmt.Println("Normal priority: Received", e.Topic())
		return nil
	}

	highPriorityListener := func(e emitter.Event) error {
		fmt.Println("High priority: Received", e.Topic())
		return nil
	}

	// Subscribe listeners with specified priorities
	e.On("user.created", normalPriorityListener) // Default is normal priority
	e.On("user.created", highPriorityListener, emitter.WithPriority(emitter.High))

	// Emit an event and observe the order of listener notification
	e.Emit("user.created", "User signup event")
}
```

Listeners with higher priority are notified first when an event occurs.

### Generating Unique IDs with `WithIDGenerator`

Implement custom ID generation for listener tracking:

```go
package main

import (
	"github.com/google/uuid"
	"github.com/kaptinlin/emitter"
)

func main() {
	// Custom ID generator using UUID v4
	uuidGenerator := func() string {
		return uuid.NewString()
	}

	// Initialize the emitter with the UUID generator
	e := emitter.NewMemoryEmitter(emitter.WithIDGenerator(uuidGenerator))

	// Listeners will now be registered with a unique UUID
}
```

Listeners are now identified by a unique UUID, providing better traceability.

### Handling Panics Gracefully with `WithPanicHandler`

Safeguard your application from unexpected panics during event handling:

```go
package main

import (
	"log"
	"github.com/kaptinlin/emitter"
)

func main() {
	// Define a panic handler that logs the occurrence
	logPanicHandler := func(p interface{}) {
		log.Printf("Panic recovered: %v", p)
		// Insert additional logic for panic recovery here
	}

	// Equip the emitter with the panic handler
	e := emitter.NewMemoryEmitter(emitter.WithPanicHandler(logPanicHandler))

	// Your emitter is now more resilient to panics
}
```

This handler ensures that panics are logged and managed without disrupting your service.

## Contributing

Contributions are welcome! Check out our [Contributing Guidelines](CONTRIBUTING.md) to get started.

## License

Emitter is licensed under the [MIT License](LICENSE.md). Feel free to use, modify, and distribute the code as you see fit.
