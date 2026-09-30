// Package input synthesizes mouse and keyboard events. Everything except
// MousePosition needs the Accessibility permission.
package input

import (
	"math"
	"strings"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

type Button int

const (
	Left Button = iota
	Right
	Middle
)

// MousePosition uses the same coordinates as screen.Display.
func MousePosition() (x, y float64, err error) { return sys.MousePosition() }

func MouseMove(x, y float64) error {
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
		return native.ErrInvalidArgument
	}
	return sys.MouseMove(x, y)
}

// Click clicks at the current mouse position.
func Click(b Button) error {
	if b < Left || b > Middle {
		return native.ErrInvalidArgument
	}
	return sys.Click(int(b))
}

// KeyDown and KeyUp take US-layout key names such as "a", "enter", "f5", "cmd".
func KeyDown(key string) error { return sys.Key(strings.ToLower(key), true) }
func KeyUp(key string) error   { return sys.Key(strings.ToLower(key), false) }

// Tap presses and releases a key.
func Tap(key string) error {
	if err := KeyDown(key); err != nil {
		return err
	}
	return KeyUp(key)
}
