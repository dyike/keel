// SPDX-License-Identifier: Unlicense OR MIT

//go:build !noopengl

package app

import (
	"unsafe"

	"github.com/dyike/keel/gio/internal/egl"
)

type androidContext struct {
	win     *window
	eglSurf egl.NativeWindowType
	*egl.Context
}

func init() {
	newAndroidGLESContext = func(w *window) (context, error) {
		ctx, err := egl.NewContext(0)
		if err != nil {
			return nil, err
		}
		return &androidContext{win: w, Context: ctx}, nil
	}
}

func (c *androidContext) Release() {
	if c.Context != nil {
		c.Context.Release()
		c.Context = nil
	}
}

func (c *androidContext) Refresh() error {
	c.Context.ReleaseSurface()
	if err := c.win.setVisual(c.Context.VisualID()); err != nil {
		return err
	}
	win, _, _ := c.win.nativeWindow()
	c.eglSurf = egl.NativeWindowType(uintptr(unsafe.Pointer(win)))
	return nil
}

func (c *androidContext) Lock() error {
	// The Android emulator creates a broken surface if it is not
	// created on the same thread as the context is made current.
	if c.eglSurf != 0 {
		if err := c.Context.CreateSurface(c.eglSurf); err != nil {
			return err
		}
		c.eglSurf = 0
	}
	return c.Context.MakeCurrent()
}

func (c *androidContext) Unlock() {
	c.Context.ReleaseCurrent()
}
