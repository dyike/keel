// Package svgicon rasterizes the small SVG subset app icons use into PNG.
package svgicon

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/image/vector"
)

// Icons have one SVG source each (the Keel icon is docs/images/keel.svg;
// cmd/keel's app placeholder is its own), and Render rasterizes them for the
// PNG sizes browsers and app bundles want, so no PNG is checked in. It
// understands what those files use: a rounded rect filled with a
// linear gradient, and paths of M, L, H, V, C and Z commands, absolute or
// relative, filled with solid colors.

type svgDoc struct {
	ViewBox   string      `xml:"viewBox,attr"`
	Gradients []svgLinear `xml:"defs>linearGradient"`
	Rects     []svgRect   `xml:"rect"`
	Paths     []svgPath   `xml:"path"`
}

// svgLinear is a gradient; the icon's runs along the diagonal, top left to
// bottom right, which is the only direction rendered.
type svgLinear struct {
	ID    string    `xml:"id,attr"`
	Stops []svgStop `xml:"stop"`
}
type svgStop struct {
	Offset string `xml:"offset,attr"`
	Color  string `xml:"stop-color,attr"`
}
type svgRect struct {
	W, H, RX, Fill string
}
type svgPath struct {
	D    string `xml:"d,attr"`
	Fill string `xml:"fill,attr"`
}

func (r *svgRect) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "width":
			r.W = a.Value
		case "height":
			r.H = a.Value
		case "rx":
			r.RX = a.Value
		case "fill":
			r.Fill = a.Value
		}
	}
	return d.Skip()
}

// Render rasterizes svg to a size×size PNG.
func Render(svg []byte, size int) ([]byte, error) {
	var doc svgDoc
	if err := xml.Unmarshal(svg, &doc); err != nil {
		return nil, err
	}
	vb := strings.Fields(doc.ViewBox)
	if len(vb) != 4 {
		return nil, fmt.Errorf("icon: viewBox %q", doc.ViewBox)
	}
	vw, _ := strconv.ParseFloat(vb[2], 64)
	scale := float32(float64(size) / vw)
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for _, r := range doc.Rects {
		w, h, rx := num(r.W), num(r.H), num(r.RX)
		z := vector.NewRasterizer(size, size)
		roundRect(z, float32(w)*scale, float32(h)*scale, float32(rx)*scale)
		src, err := doc.paint(r.Fill, size)
		if err != nil {
			return nil, err
		}
		z.Draw(img, img.Bounds(), src, image.Point{})
	}
	for _, p := range doc.Paths {
		z := vector.NewRasterizer(size, size)
		if err := tracePath(z, p.D, scale); err != nil {
			return nil, err
		}
		src, err := doc.paint(p.Fill, size)
		if err != nil {
			return nil, err
		}
		z.Draw(img, img.Bounds(), src, image.Point{})
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func num(s string) float64 { f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64); return f }

// paint is the image a fill draws from: a color or a linear gradient.
func (doc *svgDoc) paint(fill string, size int) (image.Image, error) {
	if id, ok := strings.CutPrefix(fill, "url(#"); ok {
		id = strings.TrimSuffix(id, ")")
		for _, g := range doc.Gradients {
			if g.ID == id {
				return g.image(size)
			}
		}
		return nil, fmt.Errorf("icon: no gradient %q", id)
	}
	c, err := hexColor(fill)
	if err != nil {
		return nil, err
	}
	return image.NewUniform(c), nil
}

// image renders the gradient over the whole icon; its x1..y2 are fractions
// of the box (objectBoundingBox units).
func (g svgLinear) image(size int) (image.Image, error) {
	if len(g.Stops) < 2 {
		return nil, fmt.Errorf("icon: gradient %q needs two stops", g.ID)
	}
	c0, err := hexColor(g.Stops[0].Color)
	if err != nil {
		return nil, err
	}
	c1, err := hexColor(g.Stops[len(g.Stops)-1].Color)
	if err != nil {
		return nil, err
	}
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			// x1=0 y1=0 x2=1 y2=1: project onto the diagonal.
			t := (float64(x) + float64(y)) / float64(2*(size-1))
			img.SetNRGBA(x, y, color.NRGBA{lerp8(c0.R, c1.R, t), lerp8(c0.G, c1.G, t), lerp8(c0.B, c1.B, t), 255})
		}
	}
	return img, nil
}

func lerp8(a, b uint8, t float64) uint8 {
	return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t))
}

func hexColor(s string) (color.NRGBA, error) {
	h := strings.TrimPrefix(s, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	n, err := strconv.ParseUint(h, 16, 32)
	if len(h) != 6 || err != nil {
		return color.NRGBA{}, fmt.Errorf("icon: color %q", s)
	}
	return color.NRGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}, nil
}

// roundRect traces a rounded rectangle at the origin; corners are cubic
// approximations of quarter circles.
func roundRect(z *vector.Rasterizer, w, h, r float32) {
	k := r * 0.5523
	z.MoveTo(r, 0)
	z.LineTo(w-r, 0)
	z.CubeTo(w-r+k, 0, w, r-k, w, r)
	z.LineTo(w, h-r)
	z.CubeTo(w, h-r+k, w-r+k, h, w-r, h)
	z.LineTo(r, h)
	z.CubeTo(r-k, h, 0, h-r+k, 0, h-r)
	z.LineTo(0, r)
	z.CubeTo(0, r-k, r-k, 0, r, 0)
	z.ClosePath()
}

// tracePath follows an SVG path's M, L, H, V, C and Z commands.
func tracePath(z *vector.Rasterizer, d string, scale float32) error {
	toks := pathTokens(d)
	var cx, cy, sx, sy float32
	var cmd byte
	i := 0
	next := func() (float32, error) {
		if i >= len(toks) {
			return 0, fmt.Errorf("icon: path %q ends early", d)
		}
		f, err := strconv.ParseFloat(toks[i], 32)
		i++
		return float32(f), err
	}
	for i < len(toks) {
		if t := toks[i]; len(t) == 1 && unicode.IsLetter(rune(t[0])) {
			cmd = t[0]
			i++
		}
		rel := cmd >= 'a'
		ox, oy := float32(0), float32(0)
		if rel {
			ox, oy = cx, cy
		}
		switch cmd | 0x20 { // lower case
		case 'm', 'l':
			x, err := next()
			if err != nil {
				return err
			}
			y, err := next()
			if err != nil {
				return err
			}
			cx, cy = ox+x, oy+y
			if cmd|0x20 == 'm' {
				z.MoveTo(cx*scale, cy*scale)
				sx, sy = cx, cy
				cmd-- // 'm'→'l', 'M'→'L': further pairs are line-tos
			} else {
				z.LineTo(cx*scale, cy*scale)
			}
		case 'h':
			x, err := next()
			if err != nil {
				return err
			}
			cx = ox + x
			z.LineTo(cx*scale, cy*scale)
		case 'v':
			y, err := next()
			if err != nil {
				return err
			}
			cy = oy + y
			z.LineTo(cx*scale, cy*scale)
		case 'c':
			var v [6]float32
			for j := range v {
				f, err := next()
				if err != nil {
					return err
				}
				v[j] = f
			}
			z.CubeTo((ox+v[0])*scale, (oy+v[1])*scale, (ox+v[2])*scale, (oy+v[3])*scale, (ox+v[4])*scale, (oy+v[5])*scale)
			cx, cy = ox+v[4], oy+v[5]
		case 'z':
			z.ClosePath()
			cx, cy = sx, sy
		default:
			return fmt.Errorf("icon: path command %q is not supported", cmd)
		}
	}
	return nil
}

// pathTokens splits path data into commands and numbers: "M34 9c6-6" is
// M, 34, 9, c, 6, -6.
func pathTokens(d string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range d {
		switch {
		case unicode.IsLetter(r) && r != 'e':
			flush()
			out = append(out, string(r))
		case r == ',' || unicode.IsSpace(r):
			flush()
		case r == '-' && cur.Len() > 0 && !strings.HasSuffix(cur.String(), "e"):
			flush()
			cur.WriteRune(r)
		case r == '.' && strings.Contains(cur.String(), "."):
			flush()
			cur.WriteRune(r)
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}
