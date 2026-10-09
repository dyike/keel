// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package paint

import (
	"image/color"

	"gioui.org/op"
)

// AddTinted sets an image brush multiplied by tint in premultiplied linear
// color space. White preserves the image. The immutable source and its GPU
// texture can be shared across colors on devices supporting tinted quads.
// This optional Keel extension is available on Metal builds. Other builds
// retain the ordinary ImageOp API so atlas callers can keep colored pages.
func (i ImageOp) AddTinted(o *op.Ops, tint color.NRGBA) { i.addTinted(o, tint) }
