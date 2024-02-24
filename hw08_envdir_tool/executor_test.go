package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("testing command with Return Code 1", func(t *testing.T) {
		returnCode := RunCmd([]string{"mkdir", "//"}, nil)
		require.Equal(t, 1, returnCode)
	})
}
