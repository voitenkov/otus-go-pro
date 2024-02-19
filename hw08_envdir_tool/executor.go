package main

import (
	"errors"
	"os"
	"os/exec"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	for envName, envValue := range env {
		if envValue.NeedRemove {
			os.Unsetenv(envName)
		} else {
			os.Setenv(envName, envValue.Value)
		}
	}

	command := exec.Command(cmd[0], cmd[1:]...)
	command.Env = os.Environ()
	for envName, envValue := range env {
		command.Env = append(command.Env, envName+"="+envValue.Value)
	}

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}

		return 1
	}

	return 0
}
