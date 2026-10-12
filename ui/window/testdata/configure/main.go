// Exercise background native configuration while a real window draws frames.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/window"
)

func main() {
	time.AfterFunc(20*time.Second, func() { fmt.Println("FAIL: timeout"); os.Exit(1) })
	w := window.Open(window.Options{Title: "configure during frames", Width: 640, Height: 480, Frameless: true,
		Content: el.Root(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().Child(el.Text("后台最大化与绘制指令 · Wayland"), el.Text("Run callbacks must execute between frames."))
		}))})
	go func() {
		time.Sleep(time.Second)
		for range 40 {
			core.Update(w.ToggleMaximize)
			time.Sleep(20 * time.Millisecond)
		}
		time.Sleep(200 * time.Millisecond)
		fmt.Println("PASS")
		w.Close()
	}()
	window.Main()
}
