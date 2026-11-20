// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal

package app

// macOS-specific Metal layer helpers for metal_darwin.go. Keel patch.

var (
	selSetAutoresizingMask = sel("setAutoresizingMask:")
	selSetMaxDrawableCount = sel("setMaximumDrawableCount:")
)

const (
	kCALayerWidthSizable  = 1 << 1
	kCALayerHeightSizable = 1 << 4
)

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
