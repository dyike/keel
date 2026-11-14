// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nowayland && !noopengl && (!cgo || !giocgo || purego)

package app

import (
	"os"
	"runtime"
	"testing"

	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/unit"
	"github.com/ebitengine/purego"
)

// Exercise Mesa's actual EGL state validation: swap interval requires a
// current drawing surface, not merely an allocated context and surface.
func TestWaylandEGLCurrentSurface(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" || os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("requires a Wayland desktop")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	lib, err := purego.Dlopen("libEGL.so.1", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		t.Fatal(err)
	}
	defer purego.Dlclose(lib)
	getError, err := purego.Dlsym(lib, "eglGetError")
	if err != nil {
		t.Fatal(err)
	}
	w := new(Window)
	w.Option(Title("Keel EGL regression"), Size(unit.Dp(320), unit.Dp(200)))
	for {
		switch w.Event().(type) {
		case FrameEvent:
			if _, ok := w.driver.(*window); !ok {
				t.Fatal("expected Wayland driver")
			}
			// Drop any context left current by initial window configuration.
			if w.ctx != nil {
				w.ctx.Unlock()
			}
			purego.SyscallN(getError)
			c, err := w.driver.NewContext()
			if err != nil {
				t.Fatal(err)
			}
			code, _, _ := purego.SyscallN(getError)
			if code != 0x3000 {
				t.Errorf("EGL error after context creation: %#x", code)
			}
			if err = c.Lock(); err != nil {
				t.Fatal(err)
			}
			code, _, _ = purego.SyscallN(getError)
			if code != 0x3000 {
				t.Errorf("EGL error after binding drawing surface: %#x", code)
			}
			c.Unlock()
			c.Release()
			go w.Perform(system.ActionClose)
		case DestroyEvent:
			return
		}
	}
}
