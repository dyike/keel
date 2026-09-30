//go:build darwin && !cgo

package keel

import (
	"errors"
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/ui"
	"testing"
)

func TestNoCGODesktopReturnsUnsupportedAndCleansUp(t *testing.T) {
	kit := New()
	w, err := kit.Window.New(WindowOptions{UI: ui.NewPage("test", ui.Text("test"))})
	if err != nil {
		t.Fatal(err)
	}
	if err := kit.Run(); !errors.Is(err, capability.ErrUnsupported) {
		t.Fatal(err)
	}
	if w.Host().Ctx().Err() == nil {
		t.Fatal("startup window leaked after unavailable desktop backend")
	}
}
