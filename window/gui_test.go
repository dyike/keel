package window

import (
	"errors"
	"github.com/go-gui-org/go-gui/gui"
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/ui"
	"testing"
)

func TestManagerLifecycleAndIsolatedPages(t *testing.T) {
	app := gui.NewApp()
	m, err := NewManager(app)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()
	page := ui.NewPage("First", ui.Text("one"))
	first, err := m.New(Options{UI: page})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Host().WindowCleanup()
	if _, err := m.New(Options{UI: page}); !errors.Is(err, capability.ErrConflict) {
		t.Fatal(err)
	}
	other, err := m.New(Options{UI: ui.NewPage("Second", ui.Text("two"))})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Host().WindowCleanup()
	windows, err := m.BeginRun()
	if err != nil || len(windows) != 2 {
		t.Fatal(err, len(windows))
	}
	if _, err := m.BeginRun(); !errors.Is(err, capability.ErrConflict) {
		t.Fatal(err)
	}
	child, err := m.New(Options{UI: ui.NewPage("Child", ui.Text("three"))})
	if err != nil {
		t.Fatal(err)
	}
	if child.Host() != nil {
		t.Fatal("runtime window was created off the UI thread")
	}
	count := 0
	_ = child.OnReady(func() { count++ })
	cfg := <-app.PendingOpen()
	host := gui.NewWindow(cfg)
	defer host.WindowCleanup()
	cfg.OnInit(host)
	if child.Host() != host || count != 1 {
		t.Fatal("lost runtime ready notification")
	}
	m.Shutdown()
	if _, err := m.New(Options{}); !errors.Is(err, capability.ErrClosed) {
		t.Fatal(err)
	}
}
func TestBeforeReadyAndPendingCapacity(t *testing.T) {
	m, _ := NewManager(gui.NewApp())
	defer m.Shutdown()
	w, _ := m.New(Options{})
	defer w.Host().WindowCleanup()
	if err := w.Show(); !errors.Is(err, capability.ErrNotReady) {
		t.Fatal(err)
	}
	if err := w.SetClickThrough(true); !errors.Is(err, capability.ErrUnsupported) {
		t.Fatal(err)
	}
	_, _ = m.BeginRun()
	for i := 0; i < 16; i++ {
		if _, err := m.New(Options{}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.New(Options{}); !errors.Is(err, capability.ErrNotReady) {
		t.Fatal(err)
	}
}
func TestInvalidConfig(t *testing.T) {
	for _, o := range []Options{{Width: -1}, {Height: -1}, {GUI: gui.WindowCfg{Width: -1}}, {UI: ui.NewPage("x"), View: func(*gui.Window) gui.View { return nil }}} {
		if _, err := config(o); !errors.Is(err, capability.ErrInvalidArgument) {
			t.Fatal(err)
		}
	}
	if _, err := NewManager(nil); !errors.Is(err, capability.ErrInvalidArgument) {
		t.Fatal(err)
	}
}

func TestPageReservedAcrossManagers(t *testing.T) {
	one, _ := NewManager(gui.NewApp())
	defer one.Shutdown()
	two, _ := NewManager(gui.NewApp())
	defer two.Shutdown()
	page := ui.NewPage("shared", ui.Text("state"))
	first, err := one.New(Options{UI: page})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Host().WindowCleanup()
	if _, err := two.New(Options{UI: page}); !errors.Is(err, capability.ErrConflict) {
		t.Fatal(err)
	}
	// A failed duplicate request must not leave the second manager locked.
	other, err := two.New(Options{UI: ui.NewPage("other", ui.Text("isolated"))})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Host().WindowCleanup()
}
