package kit

import (
	"image"

	giolayout "gioui.org/layout"
	"gioui.org/op/paint"
	giowidget "gioui.org/widget"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ImageView shows decoded pixels scaled to fit its width, keeping their aspect
// ratio. Without pixels it shows a placeholder with the alt text, so load
// images elsewhere and SetImage through core.Update when they arrive.
type ImageView struct {
	alt      string
	img      image.Image
	op       paint.ImageOp
	width    float32
	rounded  float32
	onClick  func()
	disabled bool
}

func Image(img image.Image, alt string) *ImageView {
	v := &ImageView{alt: alt}
	v.SetImage(img)
	return v
}

// SetImage replaces the pixels; nil shows the placeholder.
func (v *ImageView) SetImage(img image.Image) {
	v.img = img
	if img != nil && !img.Bounds().Empty() {
		v.op = paint.NewImageOp(img)
	} else {
		v.img = nil
	}
}

// Width caps the displayed width in dp; by default the image fills its
// parent's width, no wider than its own pixels.
func (v *ImageView) Width(dp float32) *ImageView   { v.width = dp; return v }
func (v *ImageView) Rounded(dp float32) *ImageView { v.rounded = dp; return v }
func (v *ImageView) OnClick(fn func()) *ImageView  { v.onClick = fn; return v }
func (v *ImageView) SetDisabled(on bool)           { v.disabled = on }

func (v *ImageView) Render(cx *el.Context) el.Element {
	state := "loaded"
	box := el.Div().Role("image").Name(v.alt).Rounded(v.rounded).Disabled(v.disabled).Items(el.Stretch)
	if v.img == nil {
		state = "loading"
		box.W(el.Dp(260)).MaxW(el.Full).H(el.Dp(64)).Bg(theme.Subtle).Center().
			Child(el.Text(v.alt).TextColor(theme.Muted).MaxLines(1))
	} else {
		b := v.img.Bounds()
		w := float32(b.Dx())
		if v.width > 0 {
			w = min(w, v.width)
		}
		op := v.op
		// Height follows the width actually given, so the aspect ratio holds.
		box.MaxW(el.Dp(w)).WFull().Child(el.Widget(core.Func(func(gtx core.C) core.D {
			size := gtx.Constraints.Max.X
			h := size * b.Dy() / max(b.Dx(), 1)
			gtx.Constraints = giolayout.Exact(image.Pt(size, h))
			return giowidget.Image{Src: op, Fit: giowidget.Contain, Scale: 1 / gtx.Metric.PxPerDp}.Layout(gtx)
		})).WFull().H(el.Dp(w * float32(b.Dy()) / float32(max(b.Dx(), 1)))))
	}
	box.Value(state)
	if v.onClick != nil && !v.disabled {
		box.CursorPointer().Focusable(true).OnClick(v.onClick)
	}
	return box
}
