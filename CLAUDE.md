# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**emitter** is a thread-safe in-memory event emitter for Go 1.26.2.
It supports exact topics, wildcard subscriptions, listener priorities, optional goroutine pooling, and panic recovery that surfaces as returned errors.

- **Module**: `github.com/kaptinlin/emitter`
- **Go Version**: `1.26.2`
- **License**: MIT
- **Usage Docs**: [README.md](README.md)
- **Examples**: [examples/README.md](examples/README.md)

## Commands

```bash
task test        # Run `go test -race ./...`
task lint        # Run golangci-lint and the go mod tidy check
task test-fuzz   # Run fuzz coverage for topic matching
task clean       # Remove local build artifacts and test cache
task verify      # Run deps, fmt, vet, lint, test, and vuln checks
```

## Architecture

```text
emitter/
├── emitter.go         # Public Emitter interface
├── memory.go          # MemoryEmitter implementation and topic routing
├── event.go           # Event interface and BaseEvent
├── topic.go           # Listener registry and priority-ordered dispatch
├── listener.go        # Listener type and listener options
├── options.go         # MemoryEmitter configuration options
├── priority.go        # Priority constants and validation
├── pool.go            # Optional pond-backed Pool adapter
├── errors.go          # Sentinel errors and PanicError
├── utils.go           # Topic validation and wildcard matching
├── examples/          # Runnable usage examples
└── benchmark_test.go  # Performance regression coverage
```

### Core Responsibilities

- `MemoryEmitter` owns topic lookup, emission mode, close state, and emitter-wide configuration.
- `Topic` owns listener storage, ordering, and per-topic dispatch.
- `BaseEvent` owns mutable payload and aborted state for a single emission.
- `Pool` is optional; `EmitSync` stays local while `Emit` can delegate work to pooled execution.

## Design Philosophy

- **KISS** — Keep the API small: register listeners, emit events, remove listeners, close the emitter.
- **SRP** — Routing lives in `MemoryEmitter`, ordering lives in `Topic`, and event state lives in `BaseEvent`.
- **Simplicity as art** — Topic matching is limited to exact names, `*`, and `**`; avoid richer subscription DSLs.
- **Errors as teachers** — Invalid topic names, missing listeners, closed emitters, and recovered listener panics all surface as typed errors.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Progressive Disclosure** — `NewMemoryEmitter`, `On`, and `EmitSync` cover the common path; setters, listener options, and `Pool` support advanced control.

## Coding Rules

### Must Follow

- Use Go 1.26.2 features when they make the code smaller or clearer.
- Follow Google Go Best Practices and Google Go Style Decisions.
- Return errors instead of panicking in production code.
- Keep public APIs thread-safe.
- Use the sentinel errors in `errors.go` for caller-visible failure modes.
- Keep wildcard semantics aligned with `utils.go`: `*` matches one segment, `**` matches zero or more segments.
- Use `WithPriority` or the emitter setters instead of reaching into internal state.
- Keep examples and README snippets aligned with the public API.

### Forbidden

- No `panic` in production code; recovered listener panics must surface as errors matching `ErrListenerPanic`.
- No working around dependency bugs — if a dependency is wrong, report it in `reports/` instead of reimplementing it inline.
- No feature creep beyond exact topics, wildcards, priorities, pooling, and error handling already supported here.
- No breaking the `AGENTS.md -> CLAUDE.md` symlink.

## Dependency Issue Reporting

When you hit a bug or limitation in a dependency library:

1. Do not work around it by reimplementing the dependency's behavior.
2. Create `reports/<dependency-name>.md`.
3. Record the dependency version, trigger scenario, expected behavior, actual behavior, and any non-code workaround.
4. Continue with work that does not depend on the broken behavior.

## Error Handling

- `ErrNilListener`, `ErrInvalidTopicName`, `ErrTopicNotFound`, `ErrListenerNotFound`, `ErrEmitterClosed`, and `ErrEmitterAlreadyClosed` are the primary sentinel errors.
- Listener panics are recovered and wrapped in `PanicError`; callers should match them with `errors.Is(err, ErrListenerPanic)`.
- `Emit` returns a channel of handled listener errors; `EmitSync` returns a slice of handled listener errors.

## Testing

- Run `task test` before finishing work in this package.
- Use `t.Parallel()` for independent tests.
- Use `testing/synctest` for time-based concurrent behavior instead of long real sleeps.
- Prefer table-driven tests for topic validation, wildcard matching, and priority behavior.
- Validate documentation examples with runnable example tests rather than tests that parse markdown files.

## Dependencies

### Runtime

- `github.com/alitto/pond` — optional goroutine pool adapter used by `PondPool`

### Development

- `github.com/stretchr/testify` — assertions in unit tests
- `github.com/google/uuid` — example-only custom ID generation

## SPECS Index

This repository does not maintain a top-level `SPECS/` directory.

## Agent Skills

This package vendors shared Go skills in `.agents/skills/`. Start with these:

| Skill | When to Use |
|-------|-------------|
| [agent-md-writing](.agents/skills/agent-md-writing/) | Refresh `CLAUDE.md` and the `AGENTS.md` symlink. |
| [readme-writing](.agents/skills/readme-writing/) | Refresh human-facing usage documentation. |
| [library-docs-maintaining](.agents/skills/library-docs-maintaining/) | Refresh `CLAUDE.md`, `AGENTS.md`, and `README.md` together. |
| [golangci-linting](.agents/skills/golangci-linting/) | Update lint configuration or fix lint failures. |
| [library-test-covering](.agents/skills/library-test-covering/) | Add or improve runtime-focused test coverage. |
| [library-code-modernizing](.agents/skills/library-code-modernizing/) | Adopt newer Go language and toolchain features safely. |
| [go-best-practices](.agents/skills/go-best-practices/) | Align code with Go style and API design conventions. |
| [committing](.agents/skills/committing/) | Prepare Conventional Commit messages for this package. |
| [library-specs-maintaining](.agents/skills/library-specs-maintaining/) | Refresh or align top-level `SPECS/` documents when the package adds them. |
| [releasing](.agents/skills/releasing/) | Prepare versioning and release workflow changes. |
