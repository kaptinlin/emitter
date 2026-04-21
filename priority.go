package emitter

// Priority controls listener execution order.
type Priority int

const (
	// Lowest is the lowest listener priority.
	Lowest Priority = 0
	// Low runs after Normal and above Lowest.
	Low Priority = 25
	// Normal is the default listener priority.
	Normal Priority = 50
	// High runs before Normal and below Highest.
	High Priority = 75
	// Highest is the highest listener priority.
	Highest Priority = 100
)

// IsValid reports whether the priority value is within the valid range.
func (p Priority) IsValid() bool {
	return p >= Lowest && p <= Highest
}
