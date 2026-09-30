package main

import (
	"flag"
	"github.com/go-gui-org/go-gui/gui/backend/soft"
	"log"
	"strings"

	"github.com/dyike/keel"
	"github.com/dyike/keel/ui"
)

func main() {
	screenshot := flag.String("screenshot", "", "write a software-rendered PNG and exit")
	flag.Parse()
	name := ui.Input(ui.InputOptions{Label: "你的名字", Placeholder: "例如：小明", MaxLength: 80})
	remember := ui.Checkbox(ui.CheckboxOptions{Label: "在问候中加上欢迎语"})
	result := ui.Text("填写名字后，点击下面的按钮。")
	greet := ui.Button("打个招呼", func() {
		value := strings.TrimSpace(name.Value())
		if value == "" {
			result.SetText("请先输入你的名字。")
			return
		}
		message := "你好，" + value + "！"
		if remember.Value() {
			message += " 欢迎使用 Keel。"
		}
		result.SetText(message)
	})
	clear := ui.NewButton(ui.ButtonOptions{Text: "清空", Variant: ui.Secondary, OnClick: func() {
		name.SetValue("")
		remember.SetValue(false)
		result.SetText("填写名字后，点击下面的按钮。")
	}})
	page := ui.NewPage("Go UI 示例", ui.Column(
		ui.Heading("用 Go 组合界面"),
		ui.Text("输入框、按钮、布局和事件回调都由 Go 定义。"),
		ui.Card(name, remember, ui.Row(greet, clear), ui.Divider(), result),
	))
	page.SetOnError(func(err error) { log.Print(err) })
	kit := keel.NewWithOptions(keel.Options{Name: "Keel Go UI"})
	win, err := kit.Window.New(keel.WindowOptions{Title: "Keel Go UI", Width: 760, Height: 600, UI: page})
	if err != nil {
		log.Fatal(err)
	}
	if *screenshot != "" {
		if err := soft.RenderToPNG(win.Host(), 2, *screenshot); err != nil {
			log.Fatal(err)
		}
		if err := kit.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := kit.Run(); err != nil {
		log.Fatal(err)
	}
}
