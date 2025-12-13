//go:build go1.25

package emitter

import (
	"testing"
	"testing/synctest"
	"time"
)

// Virtual time tests using synctest (Go 1.25+)
// These tests provide deterministic concurrent testing without relying on real time.

// TestPriorityOrderingWithDelays tests that listener priority is respected
// even when listeners have different execution delays.
func TestPriorityOrderingWithDelays(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		e := NewMemoryEmitter()

		var executionOrder []string

		_, _ = e.On("priority.test", func(evt Event) error {
			time.Sleep(50 * time.Millisecond)
			executionOrder = append(executionOrder, "high")
			return nil
		}, WithPriority(High))

		_, _ = e.On("priority.test", func(evt Event) error {
			time.Sleep(100 * time.Millisecond)
			executionOrder = append(executionOrder, "normal")
			return nil
		}, WithPriority(Normal))

		_, _ = e.On("priority.test", func(evt Event) error {
			time.Sleep(25 * time.Millisecond)
			executionOrder = append(executionOrder, "low")
			return nil
		}, WithPriority(Low))

		e.EmitSync("priority.test", nil)

		// Verify execution order respects priority (not delay time)
		if len(executionOrder) != 3 {
			t.Fatalf("Expected 3 listeners executed, got %d", len(executionOrder))
		}

		if executionOrder[0] != "high" || executionOrder[1] != "normal" || executionOrder[2] != "low" {
			t.Errorf("Unexpected execution order: %v (expected [high, normal, low])", executionOrder)
		}
	})
}

// TestConcurrentEmissions tests deterministic concurrent event emissions
// using virtual time to avoid flakiness.
func TestConcurrentEmissions(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		e := NewMemoryEmitter()

		counter := 0
		listener := func(evt Event) error {
			counter++
			return nil
		}

		_, _ = e.On("concurrent.test", listener)

		// Emit 100 events concurrently
		for range 100 {
			go func() {
				e.EmitSync("concurrent.test", "data")
			}()
		}

		// Virtual time makes this deterministic
		time.Sleep(100 * time.Millisecond)

		if counter != 100 {
			t.Errorf("Expected 100 events processed, got %d", counter)
		}
	})
}

// TestListenerRemovalDuringEmission tests thread-safe listener removal
// while events are being emitted concurrently.
func TestListenerRemovalDuringEmission(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		e := NewMemoryEmitter()

		executionCount := 0
		listener := func(evt Event) error {
			executionCount++
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		listenerID, _ := e.On("removal.test", listener)

		// Emit events before removal
		for range 3 {
			go func() {
				e.EmitSync("removal.test", nil)
			}()
		}

		time.Sleep(15 * time.Millisecond)

		// Remove the listener
		_ = e.Off("removal.test", listenerID)
		initialCount := executionCount

		// Emit events after removal - should not be processed
		for range 5 {
			go func() {
				e.EmitSync("removal.test", nil)
			}()
		}

		time.Sleep(50 * time.Millisecond)

		// Count should not have changed after removal
		if executionCount != initialCount {
			t.Errorf("Expected execution count to remain %d after removal, got %d", initialCount, executionCount)
		}
	})
}

// TestCloseWithPendingAsyncEvents tests proper cleanup when closing
// an emitter while async events are still processing.
func TestCloseWithPendingAsyncEvents(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		e := NewMemoryEmitter()

		listener := func(evt Event) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}

		_, _ = e.On("close.test", listener)

		// Start async emission
		errChan := e.Emit("close.test", "data")

		// Close emitter while event is processing
		time.Sleep(10 * time.Millisecond)
		_ = e.Close()

		// Should still complete the pending event
		for range errChan {
		}

		// Further emissions should fail
		errs := e.EmitSync("close.test", "more data")
		if len(errs) == 0 {
			t.Error("Expected error when emitting to closed emitter")
		}
	})
}
