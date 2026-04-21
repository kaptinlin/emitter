package emitter

import "strings"

const (
	// SingleWildcard matches exactly one topic segment.
	SingleWildcard = "*"
	// MultiWildcard matches zero or more topic segments.
	MultiWildcard = "**"
)

// matchTopicPattern reports whether subject matches pattern.
func matchTopicPattern(pattern, subject string) bool {
	if pattern == subject {
		return true
	}

	// No dots: only wildcards can match a different subject
	if !strings.Contains(pattern, ".") && !strings.Contains(subject, ".") {
		return pattern == SingleWildcard || pattern == MultiWildcard
	}

	patternParts := strings.Split(pattern, ".")
	subjectParts := strings.Split(subject, ".")

	// "event.**" should not match bare "event"
	if len(patternParts) > 1 &&
		patternParts[len(patternParts)-1] == MultiWildcard &&
		len(subjectParts) == 1 &&
		subjectParts[0] == patternParts[0] {
		return false
	}

	var matchParts func(p, s int) bool
	matchParts = func(p, s int) bool {
		if p == len(patternParts) && s == len(subjectParts) {
			return true
		}

		// Subject exhausted: remaining pattern parts must all be "**"
		if s == len(subjectParts) {
			for i := range len(patternParts) - p {
				if patternParts[p+i] != MultiWildcard {
					return false
				}
			}
			return true
		}

		if p == len(patternParts) {
			return false
		}

		switch patternParts[p] {
		case SingleWildcard:
			return s < len(subjectParts) && matchParts(p+1, s+1)
		case MultiWildcard:
			if p == len(patternParts)-1 {
				return true
			}
			for i := range len(subjectParts) - s + 1 {
				if matchParts(p+1, s+i) {
					return true
				}
			}
			return false
		default:
			return patternParts[p] == subjectParts[s] && matchParts(p+1, s+1)
		}
	}

	return matchParts(0, 0)
}

// isValidTopicName reports whether topicName is valid.
func isValidTopicName(topicName string) bool {
	return topicName != "" && !strings.ContainsAny(topicName, "?[")
}
