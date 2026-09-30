package main

import (
	"fmt"
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
	"github.com/dyike/keel/ui/window"
)

func main() {
	var settings *window.Window
	count := 0

	// Reuse the settings window while it is open; create a new one after it closes.
	openSettings := func() {
		if settings != nil && !settings.Closed() {
			settings.Raise()
			return
		}
		count++
		settings = window.Open(window.Options{Title: "设置", Width: 420, Height: 340, Content: settingsPage(count)})
	}

	name := widget.Input("你的名字").Hint("例如：小明")
	result := widget.Text("主窗口和设置窗口拥有各自的状态。")
	window.Open(window.Options{
		Title: "Keel 主窗口", Width: 640, Height: 420,
		Shortcuts: map[string]func(){"mod+,": openSettings},
		Content: layout.Column(
			widget.Heading("Keel · Gio"),
			layout.Card(
				name,
				layout.Row(
					widget.Button("打招呼", func() {
						if v := strings.TrimSpace(name.Value()); v != "" {
							result.SetText("你好，" + v + "！")
						} else {
							result.SetText("请先输入名字。")
						}
					}),
					widget.Button("打开设置窗口", openSettings).Secondary(),
				),
				result,
				widget.Muted("⌘ + , 打开设置。"),
			),
		),
	})
	window.Main()
}

func settingsPage(n int) core.Widget {
	enabled := widget.Checkbox("启用提示", true)
	note := widget.Input("设置备注")
	note.SetValue(fmt.Sprintf("设置窗口 #%d", n))
	status := widget.Muted("设置尚未保存。")
	return layout.Column(
		widget.Heading("设置"),
		enabled,
		note,
		widget.Button("保存", func() {
			status.SetText(fmt.Sprintf("已保存：%s，提示=%t", note.Value(), enabled.Value()))
		}),
		status,
	)
}
