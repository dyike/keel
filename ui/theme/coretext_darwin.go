//go:build darwin && !ios

package theme

/*
#cgo LDFLAGS: -framework CoreText -framework CoreGraphics -framework CoreFoundation

#include <stdlib.h>
#include <CoreText/CoreText.h>
#include <CoreGraphics/CoreGraphics.h>

// keelCTFont opens face index of the font file at path, at size pixels.
// The caller releases it with CFRelease.
static CTFontRef keelCTFont(const char *path, int index, double size) {
	CFStringRef p = CFStringCreateWithCString(NULL, path, kCFStringEncodingUTF8);
	if (!p) return NULL;
	CFURLRef url = CFURLCreateWithFileSystemPath(NULL, p, kCFURLPOSIXPathStyle, false);
	CFRelease(p);
	if (!url) return NULL;
	CFArrayRef descs = CTFontManagerCreateFontDescriptorsFromURL(url);
	CFRelease(url);
	if (!descs) return NULL;
	CTFontRef font = NULL;
	if (index >= 0 && index < CFArrayGetCount(descs)) {
		CTFontDescriptorRef d = (CTFontDescriptorRef)CFArrayGetValueAtIndex(descs, index);
		font = CTFontCreateWithFontDescriptor(d, size, NULL);
	}
	CFRelease(descs);
	return font;
}

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

import (
	"image"
	"math"
	"sync"
	"unsafe"
)

// ctFontKey identifies an opened CoreText font.
type ctFontKey struct {
	path  string
	index int
	px    float64
}

var (
	ctFontsMu sync.Mutex
	ctFonts   = map[ctFontKey]C.CTFontRef{}
)

func ctFont(path string, index int, px float64) C.CTFontRef {
	ctFontsMu.Lock()
	defer ctFontsMu.Unlock()
	k := ctFontKey{path, index, px}
	if f, ok := ctFonts[k]; ok {
		return f
	}
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	f := C.keelCTFont(cpath, C.int(index), C.double(px))
	// A bounded set of sizes is used by an app; keep fonts for its lifetime.
	ctFonts[k] = f
	return f
}

// ctPad leaves room for CoreText's antialiasing and stem darkening.
const ctPad = 2

// coreTextMask rasterizes one glyph with CoreText, as a window would show it,
// and returns the linear-light coverage that reproduces those pixels when the
// GPU blends in linear light. phaseX and phaseY are the glyph origin's
// fraction of a pixel, right and down. bounds is relative to the origin,
// y down. dark selects dark text on a light background, otherwise light on
// dark: CoreText's antialiasing differs, and so does the conversion.
func coreTextMask(path string, index int, px float64, gid uint16, phaseX, phaseY float64, dark bool) (bounds image.Rectangle, alpha []byte, ok bool) {
	f := ctFont(path, index, px)
	if f == 0 || C.keelCTColorGlyphs(f) != 0 {
		return image.Rectangle{}, nil, false
	}
	var bx, by, bw, bh C.double
	C.keelCTGlyphBounds(f, C.CGGlyph(gid), &bx, &by, &bw, &bh)
	if bw <= 0 || bh <= 0 {
		return image.Rectangle{}, nil, true // a blank glyph such as a space
	}
	// Pixel box around the ink, y down, relative to the integral origin.
	minX := int(math.Floor(float64(bx)+phaseX)) - ctPad
	maxX := int(math.Ceil(float64(bx+bw)+phaseX)) + ctPad
	minY := int(math.Floor(-float64(by+bh)+phaseY)) - ctPad
	maxY := int(math.Ceil(-float64(by)+phaseY)) + ctPad
	w, h := maxX-minX, maxY-minY
	buf := make([]byte, w*h*4)
	d := C.int(0)
	if dark {
		d = 1
	}
	// CoreText's origin is bottom left, y up.
	ox := -float64(minX) + phaseX
	oy := float64(maxY) - phaseY
	C.keelCTDrawGlyph(f, C.CGGlyph(gid), C.double(ox), C.double(oy), C.int(w), C.int(h), d, (*C.uchar)(unsafe.Pointer(&buf[0])))
	alpha = make([]byte, w*h)
	lut := &ctLight
	if dark {
		lut = &ctDark
	}
	any := false
	for i := range alpha {
		alpha[i] = lut[buf[i*4+1]] // green dominates luminance
		any = any || alpha[i] != 0
	}
	if !any {
		return image.Rectangle{}, nil, true
	}
	return image.Rect(minX, minY, maxX, maxY), alpha, true
}

// coreTextColorGlyph draws a color glyph (emoji) as CoreText shows it, in
// premultiplied sRGB, the format image textures use. bounds is relative to the
// glyph origin, y down. ok is false for fonts without color glyphs.
func coreTextColorGlyph(path string, index int, px float64, gid uint16, phaseX, phaseY float64) (bounds image.Rectangle, pix []byte, ok bool) {
	f := ctFont(path, index, px)
	if f == 0 || C.keelCTColorGlyphs(f) == 0 {
		return image.Rectangle{}, nil, false
	}
	var bx, by, bw, bh C.double
	C.keelCTGlyphBounds(f, C.CGGlyph(gid), &bx, &by, &bw, &bh)
	if bw <= 0 || bh <= 0 {
		return image.Rectangle{}, nil, true
	}
	minX := int(math.Floor(float64(bx)+phaseX)) - ctPad
	maxX := int(math.Ceil(float64(bx+bw)+phaseX)) + ctPad
	minY := int(math.Floor(-float64(by+bh)+phaseY)) - ctPad
	maxY := int(math.Ceil(-float64(by)+phaseY)) + ctPad
	w, h := maxX-minX, maxY-minY
	pix = make([]byte, w*h*4)
	C.keelCTDrawColorGlyph(f, C.CGGlyph(gid), C.double(-float64(minX)+phaseX), C.double(float64(maxY)-phaseY), C.int(w), C.int(h), (*C.uchar)(unsafe.Pointer(&pix[0])))
	return image.Rect(minX, minY, maxX, maxY), pix, true
}

// ctDark and ctLight map a CoreText pixel to the linear-light alpha that
// reproduces it: for black on white the display value is 1-c, for white on
// black it is c, where c is CoreText's sRGB-space coverage.
var ctDark, ctLight = func() (dark, light [256]byte) {
	for v := range 256 {
		s := float64(v) / 255
		dark[v] = byte(math.Round(255 * (1 - srgbToLinear(s))))
		light[v] = byte(math.Round(255 * srgbToLinear(s)))
	}
	return
}()

func srgbToLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}
