// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !nox11

package app

import (
	"testing"
	"time"

	syscall "golang.org/x/sys/unix"
)

// newTestX11Window returns a window with a notify pipe and no X connection.
func newTestX11Window(t *testing.T) *x11Window {
	t.Helper()
	pipe := make([]int, 2)
	if err := syscall.Pipe2(pipe, syscall.O_NONBLOCK|syscall.O_CLOEXEC); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		syscall.Close(pipe[0])
		syscall.Close(pipe[1])
	})
	w := &x11Window{w: &callbacks{w: new(Window)}, done: make(chan struct{})}
	w.notify.read, w.notify.write = pipe[0], pipe[1]
	return w
}

func pipeReadable(t *testing.T, fd int) bool {
	t.Helper()
	fds := []syscall.PollFd{{Fd: int32(fd), Events: syscall.POLLIN}}
	if _, err := syscall.Poll(fds, 0); err != nil {
		t.Fatal(err)
	}
	return fds[0].Revents&syscall.POLLIN != 0
}

// Xlib calls on another goroutine can move the event loop's events into
// Xlib's queue while it waits in poll, so a window never saw its close
// request. Run must leave Xlib to the event loop: queue f, wake poll, and
// run f from Event.
func TestX11RunQueuesForEventLoop(t *testing.T) {
	w := newTestX11Window(t)
	direct := false
	w.Run(func() { direct = true })
	if !direct {
		t.Fatal("Run before the event loop starts must call f directly")
	}

	w.started.Store(true)
	ran := false
	w.Run(func() {
		ran = true
		w.ProcessEvent(wakeupEvent{})
	})
	if ran {
		t.Fatal("Run executed f on the caller's goroutine")
	}
	if !pipeReadable(t, w.notify.read) {
		t.Fatal("Run did not wake the event loop's poll")
	}
	if _, ok := w.Event().(wakeupEvent); !ok || !ran {
		t.Fatal("Event did not run the queued function")
	}
}

// Destroying the window releases Window.Run callers and stops wake-ups
// before the notify pipe closes.
func TestX11RunAfterStop(t *testing.T) {
	w := newTestX11Window(t)
	w.started.Store(true)
	win := w.w.w
	win.driver = w
	w.stopRun()
	ran := make(chan bool, 1)
	go func() {
		win.Run(func() {})
		ran <- true
	}()
	select {
	case <-ran:
	case <-time.After(5 * time.Second):
		t.Fatal("Window.Run waited for a destroyed window")
	}
	if pipeReadable(t, w.notify.read) {
		t.Fatal("Run woke a destroyed window through its notify pipe")
	}
	w.stopRun() // destroy may follow shutdown; stopping twice is harmless
}
