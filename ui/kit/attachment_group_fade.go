package kit

import (
	"gioui.org/f32"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"image"
	"image/color"
)

// EdgeFade blends overflowing edges into the supplied surrounding surface.
// The 24dp overlays do not intercept input or cover the bottom scrollbar.
func (v *AttachmentGroupView) EdgeFade(surface color.NRGBA) *AttachmentGroupView {
	v.edgeFade = &surface
	return v
}
func (v *AttachmentGroupView) ClearEdgeFade() *AttachmentGroupView { v.edgeFade = nil; return v }

func (v *AttachmentGroupView) paintEdgeFade(cx *el.Context, gtx core.C) {
	if v.edgeFade == nil {
		return
	}
	offset, view, content := v.ScrollState(cx)
	if content <= view || view <= 0 {
		return
	}
	size := gtx.Constraints.Max
	width := min(gtx.Dp(unit.Dp(24)), size.X/2)
	height := max(0, size.Y-gtx.Dp(unit.Dp(scrollbarGutter)))
	if width <= 0 || height <= 0 {
		return
	}
	opaque := *v.edgeFade
	transparent := opaque
	transparent.A = 0
	band := func(x int, from, to color.NRGBA) {
		stack := clip.Rect(image.Rect(x, 0, x+width, height)).Push(gtx.Ops)
		paint.LinearGradientOp{Stop1: f32.Pt(float32(x), 0), Stop2: f32.Pt(float32(x+width), 0), Color1: from, Color2: to}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stack.Pop()
	}
	if offset > 0 {
		band(0, opaque, transparent)
	}
	if offset+view < content {
		band(size.X-width, transparent, opaque)
	}
}
