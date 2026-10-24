// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && nometal && (!cgo || purego)

package app

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/dyike/keel/gio/gpu"
	"github.com/ebitengine/purego"
)

// The OpenGL context of gl_macos.go and gl_macos.m without cgo, for
// nometal builds. Keel patch.

// NSOpenGLPixelFormatAttribute values, as in gl_macos.m.
const (
	nsOpenGLPFAOpenGLProfile                  = 99
	nsOpenGLProfileVersion3_2Core             = 0x3200
	nsOpenGLPFAColorSize                      = 8
	nsOpenGLPFAAccelerated                    = 73
	kCGLPFASupportsAutomaticGraphicsSwitching = 101
	nsOpenGLPFAAllowOfflineRenderers          = 96
)

var (
	selInitWithAttributes  = sel("initWithAttributes:")
	selInitWithFormat      = sel("initWithFormat:shareContext:")
	selSetWantsBestRes     = sel("setWantsBestResolutionOpenGLSurface:")
	selSetView             = sel("setView:")
	selClearCurrentContext = sel("clearCurrentContext")
	selUpdate              = sel("update")
	selMakeCurrentContext  = sel("makeCurrentContext")
	selCGLContextObj       = sel("CGLContextObj")
)

var cgl struct {
	flush, lock, unlock uintptr
}

type glContext struct {
	ctx  uintptr
	view uintptr
}

func newContext(w *window) (*glContext, error) {
	if cgl.flush == 0 {
		lib, err := purego.Dlopen("/System/Library/Frameworks/OpenGL.framework/OpenGL", purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			return nil, err
		}
		for _, s := range []struct {
			name string
			addr *uintptr
		}{{"glFlush", &cgl.flush}, {"CGLLockContext", &cgl.lock}, {"CGLUnlockContext", &cgl.unlock}} {
			if *s.addr, err = purego.Dlsym(lib, s.name); err != nil {
				return nil, errors.New("gl: missing symbol " + s.name + " in the OpenGL framework")
			}
		}
	}
	view := w.contextView()
	var ctx uintptr
	withPool(func() {
		attr := []uint32{
			nsOpenGLPFAOpenGLProfile, nsOpenGLProfileVersion3_2Core,
			nsOpenGLPFAColorSize, 24,
			nsOpenGLPFAAccelerated,
			// Opt-in to automatic GPU switching. CGL-only property.
			kCGLPFASupportsAutomaticGraphicsSwitching,
			nsOpenGLPFAAllowOfflineRenderers,
			0,
		}
		pixFormat := send(send(class("NSOpenGLPixelFormat"), selAlloc), selInitWithAttributes, uintptr(unsafe.Pointer(&attr[0])))
		runtime.KeepAlive(attr)
		if pixFormat == 0 {
			return
		}
		defer releaseObject(pixFormat)
		ctx = send(send(class("NSOpenGLContext"), selAlloc), selInitWithFormat, pixFormat, 0)
		if ctx == 0 {
			return
		}
		send(view, selSetWantsBestRes, 1)
		send(ctx, selSetView, view)
	})
	if ctx == 0 {
		return nil, errors.New("gl: failed to create NSOpenGLContext")
	}
	return &glContext{ctx: ctx, view: view}, nil
}

func (c *glContext) RenderTarget() (gpu.RenderTarget, error) {
	return gpu.OpenGLRenderTarget{}, nil
}

func (c *glContext) API() gpu.API {
	return gpu.OpenGL{}
}

func clearCurrentGLContext() {
	withPool(func() { send(class("NSOpenGLContext"), selClearCurrentContext) })
}

func (c *glContext) Release() {
	if c.ctx != 0 {
		clearCurrentGLContext()
		releaseObject(c.ctx)
		c.ctx = 0
	}
}

func (c *glContext) Present() error {
	// Assume the caller already locked the context.
	purego.SyscallN(cgl.flush)
	return nil
}

func (c *glContext) Lock() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	withPool(func() {
		purego.SyscallN(cgl.lock, send(c.ctx, selCGLContextObj))
		send(c.ctx, selMakeCurrentContext)
	})
	return nil
}

func (c *glContext) Unlock() {
	clearCurrentGLContext()
	withPool(func() { purego.SyscallN(cgl.unlock, send(c.ctx, selCGLContextObj)) })
}

func (c *glContext) Refresh() error {
	c.Lock()
	defer c.Unlock()
	withPool(func() { send(c.ctx, selUpdate) })
	return nil
}

func (w *window) NewContext() (context, error) {
	return newContext(w)
}

// layerFactory returns the autoreleased backing layer of a GioView.
func layerFactory(presentWithTrans bool) uintptr {
	return send(class("CALayer"), sel("layer"))
}
