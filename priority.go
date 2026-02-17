package emitter

// Priority represents the execution order for event listeners.
// Higher values execute first.
type Priority int

const (
	Lowest  Priority = 0
	Low     Priority = 25
	Normal  Priority = 50
	High    Priority = 75
	Highest Priority = 100
)

// IsValid reports whether the priority value is within the valid range.
func (p Priority) IsValid() bool {
	return p >= Lowest && p <= Highest
}
