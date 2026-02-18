# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**emitter** is a high-performance, thread-safe event emitter library for Go 1.26. Built with lock-free atomic operations and optimized data structures for high-concurrency scenarios with minimal memory allocations.

- **Module**: `github.com/kaptinlin/emitter`
- **Go Version**: 1.26
- **License**: MIT

## Commands

```bash
# Run all checks (lint + test)
make all

# Run tests with race detection
task test

# Run fuzz tests (10 seconds)
task test-fuzz

# Check for shadowed variables
make vet-shadow

# Lint code (golangci-lint + mod tidy check)
task lint

# Clean build artifacts
task clean
```

The project uses golangci-lint v2 with version pinned in `.golangci.version`. All tests run with race detection enabled.

## Architecture

### Core Components

```
emitter/
├── emitter.go          # Emitter interface definition
├── memory.go           # MemoryEmitter: thread-safe in-memory implementation
├── event.go            # Event interface and BaseEvent implementation
├── topics.go           # Topic: priority-sorted listener management
├── listener.go         # Listener function type
├── options.go          # Functional options for configuration
├── priority.go         # Priority levels (Lowest to Highest)
├── pool.go             # Goroutine pool integration (optional)
├── utils.go            # Wildcard pattern matching
└── errors.go           # Error definitions
```

### Key Types and Interfaces

**Emitter Interface** (`emitter.go`):
- `On(topic, listener, opts...)` - Register listener with optional priority
- `Off(topic, listenerID)` - Deregister listener by ID
- `Emit(topic, payload)` - Async event emission (returns error channel)
- `EmitSync(topic, payload)` - Sync event emission (blocks until complete)
- `GetTopic(name)` / `EnsureTopic(name)` - Topic management
- `Close()` - Graceful shutdown

**MemoryEmitter** (`memory.go`):
- Uses `sync.Map` for lock-free topic storage
- `atomic.Pointer` for handlers (error, panic, ID generator)
- `atomic.Bool` for closed state
- `atomic.Int32` for error channel buffer size

**Topic** (`topics.go`):
- Maintains listeners in priority-sorted order using `slices.BinarySearchFunc`
- `sync.RWMutex` for listener map protection
- `Trigger(event)` executes listeners in priority order, stops on abort

**Event Interface** (`event.go`):
- `Topic()` - Event topic name
- `Payload()` - Event data
- `SetAborted(bool)` / `IsAborted()` - Control propagation

**Listener** (`listener.go`):
- Type: `func(Event) error`

### Wildcard Pattern Matching

Supports two wildcard types (`utils.go`):
- `*` - Matches exactly one segment (e.g., `user.*` matches `user.created`)
- `**` - Matches zero or more segments (e.g., `order.**` matches `order.created.success`)

Pattern matching uses recursive algorithm with fast-path optimization for exact matches.

### Priority System

Five priority levels (`priority.go`):
- `Highest` (100)
- `High` (75)
- `Normal` (50) - default
- `Low` (25)
- `Lowest` (0)

Listeners execute in descending priority order. Use `WithPriority(level)` option when registering.

### Configuration Options

Functional options pattern (`options.go`):
- `WithErrorHandler(func(Event, error) error)` - Custom error handling
- `WithIDGenerator(func() string)` - Custom listener ID generation (default: `rand.Text()`)
- `WithPanicHandler(func(any))` - Panic recovery
- `WithPool(Pool)` - Goroutine pool for concurrent execution
- `WithErrChanBufferSize(int)` - Error channel buffer size (default: 10)

## Design Philosophy

### Performance-First Design
- **Lock-free hot paths**: `atomic.Pointer` for handler access, `sync.Map` for topic lookup
- **Zero-copy operations**: Minimal allocations in event emission and listener execution
- **Efficient data structures**: Binary search for listener insertion (O(log n)), sorted slice traversal
- **Modern Go features**: Uses Go 1.26 `slices` package, `synctest` for deterministic concurrency testing

### Thread Safety Guarantees
- All public methods are thread-safe
- MemoryEmitter uses atomic operations for state management
- Topics use `sync.RWMutex` with reader preference for high-read scenarios
- Events use `sync.RWMutex` for payload/abort state protection

### Simplicity and Composability
- Clean interface-based design
- Functional options for extensibility
- Optional goroutine pool integration (bring your own pool)
- No external dependencies for core functionality

## Coding Rules

### Go 1.26 Modern Features
- Use `slices.BinarySearchFunc`, `slices.Insert`, `slices.Delete` for slice operations
- Use `any` instead of `interface{}`
- Use `atomic.Pointer`, `atomic.Bool`, `atomic.Int32` for lock-free operations
- Use `rand.Text()` for ID generation (Go 1.24+)
- Use `for range N` syntax where appropriate

### Error Handling
- Return sentinel errors defined in `errors.go`:
  - `ErrNilListener` - nil listener passed to On()
  - `ErrInvalidTopicName` - empty or invalid topic name
  - `ErrTopicNotFound` - topic doesn't exist
  - `ErrListenerNotFound` - listener ID not found
  - `ErrEmitterClosed` - operation on closed emitter
  - `ErrEmitterAlreadyClosed` - Close() called twice
- Use `fmt.Errorf` with `%w` for error wrapping when adding context

### Concurrency Patterns
- Use `sync.Map` for concurrent map access without explicit locking
- Use `atomic.Pointer` for lock-free handler updates
- Use `sync.RWMutex` when read-heavy workloads benefit from reader concurrency
- Always use `defer` for mutex unlocks
- Recover from panics in event handlers using `defer recover()`

### Performance Guidelines
- Avoid allocations in hot paths (event emission, listener execution)
- Pre-allocate slices when size is known
- Use `strings.Builder` for string concatenation
- Prefer atomic operations over mutexes for simple state
- Use binary search for sorted slice operations

## Testing

### Test Strategy
- **Unit tests**: Component-level testing with `testify/assert` and `testify/require`
- **Parallel tests**: Use `t.Parallel()` for independent tests
- **Concurrency tests**: Use `testing/synctest` (Go 1.26) for deterministic time-based testing
- **Fuzz tests**: Pattern matching edge cases (`utils_fuzz_test.go`)
- **Race detection**: All tests run with `-race` flag via `task test`
- **Benchmarks**: Performance regression detection (`benchmark_test.go`)

### Test Conventions
- Use table-driven tests with subtests for multiple scenarios
- Use `sync.WaitGroup` and channels for synchronization, never `time.Sleep` in production tests
- Use `synctest.Test()` for tests involving time delays or concurrent operations
- Name test errors with `errTest` prefix (e.g., `errTestListenerError`)
- Use `t.Cleanup()` for resource cleanup

### Running Tests
```bash
task test           # All tests with race detection
task test-fuzz      # Fuzz tests (10s)
make vet-shadow     # Check variable shadowing
```

## Dependencies

**Production**:
- `github.com/alitto/pond` - Optional goroutine pool (used via Pool interface)

**Development**:
- `github.com/stretchr/testify` - Test assertions
- `github.com/google/uuid` - Used in examples only


## Agent Skills

This package indexes agent skills from its own .agents/skills directory (emitter/.agents/skills/):

| Skill | When to Use |
|-------|-------------|
| [agent-md-creating](.agents/skills/agent-md-creating/) | Create or update CLAUDE.md and AGENTS.md instructions for this Go package. |
| [code-simplifying](.agents/skills/code-simplifying/) | Refine recently changed Go code for clarity and consistency without behavior changes. |
| [committing](.agents/skills/committing/) | Prepare conventional commit messages for this Go package. |
| [dependency-selecting](.agents/skills/dependency-selecting/) | Evaluate and choose Go dependencies with alternatives and risk tradeoffs. |
| [go-best-practices](.agents/skills/go-best-practices/) | Apply Google Go style and architecture best practices to code changes. |
| [linting](.agents/skills/linting/) | Configure or run golangci-lint and fix lint issues in this package. |
| [modernizing](.agents/skills/modernizing/) | Adopt newer Go language and toolchain features safely. |
| [ralphy-initializing](.agents/skills/ralphy-initializing/) | Initialize or repair the .ralphy workflow configuration. |
| [ralphy-todo-creating](.agents/skills/ralphy-todo-creating/) | Generate or refine TODO tracking via the Ralphy workflow. |
| [readme-creating](.agents/skills/readme-creating/) | Create or rewrite README.md for this package. |
| [releasing](.agents/skills/releasing/) | Prepare release and semantic version workflows for this package. |
| [testing](.agents/skills/testing/) | Design or update tests (table-driven, fuzz, benchmark, and edge-case coverage). |
