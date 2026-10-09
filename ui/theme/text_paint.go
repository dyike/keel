package theme

import (
	"container/list"
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/text"
	"golang.org/x/image/math/fixed"
	"image/color"
)

// GlyphPainter paints shaped single-line text using reusable vector fragments.
// Its zero value is ready to use. Like text.Shaper, it must be used serially.
// Keep it across frames; caches retain at most 128 styles and 4096 fragments.
// Do not copy a GlyphPainter after its first use.
// Complex or overlapping glyphs and translucent colors use a whole-run path.
type GlyphPainter struct {
	// FragmentSize is the maximum number of glyphs per fragment. Zero means two.
	// Values are limited to 1–8; smaller fragments reuse more paths but draw more.
	FragmentSize int
	// SubpixelPhases rounds each glyph's horizontal position inside a
	// fragment to 1/SubpixelPhases of a pixel. Zero keeps exact positions.
	// Text whose advances are not whole pixels, such as a monospace grid at
	// fractional scale, otherwise yields up to 64 phases per glyph: more
	// fragments than the cache holds, so paths and their GPU buffers are
	// rebuilt every frame. Four phases are visually indistinguishable at
	// text sizes. Values are limited to 1–64.
	SubpixelPhases int
	shaper         *text.Shaper
	blanks         map[glyphAnchorKey]text.Glyph
	fragments      map[glyphFragmentKey]*list.Element
	recent         list.List
	fragment       [9]text.Glyph
}

type glyphFragmentKey struct {
	shaper *text.Shaper
	count  int
	ids    [9]text.GlyphID
	x      [9]fixed.Int26_6
}
type glyphFragment struct {
	key    glyphFragmentKey
	path   clip.PathSpec
	bitmap op.CallOp
}

type glyphAnchorKey struct {
	shaper *text.Shaper
	params text.Parameters
}

// Paint draws glyphs at the same origin as Shaper.Shape and Shaper.Bitmaps:
// X is relative to the first glyph and Y is the caller's baseline, not Glyph.Y.
// params must describe the font and size used to shape glyphs. The caller owns
// clipping and semantic operations. Paint may replace the shaper's iterator.
func (p *GlyphPainter) Paint(ops *op.Ops, sh *text.Shaper, params text.Parameters, gs []text.Glyph, col color.NRGBA) {
	if len(gs) == 0 {
		return
	}
	if !p.walkFragments(sh, params, gs, col, func(fragment []text.Glyph, displacement fixed.Int26_6) {
		tr := op.Affine(f32.AffineId().Offset(f32.Pt(float32(displacement)/64, 0))).Push(ops)
		p.paintFragment(ops, sh, fragment, col)
		tr.Pop()
	}) {
		paintGlyphRun(ops, sh, gs, col)
	}
}

func (p *GlyphPainter) walkFragments(sh *text.Shaper, params text.Parameters, gs []text.Glyph, col color.NRGBA, visit func([]text.Glyph, fixed.Int26_6)) bool {
	if len(gs) == 0 {
		return true
	}
	// Font loading replaces the theme shaper. Release old font/cache references.
	if p.shaper != sh {
		p.shaper = sh
		clear(p.blanks)
		clear(p.fragments)
		p.recent.Init()
	}
	eligible := col.A == 255
	for i, g := range gs {
		if g.Offset != (fixed.Point26_6{}) || g.Flags&(text.FlagTowardOrigin|text.FlagParagraphBreak) != 0 || g.Y != gs[0].Y {
			eligible = false
			break
		}
		if i > 0 && gs[i-1].X+gs[i-1].Bounds.Max.X > g.X+g.Bounds.Min.X {
			eligible = false
			break
		}
	}
	if !eligible {
		return false
	}
	// An empty space anchors each path at an integer displacement from the
	// run's origin. Fractional glyph positions stay inside Shape's cache key,
	// avoiding the extra texture filtering caused by fractional translations.
	params.Alignment, params.MinWidth, params.MaxWidth = text.Start, 0, 1<<24
	params.MaxLines, params.Truncator = 0, ""
	key := glyphAnchorKey{sh, params}
	blank, ok := p.blanks[key]
	if !ok {
		sh.LayoutString(params, " ")
		blank, ok = sh.NextGlyph()
		if !ok || blank.Bounds != (fixed.Rectangle26_6{}) {
			return false
		}
		if len(p.blanks) >= 128 {
			clear(p.blanks)
		}
		if p.blanks == nil {
			p.blanks = make(map[glyphAnchorKey]text.Glyph)
		}
		p.blanks[key] = blank
	}
	size := p.FragmentSize
	if size == 0 {
		size = 2
	}
	size = max(1, min(size, 8))
	var step fixed.Int26_6
	if p.SubpixelPhases > 0 {
		step = fixed.Int26_6(64 / max(1, min(p.SubpixelPhases, 64)))
	}
	// The visitor is synchronous and must not retain this scratch slice. Keep
	// it on the serial painter instead of allocating once for every text run.
	fragment := p.fragment[:]
	for start := 0; start < len(gs); start += size {
		part := gs[start:min(start+size, len(gs))]
		displacement := fixed.I((part[0].X - gs[0].X).Floor())
		blank.X = gs[0].X + displacement
		fragment[0] = blank
		copy(fragment[1:], part)
		if step > 0 {
			// Offsets from the anchor are non-negative: glyphs are in
			// visual order and the anchor sits at the first one's pixel.
			for i := 1; i <= len(part); i++ {
				off := fragment[i].X - blank.X
				fragment[i].X = blank.X + (off+step/2)/step*step
			}
		}
		visit(fragment[:1+len(part)], displacement)
	}
	return true
}

// Keep fragment identities stable beyond Shaper's 1000-entry whole-run cache.
// A monospace grid can exceed that count with the same glyphs at different
// subpixel phases. Bounding this LRU also bounds retained glyph path storage.
func (p *GlyphPainter) cachedFragment(sh *text.Shaper, gs []text.Glyph) (glyphFragmentKey, glyphFragment) {
	key := glyphFragmentKey{shaper: sh, count: len(gs)}
	for i, g := range gs {
		key.ids[i] = g.ID
		key.x[i] = g.X - gs[0].X
	}
	var f glyphFragment
	if elem, ok := p.fragments[key]; ok {
		p.recent.MoveToFront(elem)
		f = elem.Value.(glyphFragment)
	} else {
		f = glyphFragment{key: key, path: sh.Shape(gs), bitmap: sh.Bitmaps(gs)}
		if p.fragments == nil {
			p.fragments = make(map[glyphFragmentKey]*list.Element)
		}
		if len(p.fragments) >= 4096 {
			last := p.recent.Back()
			delete(p.fragments, last.Value.(glyphFragment).key)
			p.recent.Remove(last)
		}
		p.fragments[key] = p.recent.PushFront(f)
	}
	return key, f
}

func (p *GlyphPainter) paintFragment(ops *op.Ops, sh *text.Shaper, gs []text.Glyph, col color.NRGBA) {
	_, f := p.cachedFragment(sh, gs)
	paint.ColorOp{Color: col}.Add(ops)
	outline := clip.Outline{Path: f.path}.Op().Push(ops)
	paint.PaintOp{}.Add(ops)
	outline.Pop()
	if f.bitmap != (op.CallOp{}) {
		f.bitmap.Add(ops)
	}
}

func paintGlyphRun(ops *op.Ops, sh *text.Shaper, gs []text.Glyph, col color.NRGBA) {
	paint.ColorOp{Color: col}.Add(ops)
	outline := clip.Outline{Path: sh.Shape(gs)}.Op().Push(ops)
	paint.PaintOp{}.Add(ops)
	outline.Pop()
	if bitmap := sh.Bitmaps(gs); bitmap != (op.CallOp{}) {
		bitmap.Add(ops)
	}
}
