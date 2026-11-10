//go:build darwin && !ios

package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/window"
)

// Inspect the actual standard NSButtons, not Keel's off-screen semantics.
func main() {
	time.AfterFunc(20*time.Second, func() { fmt.Println("FAIL: native chrome timeout"); os.Exit(1) })
	shell := kit.AppShell("App").SidebarHeader(kit.Button("Action", nil))
	w := window.Open(window.Options{Title: "chrome acceptance", Width: 800, Height: 500, Frameless: true, NativeTrafficLights: true, TrafficLightLayout: &window.TrafficLightLayout{Height: 56, Left: 12, Spacing: 20}, Content: el.Root(shell)})
	go func() {
		var buttons [3]appkit.ID
		var actions [3]appkit.ID
		var centers [3]appkit.Point
		reason := "window not ready"
		check := func(capture bool) bool {
			ok := false
			appkit.MainSync(func() {
				appkit.Pool(func() {
					windows := appkit.Send(appkit.App(), "windows")
					for i := uintptr(0); i < uintptr(appkit.Send(windows, "count")); i++ {
						win := appkit.Send(windows, "objectAtIndex:", i)
						if !appkit.Equal(appkit.Send(win, "title"), "chrome acceptance") {
							continue
						}
						view := appkit.ID(appkit.Call(appkit.Sym("objc_getAssociatedObject"), uintptr(win), uintptr(appkit.Sel("gioInputView"))))
						if view == 0 {
							reason = "missing Gio input view"
							return
						}
						for n := range buttons {
							b := appkit.Send(win, "standardWindowButton:", uintptr(n))
							reason = fmt.Sprintf("button %d id=%v hidden=%v window=%v want=%v action=%v", n, b, appkit.SendBool(b, "isHidden"), appkit.Send(b, "window"), win, appkit.Send(b, "action"))
							if b == 0 || !appkit.SendBool(b, "isKindOfClass:", uintptr(appkit.Class("NSButton"))) || appkit.SendBool(b, "isHidden") || appkit.Send(b, "window") != win || appkit.Send(b, "action") == 0 {
								return
							}
							reason = fmt.Sprintf("button %d identity/action", n)
							bounds := appkit.MsgRect(b, appkit.Sel("bounds"))
							p := appkit.MsgPointFromView(b, appkit.Sel("convertPoint:toView:"), appkit.Point{X: bounds.Size.Width / 2, Y: bounds.Size.Height / 2}, view)
							if !appkit.SendBool(view, "isFlipped") {
								p.Y = appkit.MsgRect(view, appkit.Sel("bounds")).Size.Height - p.Y
							}
							if math.Abs(p.Y-28) > 1 {
								reason = fmt.Sprintf("button %d center %+v", n, p)
								return
							}
							if capture {
								buttons[n], actions[n], centers[n] = b, appkit.Send(b, "action"), p
							} else if buttons[n] != b || actions[n] != appkit.Send(b, "action") || centers[n] != p {
								reason = fmt.Sprintf("button %d changed: center %+v -> %+v", n, centers[n], p)
								return
							}
						}
						root := appkit.Send(appkit.Send(win, "contentView"), "superview")
						for _, x := range []float64{100, 300, 700} {
							p := appkit.Point{X: x, Y: 28}
							if !appkit.SendBool(view, "isFlipped") {
								p.Y = appkit.MsgRect(view, appkit.Sel("bounds")).Size.Height - p.Y
							}
							p = appkit.MsgPointFromView(view, appkit.Sel("convertPoint:toView:"), p, appkit.Send(root, "superview"))
							hit := appkit.MsgIDForPoint(root, appkit.Sel("hitTest:"), p)
							if hit != view && !appkit.SendBool(hit, "isDescendantOf:", uintptr(view)) {
								reason = fmt.Sprintf("hit at %v: %s", x, appkit.GoString(appkit.Send(hit, "description")))
								return
							}
						}
						ok = true
					}
				})
			})
			return ok
		}
		wait := func(capture bool) {
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				if check(capture) {
					return
				}
				time.Sleep(30 * time.Millisecond)
			}
			fmt.Println("FAIL: native chrome", capture, reason)
			os.Exit(1)
		}
		wait(true)
		core.Update(func() { shell.SetCollapsed(true) })
		time.Sleep(100 * time.Millisecond)
		wait(false)
		for _, size := range [][2]int{{900, 600}, {800, 500}, {1000, 640}} {
			if err := w.Resize(size[0], size[1]); err != nil {
				panic(err)
			}
			time.Sleep(100 * time.Millisecond)
			wait(false)
		}
		fmt.Println("PASS: original AppKit buttons/actions, fixed centers after collapse/resize, Gio toolbar hit testing")
		w.Close()
	}()
	window.Main()
}
