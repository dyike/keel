//go:build linux && !android

package sys

import (
	"fmt"
	"os"

	"github.com/dyike/keel/native"
	"golang.org/x/sys/unix"
)

func ProcessForegroundPID(fd uintptr) (int, error) {
	pid, err := unix.IoctlGetInt(int(fd), unix.TIOCGPGRP)
	if err != nil {
		return 0, fmt.Errorf("%w: foreground process: %v", native.ErrFailed, err)
	}
	return pid, nil
}
func ProcessDirectory(pid int) (string, error) {
	dir, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	if err != nil {
		return "", fmt.Errorf("%w: process directory: %v", native.ErrFailed, err)
	}
	return dir, nil
}
