// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal && (!cgo || purego)

package app

import (
	"errors"

	"github.com/dyike/keel/gio/gpu"
	"github.com/ebitengine/purego"
)

// The Metal context of metal_darwin.go and metal_macos.go without cgo.
// Keel patch.

const mtlPixelFormatBGRA8UnormSRGB = 81

var (
	selSetDevice              = sel("setDevice:")
	selSetPixelFormat         = sel("setPixelFormat:")
	selSetAllowsTimeout       = sel("setAllowsNextDrawableTimeout:")
	selNextDrawable           = sel("nextDrawable")
	selTexture                = sel("texture")
	selCommandBuffer          = sel("commandBuffer")
	selCommit                 = sel("commit")
	selWaitUntilScheduled     = sel("waitUntilScheduled")
	selPresent                = sel("present")
	selNewCommandQueue        = sel("newCommandQueue")
	selContentsScale          = sel("contentsScale")
	selSetDrawableSize        = sel("setDrawableSize:")
	selSetAutoresizingMask    = sel("setAutoresizingMask:")
	selSetNeedsDisplayOnBound = sel("setNeedsDisplayOnBoundsChange:")
	selSetPresentsWithTrans   = sel("setPresentsWithTransaction:")
	selSetMaxDrawableCount    = sel("setMaximumDrawableCount:")
)

const (
	kCALayerWidthSizable  = 1 << 1
	kCALayerHeightSizable = 1 << 4
)

type mtlContext struct {
	idle     metalIdleState
	dev      uintptr
	view     uintptr
	layer    uintptr
	queue    uintptr
	drawable uintptr
	texture  uintptr
}

var createMetalDevice uintptr

func newMtlContext(w *window) (*mtlContext, error) {
	if createMetalDevice == 0 {
		metal, err := purego.Dlopen("/System/Library/Frameworks/Metal.framework/Metal", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return nil, err
		}
		if createMetalDevice, err = purego.Dlsym(metal, "MTLCreateSystemDefaultDevice"); err != nil {
			return nil, err
		}
	}
	c := &mtlContext{view: w.contextView()}
	var err error
	withPool(func() {
		// MTLCreateSystemDefaultDevice returns a retained device.
		c.dev, _, _ = purego.SyscallN(createMetalDevice)
		if c.dev == 0 {
			err = errors.New("metal: MTLCreateSystemDefaultDevice failed")
			return
		}
		c.layer = getMetalLayer(c.view)
		if c.layer == 0 {
			releaseObject(c.dev)
			err = errors.New("metal: CAMetalLayer construction failed")
			return
		}
		c.queue = send(c.dev, selNewCommandQueue)
		if c.queue == 0 {
			releaseObject(c.dev)
			releaseObject(c.layer)
			err = errors.New("metal: [MTLDevice newCommandQueue] failed")
			return
		}
		send(c.layer, selSetDevice, c.dev)
		// Package gpu assumes an sRGB-encoded framebuffer.
		send(c.layer, selSetPixelFormat, mtlPixelFormatBGRA8UnormSRGB)
		// Never let nextDrawable time out and return nil.
		send(c.layer, selSetAllowsTimeout, 0)
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *mtlContext) RenderTarget() (gpu.RenderTarget, error) {
	c.resumeIdle()
	if c.drawable != 0 || c.texture != 0 {
		return nil, errors.New("metal:a previous RenderTarget wasn't Presented")
	}
	withPool(func() {
		c.drawable = retainObject(send(c.layer, selNextDrawable))
		if c.drawable != 0 {
			c.texture = retainObject(send(c.drawable, selTexture))
		}
	})
	if c.drawable == 0 {
		return nil, errors.New("metal: [CAMetalLayer nextDrawable] failed")
	}
	if c.texture == 0 {
		return nil, errors.New("metal: CADrawable.texture is nil")
	}
	return gpu.MetalRenderTarget{
		Texture: c.texture,
	}, nil
}

func (c *mtlContext) API() gpu.API {
	return gpu.Metal{
		Device:      c.dev,
		Queue:       c.queue,
		PixelFormat: mtlPixelFormatBGRA8UnormSRGB,
	}
}

func (c *mtlContext) Release() {
	c.stopIdle()
	releaseObject(c.queue)
	releaseObject(c.dev)
	releaseObject(c.layer)
	releaseObject(c.drawable)
	releaseObject(c.texture)
	*c = mtlContext{}
}

func (c *mtlContext) Present() error {
	releaseObject(c.texture)
	c.texture = 0
	withPool(func() {
		cmdBuffer := send(c.queue, selCommandBuffer)
		send(cmdBuffer, selCommit)
		send(cmdBuffer, selWaitUntilScheduled)
		send(c.drawable, selPresent)
	})
	releaseObject(c.drawable)
	c.drawable = 0
	return nil
}

func (c *mtlContext) Lock() error {
	return nil
}

func (c *mtlContext) Unlock() {}

func (c *mtlContext) Refresh() error {
	resizeDrawable(c.view, c.layer)
	return nil
}

func (w *window) NewContext() (context, error) {
	return newMtlContext(w)
}

// layerFactory returns the autoreleased backing layer of a GioView.
func layerFactory(presentWithTrans bool) uintptr {
	l := send(class("CAMetalLayer"), sel("layer"))
	send(l, selSetAutoresizingMask, kCALayerHeightSizable|kCALayerWidthSizable)
	send(l, selSetNeedsDisplayOnBound, 1)
	send(l, selSetPresentsWithTrans, boolArg(presentWithTrans))
	// Core Animation can retain two presented surfaces during a live
	// resize. Keep the third drawable available so nextDrawable does
	// not block AppKit while those surfaces wait for a transaction.
	// The idle trim still releases spare window-sized surfaces.
	send(l, selSetMaxDrawableCount, 3)
	return l
}

func getMetalLayer(view uintptr) uintptr {
	return retainObject(send(view, selLayer))
}

func resizeDrawable(view, layer uintptr) {
	size := msgRect(layer, selBounds).Size
	scale := msgF64(layer, selContentsScale)
	msgSetSize(layer, selSetDrawableSize, nsPoint{X: size.X * scale, Y: size.Y * scale})
}
