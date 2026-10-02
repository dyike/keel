package kit

import (
	"image"

	giolayout "gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	giowidget "gioui.org/widget"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ImageFit controls pixels inside a fixed Size. Without a height, aspect ratio
// determines the height regardless of Fit.
type ImageFit uint8

const (
	ImageContain ImageFit = iota
	ImageCover
	ImageFill
)

// ImageView shows decoded pixels scaled to fit its width, keeping their aspect
// ratio. Without pixels it shows a placeholder with the alt text, so load
// images elsewhere and SetImage through core.Update when they arrive.
type ImageView struct {
	alt      string
	img      image.Image
	op       paint.ImageOp
	width    float32
	height   float32
	fit      ImageFit
	err      string
	retry    func()
	preview  bool
	dialog   *DialogView
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
	v.err = ""
	if img != nil && !img.Bounds().Empty() {
		v.op = paint.NewImageOp(img)
	} else {
		v.img = nil
		v.op = paint.ImageOp{}
		if v.dialog != nil {
			v.dialog.SetValue(false)
		}
	}
}

// Width caps the displayed width in dp; by default the image fills its
// parent's width, no wider than its own pixels.
func (v *ImageView) Width(dp float32) *ImageView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.width = dp
	}
	return v
}
func (v *ImageView) Size(width, height float32) *ImageView {
	v.Width(width)
	if height >= 0 && finiteNumber(float64(height)) {
		v.height = height
	}
	return v
}
func (v *ImageView) Fit(fit ImageFit) *ImageView {
	if fit <= ImageFill {
		v.fit = fit
	}
	return v
}
func (v *ImageView) Rounded(dp float32) *ImageView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.rounded = dp
	}
	return v
}
func (v *ImageView) OnClick(fn func()) *ImageView { v.onClick = fn; return v }
func (v *ImageView) OnRetry(fn func()) *ImageView { v.retry = fn; return v }

// SetError replaces pixels with a retryable failure. SetImage clears the error.
func (v *ImageView) SetError(message string) { v.SetImage(nil); v.err = message }
func (v *ImageView) SetDisabled(on bool) {
	v.disabled = on
	if on && v.dialog != nil {
		v.dialog.SetValue(false)
	}
}

// Preview opens the current pixels in a modal when clicked or keyboard activated.
func (v *ImageView) Preview() *ImageView {
	v.preview = true
	if v.dialog == nil {
		v.dialog = Dialog(v.alt).Width(900).Body(el.ViewFunc(func(cx *el.Context) el.Element {
			w, h := cx.ViewportSize()
			preview := *v
			preview.dialog, preview.preview, preview.onClick = nil, false, nil
			preview.width, preview.height, preview.fit = max(1, w-80), max(1, h-200), ImageContain
			return preview.Render(cx)
		}))
	}
	return v
}

func (v *ImageView) Render(cx *el.Context) el.Element {
	state := "loaded"
	box := el.Div().ID(autoID("image", v)).Role("image").Name(v.alt).Rounded(v.rounded).Disabled(v.disabled).Items(el.Stretch)
	if v.img == nil {
		state = "loading"
		w, h := float32(260), float32(64)
		if v.width > 0 {
			w = v.width
		}
		if v.height > 0 {
			h = v.height
		}
		box.W(el.Dp(w)).MaxW(el.Full).MinH(el.Dp(h)).Bg(theme.Subtle).Center().
			Child(el.Text(v.alt).TextColor(theme.Muted).MaxLines(1))
		if v.err != "" {
			state = "error"
			box.Gap(theme.SpaceSm).Child(el.Text(v.err).TextSize(theme.TextSm).TextColor(theme.DangerText))
			if v.retry != nil {
				box.Child(Button(locale.Current().Retry, func() {
					if v.err == "" || v.disabled {
						return
					}
					v.err = ""
					v.retry()
				}).Name(locale.Current().Name(locale.Current().Retry, v.alt)).Variant(ButtonGhost).Render(cx))
			}
		}
	} else {
		b := v.img.Bounds()
		w := float32(b.Dx())
		if v.width > 0 {
			w = v.width
		}
		op := v.op
		height, fit := v.height, v.fit
		box.MaxW(el.Dp(w)).WFull().Child(el.Widget(core.Func(func(gtx core.C) core.D {
			width := gtx.Constraints.Max.X
			h := int(float64(width) * float64(b.Dy()) / float64(max(b.Dx(), 1)))
			if height > 0 {
				h = gtx.Dp(unit.Dp(height))
			}
			size := gtx.Constraints.Constrain(image.Pt(width, h))
			gtx.Constraints = giolayout.Exact(size)
			mode := giowidget.Contain
			if fit == ImageCover {
				mode = giowidget.Cover
			} else if fit == ImageFill {
				mode = giowidget.Fill
			}
			return giowidget.Image{Src: op, Fit: mode, Position: giolayout.Center, Scale: 1 / gtx.Metric.PxPerDp}.Layout(gtx)
		})).WFull())
	}
	box.Value(state)
	if v.img != nil && (v.onClick != nil || v.preview) && !v.disabled {
		box.CursorPointer().Focusable(true).OnClick(func() {
			if v.preview {
				v.dialog.SetValue(true)
			}
			if v.onClick != nil {
				v.onClick()
			}
		})
	}
	if v.dialog != nil {
		box.Child(v.dialog.Render(cx))
	}
	return box
}
