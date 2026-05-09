package emitter

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
)

type listenerItem struct {
	id       uint64
	listener Listener
	priority Priority
	once     bool
	fired    atomic.Bool // for Once: at-most-once across concurrent emits
}

// bucket is the internal listener registry for a single pattern.
// Dispatch uses a snapshot taken under RLock and runs listeners outside
// the lock to keep listener bodies free to call back into the Emitter.
type bucket struct {
	mu     sync.RWMutex
	sorted []*listenerItem // priority desc, then registration order
}

func newBucket() *bucket { return &bucket{} }

func (b *bucket) add(it *listenerItem) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Insert at the first position where existing priority is strictly less than the
	// new item's priority. Equal priorities are preserved in registration order.
	idx := len(b.sorted)
	for i, existing := range b.sorted {
		if existing.priority < it.priority {
			idx = i
			break
		}
	}
	b.sorted = slices.Insert(b.sorted, idx, it)
}

func (b *bucket) remove(id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sorted = slices.DeleteFunc(b.sorted, func(it *listenerItem) bool {
		return it.id == id
	})
}

// trigger dispatches ev to all registered listeners in priority order.
// Returns the slice of errors collected; the caller decides how to combine them.
func (b *bucket) trigger(ctx context.Context, ev *event) []error {
	b.mu.RLock()
	items := slices.Clone(b.sorted)
	b.mu.RUnlock()

	var errs []error
	var fired []uint64

	for _, it := range items {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		if it.once && !it.fired.CompareAndSwap(false, true) {
			continue // another emit already fired this once-listener
		}
		if err := safeCall(ctx, it.listener, ev); err != nil {
			errs = append(errs, err)
		}
		if it.once {
			fired = append(fired, it.id)
		}
		if ev.stopped {
			break
		}
	}

	if len(fired) > 0 {
		b.removeMany(fired)
	}
	return errs
}

func (b *bucket) removeMany(ids []uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sorted = slices.DeleteFunc(b.sorted, func(it *listenerItem) bool {
		return slices.Contains(ids, it.id)
	})
}

// safeCall runs l with panic recovery, wrapping any panic as a *PanicError.
func safeCall(ctx context.Context, l Listener, ev *event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			pe := &PanicError{Value: r}
			if cause, ok := r.(error); ok {
				pe.cause = cause
			}
			err = pe
		}
	}()
	return l(ctx, ev)
}
