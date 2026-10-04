package kit

import (
	"context"
	"image"
	"image/color"

	"gioui.org/f32"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
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
// ratio. Source loads pixels asynchronously; SetImage accepts decoded pixels.
type ImageView struct {
	source                   string
	cache                    *ImageCache
	cancel                   context.CancelFunc
	revision                 uint64
	loading                  bool
	imageErr                 error
	loadingContent, fallback el.View
	alt                      string
	img                      image.Image
	op                       paint.ImageOp
	media                    *imageMedia     // SVG or animated GIF from a Source; nil for a still
	ops                      []paint.ImageOp // animation frames
	frame                    int
	width                    float32
	height                   float32
	fit                      ImageFit
	err                      string
	retry                    func()
	preview                  bool
	dialog                   *DialogView
	rounded                  float32
	onClick                  func()
	disabled                 bool
}

func Image(img image.Image, alt string) *ImageView {
	v := &ImageView{alt: alt, cache: defaultImageCache}
	v.SetImage(img)
	return v
}

// SetImage replaces the pixels; nil shows the placeholder.
func (v *ImageView) SetImage(img image.Image) {
	v.stopLoad()
	v.source = ""
	v.setPixels(img)
}

// setMedia shows a decoded source: its still, and for an SVG or animated
// GIF the vector or the frames as well.
func (v *ImageView) setMedia(m *imageMedia) {
	if m == nil {
		v.setPixels(nil)
		return
	}
	v.setPixels(m.still)
	v.media, v.frame = m, 0
	if m.animated() {
		v.ops = frameOps(m.frames)
	}
}

func (v *ImageView) setPixels(img image.Image) {
	v.media, v.ops, v.frame = nil, nil, 0
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
		box.W(el.Dp(w)).MaxW(el.Full).MinH(el.Dp(h)).Bg(theme.Subtle).Center()
		if v.err != "" {
			state = "error"
			if v.fallback != nil {
				box.Child(v.fallback.Render(cx))
			} else {
				box.Gap(theme.SpaceSm).Child(el.Text(v.alt).TextColor(theme.Muted).MaxLines(1), el.Text(v.err).TextSize(theme.TextSm).TextColor(theme.DangerText))
			}
			if v.retry != nil || v.source != "" {
				box.Child(Button(locale.Current().Retry, v.Retry).Name(locale.Current().Name(locale.Current().Retry, v.alt)).Variant(ButtonGhost).Render(cx))
			}
		} else if v.loadingContent != nil {
			box.Child(v.loadingContent.Render(cx))
		} else {
			box.Child(el.Text(v.alt).TextColor(theme.Muted).MaxLines(1))
		}

	} else {
		b := v.img.Bounds()
		natural := image.Pt(b.Dx(), b.Dy())
		var svg *svgIcon
		if v.media != nil && v.media.svg != nil {
			svg = v.media.svg
			vb := svg.icon.ViewBox
			natural = image.Pt(max(1, int(vb.W+.5)), max(1, int(vb.H+.5)))
		}
		w := float32(natural.X)
		if v.width > 0 {
			w = v.width
		}
		op := v.op
		if v.media != nil && v.media.animated() {
			op = v.ops[v.frame%len(v.ops)]
			v.animate(cx)
		}
		height, fit := v.height, v.fit
		box.MaxW(el.Dp(w)).WFull().Child(el.Widget(core.Func(func(gtx core.C) core.D {
			width := gtx.Constraints.Max.X
			h := int(float64(width) * float64(natural.Y) / float64(max(natural.X, 1)))
			if height > 0 {
				h = gtx.Dp(unit.Dp(height))
			}
			size := gtx.Constraints.Constrain(image.Pt(width, h))
			gtx.Constraints = giolayout.Exact(size)
			if svg != nil {
				return layoutSVGImage(gtx, svg, natural, size, fit)
			}
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
	if v.img == nil {
		// State description and interactive fallback children are siblings: an
		// image is a semantic leaf and must not absorb the retry button.
		box.Role("group").Name("").Value("").Child(el.Div().Absolute().Left(0).Top(0).WFull().HFull().Role("image").Name(v.alt).Value(state))
	}
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

// animate shows the next GIF frame after the current one's delay. With
// reduced motion the first frame stays.
func (v *ImageView) animate(cx *el.Context) {
	if theme.ReducedMotion || v.disabled {
		v.frame = 0
		return
	}
	m, frame := v.media, v.frame
	cx.After(imageFrameKey{v, m, frame}, m.delays[frame%len(m.delays)], func() {
		if v.media == m && v.frame == frame {
			v.frame = (frame + 1) % len(m.frames)
		}
	})
}

type imageFrameKey struct {
	v     *ImageView
	m     *imageMedia
	frame int
}

// layoutSVGImage draws an SVG into size by fit, sharp at any scale.
func layoutSVGImage(gtx core.C, svg *svgIcon, natural, size image.Point, fit ImageFit) core.D {
	defer clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops).Pop()
	aspect := float64(natural.X) / float64(max(natural.Y, 1))
	sx, sy := float64(size.X), float64(size.Y)
	w, h := sx, sx/aspect
	switch {
	case fit == ImageFill:
		// Draw at the image's own aspect, then stretch it to the box.
		draw := image.Pt(size.X, max(1, int(sx/aspect+.5)))
		scaleY := float32(sy) / float32(draw.Y)
		defer op.Affine(f32.Affine2D{}.Scale(f32.Point{}, f32.Pt(1, scaleY))).Push(gtx.Ops).Pop()
		svg.layout(gtx, draw, color.NRGBA{}, true)
		return core.D{Size: size}
	case fit == ImageCover:
		if h < sy {
			w, h = sy*aspect, sy
		}
	default: // contain
		if h > sy {
			w, h = sy*aspect, sy
		}
	}
	draw := image.Pt(max(1, int(w+.5)), max(1, int(h+.5)))
	off := image.Pt((size.X-draw.X)/2, (size.Y-draw.Y)/2)
	defer op.Offset(off).Push(gtx.Ops).Pop()
	svg.layout(gtx, draw, color.NRGBA{}, true)
	return core.D{Size: size}
}
