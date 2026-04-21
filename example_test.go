package emitter_test

import (
	"fmt"

	"github.com/kaptinlin/emitter"
)

func ExampleMemoryEmitter() {
	bus := emitter.NewMemoryEmitter()

	_, err := bus.On("user.created", func(evt emitter.Event) error {
		fmt.Printf("%s %v\n", evt.Topic(), evt.Payload())
		return nil
	})
	if err != nil {
		panic(err)
	}

	errs := bus.EmitSync("user.created", "alice@example.com")
	fmt.Printf("listener errors: %d\n", len(errs))

	// Output:
	// user.created alice@example.com
	// listener errors: 0
}
