package kit

import (
	"slices"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
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
	title                                string
	body                                 el.View
	footer                               []el.View
	width                                float32
	alert                                bool
	open                                 bool
	disabled                             bool
	opened                               bool // the open state the last Render saw, to focus on opening
	focus                                string
	onClose                              func()
	keyboardOff, overlayOff, closeButton bool
	overlayClosable                      *bool
	beforeConfirm                        func() bool
	beforeCancel                         func() bool
	canceling                            bool
	confirming                           bool
	generation                           uint64
	onCancel                             func() // message dialogs: run on Esc, scrim or the cancel button
}

func Dialog(title string) *DialogView { return &DialogView{title: title, width: 420} }

func (v *DialogView) Body(b el.View) *DialogView          { v.body = b; return v }
func (v *DialogView) Footer(views ...el.View) *DialogView { v.footer = slices.Clone(views); return v }
func (v *DialogView) OnClose(fn func()) *DialogView       { v.onClose = fn; return v }
func (v *DialogView) SetTitle(s string)                   { v.title = s }
func (v *DialogView) Value() bool                         { return v.open }
func (v *DialogView) SetValue(open bool) {
	v.generation++
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

// Keyboard controls Esc dismissal, enabled by default.
func (v *DialogView) Keyboard(on bool) *DialogView { v.keyboardOff = !on; return v }

// Overlay controls scrim painting, not modality or outside-click behavior.
func (v *DialogView) Overlay(on bool) *DialogView { v.overlayOff = !on; return v }

// OverlayClosable overrides outside-click dismissal, including Persistent.
func (v *DialogView) OverlayClosable(on bool) *DialogView { v.overlayClosable = &on; return v }

// CloseButton shows an explicit header close action, hidden by default.
func (v *DialogView) CloseButton(on bool) *DialogView { v.closeButton = on; return v }

// Persistent makes it an alertdialog: a press on the scrim does not close it;
// Esc still does. ConfirmDanger dialogs are persistent.
func (v *DialogView) Persistent() *DialogView { v.alert = true; return v }

// close is a user dismissal: Esc, the scrim, or a cancel button.
func (v *DialogView) close() {
	if v.canCancel() {
		v.closeAccepted()
	}
}

// BeforeCancel may reject user dismissal; nil removes the guard. It does not
// prevent programmatic closing or cleanup when the owner disappears.
func (v *DialogView) BeforeCancel(fn func() bool) *DialogView { v.beforeCancel = fn; return v }

func (v *DialogView) canCancel() bool {
	if !v.open || v.canceling {
		return false
	}
	if v.beforeCancel == nil {
		return true
	}
	generation := v.generation
	v.canceling = true
	defer func() { v.canceling = false }()
	allowed := v.beforeCancel()
	return allowed && v.open && generation == v.generation
}

func (v *DialogView) closeAccepted() {
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

// BeforeConfirm runs before a standard message's OK action. False keeps the
// dialog open and skips onOK. Nil removes the guard. It persists across message
// reuse; custom Footer actions remain owned by the application.
func (v *DialogView) BeforeConfirm(fn func() bool) *DialogView {
	v.beforeConfirm = fn
	return v
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
	v.generation++
	v.opened = false
	id := autoID("dialog", v)
	v.title, v.alert, v.open, v.onCancel = title, variant == ButtonDanger, true, nil
	v.body = el.ViewFunc(func(*el.Context) el.Element { return el.Text(message).TextColor(theme.Muted) })
	ok := Button(okText, func() {
		if !v.open || v.disabled || v.confirming {
			return
		}
		generation := v.generation
		if v.beforeConfirm != nil {
			allowed := func() bool {
				v.confirming = true
				defer func() { v.confirming = false }()
				return v.beforeConfirm()
			}()
			if !allowed || !v.open || v.disabled || generation != v.generation {
				return
			}
		}
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
	panel := floating(theme.ElevationLg).Rounded(theme.RadiusXl).Role(role).Name(v.title).W(el.Dp(v.width)).MaxW(el.Dp(max(0, w-16))).MaxH(el.Dp(max(0, h-16))).ScrollY().P(20).Gap(theme.SpaceXl).Items(el.Stretch)
	layer := el.Modal(panel).Owner(id).BeforeDismiss(v.canCancel).OnDismiss(v.closeAccepted).Scrim(!v.overlayOff)
	if v.keyboardOff {
		layer.KeepOnEscape()
	}
	outside := !v.alert
	if v.overlayClosable != nil {
		outside = *v.overlayClosable
	}
	if !outside {
		layer.KeepOnOutsidePress()
	}
	cx.Overlay(id, layer)
	if v.title != "" || v.closeButton {
		header := el.Div().ID(id + "/header").Row().Items(el.Center).Gap(theme.SpaceMd)
		header.Child(el.Text(v.title).TextSize(theme.TextLg).Bold().Grow())
		if v.closeButton {
			header.Child(Button("", v.close).ID(id + "/close").Name(locale.Current().Close).Icon(IconClose).Variant(ButtonGhost).Size(28).Render(cx))
		}
		panel.Child(header)
	}
	if v.body != nil {
		panel.Child(el.Div().ID(id + "/body").MaxH(el.Dp(max(0, h-160))).ScrollY().Items(el.Stretch).Child(v.body.Render(cx)))
	}
	if len(v.footer) > 0 {
		row := el.Div().ID(id + "/footer").Gap(theme.SpaceMd).Justify(el.End)
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
