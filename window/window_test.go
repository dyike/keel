package window

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/dyike/keel/capability"
)

func TestNativeHandleResolvedInsideDispatch(t *testing.T) {
	dispatched := false
	d := NativeDriver{Dispatch: func(f func()) { dispatched = true; f(); dispatched = false }, Handle: func() unsafe.Pointer {
		if !dispatched {
			t.Fatal("handle read outside UI dispatch")
		}
		return nil
	}}
	w, e := New(d)
	if e != nil {
		t.Fatal(e)
	}
	if e = w.SetClickThrough(true); !errors.Is(e, capability.ErrNotReady) {
		t.Fatal(e)
	}
	if e = w.SetVibrancy(255); !errors.Is(e, capability.ErrInvalidArgument) {
		t.Fatal(e)
	}
}
func TestMissingDriver(t *testing.T) {
	if _, e := New(nil); !errors.Is(e, capability.ErrInvalidArgument) {
		t.Fatal(e)
	}
	if e := (NativeDriver{}).SetAlwaysOnTop(true); !errors.Is(e, capability.ErrInvalidArgument) {
		t.Fatal(e)
	}
}
