package emitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchTopicPattern(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		// Exact matches
		{"event.some.thing.run", "event.some.thing.run", true},
		// Single node wildcard matches
		{"event.some.*.*", "event.some.thing.run", true},
		{"event.some.*.*", "event.some.thing.do", true},
		{"event.*", "event.some", true},
		{"event.*", "event.some.thing", false},
		{"event.some.*.run", "event.some.thing.run", true},
		// Single node wildcard non-matches
		{"event.some.*.run", "event.some.thing.do", false},
		{"event.*.thing.run", "event.some.thing.do", false},
		{"*.some.thing.run", "event.some.thing.do", false},
		{"event.some.*.run", "event.some.thing", false},
		// Multi-node wildcard matches
		{"event.some.**", "event.some.thing.run", true},
		{"event.some.**", "event.some.thing.do", true},
		{"**.thing.run", "event.some.thing.run", true},
		{"event.**", "event.some.thing.run", true},
		{"event.**.run", "event.some.thing.run", true},
		{"**", "event.some.thing.run", true},
		{"**", "event", true},
		{"**", "", true},
		// Multi-node wildcard non-matches
		{"event.**", "event", false},
		{"event.**.run", "event.some.thing.do", false},
		{"event.some.thing.**", "event.some.other.thing.run", false},
		{"**.thing.run", "event.some.thing.do", false},
		// Edge cases
		{"*", "", true},
		{"*", "event", true},
		{"*", "event.some", false},
		{"*", "event.some.thing", false},
		{"event.*", "event", false},
		{"event.*", "event.some", true},
		{"event.*", "event.some.thing", false},
		{"**", "", true},
		{"**", "event.some", true},
		{"**", "event.some.thing", true},
		{"**", "event.some.thing.do", true},
		{"**", "event", true},
		{"event.**", "event.some", true},
		{"event.**", "event.some.thing", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.subject, func(t *testing.T) {
			got := matchTopicPattern(tt.pattern, tt.subject)
			assert.Equal(t, tt.want, got, "matchTopicPattern(%q, %q)", tt.pattern, tt.subject)
		})
	}
}
