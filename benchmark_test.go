package emitter

import (
	"context"
	"strconv"
	"testing"
)

func BenchmarkEmitExactNoListener(b *testing.B) {
	e := New()
	defer e.Close()
	ctx := context.Background()
	for b.Loop() {
		_ = e.Emit(ctx, "user.created", nil)
	}
}

func BenchmarkEmitExactSingleListener(b *testing.B) {
	e := New()
	defer e.Close()
	_, _ = e.On("user.created", func(context.Context, Event) error { return nil })

	ctx := context.Background()
	for b.Loop() {
		_ = e.Emit(ctx, "user.created", nil)
	}
}

func BenchmarkEmitExactManyListeners(b *testing.B) {
	const n = 16
	e := New()
	defer e.Close()
	for range n {
		_, _ = e.On("user.created", func(context.Context, Event) error { return nil })
	}

	ctx := context.Background()
	for b.Loop() {
		_ = e.Emit(ctx, "user.created", nil)
	}
}

func BenchmarkEmitWildcard(b *testing.B) {
	e := New()
	defer e.Close()
	_, _ = e.On("user.**", func(context.Context, Event) error { return nil })

	ctx := context.Background()
	for b.Loop() {
		_ = e.Emit(ctx, "user.created", nil)
	}
}

func BenchmarkEmitMixedRouting(b *testing.B) {
	e := New()
	defer e.Close()
	for i := range 8 {
		topic := "exact." + strconv.Itoa(i)
		_, _ = e.On(topic, func(context.Context, Event) error { return nil })
	}
	for i := range 4 {
		_, _ = e.On("wild."+strconv.Itoa(i)+".*", func(context.Context, Event) error { return nil })
	}
	_, _ = e.On("**", func(context.Context, Event) error { return nil })

	ctx := context.Background()
	for b.Loop() {
		_ = e.Emit(ctx, "wild.2.evt", nil)
	}
}

func BenchmarkEmitConcurrent(b *testing.B) {
	e := New()
	defer e.Close()
	_, _ = e.On("evt", func(context.Context, Event) error { return nil })

	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = e.Emit(ctx, "evt", nil)
		}
	})
}

func BenchmarkOnAndCancel(b *testing.B) {
	e := New()
	defer e.Close()
	for b.Loop() {
		s, _ := e.On("evt", func(context.Context, Event) error { return nil })
		s.Cancel()
	}
}
