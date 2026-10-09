//go:build !(darwin && !ios) && !(linux && !android)

package sys

import "github.com/dyike/keel/native"

func ProcessForegroundPID(uintptr) (int, error) { return 0, native.ErrUnsupported }
func ProcessDirectory(int) (string, error)      { return "", native.ErrUnsupported }
