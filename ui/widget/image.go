package widget

import (
	"context"
	"fmt"
	"github.com/dyike/keel/ui/locale"
	"image"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	gio "gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// ImageLoader runs off the UI thread. Applications can provide authenticated
// requests or resolve paths relative to a document instead of the working dir.
type ImageLoader func(context.Context, string) (image.Image, error)

// ImageAsset is shared image data. Read its state under the UI lock, like
// other widgets. A completed load invalidates windows through core.Update.
type ImageAsset struct {
	source   string
	pixels   image.Point
	image    paint.ImageOp
	err      error
	ready    bool
	revision uint64
}

// LoadImage asynchronously loads a local path, file URL, HTTP(S) URL or data
// URL. A nil loader uses DecodeImage. The request has a 15-second deadline.
func LoadImage(source string, loader ImageLoader) *ImageAsset {
	a := &ImageAsset{source: source, revision: 1}
	if loader == nil {
		loader = DecodeImage
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

// ImageData wraps already-decoded pixels for synchronous rendering.
func ImageData(img image.Image) *ImageAsset {
	a := &ImageAsset{ready: true, revision: 1}
	if img != nil {
		a.image = paint.NewImageOp(img)
		a.pixels = img.Bounds().Size()
	}
	return a
}
func (a *ImageAsset) Revision() uint64 {
	if a == nil {
		return 0
	}
	return a.revision
}
func (a *ImageAsset) Ready() bool { return a != nil && a.ready }
func (a *ImageAsset) Error() error {
	if a == nil {
		return nil
	}
	return a.err
}
func (a *ImageAsset) Size() image.Point {
	if a == nil {
		return image.Point{}
	}
	return a.pixels
}

// ImageView shows shared pixels at their natural size in dp, scaling down to
// fit narrow containers. Loading/failure placeholders retain accessible alt
// text. Optional click callbacks work with the keyboard as well as the mouse.
type ImageView struct {
	Asset    *ImageAsset
	Alt      string
	OnClick  func()
	click    gio.Clickable
	disabled bool
}

func Image(source, alt string) *ImageView { return &ImageView{Asset: LoadImage(source, nil), Alt: alt} }

// SetDisabled prevents pointer and keyboard activation without hiding the image.
func (v *ImageView) SetDisabled(disabled bool) { v.disabled = disabled }

func (v *ImageView) Layout(gtx C) D {
	if v.disabled {
		gtx = gtx.Disabled()
	}
	gtx.Constraints.Min = image.Point{}
	for v.click.Clicked(gtx) {
		core.Call(gtx, v.OnClick)
	}
	return core.Semantic(gtx, func(gtx C) D {
		if v.OnClick != nil {
			return v.click.Layout(gtx, func(gtx C) D {
				if gtx.Enabled() {
					pointer.CursorPointer.Add(gtx.Ops)
				}
				return v.paint(gtx)
			})
		}
		return v.paint(gtx)
	}, core.Role("image", v.status()), semantic.LabelOp(v.Alt), semantic.EnabledOp(gtx.Enabled()))
}
func (v *ImageView) status() string {
	if v.Asset == nil || !v.Asset.Ready() {
		return "loading"
	}
	if v.Asset.Error() != nil || v.Asset.Size().X == 0 {
		return "error"
	}
	return "loaded"
}
func (v *ImageView) paint(gtx C) D {
	if v.status() == "loaded" {
		src := v.Asset.Size()
		scale := float32(1)
		if maxW := gtx.Constraints.Max.X; gtx.Dp(unit.Dp(src.X)) > maxW {
			scale = float32(maxW) / float32(gtx.Dp(unit.Dp(src.X)))
		}
		img := gio.Image{Src: v.Asset.image, Fit: gio.ScaleDown, Position: giolayout.NW, Scale: scale}
		return img.Layout(gtx)
	}
	w := min(gtx.Dp(260), gtx.Constraints.Max.X)
	h := gtx.Dp(64)
	size := gtx.Constraints.Constrain(image.Pt(w, h))
	fillRounded(gtx, theme.Subtle, size.X, size.Y, gtx.Dp(6))
	label := locale.Current().ImageLoading
	if v.status() == "error" {
		label = locale.Current().ImageFailed
	}
	if v.Alt != "" {
		label += " · " + v.Alt
	}
	g := gtx
	g.Constraints = giolayout.Exact(size)
	giolayout.UniformInset(8).Layout(g, func(gtx C) D {
		return giolayout.W.Layout(gtx, func(gtx C) D {
			lb := material.Label(theme.Material, theme.SmallSize, label)
			lb.Color = theme.Muted
			lb.MaxLines = 2
			return lb.Layout(gtx)
		})
	})
	return D{Size: size}
}
