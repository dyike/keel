// Reopen opens and closes windows in quick succession. Gio v0.10.3 on macOS
// crashes when a window is closed before its native window exists; Window
// defers Close until the first frame. Args: "update" opens from a frame, like a
// button callback; "locked" from another goroutine holding the frame lock, like
// automation; "now" closes right after opening. It prints PASS, or crashes.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/window"
)

func main() {
	mode := os.Args[1]
	time.AfterFunc(20*time.Second, func() { fmt.Println("FAIL: timeout"); os.Exit(1) })
	main := window.Open(window.Options{Title: "main", Width: 300, Height: 200, Content: el.Embed(el.ViewFunc(func(*el.Context) el.Element { return el.Text("main") }))})
	closed := func(w *window.Window) bool { loop.Lock(); defer loop.Unlock(); return w.Closed() }
	go func() {
		time.Sleep(time.Second)
		for i := range 4 {
			var w *window.Window
			open := func() {
				w = window.Open(window.Options{Title: fmt.Sprint("second ", i), Width: 300, Height: 200, Content: el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
					return el.Div().Child(kit.Checkbox("c", true).Render(cx), kit.Input("i").Render(cx))
				}))})
			}
			switch mode {
			case "update": // from a window frame, like a button callback
				done := make(chan bool)
				core.Update(func() { open(); done <- true })
				<-done
			case "locked": // from another goroutine holding the frame lock, like automation
				loop.Lock()
				open()
				loop.Unlock()
			}
			if len(os.Args) < 3 || os.Args[2] != "now" {
				time.Sleep(300 * time.Millisecond)
			}
			w.Close()
			for !closed(w) {
				time.Sleep(10 * time.Millisecond)
			}
			fmt.Println("closed", i)
		}
		fmt.Println("PASS")
		core.Update(main.Close)
	}()
	window.Main()
}
