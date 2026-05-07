package emitter

import (
	"strings"
	"testing"
)

func FuzzMatchTopicPattern(f *testing.F) {
	f.Add("user.created", "user.created")
	f.Add("user.*", "user.created")
	f.Add("user.**", "user.a.b.c")
	f.Add("**", "any.deep.path")
	f.Add("a.**.z", "a.b.c.z")
	f.Add("a.*.b", "a.x.b")

	f.Fuzz(func(t *testing.T, pattern, subject string) {
		// Skip inputs that violate the topic grammar — fuzzing is for the
		// matcher, not the validator.
		if !isValidTopicName(pattern) || !isValidTopicName(subject) {
			return
		}
		// Subject must have no wildcards to be a valid emit target.
		if strings.Contains(subject, "*") {
			return
		}

		// Property: matchTopicPattern is reflexive when pattern == subject and both literal.
		if pattern == subject {
			if !matchTopicPattern(pattern, subject) {
				t.Errorf("reflexivity broken: %q !~ %q", pattern, subject)
			}
		}

		// Property: ** alone matches any valid subject.
		if pattern == "**" {
			if !matchTopicPattern(pattern, subject) {
				t.Errorf("** failed to match %q", subject)
			}
		}
	})
}
