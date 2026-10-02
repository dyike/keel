//go:build (!darwin && !windows && !linux) || android || (darwin && !cgo)

package sys

import "github.com/dyike/keel/native"

func Permission(int, bool) (bool, error)                { return false, native.ErrUnsupported }
func Displays() ([]Display, error)                      { return nil, native.ErrUnsupported }
func Capture(uint32) ([]byte, error)                    { return nil, native.ErrUnsupported }
func MousePosition() (float64, float64, error)          { return 0, 0, native.ErrUnsupported }
func MouseMove(float64, float64) error                  { return native.ErrUnsupported }
func Click(int) error                                   { return native.ErrUnsupported }
func HasKey(string) bool                                { return false }
func Key(string, bool) error                            { return native.ErrUnsupported }
func Hotkey(string, uint, func()) (func() error, error) { return nil, native.ErrUnsupported }
