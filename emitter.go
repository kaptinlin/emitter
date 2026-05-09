package emitter

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

// Emitter is an in-memory pub/sub primitive.
// All configuration is fixed at construction; runtime state changes only
// through On / Emit / Subscription.Cancel / Close.
//
// The zero value of Emitter is not usable; always construct via [New].
type Emitter struct {
	exact    sync.Map                    // string -> *bucket (exact patterns, O(1) lookup)
	wildcard atomic.Pointer[[]wildEntry] // copy-on-write list of wildcard patterns
	wildMu   sync.Mutex                  // serializes writes to wildcard

	nextID atomic.Uint64
	closed atomic.Bool
}

// wildEntry is a single wildcard registration; pre-split parts cache avoids
// re-splitting the pattern on every emit.
type wildEntry struct {
	pattern string
	parts   []string
	bucket  *bucket
}

// New constructs an Emitter. Options are reserved for future extension.
func New(opts ...Option) *Emitter {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	_ = cfg
	return &Emitter{}
}

// On registers listener to receive events whose topic matches pattern.
// Returns a Subscription whose Cancel removes the listener.
//
// pattern must satisfy the topic grammar (see package doc) and may contain
// '*' / '**' wildcards for subscription. Returns ErrInvalidTopicName when
// the pattern is malformed, ErrNilListener when listener is nil, and
// ErrEmitterClosed when the emitter has been closed.
func (e *Emitter) On(pattern string, listener Listener, opts ...ListenerOption) (Subscription, error) {
	if listener == nil {
		return nil, ErrNilListener
	}
	if !isValidTopicName(pattern) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidTopicName, pattern)
	}
	if e.closed.Load() {
		return nil, ErrEmitterClosed
	}

	o := &listenerOpts{}
	for _, opt := range opts {
		opt(o)
	}

	item := &listenerItem{
		id:       e.nextID.Add(1),
		listener: listener,
		priority: o.priority,
		once:     o.once,
	}

	if hasWildcard(pattern) {
		e.ensureWildcard(pattern).add(item)
	} else {
		e.ensureExact(pattern).add(item)
	}

	return &subscription{emitter: e, pattern: pattern, id: item.id}, nil
}

// Emit dispatches payload to listeners registered on a pattern that matches topic.
// Listeners run synchronously in priority order; their errors are joined.
//
// topic must be a literal name (no wildcards). Returns ErrEmitterClosed if
// the emitter is closed and ErrInvalidTopicName for malformed names.
// If ctx is cancelled during dispatch, ctx.Err() is included in the joined result
// and remaining listeners are skipped.
func (e *Emitter) Emit(ctx context.Context, topic string, payload any) error {
	if e.closed.Load() {
		return ErrEmitterClosed
	}
	if !isValidTopicName(topic) {
		return fmt.Errorf("%w: %q", ErrInvalidTopicName, topic)
	}
	if hasWildcard(topic) {
		return fmt.Errorf("%w: emit topic must not contain wildcards: %q", ErrInvalidTopicName, topic)
	}

	ev := &event{topic: topic, payload: payload}
	var errs []error

	if v, ok := e.exact.Load(topic); ok {
		if list := v.(*bucket).trigger(ctx, ev); len(list) > 0 {
			errs = append(errs, list...)
		}
	}

	if ev.stopped {
		return errors.Join(errs...)
	}
	if cur := e.wildcard.Load(); cur != nil && len(*cur) > 0 {
		subjectParts := strings.Split(topic, ".")
		for _, w := range *cur {
			if !matchParts(w.parts, subjectParts, 0, 0) {
				continue
			}
			if list := w.bucket.trigger(ctx, ev); len(list) > 0 {
				errs = append(errs, list...)
			}
			if ev.stopped {
				break
			}
		}
	}

	return errors.Join(errs...)
}

// Close prevents further emits. Idempotent — calling Close more than once is safe.
//
// In-flight Emit calls are not interrupted; they complete with whatever listeners
// were already snapshotted. Listeners registered before Close continue to be
// reachable through Subscription.Cancel for cleanup.
func (e *Emitter) Close() {
	e.closed.Store(true)
}

func (e *Emitter) ensureExact(pattern string) *bucket {
	if v, ok := e.exact.Load(pattern); ok {
		return v.(*bucket)
	}
	nb := newBucket()
	actual, _ := e.exact.LoadOrStore(pattern, nb)
	return actual.(*bucket)
}

func (e *Emitter) ensureWildcard(pattern string) *bucket {
	e.wildMu.Lock()
	defer e.wildMu.Unlock()
	if cur := e.wildcard.Load(); cur != nil {
		for _, w := range *cur {
			if w.pattern == pattern {
				return w.bucket
			}
		}
	}
	nb := newBucket()
	var prev []wildEntry
	if cur := e.wildcard.Load(); cur != nil {
		prev = *cur
	}
	next := append(slices.Clone(prev), wildEntry{
		pattern: pattern,
		parts:   strings.Split(pattern, "."),
		bucket:  nb,
	})
	e.wildcard.Store(&next)
	return nb
}

func (e *Emitter) removeListener(pattern string, id uint64) {
	if !hasWildcard(pattern) {
		if v, ok := e.exact.Load(pattern); ok {
			v.(*bucket).remove(id)
		}
		return
	}

	if cur := e.wildcard.Load(); cur != nil {
		for _, w := range *cur {
			if w.pattern == pattern {
				w.bucket.remove(id)
				return
			}
		}
	}
}
