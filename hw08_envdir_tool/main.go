package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	commandWithArgs := os.Args[2:]
	dir := os.Args[1]
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	envDirMap, err := ReadDir(dir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	returnCode := RunCmd(commandWithArgs, envDirMap)
	os.Exit(returnCode)
}
