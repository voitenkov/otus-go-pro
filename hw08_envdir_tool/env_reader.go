package main

import "os"

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	var envDirMap Environment
	dirEntries, err := os.ReadDir(dir)
	for _, entry := range dirEntries {
		if !entry.IsDir() {

			// envDirMap[entry.Name()] =
		}
	}

	return nil, nil
}
