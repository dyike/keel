// A resident app starts without windows, opens from core.Update, closes its last
// window, then opens another. Quit must run both window and application cleanup.
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
	time.AfterFunc(20*time.Second, func() { fmt.Println("FAIL: resident timeout"); os.Exit(1) })
	closes := 0
	var open func()
	open = func() {
		w := window.Open(window.Options{Title: "Resident regression", Width: 300, Height: 180, Content: el.Root(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Resident window") })), OnClose: func() {
			closes++
			if closes == 1 {
				core.Update(open)
			} else {
				window.Quit()
			}
		}})
		if closes == 1 && len(os.Args) > 1 && os.Args[1] == "quit-open" {
			time.AfterFunc(300*time.Millisecond, window.Quit)
		} else {
			time.AfterFunc(300*time.Millisecond, w.Close)
		}
	}
	core.Update(open)
	window.MainWithOptions(window.MainOptions{KeepAlive: true, OnQuit: func() {
		if closes != 2 {
			fmt.Printf("FAIL: closes=%d\n", closes)
			os.Exit(1)
		}
		fmt.Println("PASS")
	}})
}
