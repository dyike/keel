// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal

package app

/*
#cgo CFLAGS: -Werror -xobjective-c -fobjc-arc

#import <AppKit/AppKit.h>
#import <QuartzCore/CAMetalLayer.h>
#include <CoreFoundation/CoreFoundation.h>

CALayer *gio_layerFactory(BOOL presentWithTrans) {
	@autoreleasepool {
		CAMetalLayer *l = [CAMetalLayer layer];
		l.autoresizingMask = kCALayerHeightSizable|kCALayerWidthSizable;
		l.needsDisplayOnBoundsChange = YES;
		l.presentsWithTransaction = presentWithTrans;
		// Two window-sized drawables, not three: one less frame of memory.
		// Only frames that resize the layer present with the Core
		// Animation transaction (mtlContext.RenderTarget); with every frame
		// presenting that way, two drawables stalled a live resize while
		// Core Animation retained both.
		l.maximumDrawableCount = 2;
		return l;
	}
}

static CFTypeRef getMetalLayer(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		return CFBridgingRetain(view.layer);
	}
}

static BOOL resizeDrawable(CFTypeRef viewRef, CFTypeRef layerRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		CAMetalLayer *layer = (__bridge CAMetalLayer *)layerRef;
		CGSize size = layer.bounds.size;
		size.width *= layer.contentsScale;
		size.height *= layer.contentsScale;
		if (CGSizeEqualToSize(layer.drawableSize, size)) {
			return NO;
		}
		layer.drawableSize = size;
		return YES;
	}
}
*/
import "C"

func getMetalLayer(view C.CFTypeRef) C.CFTypeRef {
	return C.getMetalLayer(view)
}

// resizeDrawable reports whether the drawable's size changed.
func resizeDrawable(view, layer C.CFTypeRef) bool {
	return bool(C.resizeDrawable(view, layer))
}

// metalPresentOnDemand presents with the Core Animation transaction only
// the frames that resize the layer.
const metalPresentOnDemand = true
