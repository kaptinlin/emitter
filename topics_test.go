package emitter

import (
	"errors"
	"fmt"
	"sync"
	"testing"
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
	if topic == nil {
		t.Error("NewTopic() should not return nil")
	}
}

func TestAddRemoveListener(t *testing.T) {
	topic := NewTopic()
	listener1 := mockListener("1", false)
	listener2 := mockListener("2", false)

	id1 := "1"
	topic.AddListener(id1, listener1)
	if len(topic.listeners) != 1 {
		t.Error("AddListener() failed to add listener 1")
	}

	id2 := "2"
	topic.AddListener(id2, listener2)
	if len(topic.listeners) != 2 {
		t.Error("AddListener() failed to add listener 2")
	}

	if err := topic.RemoveListener(id1); err != nil {
		t.Errorf("RemoveListener() failed to remove listener 1: %v", err)
	}
	if len(topic.listeners) != 1 {
		t.Errorf("RemoveListener() failed to remove listener 1, remaining listeners: %d", len(topic.listeners))
	}

	if err := topic.RemoveListener(id2); err != nil {
		t.Errorf("RemoveListener() failed to remove listener 2: %v", err)
	}
	if len(topic.listeners) != 0 {
		t.Errorf("RemoveListener() failed to remove listener 2, remaining listeners: %d", len(topic.listeners))
	}
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
		if len(errors) != 1 {
			t.Errorf("Trigger() should return exactly 1 error, got: %d", len(errors))
		} else if errors[0].Error() != "listener error 2: listener error base" {
			t.Errorf("Trigger() should return 'listener error 2: listener error base', got: %s", errors[0].Error())
		}
	}()

	wg.Wait()
}
