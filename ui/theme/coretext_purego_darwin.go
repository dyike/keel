//go:build darwin && !ios && (!cgo || purego)

package theme

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// The CoreText rasterizer without cgo: the same CoreText and CoreGraphics
// calls as coretext_cgo_darwin.go, loaded at run time. As there, the bitmap
// context draws into a Go buffer; its lifetime is bounded by one function,
// which releases the context before returning and keeps the buffer alive
// until then. A context owning its pixels was three times slower.

type cgPoint struct{ X, Y float64 }
type cgSize struct{ W, H float64 }
type cgRect struct {
	Origin cgPoint
	Size   cgSize
}

const (
	kCTFontTraitColorGlyphs        = 1 << 13
	kCTFontOrientationHorizontal   = 1
	kCGImageAlphaPremultipliedLast = 1
	kCGBitmapByteOrder32Big        = 4 << 12
)

// cg holds CoreText and CoreGraphics entry points. Integer, pointer and
// bool signatures are called by address with SyscallN: a glyph-cache miss
// makes about seventeen calls, and RegisterFunc wrappers allocate through
// reflection. The only wrapper left returns a struct on amd64; the two text
// colors are created once.
var cg struct {
	once  sync.Once
	ready bool
	srgb  uintptr // kCGColorSpaceSRGB

	traits, drawGlyphs            uintptr
	colorSpace, colorSpaceFree    uintptr
	bitmap, release, flush        uintptr
	allowAA, shouldAA             uintptr
	allowSmooth, shouldSmooth     uintptr
	allowSubpixel, shouldSubpixel uintptr
	allowQuantize, shouldQuantize uintptr

	setFill      uintptr // CGContextSetFillColorWithColor
	white, black uintptr // CGColorRef, created once, never released
	boundsAddr   uintptr // CTFontGetBoundingRectsForGlyphs

	boundsFor func(font uintptr, orientation int, glyphs *uint16, rects *cgRect, count int) cgRect
}

func cgCall(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

func loadCG() bool {
	cg.once.Do(func() {
		ct, err := purego.Dlopen("/System/Library/Frameworks/CoreText.framework/CoreText", purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			return
		}
		g, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			return
		}
		for _, sym := range []struct {
			lib  uintptr
			name string
			addr *uintptr
		}{
			{ct, "CTFontGetSymbolicTraits", &cg.traits},
			{ct, "CTFontDrawGlyphs", &cg.drawGlyphs},
			{g, "CGColorSpaceCreateWithName", &cg.colorSpace},
			{g, "CGColorSpaceRelease", &cg.colorSpaceFree},
			{g, "CGBitmapContextCreate", &cg.bitmap},
			{g, "CGContextRelease", &cg.release},
			{g, "CGContextFlush", &cg.flush},
			{g, "CGContextSetAllowsAntialiasing", &cg.allowAA},
			{g, "CGContextSetShouldAntialias", &cg.shouldAA},
			{g, "CGContextSetAllowsFontSmoothing", &cg.allowSmooth},
			{g, "CGContextSetShouldSmoothFonts", &cg.shouldSmooth},
			{g, "CGContextSetAllowsFontSubpixelPositioning", &cg.allowSubpixel},
			{g, "CGContextSetShouldSubpixelPositionFonts", &cg.shouldSubpixel},
			{g, "CGContextSetAllowsFontSubpixelQuantization", &cg.allowQuantize},
			{g, "CGContextSetShouldSubpixelQuantizeFonts", &cg.shouldQuantize},
			{g, "CGContextSetFillColorWithColor", &cg.setFill},
			{ct, "CTFontGetBoundingRectsForGlyphs", &cg.boundsAddr},
		} {
			if *sym.addr, err = purego.Dlsym(sym.lib, sym.name); err != nil {
				return
			}
		}
		purego.RegisterLibFunc(&cg.boundsFor, ct, "CTFontGetBoundingRectsForGlyphs")
		// The two text colors as CGColors, so setting one is a pointer call.
		var color func(r, g, b, a float64) uintptr
		purego.RegisterLibFunc(&color, g, "CGColorCreateSRGB")
		cg.white, cg.black = color(1, 1, 1, 1), color(0, 0, 0, 1)
		if cg.white == 0 || cg.black == 0 {
			return
		}
		name, err := purego.Dlsym(g, "kCGColorSpaceSRGB")
		if err != nil {
			return
		}
		cg.srgb = **(**uintptr)(unsafe.Pointer(&name)) // the CFStringRef the symbol holds
		cg.ready = true
	})
	return cg.ready
}

func ctColorGlyphs(font uintptr) bool {
	return loadCG() && uint32(cgCall(cg.traits, font))&kCTFontTraitColorGlyphs != 0
}

func ctGlyphBounds(font uintptr, gid uint16) (x, y, w, h float64) {
	if !loadCG() {
		return 0, 0, 0, 0
	}
	// The per-glyph array is filled as well as the union returned; read it
	// rather than depend on the aggregate return convention.
	var r cgRect
	if runtime.GOARCH == "arm64" {
		// The returned rect comes back in vector registers, which a call
		// by address ignores; amd64 returns it through a hidden pointer.
		cgCall(cg.boundsAddr, font, kCTFontOrientationHorizontal, uintptr(unsafe.Pointer(&gid)), uintptr(unsafe.Pointer(&r)), 1)
		runtime.KeepAlive(&gid)
	} else {
		cg.boundsFor(font, kCTFontOrientationHorizontal, &gid, &r, 1)
	}
	return r.Origin.X, r.Origin.Y, r.Size.W, r.Size.H
}

// newBitmap creates a premultiplied sRGB context drawing into pix, w×h
// pixels; 0 if it cannot. Release it before pix may be collected.
func newBitmap(pix []byte, w, h int) uintptr {
	cs := cgCall(cg.colorSpace, cg.srgb)
	if cs == 0 {
		return 0
	}
	ctx := cgCall(cg.bitmap, uintptr(unsafe.Pointer(&pix[0])), uintptr(w), uintptr(h), 8, uintptr(w*4), cs, kCGImageAlphaPremultipliedLast|kCGBitmapByteOrder32Big)
	cgCall(cg.colorSpaceFree, cs)
	return ctx
}

func ctDrawGlyph(font uintptr, gid uint16, ox, oy float64, w, h int, dark bool) []byte {
	if !loadCG() {
		return nil
	}
	pix := make([]byte, w*h*4)
	ctx := newBitmap(pix, w, h)
	if ctx == 0 {
		return nil
	}
	defer runtime.KeepAlive(pix)
	defer cgCall(cg.release, ctx)
	fg := cg.white
	// The opaque background, as CGContextFillRect would paint it, written
	// directly: two reflective calls fewer per glyph.
	bg := [4]byte{0, 0, 0, 255}
	if dark {
		fg, bg = cg.black, [4]byte{255, 255, 255, 255}
	}
	for i := 0; i < len(pix); i += 4 {
		copy(pix[i:i+4], bg[:])
	}
	for _, fn := range [...]uintptr{cg.allowAA, cg.shouldAA, cg.allowSmooth, cg.shouldSmooth, cg.allowSubpixel, cg.shouldSubpixel} {
		cgCall(fn, ctx, 1)
	}
	cgCall(cg.allowQuantize, ctx, 0)
	cgCall(cg.shouldQuantize, ctx, 0)
	cgCall(cg.setFill, ctx, fg)
	drawGlyph(font, gid, ox, oy, ctx)
	return pix
}

func ctDrawColorGlyph(font uintptr, gid uint16, ox, oy float64, w, h int) []byte {
	if !loadCG() {
		return nil
	}
	pix := make([]byte, w*h*4)
	ctx := newBitmap(pix, w, h)
	if ctx == 0 {
		return nil
	}
	defer runtime.KeepAlive(pix)
	defer cgCall(cg.release, ctx)
	cgCall(cg.allowAA, ctx, 1)
	cgCall(cg.shouldAA, ctx, 1)
	cgCall(cg.shouldSubpixel, ctx, 1)
	cgCall(cg.shouldQuantize, ctx, 0)
	drawGlyph(font, gid, ox, oy, ctx)
	return pix
}

// drawGlyph draws gid with its origin at (ox, oy) and flushes ctx. The glyph
// and its position are valid for the call only.
func drawGlyph(font uintptr, gid uint16, ox, oy float64, ctx uintptr) {
	pos := cgPoint{ox, oy}
	cgCall(cg.drawGlyphs, font, uintptr(unsafe.Pointer(&gid)), uintptr(unsafe.Pointer(&pos)), 1, ctx)
	runtime.KeepAlive(&gid)
	runtime.KeepAlive(&pos)
	cgCall(cg.flush, ctx)
}
