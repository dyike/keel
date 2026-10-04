package el

import (
	"image"
	"math"
	"slices"
	"unicode/utf8"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/editorstyle"
	"github.com/dyike/keel/ui/internal/inputcontent"
	"github.com/dyike/keel/ui/theme"
)

// InputTokenRenderer lays out passive reference content in physical pixels.
// It is measured with a disabled context, then painted at the resulting inline
// position. Honor Constraints and return a baseline (distance from the bottom).
// Do not mutate application state or install input handlers; use OnTokenActivate.
// Nil restores the default label pill. Content is clipped to the input viewport.
type InputTokenRenderer func(core.C, InputToken) core.D

func (e *InputEl) TokenRenderer(fn InputTokenRenderer) *InputEl {
	e.n.input.tokenRenderer = fn
	return e
}

type inputObjectLayout struct {
	shaper   *text.Shaper
	glyphs   []editorstyle.InlineGlyph
	metrics  editorstyle.InlineFontMetrics
	revision uint64
	sizes    []core.D
	font     font.Font
}

func (d *InputDocument) presentation() inputcontent.Presentation {
	c := d.Content()
	if !d.objects {
		return c.Presentation()
	}
	spans := c.Tokens()
	// Reserve every source and label rune; private-use text remains ordinary text.
	occupied := map[rune]bool{}
	for _, r := range c.Text() {
		occupied[r] = true
	}
	for _, s := range spans {
		for _, r := range s.Token.Display() {
			occupied[r] = true
		}
	}
	labels := make([]string, len(spans))
	next := rune(0xf0000)
	for i := range labels {
		for occupied[next] || next == 0xffffe || next == 0xfffff {
			next++
		}
		if next > 0x10fffd {
			return c.Presentation()
		}
		labels[i] = string(next)
		next++
	}
	p, _ := inputcontent.LayoutPresentation(c, labels)
	return p
}

func (e *engine) tokenWidget(n *Node, g core.C, t InputToken) core.D {
	if fn := n.input.tokenRenderer; fn != nil {
		return fn(g, t)
	}
	rec := op.Record(g.Ops)
	dims := layout.Inset{Left: 4, Right: 4}.Layout(g, func(g layout.Context) layout.Dimensions {
		label := e.label(n, t.Display())
		label.MaxLines = 1
		return label.Layout(g)
	})
	call := rec.Stop()
	r := g.Dp(3)
	paint.FillShape(g.Ops, theme.Subtle, clip.RRect{Rect: image.Rectangle{Max: dims.Size}, NE: r, NW: r, SE: r, SW: r}.Op(g.Ops))
	call.Add(g.Ops)
	return dims
}

func (e *engine) prepareInputObjects(n *Node, st *elemState, maxW int) {
	d := n.input.document
	d.objects = true
	p := d.presentation()
	objects := &st.inputObjects
	objects.sizes = objects.sizes[:0]
	if len(p.Spans) == 0 {
		objects.shaper = nil
		objects.glyphs = nil
		return
	}
	pixels := max(1, e.gtx.Sp(n.textStyle.size))
	width := max(1, min(maxW, inf))
	maxAdvance, ascent, descent := 1, 1, 0
	for _, span := range p.Spans {
		dims := e.tokenWidget(n, e.measureGtx(layout.Constraints{Max: image.Pt(width, 4096)}), span.Token)
		dims.Size.X = min(width, max(1, dims.Size.X))
		dims.Size.Y = min(4096, max(1, dims.Size.Y))
		if dims.Baseline <= 0 || dims.Baseline > dims.Size.Y {
			dims.Baseline = dims.Size.Y / 5
		}
		objects.sizes = append(objects.sizes, dims)
		maxAdvance = max(maxAdvance, dims.Size.X)
		ascent = max(ascent, dims.Size.Y-dims.Baseline)
		descent = max(descent, dims.Baseline)
	}
	upem := 1024
	for upem > 16 && max(maxAdvance, ascent, descent)*upem/pixels > 32760 {
		upem /= 2
	}
	units := func(px int) uint16 {
		return uint16(min(32767, max(1, int(math.Round(float64(px*upem)/float64(pixels))))))
	}
	metrics := editorstyle.InlineFontMetrics{UnitsPerEm: uint16(upem), Ascent: int16(units(ascent)), Descent: int16(units(descent))}
	glyphs := make([]editorstyle.InlineGlyph, len(p.Spans))
	for i, span := range p.Spans {
		r, _ := utf8.DecodeRuneInString(p.Text[span.Display.Start:span.Display.End])
		glyphs[i] = editorstyle.InlineGlyph{Rune: r, Advance: units(objects.sizes[i].Size.X)}
	}
	f := textFont(n.textStyle)
	f.Typeface = font.Typeface("Keel Input Objects, " + string(f.Typeface))
	if objects.shaper != nil && slices.Equal(glyphs, objects.glyphs) && metrics == objects.metrics && theme.Revision() == objects.revision && f == objects.font {
		return
	}
	face, err := editorstyle.InlineFont("Keel Input Objects", metrics, glyphs)
	if err != nil {
		d.objects = false
		objects.shaper = nil
		return
	}
	objects.shaper = theme.NewShaper(face)
	objects.glyphs, objects.metrics, objects.revision, objects.font = glyphs, metrics, theme.Revision(), f
}

// Map an editor-layout selection into source bytes before touching the session.
func documentSelectLayout(d *InputDocument, r InputRange) error {
	source, err := d.presentation().SourceRange(r)
	if err != nil {
		return err
	}
	return d.session.SelectSource(source)
}
func documentCompositionLayout(d *InputDocument, start, end int) (int, int) {
	if start < 0 || end < 0 {
		return -1, -1
	}
	normal := d.Content().Presentation()
	b, err := inputcontent.ByteRange(normal.Text, InputRange{Start: start, End: end})
	if err != nil {
		return -1, -1
	}
	source, err := normal.SourceRange(b)
	if err != nil {
		return -1, -1
	}
	p := d.presentation()
	out, err := inputcontent.RuneRange(p.Text, InputRange{Start: p.DisplayOffset(source.Start, 0), End: p.DisplayOffset(source.End, 0)})
	if err != nil {
		return -1, -1
	}
	return out.Start, out.End
}

// The platform sees readable labels, never internal object runes.
func documentPlatformSelection(d *InputDocument) (string, int, int) {
	p := d.Content().Presentation()
	s := d.Selection()
	r, _ := inputcontent.RuneRange(p.Text, InputRange{Start: p.DisplayOffset(s.Start, 0), End: p.DisplayOffset(s.End, 0)})
	return p.Text, r.Start, r.End
}
