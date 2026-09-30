package platform

import (
	"errors"
	"testing"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
)

func TestUnsupported(t *testing.T) {
	b := unsupported{}
	_, a := b.Check(driver.Accessibility, false)
	_, c := b.Capture(1)
	_, d := b.Displays()
	_, p := b.Position()
	_, s := b.Register(driver.Chord{}, func() {})
	for _, e := range []error{a, c, d, p, s, b.Move(driver.Point{}), b.Click(0), b.Key("a", true), b.Window(nil, "top", 1)} {
		if !errors.Is(e, capability.ErrUnsupported) {
			t.Fatal(e)
		}
	}
}
