package emitter

import (
	"slices"
	"sync"
)

// Topic represents an event channel to which listeners can subscribe.
// It maintains listeners in priority-sorted order for efficient execution.
type Topic struct {
	Name              string
	mu                sync.RWMutex
	listeners         map[string]*listenerItem
	sortedListenerIDs []string
}

// NewTopic creates a new Topic with initialized listener storage.
func NewTopic() *Topic {
	return &Topic{
		listeners: make(map[string]*listenerItem),
	}
}

// addSortedListenerID inserts a listener ID at the correct position,
// maintaining descending priority order (highest first).
func (t *Topic) addSortedListenerID(id string, priority Priority) {
	index, _ := slices.BinarySearchFunc(t.sortedListenerIDs, priority, func(existingID string, target Priority) int {
		return int(target) - int(t.listeners[existingID].priority)
	})
	t.sortedListenerIDs = slices.Insert(t.sortedListenerIDs, index, id)
}

// removeSortedListenerID removes a listener ID from the sorted slice.
func (t *Topic) removeSortedListenerID(id string) {
	if idx := slices.Index(t.sortedListenerIDs, id); idx != -1 {
		t.sortedListenerIDs = slices.Delete(t.sortedListenerIDs, idx, idx+1)
	}
}

// AddListener adds a new listener to the topic with optional configuration.
func (t *Topic) AddListener(id string, listener Listener, opts ...ListenerOption) {
	t.mu.Lock()
	defer t.mu.Unlock()

	item := &listenerItem{
		listener: listener,
		priority: Normal,
	}

	for _, opt := range opts {
		opt(item)
	}

	t.listeners[id] = item
	t.addSortedListenerID(id, item.priority)
}

// RemoveListener removes a listener from the topic by its ID.
func (t *Topic) RemoveListener(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.listeners[id]; !ok {
		return ErrListenerNotFound
	}

	delete(t.listeners, id)
	t.removeSortedListenerID(id)

	return nil
}

// Trigger calls all listeners in priority order (highest first) and returns any errors.
// Execution stops if the event is aborted.
func (t *Topic) Trigger(event Event) []error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var errs []error
	for _, id := range t.sortedListenerIDs {
		item, ok := t.listeners[id]
		if !ok {
			continue
		}
		if err := item.listener(event); err != nil {
			errs = append(errs, err)
		}
		if event.IsAborted() {
			break
		}
	}
	return errs
}
