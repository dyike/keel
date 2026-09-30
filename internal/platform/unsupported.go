package platform

import (
	"unsafe"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
)

type unsupported struct{}

func (unsupported) Check(driver.Permission, bool) (driver.Status, error) {
	return driver.NotGranted, capability.ErrUnsupported
}
func (unsupported) Displays() ([]driver.Display, error) { return nil, capability.ErrUnsupported }
func (unsupported) Capture(uint32) ([]byte, error)      { return nil, capability.ErrUnsupported }
func (unsupported) Position() (driver.Point, error)     { return driver.Point{}, capability.ErrUnsupported }
func (unsupported) Move(driver.Point) error             { return capability.ErrUnsupported }
func (unsupported) Click(uint8) error                   { return capability.ErrUnsupported }
func (unsupported) Key(string, bool) error              { return capability.ErrUnsupported }
func (unsupported) Register(driver.Chord, func()) (func() error, error) {
	return nil, capability.ErrUnsupported
}
func (unsupported) Window(unsafe.Pointer, string, int) error { return capability.ErrUnsupported }
