// Raise opens two real windows and calls Raise and Close from inside UI code,
// which holds the frame lock. It prints PASS, or exits 1 if the UI deadlocks.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/widget"
	"github.com/dyike/keel/ui/window"
)

func main() {
	time.AfterFunc(15*time.Second, func() { fmt.Println("FAIL: deadlock"); os.Exit(1) })
	var second *window.Window
	step := func(d time.Duration, fn func()) { time.AfterFunc(d, func() { core.Update(fn) }) }

	main := window.Open(window.Options{Title: "main", Width: 300, Height: 200, Content: widget.Text("main")})
	step(1*time.Second, func() {
		second = window.Open(window.Options{Title: "second", Width: 300, Height: 200, Content: widget.Text("second")})
	})
	step(2*time.Second, func() { main.Raise() })   // focus moves: second window gets frames
	step(3*time.Second, func() { second.Raise() }) // the reported case: Raise from UI code
	step(4*time.Second, func() { main.Raise() })
	step(5*time.Second, func() {
		fmt.Println("PASS")
		second.Close()
		main.Close()
	})
	window.Main()
}
