// Package input exposes explicit input synthesis; it does not record input.
package input

import (
	"math"
	"strings"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
	"github.com/dyike/keel/internal/platform"
)

type Point = driver.Point
type Button uint8

const (
	Left Button = iota
	Right
	Middle
)

type Controller struct{ backend driver.Backend }

func New() *Controller                         { return &Controller{platform.New()} }
func (c *Controller) Position() (Point, error) { return c.backend.Position() }

// Move uses global logical desktop coordinates, matching screen.Display.Bounds.
func (c *Controller) Move(p Point) error {
	if math.IsNaN(p.X) || math.IsNaN(p.Y) || math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) {
		return capability.ErrInvalidArgument
	}
	return c.backend.Move(p)
}
func (c *Controller) Click(b Button) error {
	if b > Middle {
		return capability.ErrInvalidArgument
	}
	return c.backend.Click(uint8(b))
}

// KeyDown and KeyUp use physical US-layout key names, not Unicode text. Always pair them.
func (c *Controller) KeyDown(key string) error { return c.key(key, true) }
func (c *Controller) KeyUp(key string) error   { return c.key(key, false) }
func (c *Controller) key(key string, down bool) error {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return capability.ErrInvalidArgument
	}
	return c.backend.Key(key, down)
}
