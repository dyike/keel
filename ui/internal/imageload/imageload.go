// Package imageload loads images off the UI thread and draws them with a
// placeholder until they arrive. ui/markdown uses it for Markdown images.
package imageload

import (
	"context"
	"fmt"
	"image"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	gio "gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// Loader reads an image source off the UI thread. Apps provide one for
// authenticated requests or paths relative to a document.
type Loader func(context.Context, string) (image.Image, error)

// Asset is shared image data. Read its state under the UI lock. A completed
// load invalidates windows through core.Update.
type Asset struct {
	pixels   image.Point
	image    paint.ImageOp
	err      error
	ready    bool
	revision uint64
}

// Load reads a local path, file URL, HTTP(S) URL or data URL in the
// background. A nil loader uses core.DecodeImage. The request has a 15-second deadline.
func Load(source string, loader Loader) *Asset {
	a := &Asset{revision: 1}
	if loader == nil {
		loader = core.DecodeImage
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		img, err := loader(ctx, source)
		if img == nil && err == nil {
			err = fmt.Errorf("image loader returned no pixels")
		}
		core.Update(func() {
			a.ready, a.err = true, err
			if err == nil && img != nil {
				a.image = paint.NewImageOp(img)
				a.pixels = img.Bounds().Size()
			}
			a.revision++
		})
	}()
	return a
}

// Revision changes when the asset finishes loading, for render caches.
func (a *Asset) Revision() uint64 {
	if a == nil {
		return 0
	}
	return a.revision
}
func (a *Asset) Ready() bool { return a != nil && a.ready }
func (a *Asset) Error() error {
	if a == nil {
		return nil
	}
	return a.err
}

// Size is the image's size in pixels, zero until it has loaded.
func (a *Asset) Size() image.Point {
	if a == nil {
		return image.Point{}
	}
	return a.pixels
}

// View draws an asset at its natural size in dp, scaled down to fit, or a
// placeholder with the alt text while it loads or after it fails.
type View struct {
	Asset *Asset
	Alt   string
}

func (v *View) status() string {
	if v.Asset == nil || !v.Asset.Ready() {
		return "loading"
	}
	if v.Asset.Error() != nil || v.Asset.Size().X == 0 {
		return "error"
	}
	return "loaded"
}

func (v *View) Layout(gtx core.C) core.D {
	gtx.Constraints.Min = image.Point{}
	return core.Semantic(gtx, v.paint, core.Role("image", v.status()), semantic.LabelOp(v.Alt))
}

func (v *View) paint(gtx core.C) core.D {
	if v.status() == "loaded" {
		src := v.Asset.Size()
		scale := float32(1)
		if maxW := gtx.Constraints.Max.X; gtx.Dp(unit.Dp(src.X)) > maxW {
			scale = float32(maxW) / float32(gtx.Dp(unit.Dp(src.X)))
		}
		return gio.Image{Src: v.Asset.image, Fit: gio.ScaleDown, Position: layout.NW, Scale: scale}.Layout(gtx)
	}
	size := gtx.Constraints.Constrain(image.Pt(min(gtx.Dp(260), gtx.Constraints.Max.X), gtx.Dp(64)))
	paint.FillShape(gtx.Ops, theme.Subtle, clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(6)).Op(gtx.Ops))
	label := locale.Current().ImageLoading
	if v.status() == "error" {
		label = locale.Current().ImageFailed
	}
	if v.Alt != "" {
		label += " · " + v.Alt
	}
	g := gtx
	g.Constraints = layout.Exact(size)
	layout.UniformInset(8).Layout(g, func(gtx core.C) core.D {
		return layout.W.Layout(gtx, func(gtx core.C) core.D {
			lb := material.Label(theme.Material, theme.SmallSize, label)
			lb.Color, lb.MaxLines = theme.Muted, 2
			return lb.Layout(gtx)
		})
	})
	return core.D{Size: size}
}
