package main

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/window"
	"github.com/dyike/keel/ui/window/windowtest"
)

func settleHotkey(t *testing.T, app *windowtest.App, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		app.Render()
		if done() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("native operation did not settle")
}
func TestHotkeyPreviewLifecycle(t *testing.T) {
	d := newHotkeyDemo()
	var released atomic.Int32
	var callback func()
	d.register = func(_ string, fn func()) (func() error, error) {
		callback = fn
		return func() error { released.Add(1); return nil }, nil
	}
	app := windowtest.New(t, window.Options{Width: 640, Height: 480, Content: el.Root(d)})
	if d.unregister != nil || d.busy {
		t.Fatal("render registered a shortcut")
	}
	core.Update(d.toggle)
	settleHotkey(t, app, func() bool { return !d.busy && d.unregister != nil })
	callback()
	settleHotkey(t, app, func() bool { return d.hits == 1 })
	core.Update(d.close)
	settleHotkey(t, app, func() bool { return released.Load() == 1 })
	callback()
	app.Render()
	if d.hits != 1 {
		t.Fatal("late callback updated closed preview")
	}
}
func TestHotkeyLateRegistrationIsReleased(t *testing.T) {
	d := newHotkeyDemo()
	gate := make(chan struct{})
	var released atomic.Int32
	d.register = func(string, func()) (func() error, error) {
		<-gate
		return func() error { released.Add(1); return nil }, nil
	}
	app := windowtest.New(t, window.Options{Content: el.Root(d)})
	core.Update(d.toggle)
	app.Render()
	core.Update(d.close)
	app.Render()
	close(gate)
	settleHotkey(t, app, func() bool { return released.Load() == 1 })
	if d.unregister != nil {
		t.Fatal("late registration attached to closed preview")
	}
}
func TestHotkeyFailureAllowsRetry(t *testing.T) {
	d := newHotkeyDemo()
	d.register = func(string, func()) (func() error, error) { return nil, errors.New("shortcut occupied") }
	app := windowtest.New(t, window.Options{Content: el.Root(d)})
	core.Update(d.toggle)
	settleHotkey(t, app, func() bool { return !d.busy && d.status != "" })
	app.Want("shortcut occupied")
	if d.unregister != nil {
		t.Fatal("failed registration enabled shortcut")
	}
}
