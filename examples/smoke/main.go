// Smoke creates two native windows, drives Go-Gui input and buttons, opens a
// third window at runtime, and quits without system input or permission prompts.
package main

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/go-gui-org/go-gui/gui"
	"github.com/dyike/keel"
	"github.com/dyike/keel/ui"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	go func() { time.Sleep(25 * time.Second); log.Print("smoke timed out"); os.Exit(1) }()
	kit := keel.New()
	name := ui.Input(ui.InputOptions{Label: "Name"})
	result := ui.Text("Ready")
	greet := ui.Button("Greet", func() { result.SetText("Hello " + name.Value()) })
	page := ui.NewPage("Main", ui.Column(name, greet, result))
	main, err := kit.Window.New(keel.WindowOptions{Title: "Keel native smoke", UI: page})
	check(err)
	otherName := ui.Input(ui.InputOptions{Label: "Independent", Value: "second"})
	_, err = kit.Window.New(keel.WindowOptions{UI: ui.NewPage("Second", otherName)})
	check(err)
	var passed atomic.Bool
	_, err = main.RegisterShortcut("super+comma", func() {
		child, err := kit.Window.New(keel.WindowOptions{UI: ui.NewPage("Settings", ui.Text("Cmd+, opened settings"))})
		check(err)
		check(child.OnReady(func() {
			if len(kit.App.Windows()) != 3 {
				log.Fatal("settings window was not registered")
			}
			passed.Store(true)
			log.Print("PASS: native input, isolated state, Cmd+, command → settings window")
			check(kit.Quit())
		}))
	})
	check(err)
	check(main.OnReady(func() {
		// Allow a native frame to establish the input/button layout before dispatch.
		time.AfterFunc(300*time.Millisecond, func() {
			main.Host().QueueCommand(func(host *gui.Window) {
				check(host.TestType(name.ID(), "Keel"))
				// Input changed after layout, so arrange the new state before clicking.
				host.InvalidateLayout()
				host.QueueCommand(func(host *gui.Window) {
					check(host.TestClick(greet.ID()))
					if result.Text() != "Hello Keel" || otherName.Value() != "second" {
						log.Fatal("Go callback or isolated window state failed")
					}
					// Exercise the real Go-Gui keyboard dispatch on the native UI thread.
					host.EventFn(&gui.Event{Type: gui.EventKeyDown, KeyCode: gui.KeyComma, Modifiers: gui.ModSuper})
				})
			})
		})
	}))
	check(kit.Run())
	if !passed.Load() {
		log.Fatal(fmt.Errorf("smoke exited before completion"))
	}
}
