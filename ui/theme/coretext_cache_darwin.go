//go:build darwin && !ios

package theme

import (
	"image"
	"image/color"
	"math"
	"os"
	"sync"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"
)

// macOS draws text with CoreText, as native and WebKit apps do: the same
// antialiasing, stem darkening and gamma. Gio's vector outlines, blended in
// linear light, look thin and grey beside them. Glyph masks come from
// CoreText and are drawn as tinted images; text CoreText cannot draw (fonts
// built in memory, color emoji) keeps the vector path. It needs Keel's Gio
// copy (third_party): with upstream Gio the hooks below are absent and text
// keeps vector outlines.

// rasterShaper and glyphFiler are the Keel Gio patches this uses. Upstream
// Gio lacks them; detecting them keeps Keel building against it.
type rasterShaper interface {
	SetRasterHook(func(ops *op.Ops, gs []text.Glyph, col color.NRGBA, fracX float32) bool)
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
	r.SetRasterHook(func(ops *op.Ops, gs []text.Glyph, col color.NRGBA, fracX float32) bool {
		return ctPaintRun(ops, sh, gs, col, fracX, 0)
	})
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

const (
	ctPageSize   = 1024
	ctMaxPages   = 8 // 32 MiB of page pixels at most
	ctPhases     = 4 // horizontal subpixel positions per pixel
	ctMaxPPEM    = 160
	ctPagePadded = 1 // gap between masks so sampling never bleeds
)

type ctGlyphKey struct {
	sh             *text.Shaper
	id             text.GlyphID
	phaseX, phaseY uint8
	dark           bool
}

type ctGlyph struct {
	vector bool            // drawn as an outline: no file or too large
	blank  bool            // no ink
	color  *paint.ImageOp  // a color glyph (emoji), drawn untinted as CoreText shows it
	page   int             // page index while cached in a page
	rect   image.Rectangle // location in the page
	bounds image.Rectangle // pixels relative to the glyph origin, y down
}

type ctPage struct {
	img        *image.RGBA
	op         paint.ImageOp
	tint       tintedImage
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
	vectorFrom := -1 // first glyph of a pending vector sub-run
	flushVector := func(end int) {
		if vectorFrom < 0 {
			return
		}
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
			p.op = paint.NewImageOp(p.img)
			p.tint, _ = any(p.op).(tintedImage)
			p.opValid, p.opFrame = true, ct.frame
		}
		off := image.Pt(int(ix), int(iy)).Add(gl.bounds.Min).Sub(gl.rect.Min)
		tr := op.Offset(off).Push(ops)
		cl := clip.Rect(gl.rect).Push(ops)
		p.tint.AddTinted(ops, col)
		paint.PaintOp{}.Add(ops)
		cl.Pop()
		tr.Pop()
	}
	flushVector(len(gs))
	return true
}

// ctGlyphFor returns the cached mask of g, rasterizing it on first use.
func ctGlyphFor(sh *text.Shaper, g text.Glyph, phaseX, phaseY uint8, dark bool) *ctGlyph {
	key := ctGlyphKey{sh, g.ID, phaseX, phaseY, dark}
	if gl, ok := ct.glyphs[key]; ok {
		return gl
	}
	gl := &ctGlyph{vector: true}
	path, index, ppem, gid, ok := any(sh).(glyphFiler).GlyphFile(g.ID)
	if ok && ppem > 0 && ppem <= fixed.I(ctMaxPPEM) {
		if b, pix, isColor := coreTextColorGlyph(path, index, float64(ppem)/64, gid, float64(phaseX)/ctPhases, float64(phaseY)/ctPhases); isColor {
			// Emoji are few: each keeps its own image. The key's dark flag
			// duplicates them per ink tone, which is harmless.
			if pix == nil {
				gl = &ctGlyph{blank: true}
			} else {
				img := &image.RGBA{Pix: pix, Stride: b.Dx() * 4, Rect: image.Rectangle{Max: b.Size()}}
				imageOp := paint.NewImageOp(img)
				gl = &ctGlyph{color: &imageOp, bounds: b}
			}
			ct.glyphs[key] = gl
			return gl
		}
		bounds, alpha, ok := coreTextMask(path, index, float64(ppem)/64, gid, float64(phaseX)/ctPhases, float64(phaseY)/ctPhases, dark)
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
	ct.glyphs[key] = gl
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
		ct.pages = append(ct.pages, &ctPage{img: image.NewRGBA(image.Rect(0, 0, ctPageSize, ctPageSize))})
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
	*p = ctPage{img: image.NewRGBA(image.Rect(0, 0, ctPageSize, ctPageSize))}
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

// ctWhite encodes linear coverage as premultiplied white: image textures
// store sRGB-encoded premultiplied linear colors.
var ctWhite = func() (lut [256]byte) {
	for i := range lut {
		v := float64(i) / 255
		if v <= 0.0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - 0.055
		}
		lut[i] = byte(math.Round(v * 255))
	}
	return
}()

// ctWrite stores a mask as white premultiplied coverage. A page whose image
// backed a texture in an earlier frame is copied first.
func ctWrite(p *ctPage, r image.Rectangle, alpha []byte) {
	if p.opValid && p.opFrame != ct.frame {
		img := image.NewRGBA(p.img.Rect)
		copy(img.Pix, p.img.Pix)
		p.img, p.opValid = img, false
	}
	w := r.Dx()
	for y := 0; y < r.Dy(); y++ {
		off := p.img.PixOffset(r.Min.X, r.Min.Y+y)
		for x, a := range alpha[y*w : (y+1)*w] {
			px := p.img.Pix[off+x*4 : off+x*4+4 : off+x*4+4]
			v := ctWhite[a]
			px[0], px[1], px[2], px[3] = v, v, v, a
		}
	}
}

// textLuma is the gamma-encoded luma of c, 0 for black and 1 for white:
// text below one half reads as dark ink, above as light.
func textLuma(c color.NRGBA) float64 {
	return (0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)) / 255
}
