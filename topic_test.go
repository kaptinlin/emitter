package emitter

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBucketAddPriorityOrder(t *testing.T) {
	t.Parallel()
	b := newBucket()

	items := []*listenerItem{
		{id: 1, priority: Normal},
		{id: 2, priority: Highest},
		{id: 3, priority: Low},
		{id: 4, priority: High},
		{id: 5, priority: Lowest},
	}
	for _, it := range items {
		it.listener = func(context.Context, Event) error { return nil }
		b.add(it)
	}

	got := make([]uint64, 0, len(b.sorted))
	for _, it := range b.sorted {
		got = append(got, it.id)
	}
	require.Equal(t, []uint64{2, 4, 1, 3, 5}, got)
}

func TestBucketAddPreservesRegistrationOrderAtEqualPriority(t *testing.T) {
	t.Parallel()
	b := newBucket()
	for i := range 5 {
		b.add(&listenerItem{
			id:       uint64(i + 1),
			priority: Normal,
			listener: func(context.Context, Event) error { return nil },
		})
	}
	got := make([]uint64, 0, len(b.sorted))
	for _, it := range b.sorted {
		got = append(got, it.id)
	}
	require.Equal(t, []uint64{1, 2, 3, 4, 5}, got)
}

func TestBucketRemoveDeletesMatchingListener(t *testing.T) {
	t.Parallel()
	b := newBucket()
	b.add(&listenerItem{id: 1, listener: func(context.Context, Event) error { return nil }})

	b.remove(1)
	b.remove(1)
	b.remove(99)
	require.Empty(t, b.sorted)
}

func TestBucketTriggerSnapshotsListeners(t *testing.T) {
	t.Parallel()
	b := newBucket()

	// First listener registers a second one mid-dispatch — the new one must
	// not run during this trigger because dispatch operates on a snapshot.
	var ranLate bool
	b.add(&listenerItem{
		id:       1,
		priority: Normal,
		listener: func(context.Context, Event) error {
			b.add(&listenerItem{
				id:       2,
				priority: Normal,
				listener: func(context.Context, Event) error {
					ranLate = true
					return nil
				},
			})
			return nil
		},
	})

	errs := b.trigger(context.Background(), &event{topic: "x"})
	require.Empty(t, errs)
	require.False(t, ranLate, "newly added listener must not fire during in-flight trigger")
}

func TestBucketTriggerJoinsErrors(t *testing.T) {
	t.Parallel()
	b := newBucket()

	errA := errors.New("a")
	errB := errors.New("b")
	b.add(&listenerItem{id: 1, listener: func(context.Context, Event) error { return errA }})
	b.add(&listenerItem{id: 2, listener: func(context.Context, Event) error { return errB }})

	errs := b.trigger(context.Background(), &event{topic: "x"})
	require.Len(t, errs, 2)
}

func TestBucketTriggerStopHaltsLoop(t *testing.T) {
	t.Parallel()
	b := newBucket()

	var second bool
	b.add(&listenerItem{
		id: 1, priority: High,
		listener: func(_ context.Context, ev Event) error {
			ev.Stop()
			return nil
		},
	})
	b.add(&listenerItem{
		id: 2, priority: Low,
		listener: func(context.Context, Event) error {
			second = true
			return nil
		},
	})

	_ = b.trigger(context.Background(), &event{topic: "x"})
	require.False(t, second)
}

func TestBucketRemoveDuringTriggerIsSafe(t *testing.T) {
	t.Parallel()
	b := newBucket()

	var wg sync.WaitGroup
	for i := range 8 {
		b.add(&listenerItem{
			id:       uint64(i + 1),
			listener: func(context.Context, Event) error { return nil },
		})
	}

	wg.Go(func() {
		for range 100 {
			_ = b.trigger(context.Background(), &event{topic: "x"})
		}
	})
	for i := range 8 {
		wg.Go(func() {
			b.remove(uint64(i + 1))
		})
	}
	wg.Wait()
}
