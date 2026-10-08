//go:build darwin && !ios && cgo && !nometal

package main

/*
#cgo CFLAGS: -x objective-c -fblocks
#cgo LDFLAGS: -framework AppKit -framework QuartzCore -framework Metal
int check_glass(int kind, int width, int height);
*/
import "C"

import (
	"fmt"
	"image/color"
	"os"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/window"
)

func main() {
	time.AfterFunc(20*time.Second, func() { fmt.Println("FAIL: glass window timeout"); os.Exit(1) })
	keeper := window.Open(window.Options{Title: "glass test keeper", Width: 200, Height: 100})
	go func() {
		for _, style := range []window.GlassStyle{window.GlassRegular, window.GlassClear, window.GlassFrosted} {
			w := window.Open(window.Options{Title: "glass acceptance", Width: 480, Height: 320, Frameless: true, NativeTrafficLights: true, Glass: &window.GlassOptions{Style: style, CornerRadius: 16}, Content: el.Root(el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().Bg(color.NRGBA{}).Child(el.Text("Gio content on native glass"))
			}))})
			kind := 0
			if style != window.GlassFrosted && window.LiquidGlassSupported() {
				kind = 1
			}
			wait := func(width, height int) {
				deadline := time.Now().Add(5 * time.Second)
				for time.Now().Before(deadline) {
					if C.check_glass(C.int(kind), C.int(width), C.int(height)) == 1 {
						return
					}
					time.Sleep(30 * time.Millisecond)
				}
				fmt.Printf("FAIL: native glass kind=%d size=%dx%d\n", kind, width, height)
				os.Exit(1)
			}
			wait(480, 320)
			if err := w.Resize(620, 420); err != nil {
				panic(err)
			}
			wait(620, 420)
			w.Minimize()
			time.Sleep(200 * time.Millisecond)
			w.Raise()
			wait(620, 420)
			w.Close()
			time.Sleep(100 * time.Millisecond)
		}
		// Closing before the asynchronously installed backdrop is ready must
		// not leave a queued block pointing at freed context memory.
		for range 4 {
			w := window.Open(window.Options{Title: "glass early close", Glass: &window.GlassOptions{}})
			w.Close()
		}
		time.Sleep(300 * time.Millisecond)
		fmt.Println("PASS: Liquid Glass/Clear/frosted, transparent Metal surface, input hit testing, resize, minimize, teardown and early close")
		keeper.Close()
	}()
	window.Main()
}
