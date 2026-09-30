package main

import (
	"github.com/dyike/keel"
	"github.com/dyike/keel/ui"
	"log"
)

func main() {
	kit := keel.New()
	result := ui.Text("点击按钮试试。")
	page := ui.NewPage("Keel", ui.Column(ui.Heading("Hello Keel"), ui.Button("打招呼", func() { result.SetText("你好，Go-Gui！") }), result))
	if _, err := kit.Window.New(keel.WindowOptions{UI: page}); err != nil {
		log.Fatal(err)
	}
	if err := kit.Run(); err != nil {
		log.Fatal(err)
	}
}
