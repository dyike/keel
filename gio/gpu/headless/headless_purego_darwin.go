// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && (!cgo || purego)

package headless

import (
	"errors"
	"runtime"

	"github.com/dyike/keel/gio/gpu"
	_ "github.com/dyike/keel/gio/internal/cocoainit"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// The Metal context of headless_darwin.go without cgo. Keel patch.

const mtlPixelFormatRGBA8UnormSRGB = 71

type mtlContext struct {
	dev   uintptr
	queue uintptr
}

func init() {
	newContextPrimary = newMetalContext
}

func newMetalContext() (context, error) {
	metal, err := purego.Dlopen("/System/Library/Frameworks/Metal.framework/Metal", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, err
	}
	create, err := purego.Dlsym(metal, "MTLCreateSystemDefaultDevice")
	if err != nil {
		return nil, err
	}
	// The device and queue are created inside an autorelease pool, as the cgo
	// helpers do; the pool belongs to this thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))
	defer pool.Send(objc.RegisterName("drain"))
	// Both results are owned: MTLCreateSystemDefaultDevice returns retained,
	// and newCommandQueue follows the new rule.
	dev, _, _ := purego.SyscallN(create)
	if dev == 0 {
		return nil, errors.New("headless: failed to create Metal device")
	}
	queue := uintptr(objc.ID(dev).Send(objc.RegisterName("newCommandQueue")))
	if queue == 0 {
		objc.ID(dev).Send(objc.RegisterName("release"))
		return nil, errors.New("headless: failed to create MTLQueue")
	}
	return &mtlContext{dev: dev, queue: queue}, nil
}

func (c *mtlContext) API() gpu.API {
	return gpu.Metal{
		Device:      c.dev,
		Queue:       c.queue,
		PixelFormat: mtlPixelFormatRGBA8UnormSRGB,
	}
}

func (c *mtlContext) MakeCurrent() error {
	return nil
}

func (c *mtlContext) ReleaseCurrent() {}

func (d *mtlContext) Release() {
	release := objc.RegisterName("release")
	objc.ID(d.dev).Send(release)
	objc.ID(d.queue).Send(release)
	*d = mtlContext{}
}
