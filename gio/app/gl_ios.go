// SPDX-License-Identifier: Unlicense OR MIT

//go:build ios && nometal

package app

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/dyike/keel/gio/gpu"
	"github.com/dyike/keel/gio/internal/gl"
	"github.com/ebitengine/purego"
)

// The OpenGL ES context of gl_ios.go and gl_ios.m without cgo, for nometal
// builds. Keel patch.

const kEAGLRenderingAPIOpenGLES3 = 3

var (
	selInitWithAPI         = sel("initWithAPI:")
	selSetCurrentContext   = sel("setCurrentContext:")
	selRenderbufferStorage = sel("renderbufferStorage:fromDrawable:")
	selPresentRenderbuffer = sel("presentRenderbuffer:")
	eaglOnce               sync.Once
	eaglErr                error
	classEAGLContext       uintptr
)

func loadOpenGLES() error {
	eaglOnce.Do(func() {
		if _, err := purego.Dlopen("/System/Library/Frameworks/OpenGLES.framework/OpenGLES", purego.RTLD_NOW|purego.RTLD_GLOBAL); err != nil {
			eaglErr = err
			return
		}
		classEAGLContext = class("EAGLContext")
	})
	return eaglErr
}

func gioLayerClass() uintptr {
	return class("CAEAGLLayer")
}

type glContext struct {
	owner       *window
	c           *gl.Functions
	ctx         uintptr
	layer       uintptr
	init        bool
	frameBuffer gl.Framebuffer
	colorBuffer gl.Renderbuffer
}

func newContext(w *window) (*glContext, error) {
	if err := loadOpenGLES(); err != nil {
		return nil, err
	}
	var ctx, layer uintptr
	withPool(func() {
		ctx = send(send(classEAGLContext, selAlloc), selInitWithAPI, kEAGLRenderingAPIOpenGLES3)
		if ctx != 0 {
			layer = retainObject(send(w.contextView(), selLayer))
		}
	})
	if ctx == 0 {
		return nil, fmt.Errorf("failed to create EAGLContext")
	}
	api := contextAPI()
	f, err := gl.NewFunctions(api.Context, api.ES)
	if err != nil {
		releaseObject(ctx)
		releaseObject(layer)
		return nil, err
	}
	return &glContext{
		ctx:   ctx,
		owner: w,
		layer: layer,
		c:     f,
	}, nil
}

func contextAPI() gpu.OpenGL {
	return gpu.OpenGL{}
}

func (c *glContext) RenderTarget() (gpu.RenderTarget, error) {
	return gpu.OpenGLRenderTarget(c.frameBuffer), nil
}

func (c *glContext) API() gpu.API {
	return contextAPI()
}

func (c *glContext) makeCurrent(ctx uintptr) bool {
	var ok bool
	withPool(func() {
		ok = send(classEAGLContext, selSetCurrentContext, ctx)&0xff != 0
	})
	return ok
}

func (c *glContext) renderbufferStorage(layer uintptr, buffer uint32) bool {
	var ok bool
	withPool(func() {
		ok = send(c.ctx, selRenderbufferStorage, uintptr(buffer), layer)&0xff != 0
	})
	return ok
}

func (c *glContext) presentRenderbuffer(buffer uint32) bool {
	var ok bool
	withPool(func() {
		ok = send(c.ctx, selPresentRenderbuffer, uintptr(buffer))&0xff != 0
	})
	return ok
}

func (c *glContext) Release() {
	if c.ctx == 0 {
		return
	}
	c.renderbufferStorage(0, gl.RENDERBUFFER)
	c.c.DeleteFramebuffer(c.frameBuffer)
	c.c.DeleteRenderbuffer(c.colorBuffer)
	c.makeCurrent(0)
	releaseObject(c.ctx)
	releaseObject(c.layer)
	c.ctx = 0
	c.layer = 0
}

func (c *glContext) Present() error {
	if c.layer == 0 {
		panic("context is not active")
	}
	c.c.BindRenderbuffer(gl.RENDERBUFFER, c.colorBuffer)
	if !c.presentRenderbuffer(gl.RENDERBUFFER) {
		return errors.New("presentRenderBuffer failed")
	}
	return nil
}

func (c *glContext) Lock() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	if !c.makeCurrent(c.ctx) {
		return errors.New("[EAGLContext setCurrentContext] failed")
	}
	return nil
}

func (c *glContext) Unlock() {
	c.makeCurrent(0)
}

func (c *glContext) Refresh() error {
	if !c.makeCurrent(c.ctx) {
		return errors.New("[EAGLContext setCurrentContext] failed")
	}
	if !c.init {
		c.init = true
		c.frameBuffer = c.c.CreateFramebuffer()
		c.colorBuffer = c.c.CreateRenderbuffer()
	}
	if c.owner.hidden {
		// Make sure any in-flight GL commands are complete.
		c.c.Finish()
		return nil
	}
	currentRB := gl.Renderbuffer{V: uint(c.c.GetInteger(gl.RENDERBUFFER_BINDING))}
	c.c.BindRenderbuffer(gl.RENDERBUFFER, c.colorBuffer)
	if !c.renderbufferStorage(c.layer, gl.RENDERBUFFER) {
		return errors.New("renderbufferStorage failed")
	}
	c.c.BindRenderbuffer(gl.RENDERBUFFER, currentRB)
	c.c.BindFramebuffer(gl.FRAMEBUFFER, c.frameBuffer)
	c.c.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, c.colorBuffer)
	if st := c.c.CheckFramebufferStatus(gl.FRAMEBUFFER); st != gl.FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("framebuffer incomplete, status: %#x\n", st)
	}
	return nil
}

func (w *window) NewContext() (context, error) {
	return newContext(w)
}
