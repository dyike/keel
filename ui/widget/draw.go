package widget

import (
	"image"
	"image/color"

	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// fillRounded paints a w×h rounded rectangle at the current origin.
func fillRounded(gtx C, c color.NRGBA, w, h, r int) {
	paint.FillShape(gtx.Ops, c, clip.UniformRRect(image.Rect(0, 0, w, h), r).Op(gtx.Ops))
}

func paintOp(gtx C, c color.NRGBA, shape clip.Op) {
	paint.FillShape(gtx.Ops, c, shape)
}
