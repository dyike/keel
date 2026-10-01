package theme

import (
	"image/color"

	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/internal/loop"
)

// Palette contains every global color token. Start from Light or Dark when
// customizing: Apply replaces all colors, including zero (transparent) values.
type Palette struct {
	PrimaryText, DangerText, CodeBg, CodeText      color.NRGBA
	Bg, Surface, Border, Text, Muted               color.NRGBA
	Primary, PrimaryHover, Danger, DangerHover     color.NRGBA
	Success, Warning, Info                         color.NRGBA
	Subtle, SubtleHover, OnColor, Highlight, Scrim color.NRGBA
	// Chart is the categorical order for data series: slot i always means
	// series i. Validated for color-vision deficiency against Surface.
	Chart [8]color.NRGBA
}

// Light returns an independent copy of the default light palette.
func Light() Palette {
	return Palette{
		PrimaryText: RGB(0x1d4ed8), DangerText: RGB(0xb91c1c), CodeBg: RGB(0xf0f1f3), CodeText: RGB(0x1f2328), Bg: RGB(0xf5f6f8), Surface: RGB(0xffffff), Border: RGB(0xe3e5e8),
		Text: RGB(0x1f2328), Muted: RGB(0x6b7280), Primary: RGB(0x2563eb), PrimaryHover: RGB(0x1d4ed8),
		Danger: RGB(0xdc2626), DangerHover: RGB(0xb91c1c), Success: RGB(0x15803d), Warning: RGB(0xa16207), Info: RGB(0x0369a1),
		Subtle: RGB(0xeceef1), SubtleHover: RGB(0xe2e5e9), OnColor: RGB(0xffffff), Highlight: RGB(0xdbeafe), Scrim: color.NRGBA{A: 0x66},
		Chart: [8]color.NRGBA{RGB(0x2a78d6), RGB(0xeb6834), RGB(0x1baf7a), RGB(0xeda100), RGB(0xe87ba4), RGB(0x008300), RGB(0x4a3aa7), RGB(0xe34948)},
	}
}

// Dark returns an independent copy of the dark palette.
func Dark() Palette {
	return Palette{
		PrimaryText: RGB(0x93c5fd), DangerText: RGB(0xfca5a5), CodeBg: RGB(0x161b22), CodeText: RGB(0xe6edf3), Bg: RGB(0x111827), Surface: RGB(0x1f2937), Border: RGB(0x4b5563),
		Text: RGB(0xf3f4f6), Muted: RGB(0x9ca3af), Primary: RGB(0x2563eb), PrimaryHover: RGB(0x1d4ed8),
		Danger: RGB(0xdc2626), DangerHover: RGB(0xb91c1c), Success: RGB(0x4ade80), Warning: RGB(0xfacc15), Info: RGB(0x7dd3fc),
		Subtle: RGB(0x374151), SubtleHover: RGB(0x4b5563), OnColor: RGB(0xffffff), Highlight: RGB(0x1e3a5f), Scrim: color.NRGBA{A: 0x99},
		Chart: [8]color.NRGBA{RGB(0x3987e5), RGB(0xd95926), RGB(0x199e70), RGB(0xc98500), RGB(0xd55181), RGB(0x008300), RGB(0x9085e9), RGB(0xe66767)},
	}
}

// Current returns a copy of the current colors. Read it under the UI lock,
// like the public color variables, or before opening the first window.
func Current() Palette {
	return Palette{PrimaryText: PrimaryText, DangerText: DangerText, CodeBg: CodeBg, CodeText: CodeText, Bg: Bg, Surface: Surface, Border: Border, Text: Text, Muted: Muted,
		Primary: Primary, PrimaryHover: PrimaryHover, Danger: Danger, DangerHover: DangerHover,
		Success: Success, Warning: Warning, Info: Info, Subtle: Subtle, SubtleHover: SubtleHover,
		OnColor: OnColor, Highlight: Highlight, Scrim: Scrim, Chart: Chart}
}

var revision uint64

// Revision changes whenever Apply replaces the palette. Use it in custom
// render cache keys. Read under the UI lock, like Current.
func Revision() uint64 { return revision }

// Apply synchronously replaces the global palette and redraws every window.
// Call before opening windows or from a UI callback under the frame lock.
// Other goroutines must use core.Update(func() { theme.Apply(p) }). Apply
// never acquires the frame lock itself, so callbacks cannot deadlock on it.
// Fonts, the text shaper and the Material pointer remain unchanged.
func Apply(p Palette) {
	PrimaryText, DangerText, CodeBg, CodeText = p.PrimaryText, p.DangerText, p.CodeBg, p.CodeText
	Bg, Surface, Border, Text, Muted = p.Bg, p.Surface, p.Border, p.Text, p.Muted
	Primary, PrimaryHover, Danger, DangerHover = p.Primary, p.PrimaryHover, p.Danger, p.DangerHover
	Success, Warning, Info = p.Success, p.Warning, p.Info
	Subtle, SubtleHover, OnColor, Highlight, Scrim = p.Subtle, p.SubtleHover, p.OnColor, p.Highlight, p.Scrim
	Chart = p.Chart
	Material.Palette = material.Palette{Fg: Text, Bg: Surface, ContrastBg: Primary, ContrastFg: OnColor}
	revision++
	loop.InvalidateAll()
}
