package kit

import (
	"image"
	"image/color"
	"runtime"
	"slices"

	"gioui.org/io/system"
	"gioui.org/op/clip"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// TitleBarHeight is the height of a TitleBar in dp.
const TitleBarHeight = 38

// TitleBarView is a window title bar drawn by the app, for windows opened
// with window.Options{Frameless: true}. Dragging its empty area moves the
// window. On macOS the close / minimize / zoom buttons sit on the left as
// colored circles; elsewhere minimize / maximize / close sit on the right.
// Leading and Trailing add the app's own controls (a search box, buttons).
//
// In a window with a system title bar it draws only the title and the app's
// controls, as a header.
type TitleBarView struct {
	title             string
	leading, trailing []el.View
	goos              string // runtime.GOOS; tests set others
}

func TitleBar(title string) *TitleBarView { return &TitleBarView{title: title, goos: runtime.GOOS} }

// Leading adds views after the window buttons on the left.
func (v *TitleBarView) Leading(views ...el.View) *TitleBarView {
	v.leading = slices.Clone(views)
	return v
}

// Trailing adds views at the right, before the window buttons off macOS.
func (v *TitleBarView) Trailing(views ...el.View) *TitleBarView {
	v.trailing = slices.Clone(views)
	return v
}
func (v *TitleBarView) SetTitle(s string) { v.title = s }

// dragArea marks an element as the window's move handle. It must not contain
// the buttons: the platform treats any press inside it as a window drag.
func dragArea(cx *el.Context, w core.WindowControls, d *el.DivEl) *el.DivEl {
	return d.Decorate(func(gtx core.C, draw func()) {
		if gtx.Enabled() {
			origin, visible := cx.PaintGeometry()
			rect := image.Rectangle{Min: origin, Max: origin.Add(gtx.Constraints.Max)}.Intersect(visible)
			px := gtx.Metric.PxPerDp
			if px <= 0 {
				px = 1
			}
			w.TitleBarArea(float32(rect.Min.X)/px, float32(rect.Min.Y)/px, float32(rect.Dx())/px, float32(rect.Dy())/px)
			area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
			system.ActionInputOp(system.ActionMove).Add(gtx.Ops)
			area.Pop()
		}
		draw()
	})
}

func (v *TitleBarView) Render(cx *el.Context) el.Element {
	w := core.CurrentWindow()
	controls := w != nil && w.Frameless()
	mac := v.goos == "darwin"
	bar := el.Div().Role("banner").Name(v.title).Row().Items(el.Center).H(el.Dp(TitleBarHeight)).Bg(theme.Subtle).Px(8).Gap(8)
	if controls && mac {
		bar.Child(v.lights(cx, w))
	}
	for _, l := range v.leading {
		bar.Child(renderOptional(cx, l))
	}
	fg := theme.Text
	if w != nil && !w.Focused() {
		fg = theme.Muted
	}
	title := el.Text(v.title).TextSize(theme.TextMd).Bold().MaxLines(1).TextColor(fg)
	middle := el.Div().Grow().W(el.Dp(0)).H(el.Dp(TitleBarHeight)).Row().Items(el.Center).Child(title)
	if mac {
		middle.Justify(el.Center) // macOS centers window titles
	}
	if controls {
		middle.OnDoubleClick(w.ToggleMaximize)
		dragArea(cx, w, middle)
	}
	bar.Child(middle)
	for _, t := range v.trailing {
		bar.Child(renderOptional(cx, t))
	}
	if controls && !mac {
		bar.Pr(0).Child(v.buttons(cx, w))
	}
	return el.Div().Items(el.Stretch).Child(bar, el.Div().H(el.Dp(1)).Bg(theme.Border))
}

// lights draws the macOS window buttons. Their colors are the platform's,
// not the theme's; symbols appear while the pointer is over the group.
func (v *TitleBarView) lights(cx *el.Context, w core.WindowControls) el.Element {
	id := autoID("titlebar", v) + "/lights"
	text := locale.Current()
	hover := cx.Hovered(id)
	light := func(name string, c color.NRGBA, symbol string, fn func()) el.Element {
		if !w.Focused() && !hover {
			c = theme.Border
		}
		dot := el.Div().Role("button").Name(name).Size(el.Dp(12)).Rounded(theme.RadiusMd).Bg(c).Center().
			Focusable(false).OnClick(fn)
		if hover {
			dot.Child(el.Text(symbol).TextSize(9).Bold().TextColor(color.NRGBA{A: 0x99}))
		}
		return dot
	}
	zoom := text.Maximize
	if w.Maximized() {
		zoom = text.Restore
	}
	return el.Div().ID(id).Row().Gap(8).Pl(6).Pr(6).Items(el.Center).Child(
		light(text.Close, color.NRGBA{R: 0xff, G: 0x5f, B: 0x57, A: 0xff}, "×", w.Close),
		light(text.Minimize, color.NRGBA{R: 0xfe, G: 0xbc, B: 0x2e, A: 0xff}, "−", w.Minimize),
		light(zoom, color.NRGBA{R: 0x28, G: 0xc8, B: 0x40, A: 0xff}, "+", w.ToggleMaximize),
	)
}

// buttons draws minimize, maximize / restore and close at the right edge.
func (v *TitleBarView) buttons(cx *el.Context, w core.WindowControls) el.Element {
	text := locale.Current()
	btn := func(name string, icon el.Element, hover color.NRGBA, fn func()) el.Element {
		return el.Div().Role("button").Name(name).W(el.Dp(46)).H(el.Dp(TitleBarHeight)).Center().
			CursorPointer().Focusable(false).OnClick(fn).Hover(func(s *el.Style) { s.Bg(hover) }).Child(icon)
	}
	line := func(wd, ht float32) el.Element { return el.Div().W(el.Dp(wd)).H(el.Dp(ht)).Bg(theme.Text) }
	square := el.Div().Size(el.Dp(10)).Border(1, theme.Text)
	maxName := text.Maximize
	if w.Maximized() {
		maxName = text.Restore
		square = el.Div().Size(el.Dp(10)).Child(
			el.Div().Absolute().Top(0).Left(2).Size(el.Dp(8)).Border(1, theme.Text),
			el.Div().Absolute().Top(2).Left(0).Size(el.Dp(8)).Border(1, theme.Text).Bg(theme.Subtle))
	}
	return el.Div().Row().Child(
		btn(text.Minimize, line(10, 1), theme.SubtleHover, w.Minimize),
		btn(maxName, square, theme.SubtleHover, w.ToggleMaximize),
		btn(text.Close, Icon(IconClose).Size(14).Color(theme.Text).Render(cx), theme.Danger, w.Close),
	)
}

func renderOptional(cx *el.Context, v el.View) el.Element {
	if v == nil {
		return el.Div().Hidden(true)
	}
	return v.Render(cx)
}
