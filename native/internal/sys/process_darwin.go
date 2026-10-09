//go:build darwin && !ios

package sys

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/dyike/keel/native"
	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

func ProcessForegroundPID(fd uintptr) (int, error) {
	pid, err := unix.IoctlGetInt(int(fd), unix.TIOCGPGRP)
	if err != nil {
		return 0, fmt.Errorf("%w: foreground process: %v", native.ErrFailed, err)
	}
	return pid, nil
}

// struct proc_vnodepathinfo: the current directory's vnode_info (152 bytes),
// then its path; the root directory follows.
const (
	procPIDVnodePathInfo   = 9
	procVnodePathInfoSize  = 2352
	procVnodePathCwdOffset = 152
	maxPathLen             = 1024
)

func ProcessDirectory(pid int) (string, error) {
	load()
	var info [procVnodePathInfoSize]byte
	n, _, errno := purego.SyscallN(mustSym("proc_pidinfo"), uintptr(pid), procPIDVnodePathInfo, 0, uintptr(unsafe.Pointer(&info[0])), procVnodePathInfoSize)
	if int32(n) != procVnodePathInfoSize {
		return "", fmt.Errorf("%w: process directory: %v", native.ErrFailed, syscall.Errno(errno))
	}
	path := info[procVnodePathCwdOffset : procVnodePathCwdOffset+maxPathLen]
	for i, b := range path {
		if b == 0 {
			path = path[:i]
			break
		}
	}
	return string(path), nil
}
