# Emitter: A Modern Go Event Emission Library

[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/emitter)](https://goreportcard.com/report/github.com/kaptinlin/emitter)
[![Go Version](https://img.shields.io/badge/Go-1.26+-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Emitter is a high-performance, thread-safe Go library for event management that leverages modern Go features (1.26). Built with atomic operations and optimized data structures for maximum performance and reliability.

## ✨ Features

- 🚀 **High Performance**: Built with `atomic.Pointer` and modern Go optimizations
- 🧠 **In-Memory Management**: Zero external dependencies, fully self-contained
- 📋 **Priority-Based Processing**: Fine-grained control over listener execution order
- ⚡ **Concurrent & Parallel**: Goroutine pools and lock-free operations
- 🎯 **Smart Pattern Matching**: Wildcard subscriptions with `*` and `**` support
- 🛠️ **Highly Customizable**: Custom error handlers, ID generators, panic recovery
- 🔒 **Thread-Safe**: Designed for high-concurrency environments
- 🧪 **Battle-Tested**: Comprehensive test suite including fuzz testing
- 📦 **Go 1.26 Ready**: Uses latest Go features for optimal performance

## 📦 Installation

**Requirements:** Go 1.26 or higher

```sh
go get -u github.com/kaptinlin/emitter
```

## 🚀 Quick Start

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

## ⚙️ Configuration

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
| `WithPanicHandler(handler func(any))`          | Implement a panic recovery strategy.                         |
| `WithErrChanBufferSize(size int)`              | Set the buffer size for error channels in async operations.  |

## 🎯 Wildcard Event Subscription

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

## ⛔ Aborting Event Propagation

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

## 📚 Examples

- [Managing Concurrency](#managing-concurrency-with-withpool)
- [Custom Error Handling](#custom-error-handling-with-witherrorhandler)
- [Listener Prioritization](#prioritizing-listeners-with-withpriority)
- [ID Generation](#generating-unique-ids-with-withidgenerator)
- [Panic Recovery](#handling-panics-gracefully-with-withpanichandler)

### ⚡ Managing Concurrency with `WithPool`

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

### 🚨 Custom Error Handling with `WithErrorHandler`

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

### 📊 Prioritizing Listeners with `WithPriority`

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

### 🆔 Generating Unique IDs with `WithIDGenerator`

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

### 🛡️ Handling Panics Gracefully with `WithPanicHandler`

Safeguard your application from unexpected panics during event handling:

```go
package main

import (
	"log"
	"github.com/kaptinlin/emitter"
)

func main() {
	// Define a panic handler that logs the occurrence
	logPanicHandler := func(p any) {
		log.Printf("Panic recovered: %v", p)
		// Insert additional logic for panic recovery here
	}

	// Equip the emitter with the panic handler
	e := emitter.NewMemoryEmitter(emitter.WithPanicHandler(logPanicHandler))

	// Your emitter is now more resilient to panics
}
```

This handler ensures that panics are logged and managed without disrupting your service.

## 🔧 Development Commands

```bash
# Run all tests with race detection
make test

# Run fuzz tests
make test-fuzz

# Lint code
make lint

# Check for shadowed variables
make vet-shadow

# Clean build artifacts
make clean
```

## ⚡ Performance Features

Built with modern Go 1.26 features for maximum performance:

- **Lock-Free Operations**: Uses `atomic.Bool` and `atomic.Pointer` for event state and handler storage
- **Efficient Slice Operations**: Leverages Go 1.26 `slices` package with binary search for O(log n) listener removal
- **Zero-Copy Event Handling**: Minimal memory allocations in hot paths
- **Optimized Pattern Matching**: Fast-path optimization for exact matches and simple wildcards
- **Built-in Functions**: Uses `min()`/`max()` for priority clamping (Go 1.21+)
- **Efficient String Building**: Uses `strings.Builder` with pre-allocation for ID generation
- **Concurrent-Safe**: Designed for high-throughput scenarios with proper synchronization

## 🧪 Testing

This library includes comprehensive testing:

- **Unit Tests**: Full coverage with race detection
- **Parallel Tests**: Concurrent test execution
- **Fuzz Testing**: Automated edge case discovery
- **Integration Tests**: Real-world scenario validation

```bash
# Run all tests
go test -race ./...

# Run fuzz tests
go test -fuzz=FuzzMatchTopicPattern -fuzztime=30s
```

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork** the repository
2. **Clone** your fork: `git clone https://github.com/yourusername/emitter.git`
3. **Create** a feature branch: `git checkout -b feature/amazing-feature`
4. **Make** your changes and add tests
5. **Run** tests: `make test`
6. **Commit** your changes: `git commit -m 'Add amazing feature'`
7. **Push** to the branch: `git push origin feature/amazing-feature`
8. **Open** a Pull Request

Please ensure your code:
- Follows Go conventions and passes `make lint`
- Includes appropriate tests
- Updates documentation as needed

## 📄 License

Emitter is licensed under the [MIT License](LICENSE.md). Feel free to use, modify, and distribute the code as you see fit.

---

**Made with ❤️ for the Go community**
