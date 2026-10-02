//go:build darwin && !ios && cgo

package main

/*
#cgo CFLAGS: -x objective-c -fblocks
#cgo LDFLAGS: -framework AppKit
int check_positions(int move, int report);
*/
import "C"

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
			if C.check_positions(0, 0) == 1 {
				// Moving the windows and requesting further frames must not recenter them.
				C.check_positions(1, 0)
				for range 5 {
					core.Update(func() {})
					time.Sleep(50 * time.Millisecond)
				}
				if C.check_positions(2, 1) != 1 {
					fmt.Println("FAIL: later frames moved the windows")
					os.Exit(1)
				}
				fmt.Println("PASS: standard and frameless windows centered; later frames preserve position")
				os.Exit(0)
			}
			time.Sleep(50 * time.Millisecond)
		}
		C.check_positions(0, 1)
		fmt.Println("FAIL: windows did not center")
		os.Exit(1)
	}()
	window.Main()
}
