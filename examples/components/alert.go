package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("alert", "controls", func() core.Widget {
		return el.Embed(&alertGallery{closable: kit.Alert("可关闭提示").Description("关闭后可以恢复").OnClose(func() {})})
	})
}

type alertGallery struct{ closable *kit.AlertView }

func (v *alertGallery) Render(cx *el.Context) el.Element {
	return el.Div().W(el.Full).Gap(12).Child(v.closable.Render(cx), el.Div().OnClick(func() { v.closable.SetVisible(true) }).P(8).Child(el.Text("恢复提示")),
		kit.Alert("提示 Info", "有 3 条新消息，稍后也可以查看。").Render(cx),
		kit.Alert("成功 Success", "订单 SO-123 已保存。").Tone(kit.Success).Render(cx),
		kit.Alert("警告 Warning", "连接不稳定，请检查网络后重试。").Tone(kit.Warning).Render(cx),
		kit.Alert("失败 Danger", "保存失败，输入内容已保留。").Tone(kit.Danger).Render(cx),
		el.Div().W(el.Dp(180)).Child(kit.Alert("窄容器 123", "中英文 Mixed text should wrap naturally without clipping.").Render(cx)),
	)
}
