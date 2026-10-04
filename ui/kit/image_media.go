package kit

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"time"

	"gioui.org/op/paint"
	"github.com/srwiley/rasterx"

	"github.com/dyike/keel/ui/core"
)

// imageMedia is a decoded image source: a still image, the frames of an
// animated GIF, or an SVG drawn sharp at whatever size it is shown.
type imageMedia struct {
	still  image.Image // the image, the first frame, or the SVG at its natural size
	frames []image.Image
	delays []time.Duration
	svg    *svgIcon
}

// maxGIFFrameBytes bounds the composited frames of one GIF; a longer
// animation shows its first frame only.
const maxGIFFrameBytes = 96 << 20

// maxSVGStill is the longest side, in px, of an SVG's still image, used for
// its natural size and by components that need pixels (avatars, attachments).
const maxSVGStill = 512

func (m *imageMedia) animated() bool { return len(m.frames) > 1 }

// cost estimates the memory an image holds, for the cache budget.
func (m *imageMedia) cost() int64 {
	b := m.still.Bounds()
	n := int64(1 + len(m.frames))
	return int64(b.Dx()) * int64(b.Dy()) * 4 * n
}

// loadImageMedia reads and decodes an image source.
func loadImageMedia(ctx context.Context, source string) (*imageMedia, error) {
	data, err := core.ReadImageSource(ctx, source)
	if err != nil {
		return nil, err
	}
	return decodeImageMedia(data)
}

// decodeImageMedia decodes SVG, every frame of a GIF, or any format
// core.DecodeImageBytes reads.
func decodeImageMedia(data []byte) (*imageMedia, error) {
	switch {
	case looksLikeSVG(data):
		icon, err := parseSVG(data)
		if err != nil {
			return nil, fmt.Errorf("svg: %w", err)
		}
		return &imageMedia{still: rasterizeSVG(&svgIcon{icon: icon}), svg: &svgIcon{icon: icon}}, nil
	case bytes.HasPrefix(data, []byte("GIF8")):
		if m, err := decodeGIF(data); err == nil {
			return m, nil
		}
	}
	img, err := core.DecodeImageBytes(data)
	if err != nil {
		return nil, err
	}
	return &imageMedia{still: img}, nil
}

func looksLikeSVG(data []byte) bool {
	head := data[:min(len(data), 512)]
	head = bytes.TrimLeft(head, "\xef\xbb\xbf \t\r\n")
	return bytes.HasPrefix(head, []byte("<svg")) || bytes.HasPrefix(head, []byte("<?xml")) && bytes.Contains(data[:min(len(data), 4096)], []byte("<svg"))
}

// rasterizeSVG draws an SVG at its viewBox size, its longest side at most
// maxSVGStill pixels.
func rasterizeSVG(s *svgIcon) image.Image {
	b := s.icon.ViewBox
	scale := min(1, maxSVGStill/max(b.W, b.H))
	if max(b.W, b.H) < 64 {
		scale = 64 / max(b.W, b.H) // tiny viewBoxes would give a blurry still
	}
	w, h := max(1, int(b.W*scale+.5)), max(1, int(b.H*scale+.5))
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	s.icon.Transform = rasterx.Matrix2D{A: scale, D: scale, E: -b.X * scale, F: -b.Y * scale}
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	s.icon.Draw(rasterx.NewDasher(w, h, scanner), 1)
	return img
}

// decodeGIF composites every frame by its disposal method, as browsers do.
func decodeGIF(data []byte) (*imageMedia, error) {
	cfg, err := gif.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > core.MaxImagePixels {
		return nil, fmt.Errorf("gif dimensions exceed limit")
	}
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	frameBytes := int64(bounds.Dx()) * int64(bounds.Dy()) * 4
	if len(g.Image) < 2 || frameBytes*int64(len(g.Image)) > maxGIFFrameBytes {
		first := image.NewRGBA(bounds)
		draw.Draw(first, g.Image[0].Bounds(), g.Image[0], g.Image[0].Bounds().Min, draw.Over)
		return &imageMedia{still: first}, nil
	}
	canvas := image.NewRGBA(bounds)
	m := &imageMedia{}
	for i, frame := range g.Image {
		disposal := byte(0)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}
		var previous *image.RGBA
		if disposal == gif.DisposalPrevious {
			previous = cloneRGBA(canvas)
		}
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
		m.frames = append(m.frames, cloneRGBA(canvas))
		delay := 100 * time.Millisecond // browsers treat 0 and 1 as 100ms too
		if i < len(g.Delay) && g.Delay[i] > 1 {
			delay = time.Duration(g.Delay[i]) * 10 * time.Millisecond
		}
		m.delays = append(m.delays, delay)
		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, frame.Bounds(), image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			canvas = previous
		}
	}
	m.still = m.frames[0]
	return m, nil
}

func cloneRGBA(img *image.RGBA) *image.RGBA {
	c := image.NewRGBA(img.Bounds())
	copy(c.Pix, img.Pix)
	return c
}

// frameOps converts frames to paint operations once.
func frameOps(frames []image.Image) []paint.ImageOp {
	ops := make([]paint.ImageOp, len(frames))
	for i, f := range frames {
		ops[i] = paint.NewImageOp(f)
	}
	return ops
}
