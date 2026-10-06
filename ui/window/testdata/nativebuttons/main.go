//go:build darwin && !ios && cgo

package main

/*
#cgo CFLAGS: -x objective-c -fblocks
#cgo LDFLAGS: -framework AppKit
int check_native_buttons(int action, int report, double height, double left, double offsetY, double spacing);
*/
import "C"

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/window"
	"os"
	"time"
)

func main() {
	w := window.Open(window.Options{TrafficLightLayout: &window.TrafficLightLayout{Height: 44, Left: 15, Spacing: 23}, Title: "Native buttons acceptance", Width: 640, Height: 320, Frameless: true, NativeTrafficLights: true, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(el.Text("AppKit standard window buttons")) })), OnClose: func() {
		fmt.Println("PASS: AppKit buttons centered at 36/44/64dp; runtime layout, hit testing, resize, minimize and close work")
	}})
	go func() {
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			if C.check_native_buttons(0, 0, 44, 15, 0, 23) == 1 {
				C.check_native_buttons(1, 0, 44, 15, 0, 23) // Resize: Gio reconfigures the decorations.
				for range 10 {
					core.Update(func() {})
					time.Sleep(50 * time.Millisecond)
				}
				if C.check_native_buttons(0, 1, 44, 15, 0, 23) != 1 {
					os.Exit(1)
				}
				w.SetTrafficLightLayout(window.TrafficLightLayout{Height: 64, Left: 20, OffsetY: 2, Spacing: 22})
				for range 10 {
					core.Update(func() {})
					time.Sleep(50 * time.Millisecond)
				}
				if C.check_native_buttons(0, 1, 64, 20, 2, 22) != 1 {
					os.Exit(1)
				}
				w.SetTrafficLightLayout(window.TrafficLightLayout{Height: 36, Left: 12, OffsetY: -1, Spacing: 22})
				for range 10 {
					core.Update(func() {})
					time.Sleep(50 * time.Millisecond)
				}
				if C.check_native_buttons(0, 1, 36, 12, -1, 22) != 1 {
					os.Exit(1)
				}
				if C.check_native_buttons(2, 1, 36, 12, -1, 22) != 1 {
					os.Exit(1)
				}
				waitState := func(action C.int) {
					deadline := time.Now().Add(5 * time.Second)
					for time.Now().Before(deadline) {
						if C.check_native_buttons(action, 0, 36, 12, -1, 22) == 1 {
							return
						}
						time.Sleep(20 * time.Millisecond)
					}
					fmt.Println("FAIL: native minimize/restore did not finish")
					os.Exit(1)
				}
				waitState(4)
				C.check_native_buttons(5, 0, 36, 12, -1, 22)
				waitState(6)
				C.check_native_buttons(3, 1, 36, 12, -1, 22)
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		C.check_native_buttons(0, 1, 44, 15, 0, 23)
		fmt.Println("FAIL: native buttons did not appear")
		os.Exit(1)
	}()
	window.Main()
}
