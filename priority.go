package emitter

// Priority type for listener priority levels.
type Priority int

const (
	Lowest Priority = iota + 1 // Lowest priority.
	Low
	Normal
	High
	Highest
)

// IsValid checks if the priority value is within valid range.
func (p Priority) IsValid() bool {
	return p >= Lowest && p <= Highest
}
