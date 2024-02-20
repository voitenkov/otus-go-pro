package main

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("offset greater than file size should raise an error", func(t *testing.T) {
		err = Copy("testdata/input.txt", "out.txt", 10000, 1)
		require.Truef(t, errors.Is(err, ErrOffsetExceedsFileSize), "actual error %q", err)
	})

	t.Run("missing source file", func(t *testing.T) {
		err = Copy("testdata/input2.txt", "out.txt", 0, 0)
		require.Truef(t, errors.Is(err, ErrSourceFileNotFound), "actual error %q", err)
	})

	t.Run("source file with unknown length", func(t *testing.T) {
		err = Copy("/dev/urandom", "out.txt", 0, 0)
		require.Truef(t, errors.Is(err, ErrUnsupportedFile) || errors.Is(err, ErrSourceFileNotFound), "actual error %q", err)
	})

	t.Run("trying to create file with wrong symbols in filename", func(t *testing.T) {
		err = Copy("testdata/input.txt", "//", 0, 0)
		require.Truef(t, errors.Is(err, ErrCreatingDestinationFile), "actual error %q", err)
	})

	t.Run("negative offset", func(t *testing.T) {
		err = Copy("testdata/input.txt", "out.txt", -1, 1)
		require.Truef(t, errors.Is(err, ErrNegativeOffset), "actual error %q", err)
	})

	t.Run("negative limit", func(t *testing.T) {
		err = Copy("testdata/input.txt", "out.txt", 0, -1)
		require.Truef(t, errors.Is(err, ErrNegativeLimit), "actual error %q", err)
	})

	err = os.Remove("out.txt")
}
