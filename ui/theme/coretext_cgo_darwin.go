//go:build darwin && !ios && cgo && !purego

package theme

// The cgo CoreText rasterizer, kept for comparison while the purego one in
// coretext_purego_darwin.go is validated; see docs/desktop-cgo-free.md.

/*
#cgo LDFLAGS: -framework CoreText -framework CoreGraphics -framework CoreFoundation

#include <stdlib.h>
#include <CoreText/CoreText.h>
#include <CoreGraphics/CoreGraphics.h>

// keelCTColorGlyphs reports whether font draws color glyphs, such as emoji.
static int keelCTColorGlyphs(CTFontRef font) {
	return (CTFontGetSymbolicTraits(font) & kCTFontTraitColorGlyphs) != 0;
}

// keelCTGlyphBounds reports the glyph's ink box in pixels, y up.
static void keelCTGlyphBounds(CTFontRef font, CGGlyph g, double *x, double *y, double *w, double *h) {
	CGRect r = CTFontGetBoundingRectsForGlyphs(font, kCTFontOrientationHorizontal, &g, NULL, 1);
	*x = r.origin.x; *y = r.origin.y; *w = r.size.width; *h = r.size.height;
}

// keelCTDrawGlyph draws g with its origin at (ox, oy) pixels from the bottom
// left of a w×h opaque sRGB bitmap: dark text on white when dark, otherwise
// white text on black. CoreText smooths it as it would in a window.
static void keelCTDrawGlyph(CTFontRef font, CGGlyph g, double ox, double oy, int w, int h, int dark, unsigned char *out) {
	CGColorSpaceRef cs = CGColorSpaceCreateWithName(kCGColorSpaceSRGB);
	CGContextRef ctx = CGBitmapContextCreate(out, w, h, 8, w * 4, cs, kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big);
	CGColorSpaceRelease(cs);
	if (!ctx) return;
	CGContextSetRGBFillColor(ctx, dark ? 1 : 0, dark ? 1 : 0, dark ? 1 : 0, 1);
	CGContextFillRect(ctx, CGRectMake(0, 0, w, h));
	CGContextSetAllowsAntialiasing(ctx, true);
	CGContextSetShouldAntialias(ctx, true);
	CGContextSetAllowsFontSmoothing(ctx, true);
	CGContextSetShouldSmoothFonts(ctx, true);
	CGContextSetAllowsFontSubpixelPositioning(ctx, true);
	CGContextSetShouldSubpixelPositionFonts(ctx, true);
	CGContextSetAllowsFontSubpixelQuantization(ctx, false);
	CGContextSetShouldSubpixelQuantizeFonts(ctx, false);
	CGContextSetRGBFillColor(ctx, dark ? 0 : 1, dark ? 0 : 1, dark ? 0 : 1, 1);
	CGPoint pos = CGPointMake(ox, oy);
	CTFontDrawGlyphs(font, &g, &pos, 1, ctx);
	CGContextFlush(ctx);
	CGContextRelease(ctx);
}
// keelCTDrawColorGlyph draws a color glyph (emoji) with its origin at (ox, oy)
// pixels from the bottom left of a transparent w×h premultiplied RGBA bitmap.
static void keelCTDrawColorGlyph(CTFontRef font, CGGlyph g, double ox, double oy, int w, int h, unsigned char *out) {
	CGColorSpaceRef cs = CGColorSpaceCreateWithName(kCGColorSpaceSRGB);
	CGContextRef ctx = CGBitmapContextCreate(out, w, h, 8, w * 4, cs, kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big);
	CGColorSpaceRelease(cs);
	if (!ctx) return;
	CGContextSetAllowsAntialiasing(ctx, true);
	CGContextSetShouldAntialias(ctx, true);
	CGContextSetShouldSubpixelPositionFonts(ctx, true);
	CGContextSetShouldSubpixelQuantizeFonts(ctx, false);
	CGPoint pos = CGPointMake(ox, oy);
	CTFontDrawGlyphs(font, &g, &pos, 1, ctx);
	CGContextFlush(ctx);
	CGContextRelease(ctx);
}
*/
import "C"

import "unsafe"

func ctColorGlyphs(font uintptr) bool { return C.keelCTColorGlyphs(C.CTFontRef(font)) != 0 }

func ctGlyphBounds(font uintptr, gid uint16) (x, y, w, h float64) {
	var bx, by, bw, bh C.double
	C.keelCTGlyphBounds(C.CTFontRef(font), C.CGGlyph(gid), &bx, &by, &bw, &bh)
	return float64(bx), float64(by), float64(bw), float64(bh)
}

func ctDrawGlyph(font uintptr, gid uint16, ox, oy float64, w, h int, dark bool) []byte {
	buf := make([]byte, w*h*4)
	d := C.int(0)
	if dark {
		d = 1
	}
	C.keelCTDrawGlyph(C.CTFontRef(font), C.CGGlyph(gid), C.double(ox), C.double(oy), C.int(w), C.int(h), d, (*C.uchar)(unsafe.Pointer(&buf[0])))
	return buf
}

func ctDrawColorGlyph(font uintptr, gid uint16, ox, oy float64, w, h int) []byte {
	pix := make([]byte, w*h*4)
	C.keelCTDrawColorGlyph(C.CTFontRef(font), C.CGGlyph(gid), C.double(ox), C.double(oy), C.int(w), C.int(h), (*C.uchar)(unsafe.Pointer(&pix[0])))
	return pix
}
