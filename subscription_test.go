package emitter

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionIsConcrete(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	sub, err := e.On("evt", func(context.Context, Event) error { return nil })
	require.NoError(t, err)
	require.NotNil(t, sub)
	require.NotEmpty(t, sub.Topic())
}

func TestSubscriptionCancelConcurrentSafe(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	sub, err := e.On("evt", func(context.Context, Event) error { return nil })
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range 16 {
		wg.Go(sub.Cancel)
	}
	wg.Wait()
}

func TestSubscriptionCancelOnUnknownPatternIsSafe(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	// Construct a subscription whose pattern was never registered.
	// Cancel must be a no-op rather than panicking.
	s := &subscription{emitter: e, pattern: "ghost.evt", id: 999}
	require.NotPanics(t, s.Cancel)

	s2 := &subscription{emitter: e, pattern: "ghost.*", id: 999}
	require.NotPanics(t, s2.Cancel)
}

func TestSubscriptionCancelAfterCloseRemovesListener(t *testing.T) {
	t.Parallel()

	for _, pattern := range []string{"evt", "**"} {
		t.Run(pattern, func(t *testing.T) {
			t.Parallel()
			e := New()

			sub := mustOn(t, e, pattern, func(context.Context, Event) error { return nil })
			require.Len(t, e.matchingListeners("evt"), 1)

			e.Close()
			sub.Cancel()

			require.Empty(t, e.matchingListeners("evt"))
			require.NotPanics(t, sub.Cancel)
		})
	}
}
