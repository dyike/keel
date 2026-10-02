package kit

import (
	"image"
	"time"

	"github.com/dyike/keel/ui/locale"

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
	disabled bool
	openedAt time.Time
	onClose  func()
}

// Sheet creates a sheet against side (el.Right, el.Left, el.Top, el.Bottom).
func Sheet(side el.Side, title string) *SheetView {
	if side > el.Right {
		side = el.Right
	}
	return &SheetView{side: side, title: title, size: 360}
}
func (v *SheetView) Body(b el.View) *SheetView    { v.body = b; return v }
func (v *SheetView) OnClose(fn func()) *SheetView { v.onClose = fn; return v }
func (v *SheetView) SetTitle(s string)            { v.title = s }
func (v *SheetView) Value() bool                  { return v.open }
func (v *SheetView) SetValue(open bool) {
	v.open = open && !v.disabled
	if !v.open {
		v.openedAt = time.Time{}
	}
}

// Size sets the width (left/right) or height (top/bottom) in dp, 360 by default.
func (v *SheetView) Size(dp float32) *SheetView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.size = dp
	}
	return v
}

func (v *SheetView) close() {
	if !v.open {
		return
	}
	v.SetValue(false)
	if v.onClose != nil {
		v.onClose()
	}
}

func (v *SheetView) Render(cx *el.Context) el.Element {
	if !v.open {
		v.openedAt = time.Time{}
		return el.Div().Hidden(true)
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
	panel := el.Div().Role("dialog").Name(v.title).Bg(theme.Surface).Shadow(theme.ElevationLg).P(20).Gap(theme.SpaceXl).Items(el.Stretch)
	if horizontal {
		panel.W(el.Dp(v.size)).MaxW(el.Full).H(el.Full)
	} else {
		panel.H(el.Dp(v.size)).MaxH(el.Full).W(el.Full)
	}
	header := el.Div().Row().Items(el.Center).Gap(theme.SpaceMd).Child(el.Text(v.title).TextSize(theme.TextLg).Bold().Grow())
	header.Child(Button("", v.close).Name(locale.Current().Close).Icon(IconClose).Variant(ButtonGhost).Size(28).Render(cx))
	panel.Child(header)
	id := autoID("sheet", v)
	cx.Overlay(id, el.Modal(panel).Owner(id).Placement(v.side, el.Start).OnDismiss(v.close))
	if v.body != nil {
		panel.Child(el.Div().ID(id + "/body").Grow().MinH(el.Dp(0)).ScrollY().Child(v.body.Render(cx)))
	}
	if progress < 1 {
		// Slide in from the edge: shift painting and hit areas together.
		remaining := 1 - progress
		panel.Decorate(func(gtx core.C, draw func()) {
			w, h := cx.LayoutSize(panel)
			size := w
			if !horizontal {
				size = h
			}
			scale := gtx.Metric.PxPerDp
			if scale <= 0 {
				scale = 1
			}
			d := scale * size * remaining * remaining // ease out using the fitted size
			off := map[el.Side]image.Point{el.Right: {int(d), 0}, el.Left: {-int(d), 0}, el.Bottom: {0, int(d)}, el.Top: {0, -int(d)}}[v.side]
			defer op.Offset(off).Push(gtx.Ops).Pop()
			draw()
		})
	}
	return el.Div().ID(id).Absolute().Size(el.Dp(0)).Disabled(v.disabled)
}

func (v *SheetView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.SetValue(false)
	}
}
