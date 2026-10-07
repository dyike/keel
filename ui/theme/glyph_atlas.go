package theme

import (
	"image"
	"image/color"
	"math"

	"gioui.org/gpu/headless"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"
)

const (
	glyphAtlasSide  = 512
	glyphAtlasPages = 8
	glyphMaskSide   = 64
	glyphMaskLimit  = 2 << 20
)

// GlyphAtlas reuses rasterized vector glyphs in immutable image pages. It is
// opt-in for single-line text drawn at integral pixel translations, without
// extra scale or rotation. Complex runs use the normal vector drawing path.
//
// Each frame: BeginFrame, Prepare all runs, Commit, then Paint those runs.
// Preparing all runs before painting lets each changed page be uploaded once.
// Cached page pixels are bounded to 8 MiB, masks to 2 MiB and 4096 entries.
// GPU textures and recent immutable page snapshots add to that memory.
// Use serially, keep it across frames, and call Release when no longer needed.
// Do not copy a GlyphAtlas after first use.
type GlyphAtlas struct {
	// SubpixelPhases has the same meaning as GlyphPainter.SubpixelPhases.
	// Zero preserves the exact horizontal positions.
	SubpixelPhases           int
	vector                   GlyphPainter
	shaper                   *text.Shaper
	window                   *headless.Window
	scratch                  op.Ops
	failed                   bool
	committed                bool
	frame                    uint64
	masks                    map[glyphFragmentKey]*atlasMask
	maskBytes                int
	lastMaskFrame            uint64
	newMasks                 int
	rasterDraws, vectorDraws uint64
	pages                    []*atlasPage
}

type atlasMask struct {
	bounds image.Rectangle
	alpha  []byte
	colors map[color.NRGBA]atlasLocation
}

type atlasLocation struct {
	page int
	rect image.Rectangle
}

type atlasPage struct {
	pixels          *image.RGBA
	image           paint.ImageOp
	dirty           bool
	x, y, rowHeight int
	used            uint64
}

// BeginFrame starts preparing text for a frame. Changing the shaper drops all
// caches so font glyph IDs cannot be confused with those of the previous one.
func (a *GlyphAtlas) BeginFrame(sh *text.Shaper) {
	if a.shaper != sh {
		a.Release()
		a.shaper = sh
	}
	a.frame++
	a.committed = false
	a.newMasks = 0
	a.vector.FragmentSize = 1
	a.vector.SubpixelPhases = a.SubpixelPhases
}

// Prepare caches the vector glyphs and colors required by a run. params must
// match the parameters used to shape gs. It may replace the shaper's iterator.
// Cache limits or an unavailable offscreen GPU leave a run on the vector path.
func (a *GlyphAtlas) Prepare(params text.Parameters, gs []text.Glyph, col color.NRGBA) {
	if a.shaper == nil || a.committed {
		return
	}
	a.vector.walkFragments(a.shaper, params, gs, col, func(fragment []text.Glyph, _ fixed.Int26_6) {
		key, f := a.vector.cachedFragment(a.shaper, fragment)
		m, known := a.masks[key]
		if !known {
			if len(a.masks) >= 4096 || a.frame > 1 && a.newMasks >= 32 {
				return
			}
			m = a.makeMask(f.path, fragment)
			if a.masks == nil {
				a.masks = make(map[glyphFragmentKey]*atlasMask)
			}
			a.masks[key] = m
		}
		if m == nil || len(m.alpha) == 0 {
			return
		}
		if loc, ok := m.colors[col]; ok {
			a.pages[loc.page].used = a.frame
			return
		}
		if len(m.colors) >= 8 {
			return
		}
		loc, ok := a.allocate(m.bounds.Size())
		if !ok {
			return
		}
		page := a.pages[loc.page]
		lut := atlasColorTable(col)
		for y := 0; y < m.bounds.Dy(); y++ {
			for x := 0; x < m.bounds.Dx(); x++ {
				alpha := m.alpha[y*m.bounds.Dx()+x]
				off := page.pixels.PixOffset(loc.rect.Min.X+x, loc.rect.Min.Y+y)
				copy(page.pixels.Pix[off:off+3], lut[alpha][:])
				page.pixels.Pix[off+3] = alpha
			}
		}
		m.colors[col] = loc
	})
}

// Commit freezes changed image pages. Prepare must not be called again until
// the next BeginFrame. Existing ImageOps always keep their pixels immutable.
func (a *GlyphAtlas) Commit() {
	for _, p := range a.pages {
		if p.dirty {
			p.image = paint.NewImageOp(p.pixels)
			p.image.Filter = paint.FilterNearest
			p.dirty = false
		}
	}
	// The scratch renderer can retain large coverage textures even for a tiny
	// viewport. Keep only the CPU masks between frames, not its GPU resources.
	if a.window != nil && a.frame > a.lastMaskFrame+60 {
		a.window.Release()
		a.window = nil
		a.scratch.Reset()
	}
	a.committed = true
}

// Paint draws a previously prepared run at the same origin as Shaper.Shape.
// The caller owns clipping, baseline translation and semantic operations.
func (a *GlyphAtlas) Paint(ops *op.Ops, params text.Parameters, gs []text.Glyph, col color.NRGBA) {
	if len(gs) == 0 || a.shaper == nil {
		return
	}
	if !a.committed {
		a.vector.Paint(ops, a.shaper, params, gs, col)
		return
	}
	if !a.vector.walkFragments(a.shaper, params, gs, col, func(fragment []text.Glyph, displacement fixed.Int26_6) {
		key, f := a.vector.cachedFragment(a.shaper, fragment)
		tr := op.Offset(image.Pt(displacement.Round(), 0)).Push(ops)
		m := a.masks[key]
		loc, ok := atlasLocation{}, false
		if m != nil {
			loc, ok = m.colors[col]
		}
		if ok {
			a.rasterDraws++
			page := a.pages[loc.page]
			page.used = a.frame
			offset := m.bounds.Min.Sub(loc.rect.Min)
			imgTr := op.Offset(offset).Push(ops)
			cl := clip.Rect(loc.rect).Push(ops)
			page.image.Add(ops)
			paint.PaintOp{}.Add(ops)
			cl.Pop()
			imgTr.Pop()
		} else if m == nil || len(m.alpha) != 0 {
			a.vectorDraws++
			paint.ColorOp{Color: col}.Add(ops)
			cl := clip.Outline{Path: f.path}.Op().Push(ops)
			paint.PaintOp{}.Add(ops)
			cl.Pop()
		}
		f.bitmap.Add(ops)
		tr.Pop()
	}) {
		a.vectorDraws += uint64(len(gs))
		paintGlyphRun(ops, a.shaper, gs, col)
	}
}

// Release frees the scratch GPU and drops cached masks and image pages.
// The atlas may be used again by calling BeginFrame.
func (a *GlyphAtlas) Release() {
	if a.window != nil {
		a.window.Release()
	}
	phases := a.SubpixelPhases
	*a = GlyphAtlas{SubpixelPhases: phases}
}

func (a *GlyphAtlas) makeMask(path clip.PathSpec, gs []text.Glyph) *atlasMask {
	if a.failed {
		return nil
	}
	var bounds image.Rectangle
	for _, g := range gs {
		if g.Bounds.Min == g.Bounds.Max {
			continue
		}
		x := g.X - gs[0].X
		bounds = bounds.Union(image.Rect((x + g.Bounds.Min.X).Floor(), g.Bounds.Min.Y.Floor(), (x + g.Bounds.Max.X).Ceil(), g.Bounds.Max.Y.Ceil()))
	}
	if bounds.Empty() {
		return &atlasMask{}
	}
	bounds = bounds.Inset(-1)
	if bounds.Dx() > glyphMaskSide || bounds.Dy() > glyphMaskSide || a.maskBytes+bounds.Dx()*bounds.Dy() > glyphMaskLimit {
		return nil
	}
	if a.window == nil {
		w, err := headless.NewWindow(glyphMaskSide, glyphMaskSide)
		if err != nil {
			a.failed = true
			return nil
		}
		a.window = w
	}
	a.newMasks++
	a.lastMaskFrame = a.frame
	a.scratch.Reset()
	tr := op.Offset(bounds.Min.Mul(-1)).Push(&a.scratch)
	paint.ColorOp{Color: color.NRGBA{R: 255, G: 255, B: 255, A: 255}}.Add(&a.scratch)
	cl := clip.Outline{Path: path}.Op().Push(&a.scratch)
	paint.PaintOp{}.Add(&a.scratch)
	cl.Pop()
	tr.Pop()
	if err := a.window.Frame(&a.scratch); err != nil {
		a.failed = true
		return nil
	}
	img := image.NewRGBA(image.Rectangle{Max: bounds.Size()})
	if err := a.window.Screenshot(img); err != nil {
		a.failed = true
		return nil
	}
	m := &atlasMask{bounds: bounds, alpha: make([]byte, bounds.Dx()*bounds.Dy()), colors: make(map[color.NRGBA]atlasLocation)}
	for i := range m.alpha {
		m.alpha[i] = img.Pix[4*i+3]
	}
	a.maskBytes += len(m.alpha)
	return m
}

func (a *GlyphAtlas) allocate(size image.Point) (atlasLocation, bool) {
	for i, p := range a.pages {
		if rect, ok := p.place(size); ok {
			p.makeWritable()
			p.used = a.frame
			return atlasLocation{i, rect}, true
		}
	}
	if len(a.pages) < glyphAtlasPages {
		p := &atlasPage{pixels: image.NewRGBA(image.Rect(0, 0, glyphAtlasSide, glyphAtlasSide)), dirty: true, used: a.frame}
		a.pages = append(a.pages, p)
		r, _ := p.place(size)
		return atlasLocation{len(a.pages) - 1, r}, true
	}
	// A page used by this frame cannot be recycled: prepared runs refer to it.
	oldest := -1
	for i, p := range a.pages {
		if p.used != a.frame && (oldest < 0 || p.used < a.pages[oldest].used) {
			oldest = i
		}
	}
	if oldest < 0 {
		return atlasLocation{}, false
	}
	for _, m := range a.masks {
		if m != nil {
			for col, loc := range m.colors {
				if loc.page == oldest {
					delete(m.colors, col)
				}
			}
		}
	}
	p := &atlasPage{pixels: image.NewRGBA(image.Rect(0, 0, glyphAtlasSide, glyphAtlasSide)), dirty: true, used: a.frame}
	a.pages[oldest] = p
	r, _ := p.place(size)
	return atlasLocation{oldest, r}, true
}

func (p *atlasPage) place(size image.Point) (image.Rectangle, bool) {
	x, y, h := p.x, p.y, p.rowHeight
	if x+size.X > glyphAtlasSide {
		x = 0
		y += h
		h = 0
	}
	if y+size.Y > glyphAtlasSide {
		return image.Rectangle{}, false
	}
	r := image.Rectangle{Min: image.Pt(x, y), Max: image.Pt(x+size.X, y+size.Y)}
	p.x = x + size.X
	p.y = y
	p.rowHeight = max(h, size.Y)
	return r, true
}

func (p *atlasPage) makeWritable() {
	if p.dirty {
		return
	}
	pixels := image.NewRGBA(p.pixels.Rect)
	copy(pixels.Pix, p.pixels.Pix)
	p.pixels = pixels
	p.dirty = true
}

// Image textures store sRGB-encoded premultiplied linear colors. Tinting the
// coverage in linear space preserves the vector renderer's antialiasing.
func atlasColorTable(col color.NRGBA) (lut [256][3]byte) {
	for j, v := range []byte{col.R, col.G, col.B} {
		x := float64(v) / 255
		if x <= 0.04045 {
			x /= 12.92
		} else {
			x = math.Pow((x+0.055)/1.055, 2.4)
		}
		for i := range lut {
			y := x * float64(i) / 255
			if y <= 0.0031308 {
				y *= 12.92
			} else {
				y = 1.055*math.Pow(y, 1/2.4) - 0.055
			}
			lut[i][j] = byte(math.Round(y * 255))
		}
	}
	return
}

// GlyphAtlasStats reports retained CPU cache storage and cumulative draws.
// PageBytes excludes GPU textures and older snapshots referenced by Gio.
type GlyphAtlasStats struct {
	Masks, MaskBytes, Pages, PageBytes int
	RasterDraws, VectorDraws           uint64
}

func (a *GlyphAtlas) Stats() GlyphAtlasStats {
	return GlyphAtlasStats{Masks: len(a.masks), MaskBytes: a.maskBytes, Pages: len(a.pages), PageBytes: len(a.pages) * glyphAtlasSide * glyphAtlasSide * 4, RasterDraws: a.rasterDraws, VectorDraws: a.vectorDraws}
}
