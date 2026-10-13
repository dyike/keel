package tray

import (
	"errors"
	"image"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

type fakeTray struct {
	options sys.TrayOptions
	fail    error
	closes  int
}

func (f *fakeTray) Update(options sys.TrayOptions) error {
	if f.fail != nil {
		return f.fail
	}
	f.options = options
	return nil
}
func (f *fakeTray) Close() error { f.closes++; return nil }
func testTray(t *testing.T, options Options) (*Tray, *fakeTray) {
	t.Helper()
	options.Icon = image.NewNRGBA(image.Rect(0, 0, 16, 16))
	backend := new(fakeTray)
	tray, err := newTray(options, func(options sys.TrayOptions) (sys.TrayBackend, error) { backend.options = options; return backend, nil })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tray.Close() })
	return tray, backend
}
func receive(t *testing.T, ch <-chan string) string {
	t.Helper()
	select {
	case got := <-ch:
		return got
	case <-time.After(3 * time.Second):
		t.Fatal("callback did not run")
		return ""
	}
}

func TestMenuSnapshotsAndEvents(t *testing.T) {
	called := make(chan string, 8)
	items := []MenuItem{{ID: "parent", Title: "Parent", Children: []MenuItem{{ID: "child", Title: "Child", OnSelect: func() { called <- "child" }}}}, {Title: "Disabled", Disabled: true, OnSelect: func() { called <- "disabled" }}}
	tray, backend := testTray(t, Options{Menu: items, OnActivate: func() { called <- "activate" }})
	snapshot := backend.options
	parent, child := snapshot.Menu[0], snapshot.Menu[0].Children[0]
	if parent.ID == child.ID {
		t.Fatal("parent and child IDs collide")
	}
	items[0].Children[0].Title = "mutated"
	if snapshot.Menu[0].Children[0].Title != "Child" {
		t.Fatal("menu retained caller's slice")
	}
	snapshot.Event(child.ID)
	snapshot.Event(0)
	if receive(t, called) != "child" || receive(t, called) != "activate" {
		t.Fatal("events reordered")
	}
	if err := tray.SetMenu([]MenuItem{{Title: "New", OnSelect: func() { called <- "new" }}}); err != nil {
		t.Fatal(err)
	}
	snapshot.Event(child.ID) // stale menu event must not invoke a replacement action
	snapshot.Event(snapshot.Menu[1].ID)
	backend.options.Event(backend.options.Menu[0].ID)
	if receive(t, called) != "new" {
		t.Fatal("stale or disabled action invoked")
	}
}

func TestCallbackMayUpdateAndClose(t *testing.T) {
	done := make(chan string, 2)
	var tray *Tray
	var backend *fakeTray
	tray, backend = testTray(t, Options{OnActivate: func() {
		if err := tray.SetTooltip("Updated"); err != nil {
			done <- err.Error()
			return
		}
		if err := tray.Close(); err != nil {
			done <- err.Error()
			return
		}
		done <- "closed"
	}})
	backend.options.Event(0)
	if receive(t, done) != "closed" {
		t.Fatal("callback operations failed")
	}
	if err := tray.Close(); err != nil || backend.closes != 1 {
		t.Fatal("Close is not idempotent")
	}
	if err := tray.SetTooltip("closed"); !errors.Is(err, native.ErrFailed) {
		t.Fatal("closed tray accepted update")
	}
}

func TestCallbacksSerializeAndCloseDiscardsQueue(t *testing.T) {
	entered := make(chan string, 1)
	release := make(chan struct{})
	var calls atomic.Int32
	tray, backend := testTray(t, Options{OnActivate: func() { calls.Add(1); entered <- "entered"; <-release }})
	backend.options.Event(0)
	receive(t, entered)
	for range 30 {
		backend.options.Event(0)
	}
	if calls.Load() != 1 {
		t.Fatal("concurrent callbacks")
	}
	if err := tray.Close(); err != nil {
		t.Fatal(err)
	}
	close(release)
	// Posting after close must not create another callback.
	backend.options.Event(0)
	tray.mu.Lock()
	if len(tray.actions) != 0 {
		t.Fatal("retained closed actions")
	}
	tray.mu.Unlock()
}

func TestValidationAndFailedUpdatesPreserveMenu(t *testing.T) {
	tray, backend := testTray(t, Options{Menu: []MenuItem{{Title: "Old"}}})
	for _, menu := range [][]MenuItem{
		{{Title: ""}}, {{Title: "NUL\x00"}}, {{ID: "same", Title: "A"}, {ID: "same", Title: "B"}},
		{{Separator: true, Title: "invalid"}}, {{Title: "Parent", Children: []MenuItem{{Title: "child"}}, OnSelect: func() {}}},
	} {
		if err := tray.SetMenu(menu); !errors.Is(err, native.ErrInvalidArgument) {
			t.Fatalf("invalid menu accepted: %v", err)
		}
	}
	backend.fail = native.ErrFailed
	if err := tray.SetMenu([]MenuItem{{Title: "New"}}); err == nil {
		t.Fatal("backend error lost")
	}
	if backend.options.Menu[0].Title != "Old" || tray.options.Menu[0].Title != "Old" {
		t.Fatal("failed update replaced menu")
	}
	if err := tray.SetIcon(nil); !errors.Is(err, native.ErrInvalidArgument) {
		t.Fatal("nil icon accepted")
	}
	if err := tray.SetIcon(image.NewNRGBA(image.Rect(0, 0, 1025, 1))); !errors.Is(err, native.ErrInvalidArgument) {
		t.Fatal("oversized icon accepted")
	}
	if err := tray.SetTooltip("bad\x00"); !errors.Is(err, native.ErrInvalidArgument) {
		t.Fatal("NUL tooltip accepted")
	}
}

func TestConcurrentUpdates(t *testing.T) {
	tray, _ := testTray(t, Options{})
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				if err := tray.SetMenu([]MenuItem{{Title: "Item"}}); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
}
