package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

var ErrReadingDir = errors.New("error reading directory")

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	var fileName string

	dirEntries, err := os.ReadDir(dir)
	envDirMap := make(Environment)
	if err != nil {
		return nil, ErrReadingDir
	}
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			fileName = entry.Name()
			fileName = strings.Replace(fileName, "=", "", -1)

			fileInfo, err := os.Stat(dir + fileName)
			if err != nil {
				fmt.Printf("error reading info of file %v: %v\n", fileName, err)
				continue
			}

			if fileInfo.Size() == 0 {
				envDirMap[fileName] = EnvValue{Value: "", NeedRemove: true}
				continue
			}

			file, err := os.Open(dir + fileName)
			if err != nil {
				fmt.Printf("error opening file %v: %v\n", fileName, err)
				continue
			}

			line, err := readLine(file)
			if err != nil {
				fmt.Printf("error reading file %v: %v\n", fileName, err)
				continue
			}

			line = strings.TrimRight(line, " \t")
			envDirMap[entry.Name()] = EnvValue{Value: line, NeedRemove: false}
			file.Close()
		}
	}

	return envDirMap, nil
}

func readLine(reader io.Reader) (line string, err error) {
	var n int
	var lineSlice []byte
	rune := make([]byte, 1)
	for {
		n, err = reader.Read(rune)
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		if n > 0 {
			char := rune[0]
			// end of line or terminal zero
			if char == '\n' || char == 0x00 {
				break
			}
			lineSlice = append(lineSlice, char)
		}
	}

	return string(lineSlice), nil
}
