package emitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrioritySentinelOrdering(t *testing.T) {
	t.Parallel()
	require.Less(t, Lowest, Low)
	require.Less(t, Low, Normal)
	require.Less(t, Normal, High)
	require.Less(t, High, Highest)
}

func TestPriorityIsPlainInt(t *testing.T) {
	t.Parallel()
	// Custom priorities outside the sentinel range are valid by construction.
	custom := Priority(1234)
	require.Greater(t, custom, Highest)
}
