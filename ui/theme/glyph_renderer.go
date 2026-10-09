package theme

import (
	"math"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/text"
	"golang.org/x/image/math/fixed"
	"image/color"
)

// GlyphRun describes already-shaped, caller-positioned single-line text.
// Position is its baseline origin in physical pixels. The caller retains the
// glyph slice until Paint; shaping, clipping and semantic text remain its job.
type GlyphRun struct {
	Params   text.Parameters
	Glyphs   []text.Glyph
	Color    color.NRGBA
	Position f32.Point
}

// GlyphRenderer shares bounded glyph caches across changing custom text runs.
// Keep one per view and use serially: BeginFrame, Prepare all visible runs,
// Commit, then Paint. Release it when the view closes. Do not copy after use.
// Integral vertical origins use GlyphAtlas. Horizontal fractions are baked
// into glyph masks so cached pixels are never resampled. Other origins use
// vector drawing. Complex runs retain the atlas/painter
// fallbacks for overlapping ink, combining marks, bidi and translucent colors.
// Extra caller-applied scale or rotation requires GlyphPainter directly.
type GlyphRenderer struct {
	atlas   GlyphAtlas
	vector  GlyphPainter
	shaper  *text.Shaper
	blanks  map[text.Parameters]text.Glyph
	scratch []text.Glyph
}

func (r *GlyphRenderer) BeginFrame(sh *text.Shaper) {
	if r.shaper != sh {
		r.vector = GlyphPainter{}
		clear(r.blanks)
	}
	r.shaper = sh
	r.atlas.SubpixelPhases = 4
	r.atlas.BeginFrame(sh)
}

func (r *GlyphRenderer) Prepare(run GlyphRun) {
	if atlasOrigin(run.Position) {
		r.atlas.Prepare(run.Params, r.positionedGlyphs(run), run.Color)
	}
}

func (r *GlyphRenderer) Commit() { r.atlas.Commit() }

func (r *GlyphRenderer) Paint(ops *op.Ops, run GlyphRun) {
	if r.shaper == nil || len(run.Glyphs) == 0 {
		return
	}
	position := run.Position
	useAtlas := atlasOrigin(position)
	if useAtlas {
		position.X = float32(math.Floor(float64(position.X)))
	}
	tr := op.Affine(f32.AffineId().Offset(position)).Push(ops)
	if useAtlas {
		r.atlas.Paint(ops, run.Params, r.positionedGlyphs(run), run.Color)
	} else {
		// Non-integral baselines retain exact vector antialiasing. Reuse
		// individual glyphs without rounding their positions.
		r.vector.FragmentSize, r.vector.SubpixelPhases = 1, 0
		r.vector.Paint(ops, r.shaper, run.Params, run.Glyphs, run.Color)
	}
	tr.Pop()
}

// ReleaseScratch drops temporary rasterization resources when a view is idle.
// It preserves prepared masks and image pages. Call after Commit, serially.
func (r *GlyphRenderer) ReleaseScratch() { r.atlas.ReleaseScratch() }

func (r *GlyphRenderer) Release() {
	r.atlas.Release()
	*r = GlyphRenderer{}
}

func (r *GlyphRenderer) Stats() GlyphAtlasStats { return r.atlas.Stats() }

func atlasOrigin(p f32.Point) bool {
	return !math.IsNaN(float64(p.X)) && !math.IsNaN(float64(p.Y)) &&
		!math.IsInf(float64(p.X), 0) && !math.IsInf(float64(p.Y), 0) &&
		p.Y == float32(math.Trunc(float64(p.Y)))
}

// A zero-ink anchor moves the fractional origin into glyph spacing. The
// atlas rasterizes that phase once, then paints at an integer translation.
// This also preserves the origin when a complex run takes the vector fallback.
func (r *GlyphRenderer) positionedGlyphs(run GlyphRun) []text.Glyph {
	if len(run.Glyphs) == 0 || run.Position.X == float32(math.Floor(float64(run.Position.X))) {
		return run.Glyphs
	}
	params := run.Params
	params.Alignment, params.MinWidth, params.MaxWidth = text.Start, 0, 1<<24
	params.MaxLines, params.Truncator = 0, ""
	blank, ok := r.blanks[params]
	if !ok {
		r.shaper.LayoutString(params, " ")
		blank, ok = r.shaper.NextGlyph()
		if !ok {
			return run.Glyphs
		}
		if r.blanks == nil {
			r.blanks = make(map[text.Parameters]text.Glyph)
		}
		if len(r.blanks) >= 128 {
			clear(r.blanks)
		}
		r.blanks[params] = blank
	}
	phase := run.Position.X - float32(math.Floor(float64(run.Position.X)))
	blank.X = run.Glyphs[0].X - fixed.Int26_6(math.Round(float64(phase)*64))
	blank.Y, blank.Flags = run.Glyphs[0].Y, 0
	r.scratch = append(r.scratch[:0], blank)
	r.scratch = append(r.scratch, run.Glyphs...)
	return r.scratch
}
