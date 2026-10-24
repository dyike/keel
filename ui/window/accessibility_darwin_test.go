//go:build darwin && !ios && !nometal

package window

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/dyike/keel/ui/internal/loop"
)

// A client reads a real window's elements through the content view, as
// VoiceOver does, finds them by position and presses a button. No
// accessibility permission is needed: the messages stay in this process.
func TestRealWindowsAccessibility(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 to open real windows")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0])
	cmd.Env = append(os.Environ(), "KEEL_ACCESSIBILITY_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "PASS") {
		t.Fatalf("%v\n%s", err, out)
	}
	t.Log(string(out))
}

// axNode is what a client sees of an accessibility element.
type axNode struct {
	obj                appkit.ID
	role, label, value string
	frame              appkit.Rect
	children           []axNode
}

func (n axNode) String() string {
	s := fmt.Sprintf("%s %q %q @%.0f,%.0f %.0fx%.0f", n.role, n.label, n.value, n.frame.Origin.X, n.frame.Origin.Y, n.frame.Size.Width, n.frame.Size.Height)
	for _, c := range n.children {
		s += "\n  " + strings.ReplaceAll(c.String(), "\n", "\n  ")
	}
	return s
}

func axRead(obj appkit.ID) axNode {
	n := axNode{obj: obj, role: appkit.GoString(appkit.Send(obj, "accessibilityRole")),
		label: appkit.GoString(appkit.Send(obj, "accessibilityLabel")),
		frame: appkit.MsgRect(obj, appkit.Sel("accessibilityFrame"))}
	if v := appkit.Send(obj, "accessibilityValue"); v != 0 {
		n.value = appkit.GoString(appkit.Send(v, "description"))
	}
	if kids := appkit.Send(obj, "accessibilityChildren"); kids != 0 {
		for i := range int(appkit.Send(kids, "count")) {
			n.children = append(n.children, axRead(appkit.Send(kids, "objectAtIndex:", uintptr(i))))
		}
	}
	return n
}

func (n axNode) find(role, text string) *axNode {
	if n.role == role && (n.label == text || n.value == text) {
		return &n
	}
	for _, c := range n.children {
		if f := c.find(role, text); f != nil {
			return f
		}
	}
	return nil
}

func accessibilityHelper() {
	fail := func(format string, args ...any) {
		fmt.Printf("FAIL: "+format+"\n", args...)
		os.Exit(1)
	}

	saved, picked := 0, -1 // frame lock
	content := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		label := "Save"
		if saved > 0 {
			label = fmt.Sprintf("Saved %d", saved)
		}
		list := el.Div().H(el.Dp(100)).ScrollY()
		for i := range 30 {
			list.Child(el.Div().P(4).OnClick(func() { picked = i }).Child(el.Text(fmt.Sprintf("Item %d", i))))
		}
		return el.Div().P(24).Child(
			el.Text("Hello"),
			el.Div().Role("group").P(16).Child(
				el.Div().P(8).OnClick(func() { saved++ }).Child(el.Text(label)),
			),
			list,
		)
	}))
	w := Open(Options{Title: "accessibility", Width: 360, Height: 340, Content: content})
	go func() {
		<-w.shown
		// Ask on the main queue, as AppKit's run loop does: Gio's Run may
		// run while the window goroutine lays out a frame.
		onView := func(fn func(view appkit.ID)) {
			loop.Lock()
			view := appkit.Retain(appkit.ID(w.nativeView))
			loop.Unlock()
			appkit.MainSync(func() {
				defer appkit.Release(view)
				fn(view)
			})
		}
		var root axNode
		var tree string
		poll := func(what string, ok func(view appkit.ID) bool) {
			deadline := time.Now().Add(10 * time.Second)
			for time.Now().Before(deadline) {
				done := false
				onView(func(view appkit.ID) {
					root = axRead(view)
					tree = root.String()
					done = ok(view)
				})
				if done {
					return
				}
				time.Sleep(50 * time.Millisecond)
			}
			fail("%s; tree:\n%s", what, tree)
		}
		poll("elements did not appear", func(view appkit.ID) bool {
			return root.role == "AXGroup" && root.find("AXButton", "Save") != nil && root.find("AXStaticText", "Hello") != nil
		})
		// Frames match where the view draws: the button is inset 16dp in
		// the group, which is inset 24dp in the window. Read them with the
		// view: the window may still be moving to the center.
		onView(func(view appkit.ID) {
			root = axRead(view)
			tree = root.String()
			var group, button *axNode
			for _, c := range root.children {
				if c.role == "AXGroup" {
					group = &c
				}
			}
			if group == nil || len(group.children) != 1 || group.children[0].role != "AXButton" {
				fail("the button is not inside its group:\n%s", tree)
			}
			button = &group.children[0]
			win := appkit.Send(view, "window")
			gb, bb := group.frame, button.frame
			if math.Abs(bb.Origin.X-gb.Origin.X-16) > 1 || math.Abs(gb.MaxY()-bb.MaxY()-16) > 1 {
				fail("button frame %v is not 16pt inside group frame %v", bb, gb)
			}
			viewOnScreen := appkit.MsgRectRect(win, appkit.Sel("convertRectToScreen:"), appkit.MsgRectToView(view, appkit.Sel("convertRect:toView:"), appkit.MsgRect(view, appkit.Sel("bounds")), 0))
			if math.Abs(gb.Origin.X-viewOnScreen.Origin.X-24) > 1 {
				fail("group frame %v is not 24pt inside the view %v", gb, viewOnScreen)
			}
			center := appkit.Point{X: bb.Origin.X + bb.Size.Width/2, Y: bb.Origin.Y + bb.Size.Height/2}
			if hit := appkit.MsgIDForPoint(view, appkit.Sel("accessibilityHitTest:"), center); hit != button.obj {
				fail("hit test at the button's center found %v, not the button %v", hit, button.obj)
			}
			if !appkit.SendBool(button.obj, "accessibilityPerformPress") {
				fail("press refused")
			}
		})
		poll("press did not click the button", func(appkit.ID) bool { return root.find("AXButton", "Saved 1") != nil })
		// The press also gave the button keyboard focus.
		poll("the pressed button is not the focused element", func(view appkit.ID) bool {
			f := appkit.Send(view, "accessibilityFocusedUIElement")
			saved := root.find("AXButton", "Saved 1")
			return saved != nil && f == saved.obj && appkit.SendBool(f, "isAccessibilityFocused")
		})
		// A client moves focus by setting it.
		var first appkit.ID
		onView(func(view appkit.ID) {
			item := axRead(view).find("AXButton", "Item 0")
			if item == nil {
				fail("no Item 0:\n%s", tree)
			}
			first = item.obj
			appkit.Send(first, "setAccessibilityFocused:", appkit.Bool(true))
		})
		poll("setting focus did not move it to Item 0", func(view appkit.ID) bool {
			return appkit.Send(view, "accessibilityFocusedUIElement") == first && appkit.SendBool(first, "isAccessibilityFocused")
		})
		// Frames follow the window when it moves without a new frame.
		onView(func(view appkit.ID) {
			win := appkit.Send(view, "window")
			at := appkit.MsgRect(win, appkit.Sel("frame")).Origin
			appkit.MsgSetPoint(win, appkit.Sel("setFrameOrigin:"), appkit.Point{X: at.X - 40, Y: at.Y - 30})
			text := axRead(view).find("AXStaticText", "Hello")
			viewOnScreen := appkit.MsgRectRect(win, appkit.Sel("convertRectToScreen:"), appkit.MsgRectToView(view, appkit.Sel("convertRect:toView:"), appkit.MsgRect(view, appkit.Sel("bounds")), 0))
			if text == nil || math.Abs(text.frame.Origin.X-viewOnScreen.Origin.X-24) > 1 || math.Abs(viewOnScreen.MaxY()-text.frame.MaxY()-24) > 1 {
				fail("after a move, text frame %v is not 24pt inside the view %v", text, viewOnScreen)
			}
		})
		// A row scrolled out of view is in its scroll area; scrolling it into
		// view moves it inside the area, and a press then reaches it.
		var item, area axNode
		onView(func(view appkit.ID) {
			root = axRead(view)
			tree = root.String()
			for _, c := range root.children {
				if c.role == "AXScrollArea" {
					area = c
				}
			}
			found := area.find("AXButton", "Item 25")
			if found == nil {
				fail("Item 25 is not in the scroll area:\n%s", tree)
			}
			item = *found
			if item.frame.MaxY() > area.frame.Origin.Y {
				fail("Item 25 at %v should be below the scroll area %v", item.frame, area.frame)
			}
			if hit := appkit.MsgIDForPoint(view, appkit.Sel("accessibilityHitTest:"), area.frame.Origin); hit == item.obj {
				fail("hit test found a row out of view")
			}
			appkit.Send(item.obj, "accessibilityPerformAction:", uintptr(appkit.String("AXScrollToVisible")))
		})
		inside := func(r, a appkit.Rect) bool {
			return r.Origin.Y >= a.Origin.Y-1 && r.MaxY() <= a.MaxY()+1
		}
		poll("Item 25 did not scroll into view", func(view appkit.ID) bool {
			f := appkit.MsgRect(item.obj, appkit.Sel("accessibilityFrame"))
			return inside(f, area.frame)
		})
		onView(func(appkit.ID) {
			if !appkit.SendBool(item.obj, "accessibilityPerformPress") {
				fail("press on Item 25 refused")
			}
		})
		deadline := time.Now().Add(5 * time.Second)
		for {
			loop.Lock()
			p := picked
			loop.Unlock()
			if p == 25 {
				break
			}
			if time.Now().After(deadline) {
				fail("press after scrolling picked %d, not Item 25", p)
			}
			time.Sleep(20 * time.Millisecond)
		}
		fmt.Println("PASS: roles, labels, nesting, frames (also after a move), hit testing, press, focus (reported and set), rows out of view and scrolling")
		w.Close()
		os.Exit(0)
	}()
	Main()
}
