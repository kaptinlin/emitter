package emitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitAfterCloseWithZeroBufferStillReturnsError(t *testing.T) {
	t.Parallel()

	emitter := NewMemoryEmitter(WithErrChanBufferSize(0))
	require.NoError(t, emitter.Close())

	errChan := emitter.Emit("testTopic", nil)
	assert.Equal(t, 1, cap(errChan))

	err, ok := <-errChan
	require.True(t, ok)
	require.ErrorIs(t, err, ErrEmitterClosed)

	_, ok = <-errChan
	assert.False(t, ok)
}
