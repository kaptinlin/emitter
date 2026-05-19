# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**emitter** is an in-memory pub/sub primitive for the Go version declared in `go.mod`. The core dispatches synchronously in priority order, supports `*` / `**` wildcard subscriptions, recovers listener panics into typed errors, and ships with **zero third-party dependencies**.

Bounded asynchronous dispatch lives in a sibling module `emitter/pool` (combined via `go.work`) so users who never need it pay nothing for it.

- **Module**: `github.com/kaptinlin/emitter`
- **Sibling module**: `github.com/kaptinlin/emitter/pool`
- **Go version**: see `go.mod`
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

The pool sibling has its own module; lint/test it from `pool/` or via the workspace at the repo root.

## Architecture

```text
emitter/
├── go.work               # combines root + ./pool
├── go.mod                # core: stdlib only (testify is dev-only)
├── doc.go                # package doc — topic grammar, dispatch semantics
├── emitter.go            # Emitter struct: exact sync.Map + COW wildcard list
├── topic.go              # internal `bucket`: snapshot dispatch under RWMutex
├── event.go              # Event interface + internal event struct
├── listener.go           # Listener type, WithPriority, Once
├── subscription.go       # Subscription handle (idempotent Cancel)
├── priority.go           # Priority is plain int + Lowest..Highest sentinels
├── options.go            # Option (forward extension; no fields today)
├── errors.go             # sentinels + PanicError
├── utils.go              # EBNF validation + matchParts
├── typed.go              # generic helpers Subscribe[T] / Publish[T]
├── examples/             # runnable demo programs
├── pool/                 # sibling module: bounded async dispatch
│   ├── pool.go           # concrete Pool struct (pond is hidden inside)
│   └── examples/basic/   # async dispatch demo
└── *_test.go             # unit + fuzz + benchmark coverage
```

### Core responsibilities

- **`Emitter`** owns topic routing, the closed flag, and listener-id allocation. Construct with `New`.
- **`bucket`** (internal, per pattern) owns listener storage, ordering, and per-topic dispatch via a snapshot taken under `RLock`.
- **`event`** (internal, per emission) holds payload and `stopped` flag; only the `Event` interface is public.
- **`Subscription`** is the cancellation handle; `Cancel` is idempotent via `atomic.Bool`.
- **`pool.Pool`** is optional and lives in the sibling module; the underlying engine (`alitto/pond`) is an implementation detail.

### Dispatch semantics

- Listeners run **synchronously** in priority order (high → low). Equal priorities run in registration order.
- Each emit clones the listener slice under `RLock`, releases the lock, then runs listeners outside it — listeners may safely call `On` / `Cancel` on the same emitter.
- `Event.Stop()` halts the remaining listeners for the current emit only.
- Listener errors are joined via `errors.Join`. `ctx.Err()` is appended and dispatch stops if the context is cancelled mid-loop.
- Panics are recovered into `*PanicError` (which unwraps to `ErrListenerPanic` and, when applicable, the original error cause).
- `Once()` listeners use `atomic.Bool.CompareAndSwap` so concurrent emits fire them at most once; the registration is removed after dispatch.

## Design Philosophy

- **One thing well.** In-process pub/sub. No retry, no broker, no transport, no replay.
- **Synchronous core.** Async is a sibling module, never a runtime knob.
- **Concrete types > interfaces.** `Emitter` and `Pool` are concrete structs; we expose interfaces (`Event`, `Subscription`) only where polymorphism is genuinely useful.
- **No generic Bus type.** Type-safety lives in free functions `Subscribe[T]` / `Publish[T]`; the emitter stays untyped so one instance can carry many event types.
- **Zero core deps.** The root `go.mod` requires only stdlib + testify (dev-only).
- **Public surface = promise.** Internal types (`bucket`, `event`, `wildEntry`, `listenerItem`, `subscription`) stay lowercase. Don't promote them.
- **Errors as teachers.** Failure modes that callers can act on get sentinels; nothing more.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Progressive disclosure.** `New` + `On` + `Emit` + `Close` cover the 90% path; `WithPriority`, `Once`, `Subscribe[T]`, `Publish[T]`, and the `pool` module sit one step further out.
- **No runtime reconfiguration.** Configuration is fixed at construction. There are no `SetXxx` methods that race with `Emit`.
- **Idempotent destructors.** `Subscription.Cancel` and `Emitter.Close` may be called more than once. `Close` returns nothing — there is no error to surface.

## Coding Rules

### Must Follow

- Use the Go version declared in `go.mod` and features when they make code smaller or clearer (`range over int`, `WaitGroup.Go`, `slices.*`, etc.).
- Follow Google Go Best Practices and Google Go Style Decisions.
- Return errors instead of panicking in production code.
- Keep public APIs thread-safe; document any non-obvious lock ordering near the field declarations.
- Use the sentinels in `errors.go` for caller-visible failure modes — never invent ad-hoc error strings for the same condition.
- Keep wildcard semantics aligned with `utils.go`: `*` matches one segment, `**` matches zero or more segments.
- Pre-split wildcard patterns at registration time; don't re-split on every emit.
- Keep `examples/`, `pool/examples/`, `example_test.go`, and `README.md` in sync with the public API.

### Forbidden

- No `panic` in production code; recovered listener panics must surface as errors matching `ErrListenerPanic`.
- No exposing internal types: `bucket`, `event`, `wildEntry`, `listenerItem`, and the concrete `subscription` struct are private and stay private.
- No runtime setters on `Emitter` (`SetPool`, `SetErrorHandler`, etc.) — they race with `Emit`.
- No third-party imports in the **root** module. Pool's `alitto/pond` lives only inside `pool/`.
- No leaking `pool/`'s engine: `Pool`, `New`, `Submit`, `Close`, `ErrPoolFull` are the entire surface.
- No working around dependency bugs — if a dependency is wrong, report it in `reports/` instead of reimplementing it inline.
- No feature creep beyond exact topics, wildcards, priorities, optional pool, panic recovery, and typed helpers.
- No breaking the `AGENTS.md -> CLAUDE.md` symlink.

## Dependency Issue Reporting

When you hit a bug or limitation in a dependency library:

1. Do not work around it by reimplementing the dependency's behavior.
2. Create `reports/<dependency-name>.md`.
3. Record the dependency version, trigger scenario, expected behavior, actual behavior, and any non-code workaround.
4. Continue with work that does not depend on the broken behavior.

## Error Handling

- Sentinels: `ErrEmitterClosed`, `ErrInvalidTopicName`, `ErrNilListener`, `ErrListenerPanic`, `ErrPayloadType`. The pool module adds `pool.ErrPoolFull`.
- Listener panics are wrapped in `*PanicError`. Match with `errors.Is(err, ErrListenerPanic)`. If the panic value is itself an `error`, it appears in the unwrap chain too — match with `errors.Is(err, originalCause)`.
- `Emit` returns `errors.Join(...)` of all listener errors. Use `errors.Is` / `errors.As` to inspect.
- Cancelling `ctx` mid-dispatch surfaces `ctx.Err()` in the joined result and skips remaining listeners.
- `Subscribe[T]` returns `ErrPayloadType` (joined into the emit result) when the published payload's dynamic type does not match `T`. The typed listener is not invoked in that case.

## Testing

- Run `task test` before finishing work in the root module; run `go test -race ./...` from `pool/` for pool changes.
- Use `t.Parallel()` for independent tests.
- Prefer table-driven tests for grammar / matching / priority behavior.
- Topic-matching changes must keep `task test-fuzz` clean — `FuzzMatchTopicPattern` exercises both validator and matcher.
- Validate the documented behavior with example tests (`example_test.go`) — they double as godoc snippets.

## Dependencies

### Runtime

- **Root module**: stdlib only.
- **`pool/` module**: `github.com/alitto/pond` — hidden behind the `Pool` struct, never re-exported.

### Development

- `github.com/stretchr/testify` — assertions in unit tests (root and pool modules).

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
