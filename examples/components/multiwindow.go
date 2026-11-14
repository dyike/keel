package main

import (
	"fmt"
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func multiwindowContent() core.Widget {
	var settings *window.Window
	count := 0

	// Reuse the settings window while it is open; create a new one after it closes.
	openSettings := func() {
		if settings != nil && !settings.Closed() {
			settings.Raise()
			return
		}
		count++
		settings = window.Open(window.Options{Title: demoText("Settings", "设置"), Width: 420, Height: 340, Content: el.Root(newMultiwindowSettings(count))})
	}

	name := kit.Input(demoText("Your name", "你的名字")).Placeholder(demoText("For example: Alex", "例如：小明"))
	result := demoText("The main and settings windows keep independent state.", "主窗口和设置窗口拥有各自的状态。")
	return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.Shortcut("mod+,", openSettings)
		return el.Div().P(24).Gap(16).Child(
			el.Text(demoText("Keel · Gio", "Keel · Gio")).TextSize(22).Bold(),
			el.Div().P(16).Gap(12).Rounded(8).Bg(theme.Surface).Border(1, theme.Border).Child(
				name.Render(cx),
				el.Div().Row().Gap(8).Child(
					kit.Button(demoText("Say hello", "打招呼"), func() {
						if v := strings.TrimSpace(name.Value()); v != "" {
							result = fmt.Sprintf(demoText("Hello, %s!", "你好，%s！"), v)
						} else {
							result = demoText("Enter your name first.", "请先输入名字。")
						}
					}).Render(cx),
					kit.Button(demoText("Open settings", "打开设置窗口"), openSettings).Variant(kit.ButtonSecondary).Render(cx),
				),
				el.Text(result),
				el.Text(demoText("Press Ctrl/Cmd+, to open settings.", "Ctrl/Cmd + , 打开设置。")).TextSize(13).TextColor(theme.Muted),
			),
		)
	}))
}

// multiwindowSettings is one settings window; each window keeps its own state.
type multiwindowSettings struct {
	enabled *kit.CheckboxView
	note    *kit.InputView
	status  string
}

func newMultiwindowSettings(n int) *multiwindowSettings {
	p := &multiwindowSettings{enabled: kit.Checkbox(demoText("Enable tips", "启用提示"), true), note: kit.Input(demoText("Settings note", "设置备注")), status: demoText("Settings have not been saved.", "设置尚未保存。")}
	p.note.SetValue(fmt.Sprintf(demoText("Settings window #%d", "设置窗口 #%d"), n))
	return p
}

func (p *multiwindowSettings) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text(demoText("Settings", "设置")).TextSize(22).Bold(),
		p.enabled.Render(cx),
		el.Div().W(el.Dp(320)).Child(p.note.Render(cx)),
		kit.Button(demoText("Save", "保存"), func() {
			p.status = fmt.Sprintf(demoText("Saved: %s, tips=%t", "已保存：%s，提示=%t"), p.note.Value(), p.enabled.Value())
		}).Render(cx),
		el.Text(p.status).TextSize(13).TextColor(theme.Muted),
	)
}

func init() { registerSection("multiwindow", "shell", multiwindowContent) }
