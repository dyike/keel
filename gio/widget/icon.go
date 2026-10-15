// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/dyike/keel/gio/internal/f32color"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/unit"

	"golang.org/x/exp/shiny/iconvg"
)

type Icon struct {
	src []byte
	// Keel patch: one shared icon can appear at several sizes/colors in a
	// frame. Keep a bounded MRU of immutable images, not just the last one.
	images                 [iconCacheEntries]iconImage
	imageCount, imageBytes int
}

const iconCacheEntries = 8
const iconCacheBytes = 1 << 20

type iconImage struct {
	op    paint.ImageOp
	size  int
	color color.NRGBA
	bytes int
}

const defaultIconSize = unit.Dp(24)

// NewIcon returns a new Icon from IconVG data.
func NewIcon(data []byte) (*Icon, error) {
	_, err := iconvg.DecodeMetadata(data)
	if err != nil {
		return nil, err
	}
	return &Icon{src: data}, nil
}

// Layout displays the icon with its size set to the X minimum constraint.
func (ic *Icon) Layout(gtx layout.Context, color color.NRGBA) layout.Dimensions {
	sz := gtx.Constraints.Min.X
	if sz == 0 {
		sz = gtx.Dp(defaultIconSize)
	}
	size := gtx.Constraints.Constrain(image.Pt(sz, sz))
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()

	ico := ic.image(size.X, color)
	ico.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{
		Size: ico.Size(),
	}
}

func (ic *Icon) image(sz int, color color.NRGBA) paint.ImageOp {
	for i := 0; i < ic.imageCount; i++ {
		cached := ic.images[i]
		if cached.size == sz && cached.color == color {
			copy(ic.images[1:i+1], ic.images[:i])
			ic.images[0] = cached
			return cached.op
		}
	}
	m, _ := iconvg.DecodeMetadata(ic.src)
	dx, dy := m.ViewBox.AspectRatio()
	img := image.NewRGBA(image.Rectangle{Max: image.Point{X: sz, Y: int(float32(sz) * dy / dx)}})
	var ico iconvg.Rasterizer
	ico.SetDstImage(img, img.Bounds(), draw.Src)
	m.Palette[0] = f32color.NRGBAToLinearRGBA(color)
	iconvg.Decode(&ico, ic.src, &iconvg.DecodeOptions{
		Palette: &m.Palette,
	})
	op := paint.NewImageOp(img)
	bytes := len(img.Pix)
	for ic.imageCount > 0 && (ic.imageCount == iconCacheEntries || ic.imageBytes > iconCacheBytes-bytes) {
		ic.imageCount--
		ic.imageBytes -= ic.images[ic.imageCount].bytes
		ic.images[ic.imageCount] = iconImage{}
	}
	// An oversized image retains only itself, matching the old single-image
	// cache. It must not multiply the budget by eight at unusually large DPI.
	copy(ic.images[1:ic.imageCount+1], ic.images[:ic.imageCount])
	ic.images[0] = iconImage{op: op, size: sz, color: color, bytes: bytes}
	ic.imageCount++
	ic.imageBytes += bytes
	return op
}
