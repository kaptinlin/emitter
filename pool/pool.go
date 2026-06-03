// Package pool provides a bounded asynchronous emit dispatcher for
// [github.com/kaptinlin/emitter].
//
// The core emitter dispatches synchronously. For high-throughput
// fire-and-forget emission, this package offers a concrete [Pool] backed by a
// bounded worker pool with a queue and backpressure. The underlying engine is
// an implementation detail and is not part of the API surface.
//
// Typical usage:
//
//	p := pool.New(64, 1024) // 64 workers, queue cap 1024
//	defer p.Close()
//
//	if err := p.Submit(ctx, e, "user.created", payload); err != nil {
//	    if errors.Is(err, pool.ErrPoolFull) { /* shed load */ }
//	}
//
// Listener errors are logged via [log/slog]; there is no error hook.
package pool

import (
	"context"
	"errors"
	"log/slog"

	"github.com/alitto/pond"

	"github.com/kaptinlin/emitter"
)

// ErrPoolFull is returned by [Pool.Submit] when the pool's queue is at capacity
// and the submission cannot be accepted without blocking.
var ErrPoolFull = errors.New("emitter/pool: queue full")

// Pool is a bounded asynchronous dispatcher over an [emitter.Emitter].
// Concurrency is capped at maxWorkers and pending tasks at maxQueue;
// once both are exhausted, [Pool.Submit] returns [ErrPoolFull].
//
// The zero value of Pool is not usable; always construct via [New].
type Pool struct {
	p *pond.WorkerPool
}

// New constructs a Pool. maxWorkers caps concurrent dispatch goroutines;
// maxQueue caps pending submissions. Values less than 1 are clamped to 1.
func New(maxWorkers, maxQueue int) *Pool {
	maxWorkers = max(maxWorkers, 1)
	maxQueue = max(maxQueue, 1)
	return &Pool{p: pond.New(maxWorkers, maxQueue, pond.Strategy(pond.Eager()))}
}

// Submit schedules an asynchronous emit. Returns [ErrPoolFull] if the queue
// has no room. Listener errors raised during dispatch are logged via slog at
// error level with the topic and error attached; the submitter is not blocked
// on listener completion.
func (pl *Pool) Submit(ctx context.Context, e *emitter.Emitter, topic string, payload any) error {
	if !pl.p.TrySubmit(func() {
		if err := e.Emit(ctx, topic, payload); err != nil {
			slog.ErrorContext(ctx, "emitter/pool: emit failed", "topic", topic, "err", err)
		}
	}) {
		return ErrPoolFull
	}
	return nil
}

// Close stops accepting submissions and waits for in-flight tasks to finish.
// After Close, further [Pool.Submit] calls return [ErrPoolFull].
func (pl *Pool) Close() {
	pl.p.StopAndWait()
}
