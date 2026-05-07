package emitter

// Priority controls listener execution order within a topic.
// Listeners with a higher numerical priority run first; ties run in registration order.
// Any int value is valid.
type Priority int

// Common priority sentinels. Callers may also use any other int value.
const (
	Lowest  Priority = -100
	Low     Priority = -50
	Normal  Priority = 0
	High    Priority = 50
	Highest Priority = 100
)
