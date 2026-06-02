package emitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewIgnoresNilOption(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		e := New(nil)
		require.NotNil(t, e)
	})
}
