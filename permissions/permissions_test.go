package permissions

import (
	"errors"
	"testing"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
)

type fake struct {
	driver.Backend
	requested bool
}

func (f *fake) Check(k Kind, r bool) (Status, error) { f.requested = r; return NotGranted, nil }
func TestExplicitRequest(t *testing.T) {
	f := &fake{}
	m := &Manager{f}
	if _, e := m.Check(Accessibility); e != nil || f.requested {
		t.Fatal(e)
	}
	if _, e := m.Request(Accessibility); e != nil || !f.requested {
		t.Fatal(e)
	}
	if _, e := m.Check(255); !errors.Is(e, capability.ErrInvalidArgument) {
		t.Fatal(e)
	}
}
