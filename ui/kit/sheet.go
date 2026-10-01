package kit

import (
	"image"
	"time"

	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// SheetSlide is how long a sheet takes to slide in; reduced motion skips it.
const SheetSlide = 200 * time.Millisecond

// SheetView is a modal panel against one edge of the window: settings,
// details, a long form. It closes on Esc, a press on the scrim, or its close
// button, and returns focus where it was.
type SheetView struct {
	side     el.Side
	title    string
	body     el.View
	size     float32
	open     bool
	openedAt time.Time
	onClose  func()
}

// Sheet creates a sheet against side (el.Right, el.Left, el.Top, el.Bottom).
func Sheet(side el.Side, title string) *SheetView {
	return &SheetView{side: side, title: title, size: 360}
}
func (v *SheetView) Body(b el.View) *SheetView    { v.body = b; return v }
func (v *SheetView) OnClose(fn func()) *SheetView { v.onClose = fn; return v }
func (v *SheetView) SetTitle(s string)            { v.title = s }
func (v *SheetView) Value() bool                  { return v.open }
func (v *SheetView) SetValue(open bool)           { v.open = open }

// Size sets the width (left/right) or height (top/bottom) in dp, 360 by default.
func (v *SheetView) Size(dp float32) *SheetView {
	if dp > 0 {
		v.size = dp
	}
	return v
}

func (v *SheetView) close() {
	v.open = false
	if v.onClose != nil {
		v.onClose()
	}
}

func (v *SheetView) Render(cx *el.Context) el.Element {
	if !v.open {
		v.openedAt = time.Time{}
		return el.Div()
	}
	if v.openedAt.IsZero() {
		v.openedAt = cx.Now()
	}
	progress := float32(1)
	if d := cx.Now().Sub(v.openedAt); d < SheetSlide && !el.ReducedMotion() {
		progress = float32(d) / float32(SheetSlide)
		cx.Animating()
	}
	horizontal := v.side == el.Left || v.side == el.Right
	panel := el.Div().Role("dialog").Name(v.title).Bg(theme.Surface).P(20).Gap(16).Items(el.Stretch)
	if horizontal {
		panel.W(el.Dp(v.size)).MaxW(el.Full).H(el.Full)
	} else {
		panel.H(el.Dp(v.size)).MaxH(el.Full).W(el.Full)
	}
	header := el.Div().Row().Items(el.Center).Gap(8).Child(el.Text(v.title).TextSize(17).Bold().Grow())
	header.Child(Button("", v.close).Name("关闭").Icon(IconClose).Variant(ButtonGhost).Size(28).Render(cx))
	panel.Child(header)
	if v.body != nil {
		panel.Child(el.Div().Grow().ScrollY().Child(v.body.Render(cx)))
	}
	if progress < 1 {
		// Slide in from the edge: shift painting and hit areas together.
		remaining := 1 - progress
		panel.Decorate(func(gtx core.C, draw func()) {
			d := float32(gtx.Dp(1)) * v.size * remaining * remaining // ease out
			off := map[el.Side]image.Point{el.Right: {int(d), 0}, el.Left: {-int(d), 0}, el.Bottom: {0, int(d)}, el.Top: {0, -int(d)}}[v.side]
			defer op.Offset(off).Push(gtx.Ops).Pop()
			draw()
		})
	}
	cx.Overlay(autoID("sheet", v), el.Modal(panel).Placement(v.side, el.Start).OnDismiss(v.close))
	return el.Div()
}
