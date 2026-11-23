// Each composition is a function of its frame number. The player and the
// offscreen encoder call the same drawing code at different pixel sizes.
package main

import (
	"cmp"
	"image"
	"image/color"
	"math"
	"slices"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/theme"
)

const (
	fps         = 30
	videoWidth  = 1280
	videoHeight = 720
)

var templates = [...]struct {
	name, description string
	defaultPalette    int
	seconds           int
	paint             func(*op.Ops, palette, float64)
}{
	{"Orbit", "A sculptural loop in three dimensions.", 0, 4, orbit},
	{"Signal", "Interference waves in a flowing ribbon.", 1, 4, signal},
	{"Bloom", "A radial study in layered petals.", 2, 4, bloom},
	{filmName, "A short film about building interfaces in Go.", 0, filmSeconds, keelFilm},
}

type palette struct {
	name                           string
	paper, grid, light, base, dark color.NRGBA
}

var palettes = [...]palette{
	{"Cobalt", theme.RGB(0xe8eeff), theme.RGB(0xcdd9f6), theme.RGB(0xa4dbff), theme.RGB(0x2868ff), theme.RGB(0x122084)},
	{"Tangerine", theme.RGB(0xffece1), theme.RGB(0xf3d7c5), theme.RGB(0xffdd92), theme.RGB(0xff813c), theme.RGB(0xbc3928)},
	{"Iris", theme.RGB(0xf2eafa), theme.RGB(0xe2d3f1), theme.RGB(0xf0caff), theme.RGB(0xae70ef), theme.RGB(0x5e2992)},
}

type composition struct{ Scene, Palette int }

func (c composition) frames() int { return templates[c.Scene].seconds * fps }

func (c composition) valid() bool {
	return c.Scene >= 0 && c.Scene < len(templates) && c.Palette >= 0 && c.Palette < len(palettes)
}

func (c composition) draw(ops *op.Ops, size image.Point, frame int) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	defer clip.Rect(image.Rectangle{Max: size}).Push(ops).Pop()
	p := palettes[c.Palette]
	scale := float32(size.Y) / 360
	shift := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	defer op.Affine(f32.Affine2D{}.Scale(f32.Point{}, f32.Pt(scale, scale)).Offset(shift)).Push(ops).Pop()
	if templates[c.Scene].name != filmName {
		paint.Fill(ops, p.paper)
		halfW := int(math.Ceil(float64(float32(size.X) / scale / 2)))
		var grid clip.Path
		grid.Begin(ops)
		for x := -halfW / 24 * 24; x <= halfW; x += 24 {
			grid.MoveTo(f32.Pt(float32(x), -181))
			grid.LineTo(f32.Pt(float32(x), 181))
		}
		for y := -168; y <= 180; y += 24 {
			grid.MoveTo(f32.Pt(float32(-halfW), float32(y)))
			grid.LineTo(f32.Pt(float32(halfW), float32(y)))
		}
		paint.FillShape(ops, p.grid, clip.Stroke{Path: grid.End(), Width: 0.65}.Op())
	}
	phase := 2 * math.Pi * float64(frame) / float64(c.frames())
	templates[c.Scene].paint(ops, p, phase)
}

type vec3 struct{ x, y, z float64 }

func (v vec3) project() (f32.Point, float64) {
	const tilt = 0.85
	y, z := v.y*math.Cos(tilt)-v.z*math.Sin(tilt), v.y*math.Sin(tilt)+v.z*math.Cos(tilt)
	const roll = -0.35
	return f32.Pt(float32(v.x*math.Cos(roll)-y*math.Sin(roll)), float32(v.x*math.Sin(roll)+y*math.Cos(roll))), z
}

// A stack of elliptical cross-sections follows a closed, lifted orbit.
// Draw the far slices first so the front of the ring occludes the back.
func orbit(ops *op.Ops, p palette, phase float64) {
	type slice struct {
		center, radial, vertical f32.Point
		depth                    float64
	}
	var discs [84]slice
	for i := range discs {
		a := 2*math.Pi*float64(i)/float64(len(discs)) + 0.12*math.Sin(phase)
		sn, cs := math.Sincos(a)
		r := 101 + 10*math.Sin(2*a+phase)
		center, z := (vec3{r * cs, r * sn, 23 * math.Sin(2*a+phase)}).project()
		thick := 37 + 8*math.Cos(a-phase)
		radial, _ := (vec3{thick * cs, thick * sn, 0}).project()
		vertical, _ := (vec3{0, 0, thick}).project()
		discs[i] = slice{center, radial, vertical, z}
	}
	slices.SortFunc(discs[:], func(a, b slice) int { return cmp.Compare(a.depth, b.depth) })
	for _, s := range discs {
		// Cubic arcs keep the contour smooth at export resolution.
		var path clip.Path
		path.Begin(ops)
		path.MoveTo(s.center.Add(s.radial))
		u, v := s.radial, s.vertical
		for range 4 {
			path.CubeTo(s.center.Add(u).Add(v.Mul(.552285)), s.center.Add(v).Add(u.Mul(.552285)), s.center.Add(v))
			u, v = v, u.Mul(-1)
		}
		path.Close()
		spec := path.End()
		area := clip.Outline{Path: spec}.Op().Push(ops)
		paint.LinearGradientOp{Stop1: s.center.Add(f32.Pt(-13, -23)), Color1: p.light, Stop2: s.center.Add(f32.Pt(13, 25)), Color2: p.base}.Add(ops)
		paint.PaintOp{}.Add(ops)
		area.Pop()
		paint.FillShape(ops, p.dark, clip.Stroke{Path: spec, Width: 1.45}.Op())
	}
}

func signal(ops *op.Ops, p palette, phase float64) {
	for i := range 43 {
		t := float64(i) / 42
		var path clip.Path
		path.Begin(ops)
		for j := range 73 {
			u := float64(j) / 72
			x := (u - .5) * 394
			y := 72*math.Sin(u*2*math.Pi+phase+t*2.5) + 43*(t-.5) + 20*math.Cos(7*u-phase)
			pt := f32.Pt(float32(x), float32(y))
			if j == 0 {
				path.MoveTo(pt)
			} else {
				path.LineTo(pt)
			}
		}
		ink := p.base
		if i%5 == 0 {
			ink = p.dark
		}
		paint.FillShape(ops, ink, clip.Stroke{Path: path.End(), Width: 2}.Op())
	}
}

func bloom(ops *op.Ops, p palette, phase float64) {
	for layer := range 3 {
		for i := range 18 {
			a := float32(float64(i)*math.Pi/9 + float64(layer)*.18 + phase/18 + float64(layer)*.17*math.Sin(phase))
			r := float32(139 - 32*layer)
			trans := op.Affine(f32.Affine2D{}.Rotate(f32.Point{}, a)).Push(ops)
			var petal clip.Path
			petal.Begin(ops)
			petal.MoveTo(f32.Point{})
			petal.CubeTo(f32.Pt(r*.29, -r*.48), f32.Pt(r, -r*.32), f32.Pt(r, 0))
			petal.CubeTo(f32.Pt(r, r*.25), f32.Pt(r*.42, r*.22), f32.Point{})
			petal.Close()
			spec := petal.End()
			area := clip.Outline{Path: spec}.Op().Push(ops)
			paint.LinearGradientOp{Stop1: f32.Point{}, Color1: p.dark, Stop2: f32.Pt(r, 0), Color2: p.light}.Add(ops)
			paint.PaintOp{}.Add(ops)
			area.Pop()
			paint.FillShape(ops, p.base, clip.Stroke{Path: spec, Width: 1.5}.Op())
			trans.Pop()
		}
	}
}
