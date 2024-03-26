package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("dummy", func(t *testing.T) {
		var err error
		require.NoError(t, err)
	})
}
