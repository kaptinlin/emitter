package emitter

import (
	"cmp"
	"slices"
	"sync"
)

// Topic stores the listeners registered for an event pattern.
type Topic struct {
	// Name is the topic or pattern associated with the topic.
	Name              string
	mu                sync.RWMutex
	listeners         map[string]*listenerItem
	sortedListenerIDs []string
}

// NewTopic returns an empty Topic.
func NewTopic() *Topic {
	return &Topic{
		listeners: make(map[string]*listenerItem),
	}
}

// AddListener registers listener under id.
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
	index, _ := slices.BinarySearchFunc(t.sortedListenerIDs, item.priority, func(existingID string, target Priority) int {
		return cmp.Compare(int(target), int(t.listeners[existingID].priority))
	})
	t.sortedListenerIDs = slices.Insert(t.sortedListenerIDs, index, id)
}

// RemoveListener removes the listener identified by id.
func (t *Topic) RemoveListener(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.listeners[id]; !ok {
		return ErrListenerNotFound
	}

	delete(t.listeners, id)
	if idx := slices.Index(t.sortedListenerIDs, id); idx != -1 {
		t.sortedListenerIDs = slices.Delete(t.sortedListenerIDs, idx, idx+1)
	}

	return nil
}

// Trigger delivers event in priority order until all listeners run or event is aborted.
func (t *Topic) Trigger(event Event) []error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var errs []error
	for _, id := range t.sortedListenerIDs {
		item, ok := t.listeners[id]
		if !ok {
			continue
		}
		err := func() (err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					panicErr := &PanicError{Value: recovered}
					if cause, ok := recovered.(error); ok {
						panicErr.Cause = cause
					}
					err = panicErr
				}
			}()
			return item.listener(event)
		}()
		if err != nil {
			errs = append(errs, err)
		}
		if event.IsAborted() {
			break
		}
	}
	return errs
}
