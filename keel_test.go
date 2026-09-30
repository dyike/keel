package keel

import (
	"errors"
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/ui"
	"testing"
)

func TestAttachNil(t *testing.T) {
	if _, err := Attach(nil); !errors.Is(err, capability.ErrInvalidArgument) {
		t.Fatal(err)
	}
}
func TestRunNeedsWindow(t *testing.T) {
	k := New()
	defer k.Close()
	if err := k.Run(); !errors.Is(err, capability.ErrNotReady) {
		t.Fatal(err)
	}
	if err := k.Quit(); !errors.Is(err, capability.ErrNotReady) {
		t.Fatal(err)
	}
}
func TestOverlayRejectsUnsupportedTop(t *testing.T) {
	k := New()
	defer k.Close()
	if _, err := k.NewOverlay(OverlayOptions{AlwaysOnTop: true}); !errors.Is(err, capability.ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err := k.NewOverlay(OverlayOptions{Vibrancy: 255, UI: ui.NewPage("x")}); !errors.Is(err, capability.ErrInvalidArgument) {
		t.Fatal(err)
	}
}
