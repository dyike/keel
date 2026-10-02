//go:build windows

package main

import (
	"os/exec"
	"strconv"
)

func shellCommand(command string) *exec.Cmd { return exec.Command("cmd", "/C", command) }

// killTree kills the shell and everything it started, such as the program
// go run builds.
func killTree(cmd *exec.Cmd) {
	if exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run() != nil {
		cmd.Process.Kill()
	}
}
