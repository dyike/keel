package main

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
	"golang.org/x/image/vector"
)

// Each platform has its own icon shape. keel.json's icon is full-bleed
// square artwork; iconShape cuts it into the platform's plate, at the
// platform's size within the canvas, with its corner style and shadow.
type iconShape struct {
	name string
	// body is the plate's side as a fraction of the canvas, centered.
	body float64
	// radius is the corner radius as a fraction of the canvas.
	radius float64
	// smoothing > 0 draws continuous ("squircle") corners, as macOS does;
	// 0 draws circular arcs.
	smoothing float64
	// shadow, if set, is drawn under the plate: offset down and blur, as
	// fractions of the canvas, and opacity.
	shadow *iconShadow
}

type iconShadow struct{ dy, blur, alpha float64 }

var (
	// macOS (Big Sur and later): an 824px plate on a 1024px canvas with
	// 185.4px continuous corners and a 28px blur, 12px down, 50% black
	// shadow, from Apple's app icon template.
	macShape = iconShape{name: "macOS", body: 824.0 / 1024, radius: 185.4 / 1024, smoothing: 0.6,
		shadow: &iconShadow{dy: 12.0 / 1024, blur: 28.0 / 1024, alpha: 0.5}}
	// Windows 11: the square keyline of Microsoft's 48px icon grid, 42px,
	// with 2px exterior corners; no shadow, transparent around.
	winShape = iconShape{name: "Windows", body: 42.0 / 48, radius: 2.0 / 48}
	// GNOME: the square keyline of the 128px app icon template, 104px
	// with 8px corners.
	linuxShape = iconShape{name: "Linux", body: 104.0 / 128, radius: 8.0 / 128}
)

// Sizes each platform asks for.
var (
	// Every size Windows 11 requests at 100–400% scale; Microsoft's minimum
	// is 16, 24, 32, 48 and 256.
	windowsIconSizes = []int{16, 20, 24, 30, 32, 36, 40, 48, 60, 64, 72, 80, 96, 256}
	// The hicolor theme's usual app icon sizes.
	linuxIconSizes = []int{16, 24, 32, 48, 64, 128, 256, 512}
)

// shapeIcon draws art (square) as a size×size icon of this shape. With mask
// false the art keeps its own outline and transparency and is only fitted
// to the plate's box.
func (s iconShape) render(art image.Image, size int, mask bool) *image.NRGBA {
	canvas := image.NewNRGBA(image.Rect(0, 0, size, size))
	side := s.body * float64(size)
	off := (float64(size) - side) / 2
	box := image.Rect(int(math.Round(off)), int(math.Round(off)), int(math.Round(off+side)), int(math.Round(off+side)))

	scaled := image.NewNRGBA(image.Rect(0, 0, box.Dx(), box.Dy()))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), art, art.Bounds(), draw.Src, nil)

	// The plate's coverage, anti-aliased, in canvas coordinates.
	plate := image.NewAlpha(canvas.Bounds())
	if mask {
		z := vector.NewRasterizer(size, size)
		roundedRect(z, float32(off), float32(off), float32(side), float32(s.radius*float64(size)), s.smoothing)
		z.Draw(plate, plate.Bounds(), image.Opaque, image.Point{})
	} else {
		draw.Draw(plate, box, alphaOf(scaled), image.Point{}, draw.Src)
	}

	if sh := s.shadow; sh != nil {
		blurred := blurAlpha(plate, sh.blur*float64(size)/2)
		dy := int(math.Round(sh.dy * float64(size)))
		shadow := image.NewUniform(color.NRGBA{A: uint8(255 * sh.alpha)})
		draw.DrawMask(canvas, canvas.Bounds().Add(image.Pt(0, dy)), shadow, image.Point{}, blurred, image.Point{}, draw.Over)
	}
	if mask {
		draw.DrawMask(canvas, box, scaled, image.Point{}, plate, box.Min, draw.Over)
	} else {
		draw.Draw(canvas, box, scaled, image.Point{}, draw.Over)
	}
	return canvas
}

func alphaOf(img *image.NRGBA) *image.Alpha {
	a := image.NewAlpha(img.Bounds())
	for i := 0; i < len(a.Pix); i++ {
		a.Pix[i] = img.Pix[i*4+3]
	}
	return a
}

// roundedRect traces a square plate at (x, y) with side w and corner
// radius r. smoothing 0 gives circular corners; 0.6 approximates Apple's
// continuous corners, with the corner curve (after Figma's corner
// smoothing) easing into the straight edges.
func roundedRect(z *vector.Rasterizer, x, y, w, r float32, smoothing float64) {
	R := float64(r)
	xi := smoothing
	p := math.Min((1+xi)*R, float64(w)/2)
	arcMeasure := 90 * (1 - xi)
	arcLen := math.Sin(rad(arcMeasure/2)) * R * math.Sqrt2
	alpha := (90 - arcMeasure) / 2
	p3p4 := R * math.Tan(rad(alpha/2))
	beta := 45 * xi
	c := p3p4 * math.Cos(rad(beta))
	d := c * math.Tan(rad(beta))
	b := (p - arcLen - c - d) / 3
	a := 2 * b

	// One corner, the top right, relative to its corner point: the edge
	// arrives along +x on the top, leaves along +y on the right.
	type pt struct{ x, y float64 }
	start := pt{-p, 0}
	c1 := [3]pt{{-p + a, 0}, {-p + a + b, 0}, {-p + a + b + c, d}}
	e1 := c1[2]
	e2 := pt{e1.x + arcLen, e1.y + arcLen}
	c2 := [3]pt{{e2.x + d, e2.y + c}, {e2.x + d, e2.y + b + c}, {0, p}}
	// The arc between e1 and e2 lies on a circle of radius R centered on
	// the corner's diagonal, inside the plate.
	center := arcCenter(e1.x, e1.y, e2.x, e2.y, R)
	a1 := math.Atan2(e1.y-center.y, e1.x-center.x)
	a2 := math.Atan2(e2.y-center.y, e2.x-center.x)
	k := 4.0 / 3 * math.Tan((a2-a1)/4)
	arc := [3]pt{
		{e1.x - k*R*math.Sin(a1), e1.y + k*R*math.Cos(a1)},
		{e2.x + k*R*math.Sin(a2), e2.y - k*R*math.Cos(a2)},
		e2,
	}

	cx, cy := float64(x), float64(y)
	W := float64(w)
	corners := []struct {
		ox, oy float64
		m      func(pt) pt
	}{
		{cx + W, cy, func(q pt) pt { return q }},                 // top right
		{cx + W, cy + W, func(q pt) pt { return pt{-q.y, q.x} }}, // bottom right
		{cx, cy + W, func(q pt) pt { return pt{-q.x, -q.y} }},    // bottom left
		{cx, cy, func(q pt) pt { return pt{q.y, -q.x} }},         // top left
	}
	at := func(i int, q pt) (float32, float32) {
		m := corners[i].m(q)
		return float32(corners[i].ox + m.x), float32(corners[i].oy + m.y)
	}
	sx, sy := at(0, start)
	z.MoveTo(sx, sy)
	for i := range corners {
		if i > 0 {
			lx, ly := at(i, start)
			z.LineTo(lx, ly)
		}
		for _, seg := range [][3]pt{c1, arc, c2} {
			ax, ay := at(i, seg[0])
			bx, by := at(i, seg[1])
			ex, ey := at(i, seg[2])
			z.CubeTo(ax, ay, bx, by, ex, ey)
		}
	}
	z.ClosePath()
}

func rad(deg float64) float64 { return deg * math.Pi / 180 }

// arcCenter finds the center, on the corner's diagonal (y = -x, in the
// corner's frame), of a circle of radius r through both points.
func arcCenter(x1, y1, x2, y2, r float64) struct{ x, y float64 } {
	mx, my := (x1+x2)/2, (y1+y2)/2
	dx, dy := x2-x1, y2-y1
	half := math.Hypot(dx, dy) / 2
	h := math.Sqrt(math.Max(r*r-half*half, 0))
	// The perpendicular toward the plate's inside: away from the corner.
	nx, ny := -dy/(2*half), dx/(2*half)
	if (mx+nx)*(mx+nx)+(my+ny)*(my+ny) < mx*mx+my*my {
		nx, ny = -nx, -ny
	}
	return struct{ x, y float64 }{mx + nx*h, my + ny*h}
}

// blurAlpha approximates a Gaussian blur of standard deviation sigma with
// three box blurs.
func blurAlpha(src *image.Alpha, sigma float64) *image.Alpha {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	buf := make([]float64, w*h)
	for i, v := range src.Pix {
		buf[i] = float64(v)
	}
	if sigma >= 0.5 {
		for _, box := range boxesForGauss(sigma, 3) {
			r := (box - 1) / 2
			boxBlur(buf, w, h, r, true)
			boxBlur(buf, w, h, r, false)
		}
	}
	out := image.NewAlpha(src.Bounds())
	for i, v := range buf {
		out.Pix[i] = uint8(math.Max(0, math.Min(255, math.Round(v))))
	}
	return out
}

func boxesForGauss(sigma float64, n int) []int {
	wIdeal := math.Sqrt(12*sigma*sigma/float64(n) + 1)
	wl := int(math.Floor(wIdeal))
	if wl%2 == 0 {
		wl--
	}
	wu := wl + 2
	mIdeal := (12*sigma*sigma - float64(n*wl*wl) - float64(4*n*wl) - float64(3*n)) / float64(-4*wl-4)
	m := int(math.Round(mIdeal))
	sizes := make([]int, n)
	for i := range sizes {
		if i < m {
			sizes[i] = wl
		} else {
			sizes[i] = wu
		}
	}
	return sizes
}

// boxBlur runs a moving average of radius r along rows or columns,
// treating pixels outside as transparent.
func boxBlur(buf []float64, w, h, r int, rows bool) {
	if r <= 0 {
		return
	}
	n, lines := w, h
	idx := func(line, i int) int { return line*w + i }
	if !rows {
		n, lines = h, w
		idx = func(line, i int) int { return i*w + line }
	}
	tmp := make([]float64, n)
	norm := float64(2*r + 1)
	for line := 0; line < lines; line++ {
		sum := 0.0
		for i := -r; i <= r; i++ {
			if i >= 0 && i < n {
				sum += buf[idx(line, i)]
			}
		}
		for i := 0; i < n; i++ {
			tmp[i] = sum / norm
			if out := i - r; out >= 0 {
				sum -= buf[idx(line, out)]
			}
			if in := i + r + 1; in < n {
				sum += buf[idx(line, in)]
			}
		}
		for i := 0; i < n; i++ {
			buf[idx(line, i)] = tmp[i]
		}
	}
}
