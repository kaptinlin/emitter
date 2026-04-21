package emitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicAddListenerMaintainsPriorityOrdering(t *testing.T) {
	t.Parallel()

	topic := NewTopic()
	listener := func(Event) error { return nil }

	topic.AddListener("normal", listener)
	topic.AddListener("highest", listener, WithPriority(Highest))
	topic.AddListener("low", listener, WithPriority(Low))
	topic.AddListener("high", listener, WithPriority(High))

	assert.Equal(t, []string{"highest", "high", "normal", "low"}, topic.sortedListenerIDs)
}
