package emitter

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errListenerBase = errors.New("listener error base")
)

// mockListener simulates a listener function for testing.
func mockListener(id string, shouldError bool) Listener {
	return func(e Event) error {
		if shouldError {
			return fmt.Errorf("listener error %s: %w", id, errListenerBase)
		}
		return nil
	}
}

func TestNewTopic(t *testing.T) {
	topic := NewTopic()
	assert.NotNil(t, topic, "NewTopic() should not return nil")
}

func TestAddRemoveListener(t *testing.T) {
	topic := NewTopic()
	listener1 := mockListener("1", false)
	listener2 := mockListener("2", false)

	id1 := "1"
	topic.AddListener(id1, listener1)
	assert.Len(t, topic.listeners, 1, "AddListener() failed to add listener 1")

	id2 := "2"
	topic.AddListener(id2, listener2)
	assert.Len(t, topic.listeners, 2, "AddListener() failed to add listener 2")

	err := topic.RemoveListener(id1)
	require.NoError(t, err, "RemoveListener() failed to remove listener 1")
	assert.Len(t, topic.listeners, 1, "RemoveListener() failed to remove listener 1")

	err = topic.RemoveListener(id2)
	require.NoError(t, err, "RemoveListener() failed to remove listener 2")
	assert.Empty(t, topic.listeners, "RemoveListener() failed to remove listener 2")
}

func TestTriggerListeners(t *testing.T) {
	topic := NewTopic()

	type Payload struct {
		Data string
	}

	event := NewBaseEvent("test", Payload{Data: "value"}) // Assumes NewBaseEvent is modified to work without generics

	// Add listeners
	topic.AddListener("1", mockListener("1", false))
	topic.AddListener("2", mockListener("2", true))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		errors := topic.Trigger(event)
		require.Len(t, errors, 1, "Trigger() should return exactly 1 error")
		assert.Equal(t, "listener error 2: listener error base", errors[0].Error())
	}()

	wg.Wait()
}
