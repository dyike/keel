//go:build darwin && !ios

// SPDX-License-Identifier: Unlicense OR MIT
package text

import (
	"runtime"

	"golang.org/x/image/math/fixed"
)

// Glyph IDs contain a face index and can outlive the layout cache (editors,
// retained drawing). Never recycle indices. Keep exact descriptors, but bound
// the costly sized fonts; reconstruct an evicted font when an old ID is drawn.
const maxNativeFonts = 512

func (e *ctShaper) unlinkFont(f *ctNativeFace) {
	if f.prev != nil {
		f.prev.next = f.next
	} else {
		e.fontOldest = f.next
	}
	if f.next != nil {
		f.next.prev = f.prev
	} else {
		e.fontNewest = f.prev
	}
	f.prev, f.next = nil, nil
}

func (e *ctShaper) touchFont(f *ctNativeFace) {
	if f.pinned || e.fontNewest == f {
		return
	}
	if f.prev != nil || f.next != nil || e.fontOldest == f {
		e.unlinkFont(f)
	}
	f.prev = e.fontNewest
	if e.fontNewest != nil {
		e.fontNewest.next = f
	} else {
		e.fontOldest = f
	}
	e.fontNewest = f
}

func (e *ctShaper) useFont(f *ctNativeFace) uintptr {
	if f.font == 0 {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		pool := ct.poolPush()
		defer ct.poolPop(pool)
		native := ct.fontWithDescriptor(f.descriptor, 0, 0)
		if native == 0 {
			return 0
		}
		f.font = e.resources.own(native)
		e.fontCount++
	}
	e.touchFont(f)
	e.trimFonts()
	return f.font
}

func (e *ctShaper) trimFonts() {
	for e.fontCount > maxNativeFonts && e.fontOldest != nil {
		f := e.fontOldest
		if f.descriptor == 0 {
			d := nativeInit.fontDescriptor(f.font)
			var restored uintptr
			if d != 0 {
				restored = ct.fontWithDescriptor(d, 0, 0)
			}
			// A descriptor must reproduce the exact variable instance, size,
			// matrix and fallback identity. Keep unsupported fonts alive.
			exact := restored != 0 && ct.equal(restored, f.font)
			if restored != 0 {
				ct.release(restored)
			}
			if !exact {
				if d != 0 {
					ct.release(d)
				}
				e.pinFont(f)
				continue
			}
			f.descriptor = e.resources.own(d)
		}
		e.unlinkFont(f)
		e.fontCount--
		e.resources.drop(f.font)
		f.font = 0
	}
}

func (e *ctShaper) pinFont(f *ctNativeFace) {
	if !f.pinned {
		e.unlinkFont(f)
		e.fontCount--
		f.pinned = true
		// A permanent handle no longer needs a reconstruction descriptor.
		if f.descriptor != 0 {
			e.resources.drop(f.descriptor)
			f.descriptor = 0
		}
	}
}

func (e *ctShaper) pinGlyphFont(id GlyphID) (uintptr, fixed.Int26_6, uint16, bool) {
	native, size, gid, ok := e.glyphFont(id)
	if ok {
		_, i, _ := splitGlyphID(id)
		e.pinFont(e.faces[i])
	}
	return native, size, gid, ok
}
