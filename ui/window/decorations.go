package window

import (
	"image"

	"github.com/dyike/keel/third_party/gio/io/system"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	gtext "github.com/dyike/keel/third_party/gio/text"
	"github.com/dyike/keel/third_party/gio/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// On Linux the title bar may be the app's job: a Wayland compositor without
// server-side decorations (GNOME, WSLg) leaves it to the client. Gio's own
// fallback title bar uses a theme of its own, without Keel's fonts, so
// Chinese titles show as boxes and its colors ignore the Keel theme. Keel
// turns that fallback off (gioapp.Decorated(false), which X11 ignores: its
// window manager always decorates) and draws the bar itself when the
// platform reports no decorations.

// titleBarHeight is the drawn bar's height in dp, like kit.TitleBarHeight.
const titleBarHeight = 38

// askDecorations is what to request from Gio for a window with a title
// bar: true where the platform always draws one, false where Keel may draw it.
func askDecorations(frameless bool) bool { return !frameless && !clientDecorations }

// updateDecorations notes from a ConfigEvent whether Keel draws the bar.
// It reports whether that changed. Call it under the frame lock.
func (w *Window) updateDecorations(decorated, fullscreen bool) bool {
	draws := clientDecorations && !w.opts.Frameless && !decorated && !fullscreen
	changed := draws != w.drawsTitle
	w.drawsTitle = draws
	return changed
}

// layoutTitleBar draws the title bar at the top of the window and returns
// its height in px. The buttons, double click and dragging are Gio's
// widget.Decorations; only the look and the text shaper are Keel's.
func (w *Window) layoutTitleBar(gtx core.C) int {
	if acts := w.deco.Update(gtx); acts != 0 {
		w.perform(acts)
	}
	w.deco.Maximized = w.maximized
	style := material.Decorations(theme.Material, &w.deco,
		system.ActionMinimize|system.ActionMaximize|system.ActionUnmaximize|system.ActionClose|system.ActionMove, w.opts.Title)
	style.Background, style.Foreground = theme.Subtle, theme.Text
	style.Title.Color = theme.Text
	if !w.focused {
		style.Title.Color = theme.Muted
	}
	style.Title.TextSize = theme.Material.TextSize * theme.TextMd / theme.TextBody
	style.Title.Font.Weight = 600
	style.Title.MaxLines = 1
	style.Title.Truncator = "…"
	style.Title.Alignment = gtext.Start
	bar := gtx
	bar.Constraints.Min.Y = gtx.Dp(titleBarHeight)
	bar.Constraints.Max.Y = bar.Constraints.Min.Y
	bar.Constraints.Min.X = gtx.Constraints.Max.X
	dims := style.Layout(bar)
	h := max(dims.Size.Y, bar.Constraints.Min.Y)
	line := gtx.Dp(1)
	paint.FillShape(gtx.Ops, theme.Border, clip.Rect(image.Rect(0, h, gtx.Constraints.Max.X, h+line)).Op())
	return h + line
}

// belowTitleBar draws the bar if Keel owns it, then offsets gtx below it.
func (w *Window) belowTitleBar(gtx core.C) (core.C, func()) {
	if !w.drawsTitle {
		return gtx, func() {}
	}
	h := w.layoutTitleBar(gtx)
	off := op.Offset(image.Pt(0, h)).Push(gtx.Ops)
	gtx.Constraints.Max.Y = max(0, gtx.Constraints.Max.Y-h)
	gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	// The window background fills its whole clip; keep it below the bar.
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	return gtx, func() { area.Pop(); off.Pop() }
}
