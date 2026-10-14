//go:build darwin && !ios

package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/dyike/keel/ui/window"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

func check(ok bool, message string) {
	if !ok {
		fmt.Println("FAIL:", message)
		os.Exit(1)
	}
}

func main() {
	time.AfterFunc(10*time.Second, func() { check(false, "menu timeout") })
	art := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 4; y < 28; y++ {
		for x := 4; x < 28; x++ {
			art.SetNRGBA(x, y, color.NRGBA{R: 80, G: 140, B: 240, A: 255})
		}
	}
	icon, err := window.NewMenuIcon(art, true)
	check(err == nil, "template icon")
	colored, err := window.NewMenuIcon(art, false)
	check(err == nil, "color icon")
	bar, err := window.NewMenuBar(window.MenuItem{ID: "file", Title: "Keel icons", Children: []window.MenuItem{
		{ID: "run", Title: "Template", Checked: true, Icon: icon},
		{ID: "color", Title: "Color", Icon: colored},
	}})
	check(err == nil, "menu model")
	w := window.Open(window.Options{Title: "Native menu icons", Width: 420, Height: 240})
	check(w.MenuFeatures().TemplateIcons && w.MenuFeatures().NativeDismissAnimation, "capabilities")
	go func() {
		// Let Gio initialize NSApplication before installing the test menu.
		time.Sleep(400 * time.Millisecond)
		check(bar.Install() == nil, "install")
		done := make(chan struct{}, 1)
		appkit.MainAsync(func() {
			menu := appkit.Send(appkit.Send(appkit.Send(appkit.App(), "mainMenu"), "itemAtIndex:", 0), "submenu")
			first := appkit.Send(menu, "itemAtIndex:", 0)
			check(appkit.Send(first, "image") != 0 && appkit.SendBool(appkit.Send(first, "image"), "isTemplate"), "native template image")
			check(appkit.Send(first, "state") == 1, "checkmark preserved beside icon")
			second := appkit.Send(menu, "itemAtIndex:", 1)
			check(appkit.Send(second, "image") != 0 && !appkit.SendBool(appkit.Send(second, "image"), "isTemplate"), "native color image")
			// Off the frame lock: the native tracking loop must accept cancellation
			// even when it is not producing UI frames.
			var popup func(appkit.ID, appkit.SEL, appkit.ID, appkit.Point, appkit.ID) bool
			purego.RegisterLibFunc(&popup, purego.RTLD_DEFAULT, "objc_msgSend")
			// Run from an AppKit timer, outside a main dispatch-queue block:
			// libdispatch does not reenter its own queue during nested tracking.
			cls := appkit.RegisterClass("KeelMenuPopupFixture", "NSObject", nil, []objc.MethodDef{
				appkit.Method("run:", func(self appkit.ID, _ objc.SEL, _ appkit.ID) {
					for _, animated := range []bool{false, true} {
						time.AfterFunc(150*time.Millisecond, func() { bar.CancelTracking(animated) })
						popup(menu, appkit.Sel("popUpMenuPositioningItem:atLocation:inView:"), 0, appkit.Point{X: 120, Y: 120}, 0)
					}
					appkit.Release(self)
					done <- struct{}{}
				}),
			})
			object := appkit.Send(appkit.Send(cls, "alloc"), "init")
			var later func(appkit.ID, appkit.SEL, appkit.SEL, appkit.ID, float64)
			purego.RegisterLibFunc(&later, purego.RTLD_DEFAULT, "objc_msgSend")
			later(object, appkit.Sel("performSelector:withObject:afterDelay:"), appkit.Sel("run:"), 0, 0.05)
		})
		<-done
		check(bar.UpdateItem("run", func(item *window.MenuItem) { item.Icon = nil }) == nil, "remove icon")
		appkit.MainAsync(func() {
			menu := appkit.Send(appkit.Send(appkit.Send(appkit.App(), "mainMenu"), "itemAtIndex:", 0), "submenu")
			check(appkit.Send(appkit.Send(menu, "itemAtIndex:", 0), "image") == 0, "native image removed on update")
			fmt.Println("PASS: native menu images, replacement and animated/immediate cancellation")
			os.Exit(0)
		})
	}()
	window.Main()
}
