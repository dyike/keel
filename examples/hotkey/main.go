// Hotkey shows a system-wide shortcut and a background goroutine updating the UI.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dyike/keel/native/hotkey"
	"github.com/dyike/keel/native/screen"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
	"github.com/dyike/keel/ui/window"
)

func main() {
	hits := 0
	counter := widget.Heading("按 ⌘⇧K（任何应用中都有效）")
	clock := widget.Muted("")

	if _, err := hotkey.Register("cmd+shift+k", func() {
		core.Update(func() { // hotkey callbacks run off the UI; go through Update
			hits++
			counter.SetText(fmt.Sprintf("全局快捷键触发了 %d 次", hits))
		})
	}); err != nil {
		log.Print(err)
	}
	go func() {
		for t := range time.Tick(time.Second) {
			core.Update(func() { clock.SetText("后台时钟：" + t.Format("15:04:05")) })
		}
	}()

	displays := widget.Text("")
	if ds, err := screen.Displays(); err == nil {
		for _, d := range ds {
			displays.SetText(displays.Text() + fmt.Sprintf("显示器 %d：%.0f×%.0f pt，主屏=%t\n", d.ID, d.Width, d.Height, d.Primary))
		}
	} else {
		displays.SetText(err.Error())
	}

	window.Open(window.Options{Title: "Keel 原生能力", Width: 520, Height: 320,
		Content: layout.Column(counter, clock, layout.Card(displays))})
	window.Main()
}
