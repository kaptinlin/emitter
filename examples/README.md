# Emitter Examples

This directory contains practical examples demonstrating the key features of the emitter package.

## Quick Start

Run any example:
```bash
cd examples/<example-name>
go run main.go
```

## Available Examples

### 1. Basic Usage
**Directory**: `basic_usage/`

Learn the fundamentals:
- Creating an emitter instance
- Subscribing listeners to topics
- Emitting events synchronously vs asynchronously
- Handling multiple listeners on the same topic
- Proper cleanup with `Close()`

**Run**: `cd basic_usage && go run main.go`

---

### 2. Wildcard Patterns
**Directory**: `wildcards/`

Master pattern matching:
- Single-level wildcard (`*`) - matches exactly one segment
- Multi-level wildcard (`**`) - matches zero or more segments
- Combined patterns (`**.error`)
- Multiple wildcard listeners on overlapping patterns

**Examples**:
- `order.*` matches `order.created`, `order.updated` but NOT `order.item.added`
- `system.**` matches `system.cpu`, `system.cpu.high`, `system.disk.full.warning`

**Run**: `cd wildcards && go run main.go`

---

### 3. Listener Management
**Directory**: `listener_management/`

Control listener lifecycle:
- Adding and removing listeners dynamically
- Priority-based execution (`Highest`, `High`, `Normal`, `Low`, `Lowest`)
- Event abortion with `SetAborted()` to stop propagation
- Handling multiple priorities on the same topic

**Run**: `cd listener_management && go run main.go`

---

### 4. Custom Error Handling
**Directory**: `custom_error_handling/`

Implement robust error handling:
- Custom error handler configuration
- Logging errors with context
- Error swallowing vs propagation strategies
- Integration with error tracking services

**Run**: `cd custom_error_handling && go run main.go`

---

### 5. Custom ID Generator
**Directory**: `custom_id_generator/`

Customize listener identification:
- Using UUID-based ID generation
- Custom ID formats for debugging
- Tracking listeners with meaningful IDs

**Run**: `cd custom_id_generator && go run main.go`

---

### 6. Custom Panic Handling
**Directory**: `custom_panic_handing/`

Gracefully handle panics:
- Custom panic recovery handlers
- Logging panic information
- Preventing entire application crashes
- Best practices for panic handling in listeners

**Run**: `cd custom_panic_handing && go run main.go`

---

### 7. Order Processing (Real-world Example)
**Directory**: `order_processing/`

Complete workflow demonstration:
- Multi-stage order processing pipeline
- Validation with high priority
- Payment processing
- Email notifications
- Event abortion on validation failure

**Run**: `cd order_processing && go run main.go`

---

### 8. Goroutine Pool Integration
**Directory**: `pool/`

Optimize concurrency:
- Using goroutine pools to limit concurrent execution
- Integration with `github.com/alitto/pond`
- Performance benefits of pooling
- Resource management

**Run**: `cd pool && go run main.go`

---

## Example Coverage Map

| Feature | Example |
|---------|---------|
| Basic event emission | `basic_usage` |
| Sync vs Async emission | `basic_usage` |
| Single wildcard (`*`) | `wildcards` |
| Multi-level wildcard (`**`) | `wildcards` |
| Adding listeners | `listener_management`, `basic_usage` |
| Removing listeners | `listener_management` |
| Priority execution | `listener_management`, `order_processing` |
| Event abortion | `listener_management`, `order_processing` |
| Custom error handling | `custom_error_handling` |
| Custom ID generation | `custom_id_generator` |
| Panic recovery | `custom_panic_handing` |
| Goroutine pooling | `pool` |

## Learning Path

**Recommended order for beginners**:
1. `basic_usage` - Start here to understand core concepts
2. `wildcards` - Learn flexible topic matching
3. `listener_management` - Master listener lifecycle and priorities
4. `order_processing` - See a complete real-world workflow
5. `custom_error_handling` - Implement robust error handling
6. `pool` - Optimize for production use

**Advanced features**:
- `custom_id_generator` - Customize listener tracking
- `custom_panic_handing` - Handle edge cases gracefully

## Common Patterns

### Pattern 1: Validation Pipeline
```go
// High priority validator
e.On("data.process", validator, emitter.WithPriority(emitter.Highest))

// Normal priority processor
e.On("data.process", processor, emitter.WithPriority(emitter.Normal))

// Low priority logger
e.On("data.process", logger, emitter.WithPriority(emitter.Low))
```

### Pattern 2: Wildcard Monitoring
```go
// Monitor all errors across the system
e.On("**.error", errorMonitor)

// Monitor specific subsystem
e.On("api.**", apiMonitor)
```

### Pattern 3: Dynamic Listener Management
```go
// Add listener
id, _ := e.On("topic", listener)

// Store ID for later removal
listenerRegistry[id] = listener

// Remove when no longer needed
e.Off("topic", id)
```

## Testing Examples

All examples can be used as integration tests. They demonstrate expected behavior and serve as living documentation.

## Performance Considerations

- **Sync vs Async**: Use `EmitSync()` when you need immediate error handling; use `Emit()` for fire-and-forget scenarios
- **Wildcards**: More specific patterns are faster than broad wildcards like `**`
- **Priorities**: Higher priority listeners execute first but add minimal overhead
- **Pooling**: Use goroutine pools for high-throughput applications (see `pool` example)

## Contributing

When adding new examples:
1. Create a new directory under `examples/`
2. Include a complete `main.go` with clear comments
3. Update this README with a description
4. Ensure the example runs without external dependencies (except the emitter package)

## Questions?

See the main package documentation: https://pkg.go.dev/github.com/kaptinlin/emitter
