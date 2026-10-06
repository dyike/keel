package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("overlay", "controls", func() core.Widget { return el.Root(&overlayGallery{popover: true}) })
}

type overlayGallery struct {
	popover, modal bool
	text           string
}

func (v *overlayGallery) Render(cx *el.Context) el.Element {
	action := func(id, label string, fn func()) el.Element {
		return el.Div().ID(id).Name(label).Px(14).Py(8).Rounded(6).Bg(theme.Surface).Border(1, theme.Border).OnClick(fn).Child(el.Text(label))
	}
	if v.popover {
		cx.Overlay("popover", el.Anchored("anchor", el.Div().W(el.Dp(280)).P(16).Gap(8).Bg(theme.Surface).Border(1, theme.Border).Rounded(8).Role("dialog").Name(demoText("Nonmodal overlay", "非模态浮层")).Child(
			el.Text(demoText("Anchor to this frame's position", "锚定本帧位置")).Bold(), el.Text(demoText("Clicking outside closes the overlay and activates the button underneath.", "外部点击会关闭浮层，同时触发下面的按钮。")),
			action("close-popover", demoText("Close overlay", "关闭浮层"), func() { v.popover = false }),
		)).OnDismiss(func() { v.popover = false }))
	}
	if v.modal {
		cx.Overlay("modal", el.Modal(el.Div().W(el.Dp(360)).P(20).Gap(12).Bg(theme.Surface).Rounded(10).Role("dialog").Name(demoText("Modal overlay", "模态浮层")).Child(
			el.Text(demoText("Focus stays inside the overlay", "焦点限制在浮层内")).Bold(), el.Text(demoText("Tab / Shift+Tab cycles focus; Esc or clicking the backdrop closes it.", "Tab / Shift+Tab 循环，Esc 或点击遮罩关闭。")),
			el.Input().ID("name").Name(demoText("Name", "名称")).Placeholder(demoText("Enter text or numbers 123", "输入中文、数字 123")).Bind(&v.text),
			action("close-modal", demoText("Close and restore focus", "关闭并恢复焦点"), func() { v.modal = false }),
		)).OnDismiss(func() { v.modal = false }))
	}
	hover := demoText("Move the pointer onto the anchor", "将指针移到锚点上")
	if cx.Hovered("anchor") {
		hover = demoText("Anchor is hovered", "锚点正在悬停")
	}
	return el.Div().P(24).Gap(16).Child(
		el.Text(demoText("Overlays and focus constraints E4 / E5", "浮层与焦点约束 E4 / E5")).TextSize(24).Bold(),
		el.Text(demoText("Nonmodal overlays let outside clicks through; modal backdrops intercept them.", "非模态外部点击穿透；模态遮罩拦截点击。")),
		el.Div().Row().Gap(12).Child(action("anchor", demoText("Toggle anchored overlay", "切换锚定浮层"), func() { v.popover = !v.popover }), action("modal-trigger", demoText("Open modal overlay", "打开模态浮层"), func() { v.modal = true })),
		el.Text(hover).TextColor(theme.Muted),
	)
}
