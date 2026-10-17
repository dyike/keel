//go:build darwin && !ios && keelnativeqa

package window

import (
	"fmt"

	"github.com/dyike/keel/ui/internal/appkit"
)

// NativeQAFocusState describes this process's native window, without using an
// automation shadow or requiring Accessibility/Input Monitoring permission.
type NativeQAFocusState struct {
	Visible, Active, Key, InputResponder bool
	FramePresented, MaterialInstalled    bool
	Level, ActivationPolicy              int
	KeyWindow                            string
}

func qaWindowNamed(title string) appkit.ID {
	var candidate appkit.ID
	windows := appkit.Send(appkit.App(), "windows")
	for i := uintptr(0); i < uintptr(appkit.Send(windows, "count")); i++ {
		w := appkit.Send(windows, "objectAtIndex:", i)
		if appkit.Equal(appkit.Send(w, "title"), title) {
			if appkit.SendBool(w, "isKeyWindow") {
				return w
			}
			if candidate == 0 || appkit.SendBool(w, "isVisible") {
				candidate = w
			}
		}
	}
	return candidate
}

// NativeQAWindowFocus must be called outside the frame lock.
func NativeQAWindowFocus(title string) NativeQAFocusState {
	var state NativeQAFocusState
	appkit.MainSync(func() {
		app := appkit.App()
		w := qaWindowNamed(title)
		view := appkit.ID(appkit.Call(appkit.Sym("objc_getAssociatedObject"), uintptr(w), uintptr(appkit.Sel("gioInputView"))))
		state = NativeQAFocusState{
			Visible: appkit.SendBool(w, "isVisible"), Active: appkit.SendBool(app, "isActive"), Key: appkit.SendBool(w, "isKeyWindow"),
			InputResponder: view != 0 && appkit.Send(w, "firstResponder") == view,
			Level:          int(appkit.Send(w, "level")), ActivationPolicy: int(appkit.Send(app, "activationPolicy")),
			KeyWindow: appkit.GoString(appkit.Send(appkit.Send(app, "keyWindow"), "title")),
		}
		if value, ok := nativeWindows.Load(uintptr(view)); ok {
			tracked := value.(*Window)
			select {
			case <-tracked.shown:
				state.FramePresented = true
			default:
			}
			state.MaterialInstalled = tracked.opts.Glass == nil || appkit.SendBool(appkit.Send(w, "contentView"), "isKindOfClass:", uintptr(appkit.Class("NSVisualEffectView")))
		}
	})
	return state
}

// NativeQAKey sends a real AppKit key event through this application's own
// event dispatch, including the native menu key-equivalent path.
func NativeQAKey(title, text string, command bool, keyCode uint16) error {
	var err error
	appkit.MainSync(func() {
		app, w := appkit.App(), qaWindowNamed(title)
		if w == 0 || !appkit.SendBool(app, "isActive") || !appkit.SendBool(w, "isKeyWindow") {
			err = fmt.Errorf("native key target %q is not active and key", title)
			return
		}
		flags := uint(0)
		if command {
			flags = nsEventModifierFlagCommand
		}
		chars := appkit.String(text)
		stamp := appkit.MsgFloat(appkit.Send(appkit.Class("NSProcessInfo"), "processInfo"), appkit.Sel("systemUptime"))
		for _, typ := range []uint{10, 11} {
			event := appkit.MsgKeyEvent(appkit.Class("NSEvent"), appkit.Sel("keyEventWithType:location:modifierFlags:timestamp:windowNumber:context:characters:charactersIgnoringModifiers:isARepeat:keyCode:"),
				typ, appkit.Point{}, flags, stamp, int(appkit.Send(w, "windowNumber")), 0, chars, chars, false, keyCode)
			appkit.Send(app, "sendEvent:", uintptr(event))
		}
	})
	return err
}
