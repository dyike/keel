//go:build !darwin || ios

package theme

import (
	"image/color"

	"gioui.org/op"
	"gioui.org/text"
)

// UsePlatformText turns platform glyph drawing on or off. Only macOS draws
// text with its platform rasterizer (CoreText); elsewhere it has no effect.
func UsePlatformText(on bool) {}

// BeginTextFrame marks the start of a window's layout for platform glyph
// caches. el calls it.
func BeginTextFrame() {}

// PlatformText reports whether sh draws solid text with a platform
// rasterizer. Only macOS does.
func PlatformText(*text.Shaper) bool { return false }

func installPlatformText(*text.Shaper) {}

func platformPaintRun(*op.Ops, *text.Shaper, []text.Glyph, color.NRGBA, float32, float32) bool {
	return false
}
