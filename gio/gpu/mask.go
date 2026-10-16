// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"math"
)

// A single-channel mask holds linear coverage. The portable image shaders
// sample sRGB premultiplied pixels, so encode white's RGB while leaving alpha
// linear. This is the same representation previously kept by CoreText pages.
var maskWhite = func() (lut [256]byte) {
	for i := range lut {
		v := float64(i) / 255
		if v <= .0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - .055
		}
		lut[i] = byte(math.Round(v * 255))
	}
	return
}()

func expandMask(src *image.Alpha) *image.RGBA {
	dst := image.NewRGBA(src.Rect)
	for y := src.Rect.Min.Y; y < src.Rect.Max.Y; y++ {
		for x := src.Rect.Min.X; x < src.Rect.Max.X; x++ {
			a := src.Pix[src.PixOffset(x, y)]
			i := dst.PixOffset(x, y)
			v := maskWhite[a]
			dst.Pix[i], dst.Pix[i+1], dst.Pix[i+2], dst.Pix[i+3] = v, v, v, a
		}
	}
	return dst
}
