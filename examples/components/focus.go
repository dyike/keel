package main

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("focus", "controls", func() core.Widget { return el.Embed(&focusGallery{}) })
}

type focusGallery struct {
	count      int
	disabled   bool
	text, last string
}

func (v *focusGallery) Render(cx *el.Context) el.Element {
	button := func(id, label string, fn func()) el.Element {
		return el.Div().ID(id).Focusable().Name(label).P(12).Rounded(6).Bg(theme.Surface).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary).TextColor(theme.Primary) }).
			OnClick(fn).Child(el.Text(label))
	}
	return el.Div().Gap(16).OnKey(func(e el.KeyEvent) bool {
		if e.State == el.KeyPress {
			v.last = string(e.Name)
		}
		return false
	}).Child(
		el.Text("Tab / Shift+Tab 切换焦点；Space / Enter 松开时激活"),
		button("disable", "切换禁用子树", func() { v.disabled = !v.disabled }),
		el.Div().Disabled(v.disabled).Gap(12).Child(button("count", "增加计数", func() { v.count++ }),
			button("to-input", "聚焦输入框", func() { cx.Focus("editor") }),
			el.Input().ID("editor").Bind(&v.text).Placeholder("输入框与按钮共用 Tab 顺序")),
		el.Text(fmt.Sprintf("计数：%d；冒泡按键：%s", v.count, v.last)),
	)
}
