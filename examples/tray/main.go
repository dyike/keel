// Tray demonstrates a resident application that starts without a window.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/dyike/keel/native/tray"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/window"
)

func main() {
	screenshot := flag.String("screenshot", "", "render the companion window without a tray")
	smoke := flag.Bool("smoke", false, "exercise tray updates and windowless reopening, then exit")
	flag.Parse()
	paused := false
	page := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		status := "Running"
		if paused {
			status = "Paused"
		}
		return el.Div().Gap(16).Child(
			el.Text("Keel tray").TextSize(24).Bold(),
			el.Text(status),
			el.Text("Closing this window keeps the app running. Use the tray menu to reopen it or quit."),
			kit.Button("Quit", window.Quit).Render(cx),
		)
	}))
	if *screenshot != "" {
		if err := window.Screenshot(page, 520, 260, *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *smoke {
		time.AfterFunc(20*time.Second, func() { fmt.Println("FAIL: tray timeout"); os.Exit(1) })
	}
	var w *window.Window
	closes := 0
	var open func()
	open = func() {
		if w != nil && !w.Closed() {
			w.Raise()
			return
		}
		w = window.Open(window.Options{Title: "Keel tray", Width: 520, Height: 260, Content: page, OnClose: func() {
			w = nil
			closes++
			if *smoke {
				if closes == 1 {
					core.Update(open)
				} else {
					window.Quit()
				}
			}
		}})
		if *smoke {
			time.AfterFunc(400*time.Millisecond, w.Close)
		}
	}
	art := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 6; y < 26; y++ {
		for x := 6; x < 26; x++ {
			if x < 12 || (x+y > 29 && x+y < 37) || (x-y > 2 && x-y < 10) {
				art.SetNRGBA(x, y, color.NRGBA{A: 255})
			}
		}
	}
	var statusIcon *tray.Tray
	var menu func() []tray.MenuItem
	menu = func() []tray.MenuItem {
		return []tray.MenuItem{
			{ID: "open", Title: "Open window", OnSelect: func() { core.Update(open) }},
			{ID: "pause", Title: "Pause", Checked: paused, OnSelect: func() {
				core.Update(func() {
					paused = !paused
					if err := statusIcon.SetMenu(menu()); err != nil {
						log.Print(err)
					}
				})
			}},
			{Separator: true},
			{ID: "quit", Title: "Quit", OnSelect: window.Quit},
		}
	}
	var err error
	statusIcon, err = tray.New(tray.Options{Icon: art, Template: true, Tooltip: "Keel tray", Menu: menu(), OnActivate: func() { core.Update(open) }})
	if err != nil {
		log.Fatal(err)
	}
	if *smoke {
		if err = statusIcon.SetTooltip("Keel tray smoke"); err != nil {
			log.Fatal(err)
		}
		if err = statusIcon.SetIcon(art); err != nil {
			log.Fatal(err)
		}
		if err = statusIcon.SetMenu(menu()); err != nil {
			log.Fatal(err)
		}
		core.Update(open)
	}
	window.MainWithOptions(window.MainOptions{KeepAlive: true, OnQuit: func() {
		if err := statusIcon.Close(); err != nil {
			log.Print(err)
		}
		if *smoke {
			if closes != 2 {
				fmt.Printf("FAIL: closes=%d\n", closes)
				os.Exit(1)
			}
			fmt.Println("PASS")
		}
	}})
}
