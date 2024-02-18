package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	var envSlice []string
	command := exec.Command(cmd[0], cmd[1:]...)
	envSlice = command.Environ()
	for envName, envValue := range env {
		envSlice = append(envSlice, envName+"="+envValue.Value)
	}

	command.Env = envSlice
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		fmt.Printf("command.Start: %v", err)
		os.Exit(1)
	}

	fmt.Println(command.Environ())
	err = command.Wait()
	if err != nil {
		if exitCode, ok := err.(*exec.ExitError); ok {
			returnCode = exitCode.ExitCode()
			log.Printf("exit code: %d", returnCode)
		} else {
			log.Printf("command.Wait: %v", err)
			os.Exit(1)
		}
	}
	return returnCode
}
