//go:build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func configureRunProcess(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} }
func signalRunProcess(cmd *exec.Cmd, signal syscall.Signal) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	err := syscall.Kill(-cmd.Process.Pid, signal)
	if err == syscall.ESRCH {
		return os.ErrProcessDone
	}
	return err
}
func interruptRunProcess(cmd *exec.Cmd) error { return signalRunProcess(cmd, syscall.SIGINT) }
func killRunProcess(cmd *exec.Cmd) error      { return signalRunProcess(cmd, syscall.SIGKILL) }

func runStopSignals() []os.Signal { return []os.Signal{os.Interrupt, syscall.SIGTERM} }
