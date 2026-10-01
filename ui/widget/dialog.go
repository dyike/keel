package widget

import (
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/locale"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
)

// DialogBox shows a modal message over the whole window: the window dims and
// ignores clicks until a button is pressed. Enter confirms, Esc cancels.
//
// Put one in window.Options.Overlay, then call Confirm or Alert from any
// callback:
//
//	dlg := widget.Dialog()
//	window.Open(window.Options{Content: page, Overlay: dlg})
//	widget.Button("删除", func() { dlg.Confirm("删除订单", "确定删除？", remove) })
type DialogBox struct {
	open           bool
	title, message string
	onOK, onCancel func()
	ok, cancel     *Btn
	showCancel     bool
}

func Dialog() *DialogBox {
	d := &DialogBox{}
	d.ok = Button("", func() { d.finish(d.onOK) })
	d.cancel = Button("", func() { d.finish(d.onCancel) }).Secondary()
	return d
}

// Confirm asks a question with OK and Cancel; onOK runs only on OK.
func (d *DialogBox) Confirm(title, message string, onOK func()) {
	d.show(title, message, true, onOK, nil)
}

// ConfirmDanger is Confirm with a red OK button, for destructive actions.
func (d *DialogBox) ConfirmDanger(title, message, okText string, onOK func()) {
	d.show(title, message, true, onOK, nil)
	d.ok.SetText(okText)
	d.ok.Danger()
}

// Alert shows a message with a single OK button; onOK may be nil.
func (d *DialogBox) Alert(title, message string, onOK func()) {
	d.show(title, message, false, onOK, nil)
}

// OnCancel sets what runs when the dialog is cancelled (button or Esc).
func (d *DialogBox) OnCancel(fn func()) *DialogBox { d.onCancel = fn; return d }

func (d *DialogBox) IsOpen() bool { return d.open }

// Close hides the dialog without running either callback.
func (d *DialogBox) Close() { d.open = false }

func (d *DialogBox) show(title, message string, cancel bool, onOK, onCancel func()) {
	d.title, d.message, d.showCancel, d.onOK = title, message, cancel, onOK
	if onCancel != nil {
		d.onCancel = onCancel
	}
	d.ok.SetText(locale.Current().OK)
	d.cancel.SetText(locale.Current().Cancel)
	d.ok.kind = primary
	d.open = true
}

func (d *DialogBox) finish(fn func()) {
	d.open = false
	if fn != nil {
		fn()
	}
}

func (d *DialogBox) Layout(gtx C) D {
	size := gtx.Constraints.Max
	if !d.open {
		return D{}
	}
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameReturn}, key.Filter{Name: key.NameEnter}, key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if e, ok := ev.(key.Event); ok && e.State == key.Press {
			if e.Name == key.NameEscape {
				if d.showCancel {
					core.Call(gtx, func() { d.finish(d.onCancel) })
				}
			} else {
				core.Call(gtx, func() { d.finish(d.onOK) })
			}
		}
	}
	// Scrim: dims the window and swallows pointer input meant for it.
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	paint.ColorOp{Color: theme.Scrim}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	event.Op(gtx.Ops, d)
	for {
		if _, ok := gtx.Event(pointer.Filter{Target: d, Kinds: pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Scroll,
			ScrollY: pointer.ScrollRange{Min: -1e6, Max: 1e6}}); !ok {
			break
		}
	}
	area.Pop()

	gtx.Constraints = giolayout.Exact(size)
	width := min(gtx.Dp(420), size.X-gtx.Dp(48))
	return giolayout.Center.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
		return core.Semantic(gtx, d.card, core.Role("dialog"), semantic.LabelOp(d.title))
	})
}

func (d *DialogBox) card(gtx C) D {
	return layout.Frame(gtx, theme.Surface, theme.Border, 10, giolayout.UniformInset(20), func(gtx C) D {
		buttons := []core.Widget{layout.Grow(layout.Space(0))}
		if d.showCancel {
			buttons = append(buttons, d.cancel)
		}
		buttons = append(buttons, d.ok)
		return layout.Column(
			core.Func(func(gtx C) D {
				lb := material.Label(theme.Material, 17, d.title)
				lb.Color, lb.Font.Weight = theme.Text, font.Bold
				return core.Semantic(gtx, lb.Layout, semantic.LabelOp(d.title))
			}),
			Text(d.message),
			layout.Space(4),
			layout.Row(buttons...),
		).Layout(gtx)
	})
}
