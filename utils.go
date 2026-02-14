package emitter

import "strings"

const (
	SingleWildcard = "*"  // Matches exactly one topic segment.
	MultiWildcard  = "**" // Matches zero or more topic segments.
)

// matchTopicPattern checks if the given subject matches the pattern with wildcards.
// It uses a recursive algorithm to match pattern segments against subject segments,
// supporting single wildcard (*) for one segment and multi-wildcard (**) for zero or more segments.
func matchTopicPattern(pattern, subject string) bool {
	// Fast path for exact match
	if pattern == subject {
		return true
	}

	// Fast path for simple wildcard without dots
	if !strings.Contains(pattern, ".") && !strings.Contains(subject, ".") {
		return pattern == SingleWildcard || pattern == MultiWildcard
	}

	// Special case: single wildcard matches an empty string
	if pattern == SingleWildcard && subject == "" {
		return true
	}

	patternParts := strings.Split(pattern, ".")
	subjectParts := strings.Split(subject, ".")

	// Handle the case where pattern ends with ".**", it should not match just "event"
	if len(patternParts) > 1 &&
		patternParts[len(patternParts)-1] == MultiWildcard &&
		len(subjectParts) == 1 &&
		subjectParts[0] == patternParts[0] {
		return false
	}

	var matchParts func(p, s int) bool
	matchParts = func(p, s int) bool {
		// If we've reached the end of pattern parts and subject parts simultaneously, it's a match.
		if p == len(patternParts) && s == len(subjectParts) {
			return true
		}
		// If we've reached the end of the subject but the pattern has remaining parts (other than '**'), it's not a match.
		if s == len(subjectParts) {
			for i := p; i < len(patternParts); i++ {
				if patternParts[i] != MultiWildcard {
					return false
				}
			}
			return true
		}
		// If we've reached the end of the pattern but not the subject, it's not a match.
		if p == len(patternParts) {
			return false
		}
		// Match based on the current part of the pattern.
		switch patternParts[p] {
		case SingleWildcard:
			// The single wildcard should match exactly one non-empty subject part.
			return s < len(subjectParts) && matchParts(p+1, s+1)
		case MultiWildcard:
			// '**' matches any number of subject parts.
			if p == len(patternParts)-1 {
				// If '**' is the last part in the pattern, it matches the rest of the subject.
				return true
			}
			// Try to match '**' with every possible subsequent part.
			// Using Go 1.22 range-over-int for cleaner, more idiomatic code
			for i := range len(subjectParts) - s + 1 {
				if matchParts(p+1, s+i) {
					return true
				}
			}
			return false
		default:
			// Exact match required for non-wildcard parts.
			return patternParts[p] == subjectParts[s] && matchParts(p+1, s+1)
		}
	}

	return matchParts(0, 0)
}

// isValidTopicName checks whether the given topic name is valid.
// Empty strings and strings containing regex-like characters are rejected.
func isValidTopicName(topicName string) bool {
	return topicName != "" && !strings.ContainsAny(topicName, "?[")
}
