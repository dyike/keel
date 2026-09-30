package window

import (
	"errors"
	"testing"

	"github.com/go-gui-org/go-gui/gui"
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/ui"
)

func TestApplicationShortcutWhileInputFocusedAndUnregister(t *testing.T) {
	m, _ := NewManager(gui.NewApp())
	defer m.Shutdown()
	input := ui.Input(ui.InputOptions{Label: "Input"})
	page := ui.NewPage("test", input)
	win, err := m.New(Options{UI: page})
	if err != nil {
		t.Fatal(err)
	}
	host := win.Host()
	defer host.WindowCleanup()
	host.TestRender(page.View)
	host.SetFocus(input.ID())
	calls := 0
	close, err := win.RegisterShortcut("cmd+comma", func() { calls++ })
	if err != nil {
		t.Fatal(err)
	}
	press := func(mod gui.Modifier, repeat bool) {
		host.EventFn(&gui.Event{Type: gui.EventKeyDown, KeyCode: gui.KeyComma, Modifiers: mod, KeyRepeat: repeat})
	}
	press(gui.ModSuper, false)
	press(gui.ModSuper, true)
	press(gui.ModNone, false)
	if calls != 1 {
		t.Fatal("shortcut did not work while editing, or fired for a repeat/plain comma", calls)
	}
	if _, err := win.RegisterShortcut("super+,", func() {}); !errors.Is(err, capability.ErrConflict) {
		t.Fatal(err)
	}
	close()
	press(gui.ModSuper, false)
	if calls != 1 {
		t.Fatal("removed shortcut fired")
	}
	closeNew, err := win.RegisterShortcut("super+comma", func() { calls++ })
	if err != nil {
		t.Fatal(err)
	}
	defer closeNew()
	close()
	press(gui.ModSuper, false)
	if calls != 2 {
		t.Fatal("old cancellation removed new binding")
	}
}

func TestApplicationShortcutValidation(t *testing.T) {
	m, _ := NewManager(gui.NewApp())
	defer m.Shutdown()
	win, _ := m.New(Options{})
	for _, name := range []string{"comma", "super+unknown", "super+shift"} {
		if _, err := win.RegisterShortcut(name, func() {}); !errors.Is(err, capability.ErrInvalidArgument) {
			t.Fatal(name, err)
		}
	}
	if _, err := win.RegisterShortcut("super+comma", nil); !errors.Is(err, capability.ErrInvalidArgument) {
		t.Fatal(err)
	}
	win.Host().WindowCleanup()
	if _, err := win.RegisterShortcut("super+comma", func() {}); !errors.Is(err, capability.ErrClosed) {
		t.Fatal(err)
	}
}

func TestDispatchUsesSurvivingWindowAndDropsShutdownCallbacks(t *testing.T) {
	app := gui.NewApp()
	m, _ := NewManager(app)
	defer m.Shutdown()
	first, _ := m.New(Options{UI: ui.NewPage("first", ui.Text("one"))})
	defer first.Host().WindowCleanup()
	second, _ := m.New(Options{UI: ui.NewPage("second", ui.Text("two"))})
	defer second.Host().WindowCleanup()
	if err := m.Dispatch(nil); !errors.Is(err, capability.ErrInvalidArgument) {
		t.Fatal(err)
	}
	if err := m.Dispatch(func() {}); !errors.Is(err, capability.ErrNotReady) {
		t.Fatal(err)
	}
	_, _ = m.BeginRun()
	app.Register(1, first.Host())
	app.Register(2, second.Host())
	first.Host().WindowCleanup()
	app.Unregister(1)
	calls := 0
	done := make(chan error, 1)
	go func() { done <- m.Dispatch(func() { calls++ }) }()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("worker ran callback off UI thread")
	}
	second.Host().FrameFn()
	if calls != 1 {
		t.Fatal("did not dispatch on surviving window")
	}
	_ = m.Dispatch(func() { calls++ })
	m.Shutdown()
	second.Host().FrameFn()
	if calls != 1 {
		t.Fatal("callback fired after shutdown")
	}
}
