// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && (!cgo || purego)

package egl

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// The EGL binding of egl_unix.go without cgo, for Linux desktop: libEGL is
// loaded at run time, as egl_windows.go loads libEGL.dll. Keel patch.

type (
	_EGLint           = int32
	_EGLDisplay       = uintptr
	_EGLConfig        = uintptr
	_EGLContext       = uintptr
	_EGLSurface       = uintptr
	NativeDisplayType = uintptr
	NativeWindowType  = uintptr
)

const _EGL_TRUE = 1

var egl struct {
	chooseConfig, createContext, createWindowSurface uintptr
	destroyContext, destroySurface, getConfigAttrib  uintptr
	getDisplay, getError, initialize, makeCurrent    uintptr
	releaseThread, swapInterval, swapBuffers         uintptr
	terminate, queryString, waitClient               uintptr
}

var loadOnce = sync.OnceValue(loadLib)

func loadEGL() error {
	return loadOnce()
}

func loadLib() error {
	var lib uintptr
	var err error
	for _, name := range []string{"libEGL.so.1", "libEGL.so"} {
		if lib, err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL); err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("egl: %w", err)
	}
	for _, s := range []struct {
		name string
		addr *uintptr
	}{
		{"eglChooseConfig", &egl.chooseConfig},
		{"eglCreateContext", &egl.createContext},
		{"eglCreateWindowSurface", &egl.createWindowSurface},
		{"eglDestroyContext", &egl.destroyContext},
		{"eglDestroySurface", &egl.destroySurface},
		{"eglGetConfigAttrib", &egl.getConfigAttrib},
		{"eglGetDisplay", &egl.getDisplay},
		{"eglGetError", &egl.getError},
		{"eglInitialize", &egl.initialize},
		{"eglMakeCurrent", &egl.makeCurrent},
		{"eglReleaseThread", &egl.releaseThread},
		{"eglSwapInterval", &egl.swapInterval},
		{"eglSwapBuffers", &egl.swapBuffers},
		{"eglTerminate", &egl.terminate},
		{"eglQueryString", &egl.queryString},
		{"eglWaitClient", &egl.waitClient},
	} {
		if *s.addr, err = purego.Dlsym(lib, s.name); err != nil {
			return fmt.Errorf("egl: %w", err)
		}
	}
	return nil
}

// call calls an EGL function with integer and pointer arguments; pointers
// to Go memory stay valid and in place for the call.
//
//go:uintptrescapes
func call(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

// ok reports an EGLBoolean result; the upper register bits are undefined.
func ok(r uintptr) bool { return uint32(r) == _EGL_TRUE }

func eglChooseConfig(disp _EGLDisplay, attribs []_EGLint) (_EGLConfig, bool) {
	var cfg _EGLConfig
	var ncfg _EGLint
	r := call(egl.chooseConfig, disp, uintptr(unsafe.Pointer(&attribs[0])), uintptr(unsafe.Pointer(&cfg)), 1, uintptr(unsafe.Pointer(&ncfg)))
	if !ok(r) {
		return nilEGLConfig, false
	}
	return cfg, true
}

func eglCreateContext(disp _EGLDisplay, cfg _EGLConfig, shareCtx _EGLContext, attribs []_EGLint) _EGLContext {
	return call(egl.createContext, disp, cfg, shareCtx, uintptr(unsafe.Pointer(&attribs[0])))
}

func eglDestroySurface(disp _EGLDisplay, surf _EGLSurface) bool {
	return ok(call(egl.destroySurface, disp, surf))
}

func eglDestroyContext(disp _EGLDisplay, ctx _EGLContext) bool {
	return ok(call(egl.destroyContext, disp, ctx))
}

func eglGetConfigAttrib(disp _EGLDisplay, cfg _EGLConfig, attr _EGLint) (_EGLint, bool) {
	var val _EGLint
	r := call(egl.getConfigAttrib, disp, cfg, uintptr(attr), uintptr(unsafe.Pointer(&val)))
	return val, ok(r)
}

func eglGetError() _EGLint {
	return _EGLint(call(egl.getError))
}

func eglInitialize(disp _EGLDisplay) (_EGLint, _EGLint, bool) {
	var maj, min _EGLint
	r := call(egl.initialize, disp, uintptr(unsafe.Pointer(&maj)), uintptr(unsafe.Pointer(&min)))
	return maj, min, ok(r)
}

func eglMakeCurrent(disp _EGLDisplay, draw, read _EGLSurface, ctx _EGLContext) bool {
	return ok(call(egl.makeCurrent, disp, draw, read, ctx))
}

func eglReleaseThread() bool {
	return ok(call(egl.releaseThread))
}

func eglSwapBuffers(disp _EGLDisplay, surf _EGLSurface) bool {
	return ok(call(egl.swapBuffers, disp, surf))
}

func eglSwapInterval(disp _EGLDisplay, interval _EGLint) bool {
	return ok(call(egl.swapInterval, disp, uintptr(interval)))
}

func eglTerminate(disp _EGLDisplay) bool {
	return ok(call(egl.terminate, disp))
}

func eglQueryString(disp _EGLDisplay, name _EGLint) string {
	p := call(egl.queryString, disp, uintptr(name))
	if p == 0 {
		return ""
	}
	ptr := *(*unsafe.Pointer)(unsafe.Pointer(&p))
	n := 0
	for *(*byte)(unsafe.Add(ptr, n)) != 0 {
		n++
	}
	return string(unsafe.Slice((*byte)(ptr), n))
}

func eglGetDisplay(disp NativeDisplayType) _EGLDisplay {
	return call(egl.getDisplay, disp)
}

func eglCreateWindowSurface(disp _EGLDisplay, conf _EGLConfig, win NativeWindowType, attribs []_EGLint) _EGLSurface {
	return call(egl.createWindowSurface, disp, conf, win, uintptr(unsafe.Pointer(&attribs[0])))
}

func eglWaitClient() bool {
	return ok(call(egl.waitClient))
}
