package main

import (
	"flag"
	"log"
	"strings"

	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
	"github.com/dyike/keel/ui/window"
)

func main() {
	screenshot := flag.String("screenshot", "", "render a PNG to this path and exit")
	flag.Parse()

	result := widget.Text("填写名字后，点击下面的按钮。")
	name := widget.Input("你的名字").Hint("例如：小明").MaxLength(80)
	welcome := widget.Checkbox("在问候中加上欢迎语", false)

	greet := func() {
		value := strings.TrimSpace(name.Value())
		if value == "" {
			result.SetText("请先输入你的名字。")
			return
		}
		msg := "你好，" + value + "！"
		if welcome.Value() {
			msg += " 欢迎使用 Keel。"
		}
		result.SetText(msg)
	}
	name.OnSubmit(func(string) { greet() }) // Enter 也能提交

	content := layout.Column(
		widget.Heading("用 Go 组合界面"),
		widget.Muted("输入框、按钮、布局和事件回调都由 Go 定义，底层是 Gio。"),
		layout.Card(
			name,
			welcome,
			layout.Row(
				widget.Button("打个招呼", greet),
				widget.Button("清空", func() {
					name.SetValue("")
					welcome.SetValue(false)
					result.SetText("填写名字后，点击下面的按钮。")
				}).Secondary(),
			),
			layout.Divider(),
			result,
		),
	)

	if *screenshot != "" {
		if err := window.Screenshot(content, 640, 480, *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	window.Open(window.Options{Title: "Keel", Width: 640, Height: 480, Content: content})
	window.Main()
}
