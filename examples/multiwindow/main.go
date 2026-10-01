package main

import (
	"fmt"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
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
		settings = window.Open(window.Options{Title: "设置", Width: 420, Height: 340, Content: el.Root(newSettings(count))})
	}

	name := kit.Input("你的名字").Placeholder("例如：小明")
	result := "主窗口和设置窗口拥有各自的状态。"
	window.Open(window.Options{
		Title: "Keel 主窗口", Width: 640, Height: 420,
		Shortcuts: map[string]func(){"mod+,": openSettings},
		Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Child(
				el.Text("Keel · Gio").TextSize(22).Bold(),
				el.Div().P(16).Gap(12).Rounded(8).Bg(theme.Surface).Border(1, theme.Border).Child(
					name.Render(cx),
					el.Div().Row().Gap(8).Child(
						kit.Button("打招呼", func() {
							if v := strings.TrimSpace(name.Value()); v != "" {
								result = "你好，" + v + "！"
							} else {
								result = "请先输入名字。"
							}
						}).Render(cx),
						kit.Button("打开设置窗口", openSettings).Variant(kit.ButtonSecondary).Render(cx),
					),
					el.Text(result),
					el.Text("⌘ + , 打开设置。").TextSize(13).TextColor(theme.Muted),
				),
			)
		})),
	})
	window.Main()
}

// settingsPage is one settings window; each window keeps its own state.
type settingsPage struct {
	enabled *kit.CheckboxView
	note    *kit.InputView
	status  string
}

func newSettings(n int) *settingsPage {
	p := &settingsPage{enabled: kit.Checkbox("启用提示", true), note: kit.Input("设置备注"), status: "设置尚未保存。"}
	p.note.SetValue(fmt.Sprintf("设置窗口 #%d", n))
	return p
}

func (p *settingsPage) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text("设置").TextSize(22).Bold(),
		p.enabled.Render(cx),
		el.Div().W(el.Dp(320)).Child(p.note.Render(cx)),
		kit.Button("保存", func() {
			p.status = fmt.Sprintf("已保存：%s，提示=%t", p.note.Value(), p.enabled.Value())
		}).Render(cx),
		el.Text(p.status).TextSize(13).TextColor(theme.Muted),
	)
}
