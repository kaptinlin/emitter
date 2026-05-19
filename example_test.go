package emitter_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/kaptinlin/emitter"
)

// Basic exact-topic subscription and emit.
func ExampleEmitter_On() {
	e := emitter.New()
	defer e.Close()

	if _, err := e.On("user.created", func(_ context.Context, ev emitter.Event) error {
		fmt.Printf("topic=%s payload=%v\n", ev.Topic(), ev.Payload())
		return nil
	}); err != nil {
		fmt.Println(err)
		return
	}

	if err := e.Emit(context.Background(), "user.created", "u-42"); err != nil {
		fmt.Println(err)
		return
	}
	// Output: topic=user.created payload=u-42
}

// Wildcards: '*' matches one segment, '**' matches zero or more.
func ExampleEmitter_wildcards() {
	e := emitter.New()
	defer e.Close()

	if _, err := e.On("user.*", func(_ context.Context, ev emitter.Event) error {
		fmt.Printf("user.*: %s\n", ev.Topic())
		return nil
	}); err != nil {
		fmt.Println(err)
		return
	}
	if _, err := e.On("metric.**", func(_ context.Context, ev emitter.Event) error {
		fmt.Printf("metric.**: %s\n", ev.Topic())
		return nil
	}); err != nil {
		fmt.Println(err)
		return
	}

	if err := e.Emit(context.Background(), "user.created", nil); err != nil {
		fmt.Println(err)
		return
	}
	if err := e.Emit(context.Background(), "metric.cpu.idle", nil); err != nil {
		fmt.Println(err)
		return
	}
	// Output:
	// user.*: user.created
	// metric.**: metric.cpu.idle
}

// Subscriptions are cancellable; Cancel is idempotent.
func ExampleSubscription_Cancel() {
	e := emitter.New()
	defer e.Close()

	sub, err := e.On("evt", func(context.Context, emitter.Event) error {
		fmt.Println("fired")
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := e.Emit(context.Background(), "evt", nil); err != nil {
		fmt.Println(err)
		return
	}
	sub.Cancel()
	if err := e.Emit(context.Background(), "evt", nil); err != nil {
		fmt.Println(err)
		return
	}
	// Output: fired
}

// Higher priority listeners run first; Stop halts remaining listeners.
func ExampleEvent_Stop() {
	e := emitter.New()
	defer e.Close()

	if _, err := e.On("evt", func(_ context.Context, ev emitter.Event) error {
		fmt.Println("guard")
		ev.Stop()
		return nil
	}, emitter.WithPriority(emitter.High)); err != nil {
		fmt.Println(err)
		return
	}

	if _, err := e.On("evt", func(context.Context, emitter.Event) error {
		fmt.Println("never runs")
		return nil
	}, emitter.WithPriority(emitter.Low)); err != nil {
		fmt.Println(err)
		return
	}

	if err := e.Emit(context.Background(), "evt", nil); err != nil {
		fmt.Println(err)
		return
	}
	// Output: guard
}

// Listener errors are joined and returned.
func ExampleEmitter_Emit_errors() {
	e := emitter.New()
	defer e.Close()

	if _, err := e.On("evt", func(context.Context, emitter.Event) error {
		return errors.New("boom")
	}); err != nil {
		fmt.Println(err)
		return
	}

	err := e.Emit(context.Background(), "evt", nil)
	fmt.Println(err)
	// Output: boom
}

// Subscribe[T] / Publish[T] adapt typed payloads.
func ExampleSubscribe() {
	type orderShipped struct{ ID string }

	e := emitter.New()
	defer e.Close()

	if _, err := emitter.Subscribe(e, "order.shipped",
		func(_ context.Context, _ emitter.Event, p orderShipped) error {
			fmt.Println("shipped:", p.ID)
			return nil
		}); err != nil {
		fmt.Println(err)
		return
	}

	if err := emitter.Publish(context.Background(), e, "order.shipped", orderShipped{ID: "ord-7"}); err != nil {
		fmt.Println(err)
		return
	}
	// Output: shipped: ord-7
}
