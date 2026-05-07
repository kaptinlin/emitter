// Typed payloads via Subscribe[T] / Publish[T].
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kaptinlin/emitter"
)

type OrderShipped struct {
	OrderID string
	Carrier string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	e := emitter.New()
	defer e.Close()

	if _, err := emitter.Subscribe(e, "order.shipped",
		func(_ context.Context, _ emitter.Event, p OrderShipped) error {
			fmt.Printf("shipped order=%s via=%s\n", p.OrderID, p.Carrier)
			return nil
		}); err != nil {
		return err
	}

	return emitter.Publish(context.Background(), e, "order.shipped",
		OrderShipped{OrderID: "ord-7", Carrier: "FedEx"})
}
