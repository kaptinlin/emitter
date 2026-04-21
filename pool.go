package emitter

import "github.com/alitto/pond"

// Pool schedules work for Emit.
type Pool interface {
	// Submit schedules task for execution.
	Submit(task func())
	// Running returns the number of active workers.
	Running() int
	// Release stops the pool and waits for queued work.
	Release()
}

// PondPool adapts [pond.WorkerPool] to [Pool].
type PondPool struct {
	pool *pond.WorkerPool
}

// NewPondPool returns a [PondPool] with the given limits.
func NewPondPool(maxWorkers, maxCapacity int, options ...pond.Option) *PondPool {
	return &PondPool{
		pool: pond.New(maxWorkers, maxCapacity, options...),
	}
}

// Submit schedules task for execution.
func (p *PondPool) Submit(task func()) {
	p.pool.Submit(task)
}

// Running returns the number of active workers.
func (p *PondPool) Running() int {
	return p.pool.RunningWorkers()
}

// Release stops the pool and waits for queued work.
func (p *PondPool) Release() {
	p.pool.StopAndWait()
}
