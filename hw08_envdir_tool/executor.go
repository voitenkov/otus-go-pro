package main

import "os/exec"

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {

	command := exec.Command(cmd)

	return
}
