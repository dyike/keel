package main

import (
	"flag"
	"log"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

// hello is the whole app: a view is a struct, Render builds its tree, and
// callbacks change fields that the next frame shows.
type hello struct {
	name    *kit.InputView
	welcome *kit.CheckboxView
	result  string
}

func (h *hello) greet() {
	value := strings.TrimSpace(h.name.Value())
	if value == "" {
		h.result = text("Enter your name first.", "请先输入你的名字。")
		return
	}
	h.result = text("Hello, "+value+"!", "你好，"+value+"！")
	if h.welcome.Value() {
		h.result += text(" Welcome to Keel.", " 欢迎使用 Keel。")
	}
}

func (h *hello) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(16).Child(
		el.Text(text("Build interfaces in Go", "用 Go 组合界面")).TextSize(22).Bold(),
		el.Text(text("Inputs, buttons, layout, and event callbacks are all defined in Go, powered by Gio.", "输入框、按钮、布局和事件回调都由 Go 定义，底层是 Gio。")).TextSize(13).TextColor(theme.Muted),
		el.Div().P(16).Gap(12).Rounded(8).Bg(theme.Surface).Border(1, theme.Border).Child(
			h.name.Render(cx),
			h.welcome.Render(cx),
			el.Div().Row().Gap(8).Child(
				kit.Button(text("Say hello", "打个招呼"), h.greet).Render(cx),
				kit.Button(text("Clear", "清空"), func() {
					h.name.SetValue("")
					h.welcome.SetValue(false)
					h.result = text("Enter your name, then click the button below.", "填写名字后，点击下面的按钮。")
				}).Variant(kit.ButtonSecondary).Render(cx),
			),
			el.Div().H(el.Dp(1)).Bg(theme.Border),
			el.Text(h.result),
		),
	)
}

func main() {
	screenshot := flag.String("screenshot", "", "render a PNG to this path and exit")
	language := flag.String("lang", "en", "example language: en or zh-CN")
	flag.Parse()
	applyLanguage(*language)

	h := &hello{
		name:    kit.Input(text("Your name", "你的名字")).Placeholder(text("For example: Alex", "例如：小明")).MaxLength(80),
		welcome: kit.Checkbox(text("Add a welcome message", "在问候中加上欢迎语"), false),
		result:  text("Enter your name, then click the button below.", "填写名字后，点击下面的按钮。"),
	}
	h.name.OnSubmit(func(string) { h.greet() }) // Enter 也能提交
	content := el.Root(h)

	if *screenshot != "" {
		if err := window.Screenshot(content, 640, 480, *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	window.Open(window.Options{Title: text("Keel", "Keel"), Width: 640, Height: 480, Content: content})
	window.Main()
}
