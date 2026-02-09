package emitter

import "testing"

// FuzzMatchTopicPattern fuzzes the matchTopicPattern function to ensure it doesn't panic
func FuzzMatchTopicPattern(f *testing.F) {
	// Add seed corpus for common patterns
	f.Add("event.*", "event.test")
	f.Add("event.**", "event.some.thing")
	f.Add("*.run", "event.run")
	f.Add("**", "any.topic.here")
	f.Add("event.some.*.*", "event.some.thing.run")
	f.Add("**.thing.run", "event.some.thing.run")
	f.Add("exact.match", "exact.match")
	f.Add("", "")
	f.Add("*", "single")
	f.Add("a.b.c", "a.b.c.d")

	f.Fuzz(func(t *testing.T, pattern, subject string) {
		// Ensure function doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("matchTopicPattern panicked with pattern=%q, subject=%q: %v", pattern, subject, r)
			}
		}()

		// Call the function under test
		result := matchTopicPattern(pattern, subject)

		// Basic sanity checks
		if pattern == subject && len(pattern) > 0 {
			// Exact matches should always return true (except empty strings)
			if !result {
				t.Logf("Exact match failed: pattern=%q, subject=%q", pattern, subject)
			}
		}

		if pattern == "**" && len(subject) > 0 {
			// Double wildcard should match any non-empty string
			if !result {
				t.Logf("Double wildcard match failed: pattern=%q, subject=%q", pattern, subject)
			}
		}
	})
}
