// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nowayland && !noopengl && (!cgo || !giocgo || purego)

package app

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/dyike/keel/gio/internal/egl"
	"github.com/ebitengine/purego"
)

// The EGL context of egl_wayland.go without cgo: libwayland-egl is loaded
// at run time. Keel patch.

var wlEGL struct {
	once                    sync.Once
	err                     error
	create, destroy, resize uintptr
}

func loadWaylandEGL() error {
	wlEGL.once.Do(func() {
		lib, err := openLib("libwayland-egl.so.1", "libwayland-egl.so")
		if err != nil {
			wlEGL.err = fmt.Errorf("wayland: %w", err)
			return
		}
		for _, s := range []struct {
			name string
			addr *uintptr
		}{
			{"wl_egl_window_create", &wlEGL.create},
			{"wl_egl_window_destroy", &wlEGL.destroy},
			{"wl_egl_window_resize", &wlEGL.resize},
		} {
			if *s.addr, err = purego.Dlsym(lib, s.name); err != nil {
				wlEGL.err = fmt.Errorf("wayland: %w", err)
				return
			}
		}
	})
	return wlEGL.err
}

type wlContext struct {
	win *window
	*egl.Context
	eglWin unsafe.Pointer // struct wl_egl_window *
}

func init() {
	newWaylandEGLContext = func(w *window) (context, error) {
		if err := loadWaylandEGL(); err != nil {
			return nil, err
		}
		disp := egl.NativeDisplayType(uintptr(w.display()))
		ctx, err := egl.NewContext(disp)
		if err != nil {
			return nil, err
		}

		surf, width, height := w.surface()
		if surf == nil {
			return nil, errors.New("wayland: no surface")
		}
		eglWin := cptr(xcall(wlEGL.create, ptr(surf), iarg(int32(width)), iarg(int32(height))))
		if eglWin == nil {
			return nil, errors.New("wayland: wl_egl_window_create failed")
		}
		if err := ctx.CreateSurface(egl.NativeWindowType(uintptr(eglWin))); err != nil {
			return nil, err
		}
		return &wlContext{Context: ctx, win: w, eglWin: eglWin}, nil
	}
}

func (c *wlContext) Release() {
	if c.Context != nil {
		c.Context.Release()
		c.Context = nil
	}
	if c.eglWin != nil {
		xcall(wlEGL.destroy, ptr(c.eglWin))
		c.eglWin = nil
	}
}

func (c *wlContext) Refresh() error {
	surf, width, height := c.win.surface()
	if surf == nil {
		return errors.New("wayland: no surface")
	}
	xcall(wlEGL.resize, ptr(c.eglWin), iarg(int32(width)), iarg(int32(height)), 0, 0)
	return nil
}

func (c *wlContext) Lock() error {
	if err := c.Context.MakeCurrent(); err != nil {
		return err
	}
	// eglSwapInterval applies to the current draw surface. Setting it before
	// MakeCurrent fails with EGL_BAD_CONTEXT (or affects another window).
	// Wayland already schedules frames; Mesa must not wait for an extra frame
	// callback when the surface is hidden and this event loop needs to answer pings.
	c.Context.EnableVSync(false)
	return nil
}

func (c *wlContext) Unlock() {
	c.Context.ReleaseCurrent()
}
