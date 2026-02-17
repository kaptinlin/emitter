package emitter

import "github.com/alitto/pond"

// Pool defines the interface for goroutine pool implementations used by the Emitter
// to manage concurrent event handler execution.
type Pool interface {
	// Submit enqueues a task for execution in the pool.
	Submit(task func())
	// Running returns the number of currently active workers.
	Running() int
	// Release stops the pool and waits for all tasks to complete.
	Release()
}

// PondPool wraps a [pond.WorkerPool] to implement the [Pool] interface.
type PondPool struct {
	pool *pond.WorkerPool
}

// NewPondPool creates a new PondPool with the given worker and capacity limits.
func NewPondPool(maxWorkers, maxCapacity int, options ...pond.Option) *PondPool {
	return &PondPool{
		pool: pond.New(maxWorkers, maxCapacity, options...),
	}
}

// Submit enqueues a task for execution in the pool.
func (p *PondPool) Submit(task func()) {
	p.pool.Submit(task)
}

// Running returns the number of currently active workers.
func (p *PondPool) Running() int {
	return p.pool.RunningWorkers()
}

// Release stops the pool and waits for all tasks to complete.
func (p *PondPool) Release() {
	p.pool.StopAndWait()
}
