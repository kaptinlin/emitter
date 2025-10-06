package emitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseEvent(t *testing.T) {
	payload := map[string]string{"key": "value"} // Payload is a map
	event := NewBaseEvent("test_topic", payload)

	assert.Equal(t, "test_topic", event.Topic())

	retrievedPayload, ok := event.Payload().(map[string]string)
	require.True(t, ok, "Payload is not of type map[string]string")

	assert.Equal(t, "value", retrievedPayload["key"])
}

func TestBaseEventSetAbortedAndIsAborted(t *testing.T) {
	type Payload struct {
		Data string
	}

	event := NewBaseEvent("test_topic", Payload{Data: "some data"}) // Payload is a struct

	assert.False(t, event.IsAborted(), "Newly created event should not be aborted")

	event.SetAborted(true)
	assert.True(t, event.IsAborted(), "BaseEvent.Abort(true) did not abort the event")

	event.SetAborted(false)
	assert.False(t, event.IsAborted(), "BaseEvent.Abort(false) did not unabort the event")
}
