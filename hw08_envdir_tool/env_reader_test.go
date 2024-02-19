package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testReadDir         = "testdata2"
	testReadFileName    = "ENV=1"
	testReadExpectedEnv = "ENV1"
	testFile            *os.File
)

func TestReadDir(t *testing.T) {
	t.Run("Filename contains '=' character", func(t *testing.T) {
		err := os.Mkdir(testReadDir, 0o777)
		if err != nil {
			panic(err)
		}

		testFile, err = os.Create(testReadDir + "/" + testReadFileName)
		if err != nil {
			os.Remove(testReadDir)
			panic(err)
		}
		defer testReadDirCleanUp()
		testFile.Close()

		envMap, err := ReadDir(testReadDir)
		require.NoError(t, err)
		envValue, ok := envMap[testReadExpectedEnv]
		require.True(t, ok)
		require.Equal(t, "", envValue.Value)
		require.Equal(t, true, envValue.NeedRemove)
	})
}

func testReadDirCleanUp() {
	os.Remove(testReadDir + "/" + testReadFileName)
	os.Remove(testReadDir)
}
