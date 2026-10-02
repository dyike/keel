package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"slices"
)

// DialogView is a modal dialog: the page dims, ignores the pointer and loses
// keyboard focus until the dialog closes; focus then returns where it was.
//
// Build a custom one with Body and Footer:
//
//	edit := kit.Dialog("编辑订单").Body(form).Footer(cancel, save)
//	edit.SetValue(true)
//
// or reuse one instance for standard messages:
//
//	dlg := kit.Dialog("")
//	dlg.ConfirmDanger("删除订单", "确定删除？", "删除", remove)
//
// Render it in the view tree; it renders nothing in place.
type DialogView struct {
	title    string
	body     el.View
	footer   []el.View
	width    float32
	alert    bool
	open     bool
	disabled bool
	opened   bool // the open state the last Render saw, to focus on opening
	focus    string
	onClose  func()
	onCancel func() // message dialogs: run on Esc, scrim or the cancel button
}

func Dialog(title string) *DialogView { return &DialogView{title: title, width: 420} }

func (v *DialogView) Body(b el.View) *DialogView          { v.body = b; return v }
func (v *DialogView) Footer(views ...el.View) *DialogView { v.footer = slices.Clone(views); return v }
func (v *DialogView) OnClose(fn func()) *DialogView       { v.onClose = fn; return v }
func (v *DialogView) SetTitle(s string)                   { v.title = s }
func (v *DialogView) Value() bool                         { return v.open }
func (v *DialogView) SetValue(open bool) {
	v.open = open && !v.disabled
	if !v.open {
		v.opened = false
	}
}

// Width sets the dialog width in dp, 420 by default.
func (v *DialogView) Width(dp float32) *DialogView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.width = dp
	}
	return v
}

// Persistent makes it an alertdialog: a press on the scrim does not close it;
// Esc still does. ConfirmDanger dialogs are persistent.
func (v *DialogView) Persistent() *DialogView { v.alert = true; return v }

// close is a user dismissal: Esc, the scrim, or a cancel button.
func (v *DialogView) close() {
	if !v.open {
		return
	}
	v.open = false
	cancel := v.onCancel
	v.onCancel = nil
	if cancel != nil {
		cancel()
	}
	if v.onClose != nil {
		v.onClose()
	}
}

// Confirm asks a question with 取消 and 确定; onOK runs only on 确定.
// Focus starts on 确定, so Enter confirms.
func (v *DialogView) Confirm(title, message string, onOK func()) {
	v.message(title, message, locale.Current().OK, ButtonPrimary, true, onOK)
}

// ConfirmDanger asks before a destructive action. Focus starts on 取消, a
// press on the scrim is ignored, and Esc cancels.
func (v *DialogView) ConfirmDanger(title, message, okText string, onOK func()) {
	v.message(title, message, okText, ButtonDanger, true, onOK)
}

// Alert shows a message with a single 确定; onOK may be nil.
func (v *DialogView) Alert(title, message string, onOK func()) {
	v.message(title, message, locale.Current().OK, ButtonPrimary, false, onOK)
}

func (v *DialogView) message(title, message, okText string, variant ButtonVariant, cancel bool, onOK func()) {
	if v.disabled {
		return
	}
	v.opened = false
	id := autoID("dialog", v)
	v.title, v.alert, v.open, v.onCancel = title, variant == ButtonDanger, true, nil
	v.body = el.ViewFunc(func(*el.Context) el.Element { return el.Text(message).TextColor(theme.Muted) })
	ok := Button(okText, func() {
		v.open, v.onCancel = false, nil
		if onOK != nil {
			onOK()
		}
	}).ID(id + "/ok").Variant(variant)
	v.footer, v.focus = []el.View{ok}, id+"/ok"
	if cancel {
		v.footer = []el.View{Button(locale.Current().Cancel, v.close).ID(id + "/cancel").Variant(ButtonSecondary), ok}
		if variant == ButtonDanger {
			v.focus = id + "/cancel"
		}
	}
}

func (v *DialogView) Render(cx *el.Context) el.Element {
	if !v.open {
		v.opened = false
		return el.Div().Hidden(true)
	}
	role := "dialog"
	if v.alert {
		role = "alertdialog"
	}
	id := autoID("dialog", v)
	w, h := cx.ViewportSize()
	panel := surface().Role(role).Name(v.title).W(el.Dp(v.width)).MaxW(el.Dp(max(0, w-16))).MaxH(el.Dp(max(0, h-16))).ScrollY().P(20).Gap(16).Items(el.Stretch)
	layer := el.Modal(panel).Owner(id).OnDismiss(v.close)
	if v.alert {
		layer.KeepOnOutsidePress()
	}
	cx.Overlay(id, layer)
	if v.title != "" {
		panel.Child(el.Text(v.title).TextSize(17).Bold())
	}
	if v.body != nil {
		panel.Child(el.Div().ID(id + "/body").MaxH(el.Dp(max(0, h-160))).ScrollY().Items(el.Stretch).Child(v.body.Render(cx)))
	}
	if len(v.footer) > 0 {
		row := el.Div().Gap(8).Justify(el.End)
		if min(w-16, v.width) >= 360 {
			row.Row()
		}
		for _, f := range v.footer {
			if f != nil {
				row.Child(f.Render(cx))
			}
		}
		panel.Child(row)
	}
	if !v.opened && v.focus != "" {
		cx.Focus(v.focus)
	}
	v.opened = true
	return el.Div().ID(id).Absolute().Size(el.Dp(0)).Disabled(v.disabled)
}

func (v *DialogView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.SetValue(false)
		v.onCancel = nil
	}
}
