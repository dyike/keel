//go:build darwin && !ios

package theme

import (
	"container/list"
	"image"
	"image/color"
	"math"
	"os"
	"runtime"
	"sync"
	"weak"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/text"
	"golang.org/x/image/math/fixed"
)

// CoreText owns matching, shaping and rasterization for platform shapers.
// Masks keep CoreText's antialiasing, stem darkening and gamma; color fonts
// keep their RGBA images. Opaque custom Go faces retain the Go shaper and
// outline fallback, with file-backed glyphs still eligible for native masks.

// rasterShaper and glyphFiler are the Keel Gio patches this uses. Upstream
// Gio lacks them; detecting them keeps Keel building against it.
type rasterShaper interface {
	SetRasterHook(func(ops *op.Ops, gs []text.Glyph, col color.NRGBA, fracX float32) bool)
}

type glyphInstanceFiler interface {
	GlyphFileInstance(id text.GlyphID) (path string, index, instance int, ppem fixed.Int26_6, gid uint16, ok bool)
}

type glyphFiler interface {
	GlyphFile(id text.GlyphID) (path string, index int, ppem fixed.Int26_6, gid uint16, ok bool)
}

// KEEL_TEXT=vector keeps Gio's vector outlines, for comparison.
var ctDisabled = os.Getenv("KEEL_TEXT") == "vector"

// UsePlatformText turns CoreText glyph drawing on or off; it is on by default
// where Keel's Gio copy and GPU backend support it.
func UsePlatformText(on bool) {
	ct.Lock()
	ctDisabled = !on
	ct.Unlock()
}

// installPlatformText makes sh draw solid text with CoreText.
func installPlatformText(sh *text.Shaper) {
	r, ok := any(sh).(rasterShaper)
	if !ok || !atlasTintSupported {
		return
	}
	if _, ok := any(sh).(glyphFiler); !ok {
		return
	}
	hook := func(ops *op.Ops, gs []text.Glyph, col color.NRGBA, fracX float32) bool {
		return ctPaintRun(ops, sh, gs, col, fracX, 0)
	}
	// ctPaintRun consumes gs synchronously. Its atlas stores glyph IDs and
	// image/path operations, never the supplied run slice. Vector fallback's
	// Shape/Bitmaps caches also copy their glyph-key metadata.
	if borrowed, ok := any(sh).(interface{ SetBorrowedRasterHook(text.RasterHook) }); ok {
		borrowed.SetBorrowedRasterHook(hook)
	} else {
		r.SetRasterHook(hook)
	}
}

// PlatformText reports whether sh draws solid text with CoreText.
func PlatformText(sh *text.Shaper) bool {
	_, ok := any(sh).(glyphFiler)
	ct.Lock()
	defer ct.Unlock()
	return ok && atlasTintSupported && !ctDisabled
}

// platformPaintRun draws gs as Shaper.Shape places them, at the current
// integral translation plus (fracX, fracY), with CoreText. It reports false
// when unavailable; the caller then draws vector outlines.
func platformPaintRun(ops *op.Ops, sh *text.Shaper, gs []text.Glyph, col color.NRGBA, fracX, fracY float32) bool {
	if _, ok := any(sh).(glyphFiler); !ok || !atlasTintSupported {
		return false
	}
	return ctPaintRun(ops, sh, gs, col, fracX, fracY)
}

const platformFontPreparation = true

const (
	// Small immutable pages keep linear coverage at one byte per pixel.
	// Portable devices and opacity layers expand only the uploaded texture.
	// Copy-on-write preserves operations queued by another window.
	ctPageSize   = 256
	ctMaxPages   = 128 // 8 MiB of single-channel page pixels at most
	ctPhases     = 4   // horizontal subpixel positions per pixel
	ctMaxPPEM    = 160
	ctPagePadded = 1 // gap between masks so sampling never bleeds
)

type ctGlyphKey struct {
	sh             weak.Pointer[text.Shaper]
	id             text.GlyphID
	phaseX, phaseY uint8
	dark           bool
}

type ctGlyph struct {
	extra *list.Element

	vector bool            // drawn as an outline: no file or too large
	blank  bool            // no ink
	color  *paint.ImageOp  // a color glyph (emoji), drawn untinted as CoreText shows it
	page   int             // page index while cached in a page
	rect   image.Rectangle // location in the page
	bounds image.Rectangle // pixels relative to the glyph origin, y down
}

type ctPage struct {
	img        *image.Alpha
	op         paint.ImageOp
	opValid    bool
	opFrame    uint64
	x, y, rowH int
	used       uint64
	keys       []ctGlyphKey
}

var ct struct {
	sync.Mutex
	frame  uint64
	glyphs map[ctGlyphKey]*ctGlyph
	pages  []*ctPage
	extra  *ctExtraCache
	// quads is ctPaintRun's scratch for a run of masks on one page.
	quads []paint.ImageQuad
}

// BeginTextFrame marks the start of a window's layout. Glyph pages changed
// after the previous frame are copied before they change again, so textures
// the GPU already holds never show different pixels. el calls it.
func BeginTextFrame() {
	ct.Lock()
	ct.frame++
	ct.Unlock()
}

// ctPaintRun draws a run with CoreText masks.
func ctPaintRun(ops *op.Ops, sh *text.Shaper, gs []text.Glyph, col color.NRGBA, fracX, fracY float32) bool {
	if len(gs) == 0 || sh == nil {
		return false
	}
	ct.Lock()
	defer ct.Unlock()
	if ctDisabled {
		return false
	}
	if ct.glyphs == nil {
		ct.glyphs = map[ctGlyphKey]*ctGlyph{}
	}
	dark := textLuma(col) < 0.5
	// Masks on one page are painted with one operation; see PaintTintedQuads.
	var quadPage *ctPage
	quads := ct.quads[:0]
	flushQuads := func() {
		if len(quads) > 0 {
			quadPage.op.PaintTintedQuads(ops, col, quads)
			quads = quads[:0]
		}
	}
	defer func() { ct.quads = quads[:0] }()
	vectorFrom := -1 // first glyph of a pending vector sub-run
	flushVector := func(end int) {
		if vectorFrom < 0 {
			return
		}
		flushQuads()
		sub := gs[vectorFrom:end]
		x := fracX + float32(sub[0].X-gs[0].X)/64
		tr := op.Affine(f32.AffineId().Offset(f32.Pt(x, fracY))).Push(ops)
		paintGlyphRun(ops, sh, sub, col)
		tr.Pop()
		vectorFrom = -1
	}
	for i, g := range gs {
		x := float64(fracX) + float64(g.X-gs[0].X-g.Offset.X)/64
		ix := math.Floor(x)
		phase := int(math.Round((x - ix) * ctPhases))
		if phase == ctPhases {
			ix, phase = ix+1, 0
		}
		y := float64(fracY) - float64(g.Offset.Y)/64
		iy := math.Floor(y)
		phaseY := int(math.Round((y - iy) * ctPhases))
		if phaseY == ctPhases {
			iy, phaseY = iy+1, 0
		}
		gl := ctGlyphFor(sh, g, uint8(phase), uint8(phaseY), dark)
		if gl.vector {
			if vectorFrom < 0 {
				vectorFrom = i
			}
			continue
		}
		flushVector(i)
		if gl.blank {
			continue
		}
		if gl.color != nil {
			flushQuads()
			tr := op.Offset(image.Pt(int(ix), int(iy)).Add(gl.bounds.Min)).Push(ops)
			cl := clip.Rect(image.Rectangle{Max: gl.bounds.Size()}).Push(ops)
			gl.color.Add(ops)
			paint.PaintOp{}.Add(ops)
			cl.Pop()
			tr.Pop()
			continue
		}
		p := ct.pages[gl.page]
		p.used = ct.frame
		if !p.opValid {
			p.op = paint.NewMaskOp(p.img)
			p.opValid, p.opFrame = true, ct.frame
		}
		if p != quadPage {
			flushQuads()
			quadPage = p
		}
		off := image.Pt(int(ix), int(iy)).Add(gl.bounds.Min).Sub(gl.rect.Min)
		quads = append(quads, paint.ImageQuad{Offset: off, Rect: gl.rect})
	}
	flushVector(len(gs))
	flushQuads()
	return true
}

// ctGlyphFor returns the cached mask of g, rasterizing it on first use.
func ctGlyphFor(sh *text.Shaper, g text.Glyph, phaseX, phaseY uint8, dark bool) *ctGlyph {
	// Native font handles are borrowed. A weak cache key must not let their
	// shaper be cleaned up while a raster allocation or native call runs.
	defer runtime.KeepAlive(sh)
	key := ctGlyphKey{weak.Make(sh), g.ID, phaseX, phaseY, dark}
	if gl, ok := ct.glyphs[key]; ok {
		if gl.extra != nil {
			ct.extra.order.MoveToFront(gl.extra)
		}
		return gl
	}
	gl := &ctGlyph{vector: true}
	var native uintptr
	var ppem fixed.Int26_6
	var gid uint16
	var ok bool
	native, ppem, gid, ok = sh.GlyphNativeFont(g.ID)
	var path string
	var index, instance int
	if !ok {
		path, index, ppem, gid, ok = any(sh).(glyphFiler).GlyphFile(g.ID)
		if f, exists := any(sh).(glyphInstanceFiler); exists {
			path, index, instance, ppem, gid, ok = f.GlyphFileInstance(g.ID)
		}
	}
	colorGlyph := func() (image.Rectangle, []byte, bool) {
		if native != 0 {
			return coreTextNativeColorGlyph(native, gid, float64(phaseX)/ctPhases, float64(phaseY)/ctPhases)
		}
		return coreTextColorGlyph(path, index, float64(ppem)/64, gid, float64(phaseX)/ctPhases, float64(phaseY)/ctPhases, instance)
	}
	maskGlyph := func() (image.Rectangle, []byte, bool) {
		if native != 0 {
			return coreTextNativeMask(native, gid, float64(phaseX)/ctPhases, float64(phaseY)/ctPhases, dark)
		}
		return coreTextMask(path, index, float64(ppem)/64, gid, float64(phaseX)/ctPhases, float64(phaseY)/ctPhases, dark, instance)
	}

	if ok && ppem > 0 && ppem <= fixed.I(ctMaxPPEM) {
		if b, pix, isColor := colorGlyph(); isColor {
			// Emoji are few: each keeps its own image. The key's dark flag
			// duplicates them per ink tone, which is harmless.
			if pix == nil {
				gl = &ctGlyph{blank: true}
			} else {
				img := &image.RGBA{Pix: pix, Stride: b.Dx() * 4, Rect: image.Rectangle{Max: b.Size()}}
				imageOp := paint.NewImageOp(img)
				gl = &ctGlyph{color: &imageOp, bounds: b}
			}
			ctRememberExtra(key, gl, len(pix))
			return gl
		}
		bounds, alpha, ok := maskGlyph()
		switch {
		case !ok:
		case alpha == nil:
			gl = &ctGlyph{blank: true}
		default:
			if page, rect, ok := ctPlace(bounds.Size(), key); ok {
				ctWrite(ct.pages[page], rect, alpha)
				gl = &ctGlyph{page: page, rect: rect, bounds: bounds}
			} else {
				// Every page is in use this frame: draw it as an outline,
				// without caching, and try again next frame.
				return gl
			}
		}
	}
	if gl.blank || gl.vector {
		ctRememberExtra(key, gl, 0)
	} else {
		ct.glyphs[key] = gl
	}
	return gl
}

// ctPlace finds room for a mask, evicting the least recently used page that
// this frame has not drawn from when all are full.
func ctPlace(size image.Point, key ctGlyphKey) (int, image.Rectangle, bool) {
	if size.X+ctPagePadded > ctPageSize || size.Y+ctPagePadded > ctPageSize {
		return 0, image.Rectangle{}, false
	}
	for i, p := range ct.pages {
		if r, ok := p.place(size); ok {
			p.keys = append(p.keys, key)
			return i, r, true
		}
	}
	if len(ct.pages) < ctMaxPages {
		ct.pages = append(ct.pages, &ctPage{img: image.NewAlpha(image.Rect(0, 0, ctPageSize, ctPageSize))})
		i := len(ct.pages) - 1
		r, _ := ct.pages[i].place(size)
		ct.pages[i].keys = append(ct.pages[i].keys, key)
		return i, r, true
	}
	victim := -1
	for i, p := range ct.pages {
		if p.used != ct.frame && (victim < 0 || p.used < ct.pages[victim].used) {
			victim = i
		}
	}
	if victim < 0 {
		return 0, image.Rectangle{}, false
	}
	p := ct.pages[victim]
	for _, k := range p.keys {
		delete(ct.glyphs, k)
	}
	// A fresh image: the old one may back textures drawn in earlier frames.
	*p = ctPage{img: image.NewAlpha(image.Rect(0, 0, ctPageSize, ctPageSize))}
	r, _ := p.place(size)
	p.keys = append(p.keys, key)
	return victim, r, true
}

func (p *ctPage) place(size image.Point) (image.Rectangle, bool) {
	w, h := size.X+ctPagePadded, size.Y+ctPagePadded
	if p.x+w > ctPageSize {
		p.x, p.y, p.rowH = 0, p.y+p.rowH, 0
	}
	if p.y+h > ctPageSize {
		return image.Rectangle{}, false
	}
	r := image.Rect(p.x, p.y, p.x+size.X, p.y+size.Y)
	p.x += w
	p.rowH = max(p.rowH, h)
	return r, true
}

// ctWrite stores one linear coverage byte per pixel. A page whose image
// backed a texture in an earlier frame is copied first.
func ctWrite(p *ctPage, r image.Rectangle, alpha []byte) {
	if p.opValid && p.opFrame != ct.frame {
		img := image.NewAlpha(p.img.Rect)
		copy(img.Pix, p.img.Pix)
		p.img, p.opValid = img, false
	}
	w := r.Dx()
	for y := 0; y < r.Dy(); y++ {
		off := p.img.PixOffset(r.Min.X, r.Min.Y+y)
		copy(p.img.Pix[off:off+w], alpha[y*w:(y+1)*w])
	}
}

// textLuma is the gamma-encoded luma of c, 0 for black and 1 for white:
// text below one half reads as dark ink, above as light.
func textLuma(c color.NRGBA) float64 {
	return (0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)) / 255
}

// Glyphs outside atlas pages (color images, blanks and vector fallback
// markers) need their own bound: atlas-page eviction never visits them.
// Keys weakly reference shapers so closing windows can release native fonts.
const ctExtraBudget = 2 << 20
const ctExtraEntries = 512

type ctExtraEntry struct {
	key   ctGlyphKey
	bytes int
}
type ctExtraCache struct {
	order list.List
	bytes int
}

func ctRememberExtra(key ctGlyphKey, gl *ctGlyph, pixels int) {
	if ct.extra == nil {
		ct.extra = new(ctExtraCache)
	}
	if pixels > ctExtraBudget {
		return
	} // draw oversized color glyphs uncached
	gl.extra = ct.extra.order.PushFront(ctExtraEntry{key, pixels})
	ct.extra.bytes += pixels
	ct.glyphs[key] = gl
	for ct.extra.bytes > ctExtraBudget || ct.extra.order.Len() > ctExtraEntries {
		last := ct.extra.order.Back()
		entry := last.Value.(ctExtraEntry)
		delete(ct.glyphs, entry.key)
		ct.extra.bytes -= entry.bytes
		ct.extra.order.Remove(last)
	}
}
