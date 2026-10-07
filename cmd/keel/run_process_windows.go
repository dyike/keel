package main

import (
	"os"
	"os/exec"
	"strconv"
)

func configureRunProcess(cmd *exec.Cmd)       {}
func interruptRunProcess(cmd *exec.Cmd) error { return killRunProcess(cmd) }
func killRunProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	// taskkill includes compiler children; Process.Kill alone leaves them running.
	if err := exec.Command("taskkill", "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F").Run(); err != nil {
		return cmd.Process.Kill()
	}
	return nil
}

func runStopSignals() []os.Signal { return []os.Signal{os.Interrupt} }
