package kit

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const maxSVGBytes = 1 << 20

type svgIcon struct {
	icon     *oksvg.SvgIcon
	size     image.Point
	color    color.NRGBA
	original bool
	valid    bool
	image    paint.ImageOp
}

// SVGIcon parses SVG bytes synchronously. Keep the returned view for reuse;
// parsing should not happen every Render. Input is limited to 1MiB. Unsupported
// SVG elements return an error. As with built-in icons, Color tints the alpha
// silhouette; OriginalColors(true) preserves the SVG's colors instead.
func SVGIcon(data []byte) (*IconView, error) {
	if len(data) == 0 || len(data) > maxSVGBytes {
		return nil, fmt.Errorf("kit.SVGIcon: SVG must contain 1..%d bytes", maxSVGBytes)
	}
	parsed, err := parseSVG(data)
	if err != nil {
		return nil, fmt.Errorf("kit.SVGIcon: %w", err)
	}
	return &IconView{size: 18, svg: &svgIcon{icon: parsed}}, nil
}

// parseSVG reads an SVG document with a usable viewBox.
func parseSVG(data []byte) (*oksvg.SvgIcon, error) {
	parsed, err := oksvg.ReadReplacingCurrentColor(bytes.NewReader(data), "black", oksvg.StrictErrorMode)
	if err != nil {
		return nil, err
	}
	b := parsed.ViewBox
	if b.W <= 0 || b.H <= 0 || !finiteNumber(b.W) || !finiteNumber(b.H) || !finiteNumber(b.X) || !finiteNumber(b.Y) {
		return nil, fmt.Errorf("invalid viewBox")
	}
	return parsed, nil
}

// SVGIconFile reads a local SVG file once. It does not watch files or fetch URLs.
func SVGIconFile(path string) (*IconView, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxSVGBytes+1))
	if err != nil {
		return nil, err
	}
	return SVGIcon(data)
}

// OriginalColors preserves an SVG's source colors (currentColor becomes black).
// It has no effect on built-in or Gio vector icons. False restores Color/theme tint.
func (i *IconView) OriginalColors(on bool) *IconView { i.originalColors = on; return i }

func (s *svgIcon) layout(gtx core.C, size image.Point, tint color.NRGBA, original bool) core.D {
	if size.X <= 0 || size.Y <= 0 {
		return core.D{Size: size}
	}
	// Rasterize at physical pixel density, retaining at most one 2048² image.
	factor := math.Min(1, 2048/math.Max(float64(size.X), float64(size.Y)))
	rasterSize := image.Pt(max(1, int(float64(size.X)*factor)), max(1, int(float64(size.Y)*factor)))
	if !s.valid || s.size != rasterSize || s.color != tint || s.original != original {
		img := image.NewRGBA(image.Rectangle{Max: rasterSize})
		b := s.icon.ViewBox
		scale := math.Min(float64(rasterSize.X)/b.W, float64(rasterSize.Y)/b.H)
		x, y := (float64(rasterSize.X)-b.W*scale)/2, (float64(rasterSize.Y)-b.H*scale)/2
		s.icon.Transform = rasterx.Matrix2D{A: scale, D: scale, E: x - b.X*scale, F: y - b.Y*scale}
		scanner := rasterx.NewScannerGV(rasterSize.X, rasterSize.Y, img, img.Bounds())
		s.icon.Draw(rasterx.NewDasher(rasterSize.X, rasterSize.Y, scanner), 1)
		convertSVGColors(img, tint, original)
		s.image, s.size, s.color, s.original, s.valid = paint.NewImageOp(img), rasterSize, tint, original, true
	}
	defer clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops).Pop()
	defer op.Affine(f32.Affine2D{}.Scale(f32.Point{}, f32.Pt(float32(size.X)/float32(rasterSize.X), float32(size.Y)/float32(rasterSize.Y)))).Push(gtx.Ops).Pop()
	s.image.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return core.D{Size: size}
}

// Gio samples image RGB in linear space before alpha blending. Go rasterizers
// premultiply in sRGB instead; convert once when refreshing the cached image.
func convertSVGColors(img *image.RGBA, tint color.NRGBA, original bool) {
	encode := func(c, a uint8) uint8 {
		if a == 255 {
			return c
		}
		if a == 0 {
			return 0
		}
		x := float64(c) / 255
		if x <= 0.04045 {
			x /= 12.92
		} else {
			x = math.Pow((x+0.055)/1.055, 2.4)
		}
		x *= float64(a) / 255
		if x <= 0.0031308 {
			x *= 12.92
		} else {
			x = 1.055*math.Pow(x, 1/2.4) - 0.055
		}
		return uint8(math.Round(x * 255))
	}
	var palette [256]color.RGBA
	if !original {
		for i := range palette {
			a := uint8(uint32(i) * uint32(tint.A) / 255)
			palette[i] = color.RGBA{R: encode(tint.R, a), G: encode(tint.G, a), B: encode(tint.B, a), A: a}
		}
	}
	for p := 0; p < len(img.Pix); p += 4 {
		a := img.Pix[p+3]
		if !original {
			c := palette[a]
			img.Pix[p], img.Pix[p+1], img.Pix[p+2], img.Pix[p+3] = c.R, c.G, c.B, c.A
		} else if a > 0 && a < 255 {
			for channel := range 3 {
				straight := uint8(min(255, (uint32(img.Pix[p+channel])*255+uint32(a)/2)/uint32(a)))
				img.Pix[p+channel] = encode(straight, a)
			}
		}
	}
}
