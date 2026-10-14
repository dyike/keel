// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package app

import (
	"errors"

	"github.com/dyike/keel/gio/gpu"
)

/*
#cgo CFLAGS: -Werror -xobjective-c -fobjc-arc
#cgo LDFLAGS: -framework QuartzCore -framework Metal

#import <Metal/Metal.h>
#import <QuartzCore/CAMetalLayer.h>

#include <CoreFoundation/CoreFoundation.h>

static CFTypeRef createMetalDevice(void) {
	@autoreleasepool {
		id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
		return CFBridgingRetain(dev);
	}
}

static void setupLayer(CFTypeRef layerRef, CFTypeRef devRef) {
	@autoreleasepool {
		CAMetalLayer *layer = (__bridge CAMetalLayer *)layerRef;
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		layer.device = dev;
		// Package gpu assumes an sRGB-encoded framebuffer.
		layer.pixelFormat = MTLPixelFormatBGRA8Unorm_sRGB;
		if (@available(iOS 11.0, *)) {
			// Never let nextDrawable time out and return nil.
			layer.allowsNextDrawableTimeout = NO;
		}
	}
}

static CFTypeRef nextDrawable(CFTypeRef layerRef) {
	@autoreleasepool {
		CAMetalLayer *layer = (__bridge CAMetalLayer *)layerRef;
		return CFBridgingRetain([layer nextDrawable]);
	}
}

static CFTypeRef drawableTexture(CFTypeRef drawableRef) {
	@autoreleasepool {
		id<CAMetalDrawable> drawable = (__bridge id<CAMetalDrawable>)drawableRef;
		return CFBridgingRetain(drawable.texture);
	}
}

static void presentDrawable(CFTypeRef queueRef, CFTypeRef drawableRef, int transaction) {
	@autoreleasepool {
		id<MTLDrawable> drawable = (__bridge id<MTLDrawable>)drawableRef;
		id<MTLCommandQueue> queue = (__bridge id<MTLCommandQueue>)queueRef;
		id<MTLCommandBuffer> cmdBuffer = [queue commandBuffer];
		if (!transaction) {
			// Present once the GPU has drawn it, without waiting here.
			[cmdBuffer presentDrawable:drawable];
			[cmdBuffer commit];
			return;
		}
		[cmdBuffer commit];
		[cmdBuffer waitUntilScheduled];
		[drawable present];
	}
}

static BOOL layerPresentsWithTransaction(CFTypeRef layerRef) {
	@autoreleasepool {
		CAMetalLayer *layer = (__bridge CAMetalLayer *)layerRef;
		return layer.presentsWithTransaction;
	}
}

static void setLayerPresentsWithTransaction(CFTypeRef layerRef, BOOL on) {
	@autoreleasepool {
		CAMetalLayer *layer = (__bridge CAMetalLayer *)layerRef;
		layer.presentsWithTransaction = on;
	}
}

static CFTypeRef newCommandQueue(CFTypeRef devRef) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		return CFBridgingRetain([dev newCommandQueue]);
	}
}
*/
import "C"

type mtlContext struct {
	idle     metalIdleState
	dev      C.CFTypeRef
	view     C.CFTypeRef
	layer    C.CFTypeRef
	queue    C.CFTypeRef
	drawable C.CFTypeRef
	texture  C.CFTypeRef
	// transaction is whether the layer was made to present with the Core
	// Animation transaction; onDemand limits that to frames that resize.
	transaction, onDemand bool
	// resized is set by a size change until the next frame presents.
	resized bool
	// synced is whether the current frame presents with the transaction.
	synced bool
}

func newMtlContext(w *window) (*mtlContext, error) {
	dev := C.createMetalDevice()
	if dev == 0 {
		return nil, errors.New("metal: MTLCreateSystemDefaultDevice failed")
	}
	view := w.contextView()
	layer := getMetalLayer(view)
	if layer == 0 {
		C.CFRelease(dev)
		return nil, errors.New("metal: CAMetalLayer construction failed")
	}
	queue := C.newCommandQueue(dev)
	if queue == 0 {
		C.CFRelease(dev)
		C.CFRelease(layer)
		return nil, errors.New("metal: [MTLDevice newCommandQueue] failed")
	}
	C.setupLayer(layer, dev)
	c := &mtlContext{
		dev:   dev,
		view:  view,
		layer: layer,
		queue: queue,
	}
	c.transaction = bool(C.layerPresentsWithTransaction(layer))
	c.onDemand = c.transaction && metalPresentOnDemand
	c.synced, c.resized = c.transaction, true
	return c, nil
}

func (c *mtlContext) RenderTarget() (gpu.RenderTarget, error) {
	if c.resumeIdle() {
		c.resized = true
	}
	if c.drawable != 0 || c.texture != 0 {
		return nil, errors.New("metal:a previous RenderTarget wasn't Presented")
	}
	if c.onDemand {
		// A frame that resizes the layer shows with the window's other
		// changes, as during a live resize. Others present as soon as the
		// GPU has them, so the display returns drawables without waiting
		// for a transaction, and two of them suffice.
		if sync := c.resized; sync != c.synced {
			C.setLayerPresentsWithTransaction(c.layer, C.BOOL(sync))
			c.synced = sync
		}
	}
	c.drawable = C.nextDrawable(c.layer)
	if c.drawable == 0 {
		return nil, errors.New("metal: [CAMetalLayer nextDrawable] failed")
	}
	c.texture = C.drawableTexture(c.drawable)
	if c.texture == 0 {
		return nil, errors.New("metal: CADrawable.texture is nil")
	}
	return gpu.MetalRenderTarget{
		Texture: uintptr(c.texture),
	}, nil
}

func (c *mtlContext) API() gpu.API {
	return gpu.Metal{
		Device:      uintptr(c.dev),
		Queue:       uintptr(c.queue),
		PixelFormat: int(C.MTLPixelFormatBGRA8Unorm_sRGB),
	}
}

func (c *mtlContext) Release() {
	c.stopIdle()
	C.CFRelease(c.queue)
	C.CFRelease(c.dev)
	C.CFRelease(c.layer)
	if c.drawable != 0 {
		C.CFRelease(c.drawable)
	}
	if c.texture != 0 {
		C.CFRelease(c.texture)
	}
	*c = mtlContext{}
}

func (c *mtlContext) Present() error {
	C.CFRelease(c.texture)
	c.texture = 0
	// Only on-demand layers present without the transaction's sequence;
	// others keep the original wait-then-present.
	transaction := 1
	if c.onDemand && !c.synced {
		transaction = 0
	}
	C.presentDrawable(c.queue, c.drawable, C.int(transaction))
	c.resized = false
	C.CFRelease(c.drawable)
	c.drawable = 0
	return nil
}

func (c *mtlContext) Lock() error {
	return nil
}

func (c *mtlContext) Unlock() {}

func (c *mtlContext) Refresh() error {
	if resizeDrawable(c.view, c.layer) {
		c.resized = true
	}
	return nil
}

func (w *window) NewContext() (context, error) {
	return newMtlContext(w)
}
