// Package process provides terminal foreground process and working-directory
// queries without a UI dependency. macOS and Linux are supported; other builds
// return native.ErrUnsupported. Query results may change as processes exit.
package process

import (
	"fmt"
	"os"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

// ForegroundPID returns a terminal's foreground process group ID. On Unix this
// is the group leader PID, not necessarily the deepest running child process.
// The caller retains ownership of terminal and must keep it open during the call.
func ForegroundPID(terminal *os.File) (int, error) {
	if terminal == nil {
		return 0, fmt.Errorf("%w: nil terminal", native.ErrInvalidArgument)
	}
	return sys.ProcessForegroundPID(terminal.Fd())
}

// Directory reads the process's current working directory, without invoking a shell.
func Directory(pid int) (string, error) {
	if pid <= 0 || uint64(pid) > 2147483647 {
		return "", fmt.Errorf("%w: invalid PID", native.ErrInvalidArgument)
	}
	return sys.ProcessDirectory(pid)
}
