package input

import (
	"errors"
	"math"
	"testing"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
)

type fake struct {
	driver.Backend
	key  string
	down bool
}

func (f *fake) Key(k string, d bool) error { f.key = k; f.down = d; return nil }
func TestValidationAndKeyNormalization(t *testing.T) {
	f := &fake{}
	c := &Controller{f}
	for _, p := range []Point{{X: math.NaN(), Y: 0}, {X: 0, Y: math.Inf(1)}} {
		if e := c.Move(p); !errors.Is(e, capability.ErrInvalidArgument) {
			t.Fatal(e)
		}
	}
	if e := c.Click(255); !errors.Is(e, capability.ErrInvalidArgument) {
		t.Fatal(e)
	}
	if e := c.KeyDown("  "); !errors.Is(e, capability.ErrInvalidArgument) {
		t.Fatal(e)
	}
	if e := c.KeyDown(" ENTER "); e != nil || f.key != "enter" || !f.down {
		t.Fatal(e, f)
	}
	if e := c.KeyUp("enter"); e != nil || f.down {
		t.Fatal(e, f)
	}
}
