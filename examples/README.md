# Examples

Runnable examples for `github.com/kaptinlin/emitter`. Each subdirectory is its own `main` package; run with `go run .`.

| Directory | What it shows |
| --- | --- |
| [`basic/`](basic/) | The smallest end-to-end loop — register one listener, emit one event. |
| [`wildcards/`](wildcards/) | Wildcard subscriptions, priority ordering, and `Event.Stop()`. |
| [`typed/`](typed/) | `Subscribe[T]` / `Publish[T]` for type-safe payloads at the boundary. |
| [`../pool/examples/basic/`](../pool/examples/basic/) | Bounded asynchronous dispatch via the [`emitter/pool`](../pool/) sibling module. |

Each example is small enough to read in under a minute. For executable godoc snippets, see the `Example*` functions in [`example_test.go`](../example_test.go).
