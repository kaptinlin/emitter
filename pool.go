package emitter

import "github.com/alitto/pond"

type Pool interface {
	Submit(task func())
	Running() int
	Release()
}

type PondPool struct {
	pool *pond.WorkerPool
}

func NewPondPool(maxWorkers, maxCapacity int, options ...pond.Option) *PondPool {
	return &PondPool{
		pool: pond.New(maxWorkers, maxCapacity, options...),
	}
}

func (p *PondPool) Submit(task func()) {
	p.pool.Submit(task)
}

func (p *PondPool) Running() int {
	return p.pool.RunningWorkers()
}

func (p *PondPool) Release() {
	p.pool.StopAndWait()
}
