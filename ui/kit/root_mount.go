package kit

import "github.com/dyike/keel/ui/el"

// Window-level overlays, as GPUI's root view hosts them: Show and
// WindowNotifier attach to the window's root (el.Context.Mount), so the
// dialog, sheet or notifier need not be rendered in the view tree. Use
// either Show or Render for one instance, not both.

// unmountWhenClosed renders an overlay and unmounts it once it is closed.
type unmountWhenClosed struct {
	key    any
	view   el.View
	isOpen func() bool
}

func (u unmountWhenClosed) Render(cx *el.Context) el.Element {
	e := u.view.Render(cx)
	if !u.isOpen() {
		cx.Unmount(u.key)
	}
	return e
}

// Show opens the dialog in the window of cx without placing it in a view
// tree; it leaves the window when closed.
func (v *DialogView) Show(cx *el.Context) {
	v.SetValue(true)
	if v.open {
		cx.Mount(v, unmountWhenClosed{v, v, v.Value})
	}
}

// Show opens the sheet in the window of cx without placing it in a view
// tree; it leaves the window when closed.
func (v *SheetView) Show(cx *el.Context) {
	v.SetValue(true)
	if v.open {
		cx.Mount(v, unmountWhenClosed{v, v, v.Value})
	}
}

type windowNotifierKey struct{}

// WindowNotifier is the window's own Notifier, created and mounted on first
// use: kit.WindowNotifier(cx).Notify(...) needs no Notifier in the view
// tree. Configure it like any Notifier.
func WindowNotifier(cx *el.Context) *NotifierView {
	if n, ok := cx.MountedView(windowNotifierKey{}).(*NotifierView); ok {
		return n
	}
	n := Notifier()
	cx.Mount(windowNotifierKey{}, n)
	return n
}
