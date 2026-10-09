package el

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"image"
	"math"
)

// dashedBorderPath follows a rounded perimeter with 4dp dashes and 3dp gaps.
// Curves are sampled at most one physical pixel apart so the pattern remains
// continuous around corners and scales with the display metric.
func dashedBorderPath(ops *op.Ops, rect image.Rectangle, radius, dash, gap float32) clip.PathSpec {
	var path clip.Path
	path.Begin(ops)
	w, h := float32(rect.Dx()), float32(rect.Dy())
	if w <= 0 || h <= 0 {
		return path.End()
	}
	r := min(max(radius, 0), min(w, h)/2)
	horizontal, vertical := w-2*r, h-2*r
	arc := float32(math.Pi/2) * r
	lengths := [8]float32{horizontal, arc, vertical, arc, horizontal, arc, vertical, arc}
	total := 2*(horizontal+vertical) + 4*arc
	point := func(distance float32) f32.Point {
		segment := 0
		for segment < 7 && distance > lengths[segment] {
			distance -= lengths[segment]
			segment++
		}
		var x, y float32
		switch segment {
		case 0:
			x, y = r+distance, 0
		case 2:
			x, y = w, r+distance
		case 4:
			x, y = w-r-distance, h
		case 6:
			x, y = 0, h-r-distance
		default:
			centers := [4]f32.Point{{X: w - r, Y: r}, {X: w - r, Y: h - r}, {X: r, Y: h - r}, {X: r, Y: r}}
			corner := segment / 2
			angle := float64(corner-1) * math.Pi / 2
			if r > 0 {
				angle += float64(distance / r)
			}
			x = centers[corner].X + r*float32(math.Cos(angle))
			y = centers[corner].Y + r*float32(math.Sin(angle))
		}
		return f32.Pt(float32(rect.Min.X)+x, float32(rect.Min.Y)+y)
	}
	dash, gap = max(dash, 1), max(gap, 1)
	for start := float32(0); start < total; start += dash + gap {
		end := min(start+dash, total)
		path.MoveTo(point(start))
		for d := start + 1; d < end; d++ {
			path.LineTo(point(d))
		}
		path.LineTo(point(end))
	}
	return path.End()
}
