// Basic emit and subscribe.
package main

import (
	"context"
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

	sub, err := e.On("user.created", func(_ context.Context, ev emitter.Event) error {
		fmt.Printf("topic=%s payload=%v\n", ev.Topic(), ev.Payload())
		return nil
	})
	if err != nil {
		return err
	}
	defer sub.Cancel()

	return e.Emit(context.Background(), "user.created", "alice@example.com")
}
