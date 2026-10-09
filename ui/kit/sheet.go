package kit

import (
	"image"
	"slices"
	"time"

	"github.com/dyike/keel/ui/locale"

	"github.com/dyike/keel/third_party/gio/op"
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
	side                                                el.Side
	title                                               string
	body                                                el.View
	size                                                float32
	marginTop                                           float32
	panelStyle                                          func(*el.DivEl)
	resizeOff                                           bool
	resizing                                            bool
	resizeGrab, resizeStart, resizePainted, resizeLimit float32
	onResize                                            func(float32)
	open                                                bool
	disabled                                            bool
	openedAt                                            time.Time
	onClose                                             func()
	footer                                              []el.View
	keyboardOff, overlayOff, outsideOff, closeOff       bool
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
		v.resizing = false
	}
}

// Footer adds fixed actions below the scrollable body and copies the slice.
func (v *SheetView) Footer(views ...el.View) *SheetView { v.footer = slices.Clone(views); return v }
func (v *SheetView) Keyboard(on bool) *SheetView        { v.keyboardOff = !on; return v }
func (v *SheetView) Overlay(on bool) *SheetView         { v.overlayOff = !on; return v }
func (v *SheetView) OverlayClosable(on bool) *SheetView { v.outsideOff = !on; return v }
func (v *SheetView) CloseButton(on bool) *SheetView     { v.closeOff = !on; return v }

// Size sets the width (left/right) or height (top/bottom) in dp, 360 by default.
func (v *SheetView) Size(dp float32) *SheetView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.resizing = false
		v.size = dp
	}
	return v
}

// PanelStyle refines the panel's default colors, border, padding and gap each
// frame. Do not retain the element. Size, bounds and dialog identity are applied
// afterward; use Body and Footer for content. Nil restores the default style.
func (v *SheetView) PanelStyle(fn func(*el.DivEl)) *SheetView {
	v.panelStyle = fn
	return v
}

// MarginTop reserves space above the panel in dp; zero restores full height.
// Invalid values are ignored. The modal scrim still covers the whole window.
func (v *SheetView) MarginTop(dp float32) *SheetView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.marginTop = dp
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
	id := autoID("sheet", v)
	panel := el.Div().Bg(theme.Surface).Shadow(theme.ElevationLg).P(20).Gap(theme.SpaceXl).Items(el.Stretch)
	if v.panelStyle != nil {
		v.panelStyle(panel)
	}
	panel.ID(id + "/panel").Role("dialog").Name(v.title)
	if horizontal {
		panel.W(el.Dp(v.size)).MaxW(el.Full).H(el.Full)
	} else {
		panel.H(el.Dp(v.size)).MaxH(el.Full).W(el.Full)
	}
	header := el.Div().ID(id + "/header").NoShrink().Row().Items(el.Center).Gap(theme.SpaceMd).Child(el.Text(v.title).TextSize(theme.TextLg).Bold().Grow())
	if !v.closeOff {
		header.Child(Button("", v.close).ID(id + "/close").Name(locale.Current().Close).Icon(IconClose).Variant(ButtonGhost).Size(28).Render(cx))
	}
	panel.Child(header)
	layer := el.Modal(panel).Owner(id).Placement(v.side, el.Start).TopInset(v.marginTop).OnDismiss(v.close).Scrim(!v.overlayOff)
	if v.keyboardOff {
		layer.KeepOnEscape()
	}
	if v.outsideOff {
		layer.KeepOnOutsidePress()
	}
	cx.Overlay(id, layer)
	if v.body != nil {
		panel.Child(el.Div().ID(id + "/body").Grow().MinH(el.Dp(0)).ScrollY().Child(v.body.Render(cx)))
	}
	if len(v.footer) > 0 {
		if v.body == nil {
			panel.Child(el.Div().ID(id + "/spacer").Grow())
		}
		footer := el.Div().ID(id + "/footer").NoShrink().Row().Wrap().Gap(theme.SpaceMd).Justify(el.End)
		for _, view := range v.footer {
			if view != nil {
				footer.Child(view.Render(cx))
			}
		}
		panel.Child(footer)
	}
	if !v.resizeOff {
		panel.Child(v.resizeHandle(id, cx))
	}
	panel.Decorate(func(gtx core.C, draw func()) {
		w, h := cx.LayoutSize(panel)
		v.resizePainted = w
		if !horizontal {
			v.resizePainted = h
		}
		if progress >= 1 {
			draw()
			return
		}
		// Slide in from the edge: shift painting and hit areas together.
		remaining := 1 - progress
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
	return el.Div().ID(id).Absolute().Size(el.Dp(0)).Disabled(v.disabled)
}

func (v *SheetView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.SetValue(false)
	}
}
