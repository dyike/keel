// Minimized checks that core.Update runs while the only window is hidden.
// A window nobody can see draws no frames on macOS (Keel's occlusion patch),
// so updates must also drain on wakeups. It prints PASS, or FAIL on a stall.
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
	time.AfterFunc(10*time.Second, func() { fmt.Println("FAIL: update stalled while minimized"); os.Exit(1) })
	w := window.Open(window.Options{Title: "minimized", Width: 300, Height: 200, Content: el.Embed(el.ViewFunc(func(*el.Context) el.Element { return el.Text("minimized") }))})
	time.AfterFunc(time.Second, func() { core.Update(w.Minimize) })
	// Give AppKit time to finish the minimize animation and report occlusion.
	time.AfterFunc(3*time.Second, func() {
		core.Update(func() {
			fmt.Println("PASS")
			w.Close()
		})
	})
	window.Main()
}
