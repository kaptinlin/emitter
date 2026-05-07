package emitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithPrioritySetsPriority(t *testing.T) {
	t.Parallel()
	o := &listenerOpts{}
	WithPriority(High)(o)
	require.Equal(t, High, o.priority)
}

func TestOnceSetsOnceFlag(t *testing.T) {
	t.Parallel()
	o := &listenerOpts{}
	Once()(o)
	require.True(t, o.once)
}

func TestZeroOptsAreNormalAndNotOnce(t *testing.T) {
	t.Parallel()
	o := &listenerOpts{}
	require.Equal(t, Normal, o.priority)
	require.False(t, o.once)
}
