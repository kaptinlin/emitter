package emitter

import "strings"

const (
	singleWildcard = "*"
	multiWildcard  = "**"
)

// isValidTopicName reports whether s satisfies the topic grammar.
//
//	topic    := segment ('.' segment)*
//	segment  := name | wildcard
//	name     := [a-zA-Z0-9_-]+
//	wildcard := '*' | '**'
func isValidTopicName(s string) bool {
	if s == "" {
		return false
	}
	for seg := range strings.SplitSeq(s, ".") {
		if !isValidSegment(seg) {
			return false
		}
	}
	return true
}

func isValidSegment(s string) bool {
	if s == singleWildcard || s == multiWildcard {
		return true
	}
	if s == "" {
		return false
	}
	for _, r := range s {
		if !isNameByte(r) {
			return false
		}
	}
	return true
}

func isNameByte(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
	case r >= 'A' && r <= 'Z':
	case r >= '0' && r <= '9':
	case r == '_' || r == '-':
	default:
		return false
	}
	return true
}

// hasWildcard reports whether pattern contains a wildcard segment.
// Cheaper than splitting; a literal '*' character can only appear as a
// wildcard segment under the topic grammar.
func hasWildcard(pattern string) bool {
	return strings.Contains(pattern, "*")
}

// matchParts reports whether sp matches pattern parts pp from indices p, s.
// '*' matches exactly one segment; '**' matches zero or more segments.
func matchParts(pp, sp []string, p, s int) bool {
	if p == len(pp) {
		return s == len(sp)
	}
	if pp[p] == multiWildcard {
		// ** matches zero or more segments.
		for i := s; i <= len(sp); i++ {
			if matchParts(pp, sp, p+1, i) {
				return true
			}
		}
		return false
	}
	if s == len(sp) {
		return false
	}
	if pp[p] == singleWildcard {
		return matchParts(pp, sp, p+1, s+1)
	}
	if pp[p] == sp[s] {
		return matchParts(pp, sp, p+1, s+1)
	}
	return false
}
