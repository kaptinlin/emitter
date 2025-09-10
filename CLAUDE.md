# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **high-performance, thread-safe event emitter library** for Go 1.24+ that leverages modern Go features for optimal performance and reliability. The library is designed for high-concurrency scenarios with lock-free operations and efficient memory management.

## Common Development Commands

- **Build and test**: `make all` (runs lint and test)
- **Run tests**: `make test` (runs tests with race detection)
- **Run fuzz tests**: `make test-fuzz` (runs fuzz testing for 10 seconds)
- **Check shadowed variables**: `make vet-shadow` (runs go vet -shadow)
- **Lint code**: `make lint` (runs golangci-lint and mod tidy checks)
- **Clean**: `make clean` (removes ./bin directory)

The project uses golangci-lint with version pinned in `.golangci.version`. Tests are run with race detection enabled.

## Go Version and Modern Optimizations

This project targets **Go 1.24** and leverages cutting-edge Go features:

### Performance Optimizations
- **`atomic.Pointer`**: Lock-free access to handlers (error, panic, ID generators)
- **`slices.DeleteFunc`**: Efficient slice operations for listener management (Go 1.21+)
- **Zero-copy operations**: Minimal memory allocations in hot paths
- **Concurrent-safe design**: Optimized for high-throughput scenarios

### Testing Improvements
- **Parallel testing**: Concurrent test execution with `t.Parallel()` and subtests
- **Fuzz testing**: Automated edge case discovery for pattern matching
- **Proper synchronization**: Uses `sync.WaitGroup` and channels instead of `time.Sleep`
- **Race detection**: All tests run with `-race` flag enabled

### Code Quality
- **Type safety**: Strong typing with proper validation
- **Modern idioms**: Uses `any` instead of `interface{}`
- **Boundary validation**: Priority ranges are validated and clamped

## Architecture Overview

Core architectural components designed for high performance and thread safety:

### Core Interfaces and Types

- **Emitter interface** (`emitter.go`): Defines the contract for event management with methods for listener registration, event emission (sync/async), and configuration
- **MemoryEmitter** (`memory.go`): Thread-safe in-memory implementation using sync.Map for topics storage
- **Event interface** (`event.go`): Represents events with topic, payload, and abort functionality via BaseEvent
- **Topic** (`topics.go`): Manages listeners with priority-based ordering using sorted slices
- **Listener** (`listener.go`): Function type `func(Event) error` for event handlers

### Key Features

1. **Wildcard Topic Matching** (`utils.go`): Supports `*` (single segment) and `**` (multiple segments) patterns
2. **Priority System** (`priority.go`): Five levels (Lowest to Highest) with listeners executed in priority order
3. **Concurrency Management** (`pool.go`): Optional goroutine pool integration using github.com/alitto/pond
4. **Event Propagation Control**: Events can be aborted to stop further listener execution

### Configuration System

The emitter uses a functional options pattern (`options.go`) with these configurations:
- Custom error handlers
- ID generators for listener tracking  
- Panic handlers for recovery
- Goroutine pools for concurrent execution
- Error channel buffer sizes

### Thread Safety and Concurrency

- **MemoryEmitter**: Uses `sync.Map` for lock-free topic storage
- **Handler Storage**: `atomic.Pointer` for lock-free handler access
- **Topics**: `sync.RWMutex` for listener management with reader preference
- **Events**: `sync.RWMutex` for payload/abort state protection
- **Emitter State**: `atomic.Bool` for closed state management
- **Priority Validation**: Atomic boundary checking for listener priorities

## Dependencies

- `github.com/alitto/pond`: Goroutine pool management (optional)
- `github.com/google/uuid`: Used in examples for ID generation

## Testing Strategy

### Test Types
1. **Unit Tests**: Individual component testing with race detection
2. **Integration Tests**: End-to-end event flow testing
3. **Parallel Tests**: Concurrent execution validation
4. **Fuzz Tests**: Pattern matching edge case discovery
5. **Benchmark Tests**: Performance regression detection

### Test Commands
- `make test`: Run all tests with race detection
- `make test-fuzz`: Run fuzz tests for 10 seconds
- `make vet-shadow`: Check for variable shadowing

## Performance Characteristics

### Optimized Hot Paths
- Event emission: Lock-free topic lookup with `sync.Map`
- Handler access: Atomic pointer dereferencing
- Listener iteration: Pre-sorted slice traversal
- Pattern matching: Optimized string operations

### Memory Efficiency
- Minimal allocations in event processing
- Reused data structures where possible
- Efficient slice operations using Go 1.21+ features
- Zero-copy event payload handling

## Development Guidelines

### Code Style
- Follow Go 1.24+ conventions
- Use `any` instead of `interface{}`
- Prefer atomic operations over mutexes where possible
- Implement proper error handling and validation

### Performance Considerations
- Avoid allocations in hot paths
- Use atomic operations for frequently accessed data
- Prefer lock-free data structures
- Implement efficient algorithms for core operations

### Testing Requirements
- All new features must include tests
- Tests must pass with race detection enabled
- Consider adding fuzz tests for string/pattern operations
- Benchmark critical performance paths