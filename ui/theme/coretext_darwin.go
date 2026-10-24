//go:build darwin && !ios

package theme

import (
	"image"
	"math"
	"sync"
)

// ctFontKey identifies an opened CoreText font.
type ctFontKey struct {
	path     string
	index    int
	instance int
	px       float64
}

var (
	ctFontsMu sync.Mutex
	ctFonts   = map[ctFontKey]uintptr{}
)

func ctFont(path string, index int, px float64, instance ...int) uintptr {
	ctFontsMu.Lock()
	defer ctFontsMu.Unlock()
	var inst int
	if len(instance) > 0 {
		inst = instance[0]
	}
	k := ctFontKey{path: path, index: index, px: px, instance: inst}
	if f, ok := ctFonts[k]; ok {
		return f
	}
	f := openCTFontFace(path, index, px, inst)
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
func coreTextMask(path string, index int, px float64, gid uint16, phaseX, phaseY float64, dark bool, instance ...int) (bounds image.Rectangle, alpha []byte, ok bool) {
	return coreTextNativeMask(ctFont(path, index, px, instance...), gid, phaseX, phaseY, dark)
}

func coreTextNativeMask(native uintptr, gid uint16, phaseX, phaseY float64, dark bool) (bounds image.Rectangle, alpha []byte, ok bool) {
	f := native
	if f == 0 || ctColorGlyphs(f) {
		return image.Rectangle{}, nil, false
	}
	bx, by, bw, bh := ctGlyphBounds(f, gid)
	if bw <= 0 || bh <= 0 {
		return image.Rectangle{}, nil, true // a blank glyph such as a space
	}
	// Pixel box around the ink, y down, relative to the integral origin.
	minX := int(math.Floor(bx+phaseX)) - ctPad
	maxX := int(math.Ceil(bx+bw+phaseX)) + ctPad
	minY := int(math.Floor(-(by+bh)+phaseY)) - ctPad
	maxY := int(math.Ceil(-by+phaseY)) + ctPad
	w, h := maxX-minX, maxY-minY
	// CoreText's origin is bottom left, y up.
	ox := -float64(minX) + phaseX
	oy := float64(maxY) - phaseY
	buf := ctDrawGlyph(f, gid, ox, oy, w, h, dark)
	if buf == nil {
		return image.Rectangle{}, nil, false
	}
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
func coreTextColorGlyph(path string, index int, px float64, gid uint16, phaseX, phaseY float64, instance ...int) (bounds image.Rectangle, pix []byte, ok bool) {
	return coreTextNativeColorGlyph(ctFont(path, index, px, instance...), gid, phaseX, phaseY)
}

func coreTextNativeColorGlyph(native uintptr, gid uint16, phaseX, phaseY float64) (bounds image.Rectangle, pix []byte, ok bool) {
	f := native
	if f == 0 || !ctColorGlyphs(f) {
		return image.Rectangle{}, nil, false
	}
	bx, by, bw, bh := ctGlyphBounds(f, gid)
	if bw <= 0 || bh <= 0 {
		return image.Rectangle{}, nil, true
	}
	minX := int(math.Floor(bx+phaseX)) - ctPad
	maxX := int(math.Ceil(bx+bw+phaseX)) + ctPad
	minY := int(math.Floor(-(by+bh)+phaseY)) - ctPad
	maxY := int(math.Ceil(-by+phaseY)) + ctPad
	w, h := maxX-minX, maxY-minY
	pix = ctDrawColorGlyph(f, gid, -float64(minX)+phaseX, float64(maxY)-phaseY, w, h)
	if pix == nil {
		return image.Rectangle{}, nil, false
	}
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
