package kit

import (
	"image"
	"image/color"

	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Segmented connects adjacent buttons. Explicit positive Gap separates them again.
// Selection behavior is unchanged: single by default, multiple with Multiple.
func (v *ToggleGroupView) Segmented(on bool) *ToggleGroupView { v.segmented = on; return v }

// Gap sets finite nonnegative spacing in dp. Segmented defaults to zero;
// ordinary groups default to theme.SpaceXs. ResetGap restores that default.
func (v *ToggleGroupView) Gap(dp float32) *ToggleGroupView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.gap = &dp
	}
	return v
}
func (v *ToggleGroupView) ResetGap() *ToggleGroupView { v.gap = nil; return v }

func decorateToggleSegment(b *el.DivEl, cx *el.Context, id string, first, last, on bool, variant ToggleVariant) {
	// Preserve the usual 1dp layout inset while painting the connected border ourselves.
	b.Rounded(0).Border(1, color.NRGBA{}).FocusStyle(func(s *el.Style) { s.BorderColor(color.NRGBA{}) })
	b.Decorate(func(gtx core.C, draw func()) {
		rect := image.Rectangle{Max: gtx.Constraints.Max}
		radius := min(gtx.Dp(unit.Dp(theme.RadiusMd)), min(rect.Dx(), rect.Dy())/2)
		shape := clip.RRect{Rect: rect}
		if first {
			shape.NW, shape.SW = radius, radius
		}
		if last {
			shape.NE, shape.SE = radius, radius
		}
		defer shape.Push(gtx.Ops).Pop()
		draw()
		focused := cx.Focused(id)
		if variant == ToggleGhost && !focused {
			return
		}
		border := theme.Border
		if on || focused {
			border = theme.Primary
		}
		bw := gtx.Dp(1)
		if bw <= 0 {
			return
		}
		half := bw / 2
		shape.Rect = image.Rectangle{Min: image.Pt(half, half), Max: rect.Max.Sub(image.Pt(bw-half, bw-half))}
		if first {
			shape.NW, shape.SW = max(radius-half, 0), max(radius-half, 0)
		}
		if last {
			shape.NE, shape.SE = max(radius-half, 0), max(radius-half, 0)
		}
		// Only the preceding segment draws a shared separator. Focus keeps a full outline.
		visible := rect
		if !first && !focused && !on {
			visible.Min.X = bw
		}
		defer clip.Rect(visible).Push(gtx.Ops).Pop()
		paint.FillShape(gtx.Ops, border, clip.Stroke{Path: shape.Path(gtx.Ops), Width: float32(bw)}.Op())
	})
}
