package emitter

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

type userCreated struct {
	ID   string
	Name string
}

func TestSubscribeTypedDelivery(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var got userCreated
	_, err := Subscribe(e, "user.created", func(_ context.Context, _ Event, p userCreated) error {
		got = p
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, Publish(context.Background(), e, "user.created", userCreated{ID: "u1", Name: "Ada"}))
	if diff := cmp.Diff(userCreated{ID: "u1", Name: "Ada"}, got); diff != "" {
		t.Errorf("payload mismatch (-want +got):\n%s", diff)
	}
}

func TestSubscribeDoesNotMutateListenerOptions(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	options := make([]ListenerOption, 1, 2)
	options[0] = WithPriority(High)
	shared := options[:cap(options)]

	_, err := Subscribe(e, "evt", func(context.Context, Event, int) error {
		return nil
	}, options...)
	require.NoError(t, err)
	require.Nil(t, shared[1], "Subscribe must not append into caller-owned option storage")
}

func TestSubscribePayloadTypeMismatch(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	_, err := Subscribe(e, "evt", func(context.Context, Event, userCreated) error {
		t.Fatal("listener must not be invoked on type mismatch")
		return nil
	})
	require.NoError(t, err)

	emitErr := e.Emit(context.Background(), "evt", "not the right type")
	require.ErrorIs(t, emitErr, ErrPayloadType)
	require.ErrorContains(t, emitErr, `topic "evt" expected emitter.userCreated, got string: emitter: payload type mismatch`)
}

func TestSubscribeOnceTypeMismatchDoesNotConsumeListener(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var got int
	_, err := Subscribe(e, "evt", func(_ context.Context, _ Event, p int) error {
		got = p
		return nil
	}, Once())
	require.NoError(t, err)

	require.ErrorIs(t, Publish(context.Background(), e, "evt", "not the right type"), ErrPayloadType)
	require.Zero(t, got)

	require.NoError(t, Publish(context.Background(), e, "evt", 42))
	require.Equal(t, 42, got)

	require.NoError(t, Publish(context.Background(), e, "evt", 99))
	require.Equal(t, 42, got)
}

func TestSubscribeInterfacePayloadAcceptsNil(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	var called bool
	_, err := Subscribe[error](e, "evt", func(_ context.Context, _ Event, p error) error {
		called = true
		require.NoError(t, p)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, Publish[error](context.Background(), e, "evt", nil))
	require.True(t, called)
}

func TestSubscribeNilCallback(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	_, err := Subscribe[int](e, "evt", nil)
	require.ErrorIs(t, err, ErrNilListener)
}

func TestSubscribeRejectsInvalidPattern(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	_, err := Subscribe(e, "evt..bad", func(context.Context, Event, int) error {
		return nil
	})
	require.ErrorIs(t, err, ErrInvalidTopicName)
}

func TestPublishPropagatesEmitterErrors(t *testing.T) {
	t.Parallel()
	e := New()
	e.Close()

	require.ErrorIs(t, Publish(context.Background(), e, "evt", 1), ErrEmitterClosed)
}

func TestPublishIsTypeSafeOnEntry(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	type orderShipped struct{ Order int }

	got := make(chan orderShipped, 1)
	_, err := Subscribe(e, "order.shipped", func(_ context.Context, _ Event, p orderShipped) error {
		got <- p
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, Publish(context.Background(), e, "order.shipped", orderShipped{Order: 42}))
	if diff := cmp.Diff(orderShipped{Order: 42}, <-got); diff != "" {
		t.Errorf("payload mismatch (-want +got):\n%s", diff)
	}
}

func TestSubscribeListenerErrorPropagates(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	wantErr := errors.New("typed handler failed")
	_, err := Subscribe(e, "evt", func(context.Context, Event, int) error {
		return wantErr
	})
	require.NoError(t, err)

	require.ErrorIs(t, Publish(context.Background(), e, "evt", 1), wantErr)
}

func TestSubscribeWildcardPattern(t *testing.T) {
	t.Parallel()
	e := New()
	defer e.Close()

	got := make(chan int, 4)
	_, err := Subscribe(e, "metric.**", func(_ context.Context, _ Event, p int) error {
		got <- p
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, Publish(context.Background(), e, "metric.cpu", 1))
	require.NoError(t, Publish(context.Background(), e, "metric.mem.used", 2))
	close(got)
	var collected []int
	for v := range got {
		collected = append(collected, v)
	}
	if diff := cmp.Diff([]int{1, 2}, collected); diff != "" {
		t.Errorf("payloads mismatch (-want +got):\n%s", diff)
	}
}
