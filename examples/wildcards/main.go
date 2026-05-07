// Wildcards, priority, and Stop.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	e := emitter.New()
	defer e.Close()

	if _, err := e.On("**", func(_ context.Context, ev emitter.Event) error {
		fmt.Printf("audit: %s\n", ev.Topic())
		if ev.Topic() == "user.banned" {
			ev.Stop()
		}
		return nil
	}, emitter.WithPriority(emitter.Highest)); err != nil {
		return err
	}

	if _, err := e.On("user.*", func(_ context.Context, ev emitter.Event) error {
		fmt.Printf("user listener: %s\n", ev.Topic())
		return nil
	}); err != nil {
		return err
	}

	ctx := context.Background()
	return errors.Join(
		e.Emit(ctx, "user.created", nil),
		e.Emit(ctx, "user.banned", nil),
	)
}
