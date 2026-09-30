package screen

import (
	"errors"
	"testing"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
)

type fake struct{ driver.Backend }

func (fake) Capture(id uint32) ([]byte, error) { return nil, capability.ErrPermissionDenied }
func TestCaptureErrors(t *testing.T) {
	m := &Manager{fake{}}
	if _, e := m.Capture(0); !errors.Is(e, capability.ErrInvalidArgument) {
		t.Fatal(e)
	}
	if _, e := m.Capture(1); !errors.Is(e, capability.ErrPermissionDenied) {
		t.Fatal(e)
	}
}
