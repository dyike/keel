// Hotkey shows a system-wide shortcut and a background goroutine updating the UI.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dyike/keel/native/hotkey"
	"github.com/dyike/keel/native/screen"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func main() {
	language := flag.String("lang", "en", "example language: en or zh-CN")
	flag.Parse()
	applyLanguage(*language)
	hits := 0
	counter := text("Press ⌘⇧K from any app", "按 ⌘⇧K（任何应用中都有效）")
	clock := ""

	if _, err := hotkey.Register("cmd+shift+k", func() {
		core.Update(func() { // hotkey callbacks run off the UI; go through Update
			hits++
			counter = fmt.Sprintf(text("Global shortcut triggered %d times", "全局快捷键触发了 %d 次"), hits)
		})
	}); err != nil {
		log.Print(err)
	}
	go func() {
		for t := range time.Tick(time.Second) {
			core.Update(func() { clock = text("Background clock: ", "后台时钟：") + t.Format("15:04:05") })
		}
	}()

	displays := ""
	if ds, err := screen.Displays(); err == nil {
		for _, d := range ds {
			displays += fmt.Sprintf(text("Display %d: %.0f×%.0f pt, primary=%t\n", "显示器 %d：%.0f×%.0f pt，主屏=%t\n"), d.ID, d.Width, d.Height, d.Primary)
		}
	} else {
		displays = err.Error()
	}

	window.Open(window.Options{Title: text("Keel Native Capabilities", "Keel 原生能力"), Width: 520, Height: 320,
		Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Child(
				el.Text(counter).TextSize(22).Bold(),
				el.Text(clock).TextSize(13).TextColor(theme.Muted),
				el.Div().P(16).Rounded(8).Bg(theme.Surface).Border(1, theme.Border).Child(el.Text(strings.TrimSpace(displays))),
			)
		}))})
	window.Main()
}
