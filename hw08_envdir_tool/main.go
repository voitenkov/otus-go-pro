package main

import (
	"fmt"
	"os"
)

func main() {
	commandWithArgs := os.Args[2:]
	envDirMap, _ := ReadDir(os.Args[1])
	for env, envValue := range envDirMap {
		os.Unsetenv(env)
		if envValue.NeedRemove {
			delete(envDirMap, env)
		} else {
			os.Setenv(env, envValue.Value)
		}
	}

	returnCode := RunCmd(commandWithArgs, envDirMap)
	fmt.Println(returnCode)
}
