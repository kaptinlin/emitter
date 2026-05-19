package emitter

import (
	"context"
	"strconv"
	"sync"
	"testing"
)

func mustBenchmarkOn(b *testing.B, e *Emitter, pattern string, listener Listener, opts ...ListenerOption) Subscription {
	b.Helper()
	sub, err := e.On(pattern, listener, opts...)
	if err != nil {
		b.Fatal(err)
	}
	return sub
}

func BenchmarkEmitExactNoListener(b *testing.B) {
	e := New()
	defer e.Close()
	ctx := context.Background()
	for b.Loop() {
		if err := e.Emit(ctx, "user.created", nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEmitExactSingleListener(b *testing.B) {
	e := New()
	defer e.Close()
	mustBenchmarkOn(b, e, "user.created", func(context.Context, Event) error { return nil })

	ctx := context.Background()
	for b.Loop() {
		if err := e.Emit(ctx, "user.created", nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEmitExactManyListeners(b *testing.B) {
	const n = 16
	e := New()
	defer e.Close()
	for range n {
		mustBenchmarkOn(b, e, "user.created", func(context.Context, Event) error { return nil })
	}

	ctx := context.Background()
	for b.Loop() {
		if err := e.Emit(ctx, "user.created", nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEmitWildcard(b *testing.B) {
	e := New()
	defer e.Close()
	mustBenchmarkOn(b, e, "user.**", func(context.Context, Event) error { return nil })

	ctx := context.Background()
	for b.Loop() {
		if err := e.Emit(ctx, "user.created", nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEmitMixedRouting(b *testing.B) {
	e := New()
	defer e.Close()
	for i := range 8 {
		topic := "exact." + strconv.Itoa(i)
		mustBenchmarkOn(b, e, topic, func(context.Context, Event) error { return nil })
	}
	for i := range 4 {
		mustBenchmarkOn(b, e, "wild."+strconv.Itoa(i)+".*", func(context.Context, Event) error { return nil })
	}
	mustBenchmarkOn(b, e, "**", func(context.Context, Event) error { return nil })

	ctx := context.Background()
	for b.Loop() {
		if err := e.Emit(ctx, "wild.2.evt", nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEmitConcurrent(b *testing.B) {
	e := New()
	defer e.Close()
	mustBenchmarkOn(b, e, "evt", func(context.Context, Event) error { return nil })

	ctx := context.Background()
	var (
		errMu    sync.Mutex
		firstErr error
	)
	recordErr := func(err error) {
		errMu.Lock()
		defer errMu.Unlock()
		if firstErr == nil {
			firstErr = err
		}
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := e.Emit(ctx, "evt", nil); err != nil {
				recordErr(err)
				return
			}
		}
	})
	if firstErr != nil {
		b.Fatal(firstErr)
	}
}

func BenchmarkOnAndCancel(b *testing.B) {
	e := New()
	defer e.Close()
	for b.Loop() {
		s, err := e.On("evt", func(context.Context, Event) error { return nil })
		if err != nil {
			b.Fatal(err)
		}
		s.Cancel()
	}
}
