package kit

import (
	"image"
	"image/color"
	"time"

	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ResizableHandleAppearance describes visual line widths in dp, independently
// of the fixed 6dp hit area. Widths are clamped to 0..6; zero hides that state.
type ResizableHandleAppearance struct {
	Idle, Hover, Pressed, Dragging float32
	Color, ActiveColor             color.NRGBA
	Duration                       time.Duration
}

// HandleAppearance customizes the theme-derived defaults on each render.
// nil restores defaults. It must not mutate the element tree.
func (v *ResizableView) HandleAppearance(fn func(ResizableHandleAppearance) ResizableHandleAppearance) *ResizableView {
	v.handleAppearance = fn
	return v
}

func (v *ResizableView) decorateHandle(cx *el.Context, id string, handle *el.DivEl) {
	a := ResizableHandleAppearance{Idle: 1, Hover: 2, Pressed: 3, Dragging: 4, Color: theme.Border, ActiveColor: theme.Primary, Duration: 150 * time.Millisecond}
	if v.handleAppearance != nil {
		a = v.handleAppearance(a)
	}
	width, c := a.Idle, a.Color
	if !v.disabled && !v.hiddenFirst && !v.hiddenSecond && cx.Enabled(id) {
		switch {
		case v.dragging:
			width, c = a.Dragging, a.ActiveColor
		case v.pressed:
			width, c = a.Pressed, a.ActiveColor
		case cx.Hovered(id) || cx.Focused(id):
			width, c = a.Hover, a.ActiveColor
		}
	}
	if !finiteNumber(float64(width)) {
		width = 1
	}
	width = min(max(width, 0), handleSize)
	if a.Duration <= 0 {
		v.handleMotion = valueMotion{}
	}
	width = v.handleMotion.sample(cx, width, max(a.Duration, time.Nanosecond))
	handle.Decorate(func(gtx core.C, draw func()) {
		draw()
		size := gtx.Constraints.Max
		line := gtx.Dp(unit.Dp(width))
		if line <= 0 {
			return
		}
		r := image.Rectangle{Max: size}
		if v.vertical {
			line = min(line, size.Y)
			r.Min.Y = (size.Y - line) / 2
			r.Max.Y = r.Min.Y + line
		} else {
			line = min(line, size.X)
			r.Min.X = (size.X - line) / 2
			r.Max.X = r.Min.X + line
		}
		paint.FillShape(gtx.Ops, c, clip.Rect(r).Op())
	})
}
