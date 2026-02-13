package emitter

import (
	"fmt"
	"testing"
)

// BenchmarkEmitSync benchmarks synchronous event emission with a single listener.
func BenchmarkEmitSync(b *testing.B) {
	e := NewMemoryEmitter()
	_, _ = e.On("bench.topic", func(evt Event) error {
		return nil
	})

	b.ResetTimer()
	for b.Loop() {
		e.EmitSync("bench.topic", "payload")
	}
}

// BenchmarkEmitAsync benchmarks asynchronous event emission with a single listener.
func BenchmarkEmitAsync(b *testing.B) {
	e := NewMemoryEmitter()
	_, _ = e.On("bench.topic", func(evt Event) error {
		return nil
	})

	b.ResetTimer()
	for b.Loop() {
		errChan := e.Emit("bench.topic", "payload")
		for range errChan {
		}
	}
}

// BenchmarkEmitSyncMultipleListeners benchmarks synchronous emission
// with multiple listeners at different priorities.
func BenchmarkEmitSyncMultipleListeners(b *testing.B) {
	e := NewMemoryEmitter()
	for _, p := range []Priority{Highest, High, Normal, Low, Lowest} {
		_, _ = e.On("bench.topic", func(evt Event) error {
			return nil
		}, WithPriority(p))
	}

	b.ResetTimer()
	for b.Loop() {
		e.EmitSync("bench.topic", "payload")
	}
}

// BenchmarkOnOff benchmarks listener registration and removal.
func BenchmarkOnOff(b *testing.B) {
	e := NewMemoryEmitter()
	listener := func(evt Event) error { return nil }

	b.ResetTimer()
	for b.Loop() {
		id, _ := e.On("bench.topic", listener)
		_ = e.Off("bench.topic", id)
	}
}

// BenchmarkMatchTopicPattern benchmarks wildcard topic pattern matching.
func BenchmarkMatchTopicPattern(b *testing.B) {
	patterns := []struct {
		name    string
		pattern string
		subject string
	}{
		{"exact", "event.user.created", "event.user.created"},
		{"single_wildcard", "event.*.created", "event.user.created"},
		{"multi_wildcard", "event.**", "event.user.created"},
		{"complex", "**.user.*", "app.event.user.created"},
	}

	for _, p := range patterns {
		b.Run(p.name, func(b *testing.B) {
			for b.Loop() {
				matchTopicPattern(p.pattern, p.subject)
			}
		})
	}
}

// BenchmarkNewBaseEvent benchmarks event creation.
func BenchmarkNewBaseEvent(b *testing.B) {
	for b.Loop() {
		NewBaseEvent("bench.topic", "payload")
	}
}

// BenchmarkEmitSyncWildcard benchmarks synchronous emission with
// wildcard pattern matching.
func BenchmarkEmitSyncWildcard(b *testing.B) {
	e := NewMemoryEmitter()
	_, _ = e.On("bench.**", func(evt Event) error {
		return nil
	})

	b.ResetTimer()
	for b.Loop() {
		e.EmitSync("bench.topic.deep", "payload")
	}
}

// BenchmarkEmitSyncParallel benchmarks synchronous emission under
// concurrent load.
func BenchmarkEmitSyncParallel(b *testing.B) {
	e := NewMemoryEmitter()
	_, _ = e.On("bench.topic", func(evt Event) error {
		return nil
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			e.EmitSync("bench.topic", "payload")
		}
	})
}

// BenchmarkTopicAddRemoveListener benchmarks adding and removing
// listeners directly on a Topic.
func BenchmarkTopicAddRemoveListener(b *testing.B) {
	topic := NewTopic()
	listener := func(evt Event) error { return nil }

	b.ResetTimer()
	for b.Loop() {
		id := fmt.Sprintf("listener-%d", b.N)
		topic.AddListener(id, listener, WithPriority(Normal))
		_ = topic.RemoveListener(id)
	}
}
