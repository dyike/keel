package theme

import (
	"image"
	"image/color"
	"math"

	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/text"
	"golang.org/x/image/math/fixed"
)

const (
	glyphAtlasSide  = 512
	glyphAtlasPages = 8
	glyphMaskSide   = 64
	glyphMaskLimit  = 2 << 20
	glyphMaskBatch  = 64
	glyphMaskSheet  = 8 * glyphMaskSide
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
	readback                 *image.RGBA
	pending                  []pendingMask
	maskBatches              uint64
	failed                   bool
	committed                bool
	frame                    uint64
	masks                    map[atlasMaskKey]*atlasMask
	maskBytes                int
	lastMaskFrame            uint64
	newMasks                 int
	rasterDraws, vectorDraws uint64
	pages                    []*atlasPage
}

// An atlas fragment always contains one anchor and one glyph. Shaper changes
// clear the atlas; retaining the general nine-glyph vector key is unnecessary.
type atlasMaskKey struct {
	anchor, glyph text.GlyphID
	phase         fixed.Int26_6
}

func maskKey(gs []text.Glyph) atlasMaskKey {
	return atlasMaskKey{gs[0].ID, gs[1].ID, gs[1].X - gs[0].X}
}

type pendingMask struct {
	mask *atlasMask
	path clip.PathSpec
}

type atlasMask struct {
	ready       bool
	wanted      [8]color.NRGBA
	wantedCount int
	bitmap      op.CallOp
	bounds      image.Rectangle
	alpha       []byte
	colors      map[color.NRGBA]atlasLocation
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
		key := maskKey(fragment)
		m, known := a.masks[key]
		if !known {
			if len(a.masks) >= 4096 || a.frame > 1 && a.newMasks >= 32 {
				return
			}
			_, f := a.vector.cachedFragment(a.shaper, fragment)
			m = a.makeMask(f.path, fragment)
			if m != nil {
				m.bitmap = f.bitmap
			}
			if a.masks == nil {
				a.masks = make(map[atlasMaskKey]*atlasMask)
			}
			a.masks[key] = m
		}
		if m == nil || len(m.alpha) == 0 {
			return
		}
		a.prepareColor(m, col)
	})
}

func (a *GlyphAtlas) prepareColor(m *atlasMask, col color.NRGBA) {
	if !m.ready {
		for _, c := range m.wanted[:m.wantedCount] {
			if c == col {
				return
			}
		}
		if m.wantedCount < len(m.wanted) {
			m.wanted[m.wantedCount] = col
			m.wantedCount++
		}
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
}

// Commit freezes changed image pages. Prepare must not be called again until
// the next BeginFrame. Existing ImageOps always keep their pixels immutable.
func (a *GlyphAtlas) Commit() {
	a.flushMasks()
	for _, p := range a.pages {
		if p.dirty {
			p.image = paint.NewImageOp(p.pixels)
			p.image.Filter = paint.FilterNearest
			p.dirty = false
		}
	}
	// The scratch renderer can retain large coverage textures even for a tiny
	// viewport. Release the initial population immediately: a static view may
	// never draw enough frames for the later inactivity threshold.
	if a.window != nil && (a.frame == 1 || a.frame > a.lastMaskFrame+60) {
		a.ReleaseScratch()
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
		key := maskKey(fragment)
		m := a.masks[key]
		loc, ok := atlasLocation{}, false
		if m != nil {
			loc, ok = m.colors[col]
		}
		if ok {
			a.rasterDraws++
			page := a.pages[loc.page]
			page.used = a.frame
			offset := m.bounds.Min.Sub(loc.rect.Min).Add(image.Pt(displacement.Round(), 0))
			imgTr := op.Offset(offset).Push(ops)
			cl := clip.Rect(loc.rect).Push(ops)
			page.image.Add(ops)
			paint.PaintOp{}.Add(ops)
			cl.Pop()
			imgTr.Pop()
			if m.bitmap != (op.CallOp{}) {
				tr := op.Offset(image.Pt(displacement.Round(), 0)).Push(ops)
				m.bitmap.Add(ops)
				tr.Pop()
			}
		} else if m == nil || len(m.alpha) != 0 {
			a.vectorDraws++
			_, f := a.vector.cachedFragment(a.shaper, fragment)
			tr := op.Offset(image.Pt(displacement.Round(), 0)).Push(ops)
			paint.ColorOp{Color: col}.Add(ops)
			cl := clip.Outline{Path: f.path}.Op().Push(ops)
			paint.PaintOp{}.Add(ops)
			cl.Pop()
			f.bitmap.Add(ops)
			tr.Pop()
		} else if m.bitmap != (op.CallOp{}) {
			tr := op.Offset(image.Pt(displacement.Round(), 0)).Push(ops)
			m.bitmap.Add(ops)
			tr.Pop()
		}
	}) {
		a.vectorDraws += uint64(len(gs))
		paintGlyphRun(ops, a.shaper, gs, col)
	}
}

// ReleaseScratch frees temporary rasterization resources while preserving
// cached masks and image pages. Call serially after Commit, for example when
// the UI becomes idle. Future new glyphs recreate the renderer as needed.
func (a *GlyphAtlas) ReleaseScratch() {
	if len(a.pending) != 0 {
		return
	} // queued paths still need rasterization
	if a.window != nil {
		a.window.Release()
		a.window = nil
	}
	a.scratch = op.Ops{}
	a.readback = nil
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
	if len(a.pending) == glyphMaskBatch {
		a.flushMasks()
	}
	if a.failed {
		return nil
	}
	a.newMasks++
	a.lastMaskFrame = a.frame
	m := &atlasMask{bounds: bounds, alpha: make([]byte, bounds.Dx()*bounds.Dy()), colors: make(map[color.NRGBA]atlasLocation)}
	a.maskBytes += len(m.alpha)
	a.pending = append(a.pending, pendingMask{mask: m, path: path})
	return m
}

// Rasterize up to 64 isolated tiles with one GPU submission and readback.
// Mask storage is reserved before enqueueing, so pending work respects the
// same memory and per-frame glyph budgets as completed masks.
func (a *GlyphAtlas) flushMasks() {
	if len(a.pending) == 0 {
		return
	}
	defer func() { clear(a.pending); a.pending = a.pending[:0] }()
	if a.failed {
		return
	}
	if a.window == nil {
		w, err := headless.NewWindow(glyphMaskSheet, glyphMaskSheet)
		if err != nil {
			a.failed = true
			return
		}
		a.window = w
	}
	a.scratch.Reset()
	for i, p := range a.pending {
		tile := image.Pt(i%8*glyphMaskSide, i/8*glyphMaskSide)
		area := clip.Rect(image.Rectangle{Min: tile, Max: tile.Add(image.Pt(glyphMaskSide, glyphMaskSide))}).Push(&a.scratch)
		tr := op.Offset(tile.Sub(p.mask.bounds.Min)).Push(&a.scratch)
		paint.ColorOp{Color: color.NRGBA{R: 255, G: 255, B: 255, A: 255}}.Add(&a.scratch)
		cl := clip.Outline{Path: p.path}.Op().Push(&a.scratch)
		paint.PaintOp{}.Add(&a.scratch)
		cl.Pop()
		tr.Pop()
		area.Pop()
	}
	if err := a.window.Frame(&a.scratch); err != nil {
		a.failed = true
		return
	}
	if a.readback == nil {
		a.readback = image.NewRGBA(image.Rect(0, 0, glyphMaskSheet, glyphMaskSheet))
	}
	if err := a.window.Screenshot(a.readback); err != nil {
		a.failed = true
		return
	}
	a.maskBatches++
	for i, p := range a.pending {
		m := p.mask
		tile := image.Pt(i%8*glyphMaskSide, i/8*glyphMaskSide)
		hasInk := false
		for y := 0; y < m.bounds.Dy(); y++ {
			for x := 0; x < m.bounds.Dx(); x++ {
				alpha := a.readback.Pix[a.readback.PixOffset(tile.X+x, tile.Y+y)+3]
				m.alpha[y*m.bounds.Dx()+x] = alpha
				hasInk = hasInk || alpha != 0
			}
		}
		// Gio's bitmap macro is empty for an outline glyph (the other glyph is
		// our empty anchor). Keep actual bitmap glyphs on their original path.
		if hasInk {
			m.bitmap = op.CallOp{}
			a.maskBytes -= m.trimTransparentBorder()
		}
		m.ready = true
		for _, col := range m.wanted[:m.wantedCount] {
			a.prepareColor(m, col)
		}
		m.wantedCount = 0
	}
}

// trimTransparentBorder removes only zero-coverage pixels after rasterization.
// The rasterizer still receives its safety border; cached color variants and
// textures need not retain it. Return the number of released alpha bytes.
func (m *atlasMask) trimTransparentBorder() int {
	width, height := m.bounds.Dx(), m.bounds.Dy()
	left, top, right, bottom := width, height, 0, 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if m.alpha[y*width+x] != 0 {
				left = min(left, x)
				top = min(top, y)
				right = max(right, x+1)
				bottom = max(bottom, y+1)
			}
		}
	}
	// Empty masks may represent bitmap glyphs; leave their metadata untouched.
	if right <= left || bottom <= top || left == 0 && top == 0 && right == width && bottom == height {
		return 0
	}
	cropped := make([]byte, (right-left)*(bottom-top))
	for y := top; y < bottom; y++ {
		copy(cropped[(y-top)*(right-left):], m.alpha[y*width+left:y*width+right])
	}
	released := len(m.alpha) - len(cropped)
	m.alpha = cropped
	m.bounds = image.Rect(m.bounds.Min.X+left, m.bounds.Min.Y+top, m.bounds.Min.X+right, m.bounds.Min.Y+bottom)
	return released
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
	// MaskBatches counts successful offscreen submissions and readbacks.
	MaskBatches uint64
}

func (a *GlyphAtlas) Stats() GlyphAtlasStats {
	return GlyphAtlasStats{Masks: len(a.masks), MaskBytes: a.maskBytes, Pages: len(a.pages), PageBytes: len(a.pages) * glyphAtlasSide * glyphAtlasSide * 4, RasterDraws: a.rasterDraws, VectorDraws: a.vectorDraws, MaskBatches: a.maskBatches}
}
