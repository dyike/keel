package main

import (
	"github.com/dyike/keel"
	"github.com/dyike/keel/ui"
	"github.com/dyike/keel/window"
	"log"
)

func main() {
	kit := keel.NewWithOptions(keel.Options{ExitOnMainClose: true})
	var overlay *window.Window
	var err error
	_, err = kit.Window.New(keel.WindowOptions{Title: "Keel", UI: ui.NewPage("主窗口", ui.Button("显示工具窗口", func() {
		if err := overlay.Show(); err != nil {
			log.Print(err)
		}
	}))})
	if err != nil {
		log.Fatal(err)
	}
	overlay, err = kit.NewOverlay(keel.OverlayOptions{Title: "工具窗口", UI: ui.NewPage("工具", ui.Column(ui.Heading("工具窗口"), ui.Text("关闭按钮隐藏窗口；主窗口可以重新显示。"), ui.Button("隐藏", func() { _ = overlay.Hide() })))})
	if err != nil {
		log.Fatal(err)
	}
	if err := kit.Run(); err != nil {
		log.Fatal(err)
	}
}
