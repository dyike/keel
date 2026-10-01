// Package theme holds the colors, text sizes and fonts every UI module reads.
// Apply changes the global palette under the UI frame lock.
package theme

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Colors. Components read them at layout time.
var (
	Bg           = RGB(0xf5f6f8) // window background
	Surface      = RGB(0xffffff) // cards and fields
	Border       = RGB(0xe3e5e8) // borders and dividers
	Text         = RGB(0x1f2328) // body text
	Muted        = RGB(0x6b7280) // secondary text, hints, unchecked icons
	Primary      = RGB(0x2563eb) // primary buttons, links, focus, checked icons
	DangerText   = RGB(0xb91c1c) // danger text on surfaces
	PrimaryText  = RGB(0x1d4ed8) // text on selected surfaces
	CodeBg       = RGB(0xf0f1f3)
	CodeText     = RGB(0x1f2328)
	PrimaryHover = RGB(0x1d4ed8)
	DangerHover  = RGB(0xb91c1c)
	SubtleHover  = RGB(0xe2e5e9)
	Success      = RGB(0x15803d)        // positive status
	Warning      = RGB(0xa16207)        // caution status
	Info         = RGB(0x0369a1)        // informational status
	Danger       = RGB(0xdc2626)        // danger buttons
	Subtle       = RGB(0xeceef1)        // secondary buttons
	OnColor      = RGB(0xffffff)        // text on Primary and Danger
	Highlight    = RGB(0xdbeafe)        // selected rows and options
	Scrim        = color.NRGBA{A: 0x66} // dims the window behind a dialog
)

// Text sizes.
const (
	BodySize    unit.Sp = 15
	SmallSize   unit.Sp = 13
	HeadingSize unit.Sp = 22
)

// Face lists font families in priority order. Pinning a CJK family avoids tofu
// from a system fallback font that lacks some simplified Chinese glyphs.
const Face font.Typeface = "PingFang SC, Hiragino Sans GB, Microsoft YaHei, Noto Sans CJK SC, Noto Sans SC, Go"

// CJKNudge moves text down inside boxes: CJK system fonts have tall ascents, so
// their glyphs otherwise sit high in buttons and fields. Tuned for macOS.
const CJKNudge unit.Dp = 2

// TextShift moves glyphs down by this fraction of the text size when they are
// drawn. Gio places the first baseline one ascent below the top and ends the
// box one descent below the last baseline; CJK system fonts have a large
// descent, so text otherwise sits visibly high in its box (measured: 2px of
// space above the ink and 17px below, for 15sp at 2×). 0.22 balances CJK and
// Latin text. Used by ui/el and ui/markdown; ui/widget still uses CJKNudge.
const TextShift = 0.22

// Shift returns TextShift for a text size, in pixels.
func Shift(m unit.Metric, size unit.Sp) int { return int(float32(m.Sp(size))*TextShift + 0.5) }

// Material is the underlying Gio theme: text shaper and icons.
var Material = newMaterial()

func newMaterial() *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette = material.Palette{Fg: Text, Bg: Surface, ContrastBg: Primary, ContrastFg: OnColor}
	th.Face = Face
	th.TextSize = BodySize
	return th
}

// RGB converts 0xRRGGBB to an opaque color.
func RGB(c uint32) color.NRGBA {
	return color.NRGBA{R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c), A: 0xff}
}
