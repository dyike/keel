// Package theme holds the colors, text sizes and fonts every UI module reads.
// Apply changes the global palette under the UI frame lock.
package theme

import (
	"image/color"

	"gioui.org/font"
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
	Success      = RGB(0x15803d)                                   // positive status
	Warning      = RGB(0xa16207)                                   // caution status
	Info         = RGB(0x0369a1)                                   // informational status
	Danger       = RGB(0xdc2626)                                   // danger buttons
	Subtle       = RGB(0xeceef1)                                   // secondary buttons
	OnColor      = RGB(0xffffff)                                   // text on Primary and Danger
	Highlight    = RGB(0xdbeafe)                                   // selected rows and options
	Scrim        = color.NRGBA{A: 0x66}                            // dims the window behind a dialog
	Shadow       = color.NRGBA{R: 0x10, G: 0x18, B: 0x28, A: 0x2e} // tints raised surfaces' shadows
	// Chart is the categorical order for data series; see Palette.Chart.
	Chart = Light().Chart
	// BgGradient and PrimaryGradient are optional; see Palette.
	BgGradient, PrimaryGradient Gradient
)

// Text sizes.
const (
	BodySize    unit.Sp = 15
	SmallSize   unit.Sp = 13
	HeadingSize unit.Sp = 22

	// ControlHeight is the height of every single-line field: inputs,
	// selects, pickers and search boxes, so fields side by side line up.
	ControlHeight unit.Dp = 36
)

// Face lists font families in priority order. Pinning a CJK family avoids tofu
// from a system fallback font that lacks some simplified Chinese glyphs.
const Face font.Typeface = "PingFang SC, Hiragino Sans GB, Microsoft YaHei, Noto Sans CJK SC, Noto Sans SC, Go, " + EmojiFace

// MonoFace lists monospaced families for code, numbers in columns and
// keycaps; CJK falls back to Face's fonts.
const MonoFace font.Typeface = "SF Mono, Menlo, Cascadia Mono, Consolas, DejaVu Sans Mono, Go Mono, PingFang SC, Microsoft YaHei, Noto Sans CJK SC, " + EmojiFace

// EmojiFace lists platform emoji families for use at the end of a custom
// typeface list. Naming these explicitly lets common-script emoji reach the
// system's emoji font instead of an arbitrary missing-glyph fallback.
// Actual color rendering depends on the font format supported by Gio.
const EmojiFace font.Typeface = "Apple Color Emoji, Segoe UI Emoji, Noto Color Emoji, Noto Emoji"

// Material is the underlying Gio theme: text shaper and icons.
var Material = newMaterial()

// lazyShaper returns a shaper of faces and the system's fonts that sets
// itself up when it first shapes text, as text.Shaper does when made without
// NewShaper. Setting up reads the system's font index: some 70 ms and 60 MB of
// allocations that a process drawing no text, such as a helper started from
// the same binary, need not pay at startup.
func lazyShaper(faces []font.FontFace) *text.Shaper {
	sh := new(text.Shaper)
	text.WithCollection(faces)(sh)
	return sh
}

func newMaterial() *material.Theme {
	th := material.NewTheme()
	th.Shaper = lazyShaper(fallbackFaces())
	th.Palette = material.Palette{Fg: Text, Bg: Surface, ContrastBg: Primary, ContrastFg: OnColor}
	th.Face = Face
	th.TextSize = BodySize
	return th
}

// RGB converts 0xRRGGBB to an opaque color.
func RGB(c uint32) color.NRGBA {
	return color.NRGBA{R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c), A: 0xff}
}
