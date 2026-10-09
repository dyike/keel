//go:build darwin && !ios

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/window"
)

func main() {
	window.Open(window.Options{Title: "Center standard", Width: 400, Height: 300})
	window.Open(window.Options{Title: "Center frameless", Width: 480, Height: 320, Frameless: true})
	go func() {
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			if checkPositions(false, false, false) {
				// Moving the windows and requesting further frames must not recenter them.
				checkPositions(true, false, false)
				for range 5 {
					core.Update(func() {})
					time.Sleep(50 * time.Millisecond)
				}
				if !checkPositions(false, true, true) {
					fmt.Println("FAIL: later frames moved the windows")
					os.Exit(1)
				}
				fmt.Println("PASS: standard and frameless windows centered; later frames preserve position")
				os.Exit(0)
			}
			time.Sleep(50 * time.Millisecond)
		}
		checkPositions(false, true, false)
		fmt.Println("FAIL: windows did not center")
		os.Exit(1)
	}()
	window.Main()
}
