package el

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/theme"
	"image"
)

type layerArrow struct {
	points [3]f32.Point
	bounds image.Rectangle
}

func newLayerArrow(panel, anchor image.Rectangle, side Side, align Align, depth, radius int) *layerArrow {
	horizontal := side == Top || side == Bottom
	lo, hi, a, b := panel.Min.Y, panel.Max.Y, anchor.Min.Y, anchor.Max.Y
	if horizontal {
		lo, hi, a, b = panel.Min.X, panel.Max.X, anchor.Min.X, anchor.Max.X
	}
	depth = min(depth, max(0, (hi-lo)/2))
	inset := min(max(depth, radius+depth), (hi-lo)/2)
	center := (a + b) / 2
	if align == Start {
		center = a + inset
	} else if align == End {
		center = b - inset
	}
	center = max(lo+inset, min(center, hi-inset))
	var points [3]image.Point
	switch side {
	case Top:
		points = [3]image.Point{image.Pt(center-depth, panel.Max.Y), image.Pt(center, panel.Max.Y+depth), image.Pt(center+depth, panel.Max.Y)}
	case Left:
		points = [3]image.Point{image.Pt(panel.Max.X, center-depth), image.Pt(panel.Max.X+depth, center), image.Pt(panel.Max.X, center+depth)}
	case Right:
		points = [3]image.Point{image.Pt(panel.Min.X, center-depth), image.Pt(panel.Min.X-depth, center), image.Pt(panel.Min.X, center+depth)}
	default:
		points = [3]image.Point{image.Pt(center-depth, panel.Min.Y), image.Pt(center, panel.Min.Y-depth), image.Pt(center+depth, panel.Min.Y)}
	}
	arrow := &layerArrow{}
	for i, p := range points {
		arrow.points[i] = f32.Pt(float32(p.X), float32(p.Y))
		arrow.bounds = arrow.bounds.Union(image.Rectangle{Min: p, Max: p.Add(image.Pt(1, 1))})
	}
	return arrow
}

func (a *layerArrow) paint(ops *op.Ops, style Style) {
	c := theme.Surface
	if style.bg != nil {
		c = *style.bg
	}
	var p clip.Path
	p.Begin(ops)
	p.MoveTo(a.points[0])
	p.LineTo(a.points[1])
	p.LineTo(a.points[2])
	p.Close()
	paint.FillShape(ops, c, clip.Outline{Path: p.End()}.Op())
}
