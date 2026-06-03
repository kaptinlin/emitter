package emitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidTopicName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		topic   string
		isValid bool
	}{
		{"single segment", "user", true},
		{"two segments", "user.created", true},
		{"alnum and underscore", "user_v2.created_at", true},
		{"hyphen allowed", "user-service.event-type", true},
		{"digits allowed", "user.v2.event42", true},
		{"single wildcard segment", "user.*", true},
		{"multi wildcard segment", "user.**", true},
		{"interior single wildcard", "user.*.created", true},
		{"interior multi wildcard", "user.**.created", true},
		{"just wildcards", "**", true},

		{"empty string", "", false},
		{"trailing dot", "user.", false},
		{"leading dot", ".user", false},
		{"double dot", "user..created", false},
		{"space in segment", "user .created", false},
		{"slash in segment", "user/created", false},
		{"unicode in segment", "用户.created", false},
		{"three stars", "user.***", false},
		{"star mixed in segment", "user.fo*o", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.isValid, isValidTopicName(tc.topic))
		})
	}
}

func TestHasWildcard(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		pattern string
		want    bool
	}{
		{"literal topic", "user.created", false},
		{"single wildcard", "user.*", true},
		{"multi wildcard", "**", true},
		{"interior multi wildcard", "a.**.b", true},
		{"many literal segments", "a.b.c", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, hasWildcard(tc.pattern))
		})
	}
}

func matchTopicPattern(pattern, subject string) bool {
	pp := splitTopic(pattern)
	sp := splitTopic(subject)
	return matchParts(pp, sp, 0, 0)
}

func splitTopic(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func TestMatchTopicPattern(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		pattern string
		subject string
		want    bool
	}{
		{"exact match", "user.created", "user.created", true},
		{"exact mismatch", "user.created", "user.deleted", false},
		{"single wildcard one segment", "user.*", "user.created", true},
		{"single wildcard wrong arity", "user.*", "user.created.v2", false},
		{"single wildcard at start", "*.created", "user.created", true},
		{"interior single wildcard", "user.*.v2", "user.created.v2", true},
		{"multi wildcard tail empty", "user.**", "user", true},
		{"multi wildcard tail one", "user.**", "user.created", true},
		{"multi wildcard tail many", "user.**", "user.a.b.c", true},
		{"multi wildcard tail mismatch prefix", "user.**", "admin.a.b.c", false},
		{"multi wildcard interior", "user.**.v2", "user.v2", true},
		{"multi wildcard interior with span", "user.**.v2", "user.a.b.v2", true},
		{"multi wildcard interior mismatch after span", "user.**.v2", "user.a.b.v3", false},
		{"multi wildcard alone matches all", "**", "anything.goes.here", true},
		{"multi wildcard alone matches single", "**", "x", true},
		{"single wildcard does not match zero", "user.*.v2", "user.v2", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, matchTopicPattern(tc.pattern, tc.subject))
		})
	}
}
