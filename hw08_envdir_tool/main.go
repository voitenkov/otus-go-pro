package main

import (
	"os"
	"strings"
)

func main() {
	commandWithArgs := os.Args[2:]
	dir := os.Args[1]
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	envDirMap, _ := ReadDir(dir)
	for env, envValue := range envDirMap {
		os.Unsetenv(env)
		if envValue.NeedRemove {
			delete(envDirMap, env)
		} else {
			os.Setenv(env, envValue.Value)
		}
	}

	returnCode := RunCmd(commandWithArgs, envDirMap)
	os.Exit(returnCode)
}
