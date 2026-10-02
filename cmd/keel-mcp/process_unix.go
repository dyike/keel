//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// shellCommand runs command in its own process group, so killTree also
// kills the children go run starts.
func shellCommand(command string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

func killTree(cmd *exec.Cmd) { syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
