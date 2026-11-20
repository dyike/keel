// SPDX-License-Identifier: Unlicense OR MIT

//go:build ios && !nometal

package app

// iOS-specific Metal layer helpers for metal_darwin.go without cgo. Keel patch.

func gioLayerClass() uintptr {
	return class("CAMetalLayer")
}

func getMetalLayer(view uintptr) uintptr {
	l := send(view, selLayer)
	if l == 0 {
		return 0
	}
	send(l, selSetNeedsDisplayOnBound, 1)
	send(l, selSetPresentsWithTrans, 1)
	return retainObject(l)
}

func resizeDrawable(view, layer uintptr) {
	withPool(func() {
		scale := msgF64(view, selContentScaleFactor)
		msgSetF64(layer, selSetContentsScale, scale)
		size := msgRect(layer, selBounds).Size
		msgSetSize(layer, selSetDrawableSize, nsPoint{X: size.X * scale, Y: size.Y * scale})
	})
}
