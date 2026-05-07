package pool_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/emitter"
	"github.com/kaptinlin/emitter/pool"
)

func TestPoolDispatchesAsync(t *testing.T) {
	t.Parallel()
	p := pool.New(4, 16)
	defer p.Close()

	e := emitter.New()
	defer e.Close()

	done := make(chan struct{})
	_, err := e.On("evt", func(context.Context, emitter.Event) error {
		close(done)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, p.Submit(context.Background(), e, "evt", nil))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listener did not run within 1s")
	}
}

func TestPoolFullReturnsErrPoolFull(t *testing.T) {
	t.Parallel()
	// 1 worker, queue cap 1; the worker will block while we saturate the queue.
	p := pool.New(1, 1)
	defer p.Close()

	e := emitter.New()
	defer e.Close()

	hold := make(chan struct{})
	_, _ = e.On("evt", func(context.Context, emitter.Event) error {
		<-hold
		return nil
	})

	// First submit occupies the worker; second fills the queue.
	require.NoError(t, p.Submit(context.Background(), e, "evt", nil))
	require.NoError(t, p.Submit(context.Background(), e, "evt", nil))

	// Subsequent submits while both worker & queue are full must error.
	gotFull := false
	for range 16 {
		if err := p.Submit(context.Background(), e, "evt", nil); err != nil {
			require.ErrorIs(t, err, pool.ErrPoolFull)
			gotFull = true
			break
		}
	}
	require.True(t, gotFull, "expected ErrPoolFull while pool saturated")

	close(hold)
}

func TestPoolCloseWaitsForInFlight(t *testing.T) {
	t.Parallel()
	p := pool.New(2, 8)

	e := emitter.New()
	defer e.Close()

	var fired atomic.Int32
	_, _ = e.On("evt", func(context.Context, emitter.Event) error {
		time.Sleep(20 * time.Millisecond)
		fired.Add(1)
		return nil
	})

	for range 4 {
		require.NoError(t, p.Submit(context.Background(), e, "evt", nil))
	}
	p.Close()
	require.Equal(t, int32(4), fired.Load())
}

func TestPoolSwallowsListenerErrors(t *testing.T) {
	t.Parallel()
	// Listener errors are logged via slog, not surfaced to the submitter.
	p := pool.New(1, 4)
	defer p.Close()

	e := emitter.New()
	defer e.Close()

	var ran sync.WaitGroup
	ran.Add(1)
	_, _ = e.On("evt", func(context.Context, emitter.Event) error {
		defer ran.Done()
		return errSentinel
	})

	require.NoError(t, p.Submit(context.Background(), e, "evt", nil))
	ran.Wait()
}

var errSentinel = errSentinelType("listener said no")

type errSentinelType string

func (e errSentinelType) Error() string { return string(e) }
