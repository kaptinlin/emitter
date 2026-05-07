package emitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventTopic(t *testing.T) {
	t.Parallel()
	ev := &event{topic: "user.created"}
	require.Equal(t, "user.created", ev.Topic())
}

func TestEventPayload(t *testing.T) {
	t.Parallel()
	ev := &event{payload: 42}
	require.Equal(t, 42, ev.Payload())
}

func TestEventPayloadNil(t *testing.T) {
	t.Parallel()
	ev := &event{}
	require.Nil(t, ev.Payload())
}

func TestEventStop(t *testing.T) {
	t.Parallel()
	ev := &event{}
	require.False(t, ev.stopped)
	ev.Stop()
	require.True(t, ev.stopped)
}

func TestEventStopIdempotent(t *testing.T) {
	t.Parallel()
	ev := &event{}
	ev.Stop()
	ev.Stop()
	ev.Stop()
	require.True(t, ev.stopped)
}
