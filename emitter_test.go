package emitter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmitterBasicEmit(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var got string
	_, err := e.On("user.created", func(_ context.Context, ev Event) error {
		got = ev.Topic()
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, e.Emit(context.Background(), "user.created", nil))
	require.Equal(t, "user.created", got)
}

func TestEmitterPayload(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var got any
	_, err := e.On("data", func(_ context.Context, ev Event) error {
		got = ev.Payload()
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, e.Emit(context.Background(), "data", 42))
	require.Equal(t, 42, got)
}

func TestEmitterNoListener(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	require.NoError(t, e.Emit(context.Background(), "missing", nil))
}

func TestOnNilListenerRejected(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	_, err := e.On("ok", nil)
	require.ErrorIs(t, err, ErrNilListener)
}

func TestOnInvalidPattern(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	for _, bad := range []string{"", "user.", ".user", "user..created", "用户"} {
		_, err := e.On(bad, func(context.Context, Event) error { return nil })
		require.ErrorIsf(t, err, ErrInvalidTopicName, "pattern %q should be invalid", bad)
	}
}

func TestEmitInvalidTopic(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	for _, bad := range []string{"", ".x", "x..y"} {
		require.ErrorIsf(t, e.Emit(context.Background(), bad, nil), ErrInvalidTopicName,
			"topic %q should be rejected", bad)
	}
}

func TestEmitWildcardTopicRejected(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	require.ErrorIs(t, e.Emit(context.Background(), "user.*", nil), ErrInvalidTopicName)
	require.ErrorIs(t, e.Emit(context.Background(), "**", nil), ErrInvalidTopicName)
}

func TestEmitterClosed(t *testing.T) {
	t.Parallel()
	e := New()
	e.Close()

	require.ErrorIs(t, e.Emit(context.Background(), "x", nil), ErrEmitterClosed)
	_, err := e.On("x", func(context.Context, Event) error { return nil })
	require.ErrorIs(t, err, ErrEmitterClosed)
}

func TestEmitterCloseIdempotent(t *testing.T) {
	t.Parallel()
	e := New()
	e.Close()
	e.Close()
	e.Close()
}

func TestWildcardSubscriptionSingle(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var got []string
	_, err := e.On("user.*", func(_ context.Context, ev Event) error {
		got = append(got, ev.Topic())
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, e.Emit(context.Background(), "user.created", nil))
	require.NoError(t, e.Emit(context.Background(), "user.deleted", nil))
	require.NoError(t, e.Emit(context.Background(), "order.created", nil)) // not matched
	require.Equal(t, []string{"user.created", "user.deleted"}, got)
}

func TestWildcardSubscriptionMulti(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var count atomic.Int32
	_, err := e.On("user.**", func(_ context.Context, _ Event) error {
		count.Add(1)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, e.Emit(context.Background(), "user", nil))
	require.NoError(t, e.Emit(context.Background(), "user.created", nil))
	require.NoError(t, e.Emit(context.Background(), "user.a.b.c", nil))
	require.Equal(t, int32(3), count.Load())
}

func TestExactAndWildcardBothFire(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var seen []string
	_, _ = e.On("user.created", func(_ context.Context, _ Event) error {
		seen = append(seen, "exact")
		return nil
	})
	_, _ = e.On("user.*", func(_ context.Context, _ Event) error {
		seen = append(seen, "wild")
		return nil
	})
	_, _ = e.On("**", func(_ context.Context, _ Event) error {
		seen = append(seen, "all")
		return nil
	})

	require.NoError(t, e.Emit(context.Background(), "user.created", nil))
	require.ElementsMatch(t, []string{"exact", "wild", "all"}, seen)
}

func TestStopHaltsRemainingListeners(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	// Higher priority runs first; it stops, lower priority must not fire.
	var firedHigh, firedLow bool
	_, _ = e.On("evt", func(_ context.Context, ev Event) error {
		firedHigh = true
		ev.Stop()
		return nil
	}, WithPriority(High))
	_, _ = e.On("evt", func(_ context.Context, _ Event) error {
		firedLow = true
		return nil
	}, WithPriority(Low))

	require.NoError(t, e.Emit(context.Background(), "evt", nil))
	require.True(t, firedHigh)
	require.False(t, firedLow)
}

func TestStopBlocksWildcardAfterExact(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var wildHit bool
	_, _ = e.On("evt", func(_ context.Context, ev Event) error {
		ev.Stop()
		return nil
	})
	_, _ = e.On("**", func(_ context.Context, _ Event) error {
		wildHit = true
		return nil
	})

	require.NoError(t, e.Emit(context.Background(), "evt", nil))
	require.False(t, wildHit, "wildcard listener should not fire after Stop")
}

func TestListenerErrorsJoined(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	errA := errors.New("A failed")
	errB := errors.New("B failed")
	_, _ = e.On("evt", func(context.Context, Event) error { return errA })
	_, _ = e.On("evt", func(context.Context, Event) error { return errB })

	err := e.Emit(context.Background(), "evt", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
}

func TestPanicRecovered(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	_, _ = e.On("evt", func(context.Context, Event) error {
		panic("boom")
	})

	err := e.Emit(context.Background(), "evt", nil)
	require.ErrorIs(t, err, ErrListenerPanic)

	var pe *PanicError
	require.True(t, errors.As(err, &pe))
	require.Equal(t, "boom", pe.Value)
}

func TestPanicWithErrorChainsCause(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	cause := errors.New("inner cause")
	_, _ = e.On("evt", func(context.Context, Event) error {
		panic(cause)
	})

	err := e.Emit(context.Background(), "evt", nil)
	require.ErrorIs(t, err, ErrListenerPanic)
	require.ErrorIs(t, err, cause)
}

func TestContextCancelStopsDispatch(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	ctx, cancel := context.WithCancel(context.Background())
	var ranAfterCancel bool

	_, _ = e.On("evt", func(_ context.Context, _ Event) error {
		cancel()
		return nil
	}, WithPriority(High))
	_, _ = e.On("evt", func(_ context.Context, _ Event) error {
		ranAfterCancel = true
		return nil
	}, WithPriority(Low))

	err := e.Emit(ctx, "evt", nil)
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, ranAfterCancel)
}

func TestSubscriptionCancelRemovesListener(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var hits int
	sub, err := e.On("evt", func(context.Context, Event) error {
		hits++
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, e.Emit(context.Background(), "evt", nil))
	sub.Cancel()
	require.NoError(t, e.Emit(context.Background(), "evt", nil))
	require.Equal(t, 1, hits)
}

func TestSubscriptionCancelIdempotent(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	sub, err := e.On("evt", func(context.Context, Event) error { return nil })
	require.NoError(t, err)
	sub.Cancel()
	sub.Cancel() // must be safe
}

func TestWildcardSubscriptionCancel(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var hits int
	sub, err := e.On("user.*", func(context.Context, Event) error {
		hits++
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, e.Emit(context.Background(), "user.created", nil))
	sub.Cancel()
	require.NoError(t, e.Emit(context.Background(), "user.created", nil))
	require.Equal(t, 1, hits)
}

func TestSubscriptionTopic(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	sub, err := e.On("user.*", func(context.Context, Event) error { return nil })
	require.NoError(t, err)
	require.Equal(t, "user.*", sub.Topic())
}

func TestConcurrentEmitAndCancelSafe(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var wg sync.WaitGroup
	subs := make([]Subscription, 0, 32)
	for range 32 {
		s, err := e.On("evt", func(context.Context, Event) error { return nil })
		require.NoError(t, err)
		subs = append(subs, s)
	}

	for range 32 {
		wg.Go(func() {
			for range 100 {
				_ = e.Emit(context.Background(), "evt", nil)
			}
		})
	}
	for _, s := range subs {
		wg.Go(s.Cancel)
	}
	wg.Wait()
}

func TestPriorityOrdering(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var order []string
	_, _ = e.On("evt", func(context.Context, Event) error {
		order = append(order, "low")
		return nil
	}, WithPriority(Low))
	_, _ = e.On("evt", func(context.Context, Event) error {
		order = append(order, "highest")
		return nil
	}, WithPriority(Highest))
	_, _ = e.On("evt", func(context.Context, Event) error {
		order = append(order, "normal")
		return nil
	})
	_, _ = e.On("evt", func(context.Context, Event) error {
		order = append(order, "high")
		return nil
	}, WithPriority(High))

	require.NoError(t, e.Emit(context.Background(), "evt", nil))
	require.Equal(t, []string{"highest", "high", "normal", "low"}, order)
}

func TestEqualPriorityRegistrationOrder(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var order []int
	for i := range 5 {
		_, err := e.On("evt", func(context.Context, Event) error {
			order = append(order, i)
			return nil
		})
		require.NoError(t, err)
	}

	require.NoError(t, e.Emit(context.Background(), "evt", nil))
	require.Equal(t, []int{0, 1, 2, 3, 4}, order)
}

func TestOnceFiresExactlyOnce(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var count int
	_, err := e.On("evt", func(context.Context, Event) error {
		count++
		return nil
	}, Once())
	require.NoError(t, err)

	for range 5 {
		require.NoError(t, e.Emit(context.Background(), "evt", nil))
	}
	require.Equal(t, 1, count)
}

func TestOnceConcurrentFiresAtMostOnce(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var fired atomic.Int32
	_, err := e.On("evt", func(context.Context, Event) error {
		fired.Add(1)
		return nil
	}, Once())
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range 32 {
		wg.Go(func() {
			_ = e.Emit(context.Background(), "evt", nil)
		})
	}
	wg.Wait()
	require.Equal(t, int32(1), fired.Load())
}

func TestErrorMessagesContainTopic(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	_, err := e.On("bad..topic", func(context.Context, Event) error { return nil })
	require.Error(t, err)
	require.Contains(t, fmt.Sprint(err), "bad..topic")
}
