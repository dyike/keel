// Package driver is the platform contract; public packages never select an OS.
package driver

import "unsafe"

type Point struct{ X, Y float64 }
type Rect struct{ X, Y, Width, Height float64 }
type Display struct {
	ID                      uint32
	Bounds                  Rect
	PixelWidth, PixelHeight int
	Primary                 bool
}
type Permission uint8

const (
	Accessibility Permission = iota
	ScreenRecording
	InputMonitoring
)

type Status uint8

const (
	NotGranted Status = iota
	Granted
)

type Modifiers uint8

const (
	Control Modifiers = 1 << iota
	Alt
	Shift
	Super
)

type Chord struct {
	Key       string
	Modifiers Modifiers
}
type Backend interface {
	Check(Permission, bool) (Status, error)
	Displays() ([]Display, error)
	Capture(uint32) ([]byte, error)
	Position() (Point, error)
	Move(Point) error
	Click(uint8) error
	Key(string, bool) error
	Register(Chord, func()) (func() error, error)
	Window(unsafe.Pointer, string, int) error
}
