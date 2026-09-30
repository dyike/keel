package shortcut

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
)

type fakeBackend struct {
	driver.Backend
	register func(driver.Chord, func()) (func() error, error)
}

func (f fakeBackend) Register(c driver.Chord, h func()) (func() error, error) {
	return f.register(c, h)
}
func manager(f fakeBackend) *Manager {
	return &Manager{backend: f, entries: make(map[Chord]*Registration)}
}
func TestParse(t *testing.T) {
	for _, s := range []string{" CMD + Shift + K ", "super+shift+k", "k+command+shift"} {
		c, e := Parse(s)
		if e != nil || c.Key != "k" || c.Modifiers != Super|Shift {
			t.Fatalf("%q: %+v %v", s, c, e)
		}
	}
	for _, s := range []string{"", "k", "cmd", "cmd++k", "ctrl+control+k", "cmd+k+j"} {
		if _, e := Parse(s); !errors.Is(e, capability.ErrInvalidArgument) {
			t.Fatalf("accepted %q", s)
		}
	}
}
func TestLifecycle(t *testing.T) {
	var count atomic.Int32
	m := manager(fakeBackend{register: func(driver.Chord, func()) (func() error, error) {
		return func() error { count.Add(1); return nil }, nil
	}})
	r, e := m.Register("cmd+k", func() {})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Register("super+k", func() {}); !errors.Is(e, capability.ErrConflict) {
		t.Fatal(e)
	}
	if e = r.Close(); e != nil {
		t.Fatal(e)
	}
	if e = r.Close(); e != nil {
		t.Fatal(e)
	}
	if count.Load() != 1 {
		t.Fatal(count.Load())
	}
	if e = m.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Register("cmd+j", func() {}); !errors.Is(e, capability.ErrClosed) {
		t.Fatal(e)
	}
}
func TestCloseDuringRegister(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var count atomic.Int32
	m := manager(fakeBackend{register: func(driver.Chord, func()) (func() error, error) {
		close(started)
		<-release
		return func() error { count.Add(1); return nil }, nil
	}})
	result := make(chan error, 1)
	go func() { _, e := m.Register("cmd+k", func() {}); result <- e }()
	<-started
	if e := m.Close(); e != nil {
		t.Fatal(e)
	}
	close(release)
	if e := <-result; !errors.Is(e, capability.ErrClosed) {
		t.Fatal(e)
	}
	if count.Load() != 1 {
		t.Fatal("leaked hotkey")
	}
}
func TestCloseRetriesFailure(t *testing.T) {
	tries := 0
	m := manager(fakeBackend{register: func(driver.Chord, func()) (func() error, error) {
		return func() error {
			tries++
			if tries == 1 {
				return capability.ErrNative
			}
			return nil
		}, nil
	}})
	if _, e := m.Register("cmd+k", func() {}); e != nil {
		t.Fatal(e)
	}
	if e := m.Close(); !errors.Is(e, capability.ErrNative) {
		t.Fatal(e)
	}
	if e := m.Close(); e != nil {
		t.Fatal(e)
	}
	if tries != 2 {
		t.Fatal(tries)
	}
}
func TestConcurrentCloseDoesNotBlockUI(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	m := manager(fakeBackend{register: func(driver.Chord, func()) (func() error, error) {
		return func() error { close(entered); <-release; return nil }, nil
	}})
	r, e := m.Register("cmd+k", func() {})
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := r.Close(); e != nil {
			t.Error(e)
		}
	}()
	<-entered
	if e := r.Close(); !errors.Is(e, capability.ErrNotReady) {
		t.Error(e)
	}
	close(release)
	wg.Wait()
}
func TestFailedRegisterReleasesReservation(t *testing.T) {
	m := manager(fakeBackend{register: func(driver.Chord, func()) (func() error, error) { return nil, capability.ErrUnsupported }})
	for i := 0; i < 2; i++ {
		if _, e := m.Register("cmd+k", func() {}); !errors.Is(e, capability.ErrUnsupported) {
			t.Fatal(e)
		}
	}
}
