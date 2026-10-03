package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/markdown"
)

func init() {
	registerSection("alert", "controls", func() core.Widget {
		return el.Embed(&alertGallery{closable: kit.Alert("可关闭提示").Description("关闭后可以恢复").OnClose(func() {})})
	})
}

type alertGallery struct{ closable *kit.AlertView }

func (v *alertGallery) Render(cx *el.Context) el.Element {
	return el.Div().W(el.Full).Gap(12).Child(
		kit.Alert("维护横幅").Banner(true).Size(kit.AlertSizeSmall).Icon(kit.IconCalendar).Description("今晚 22:00 维护，请提前保存工作。").Render(cx),
		kit.Alert("富内容").Content(markdown.New("**支付未完成**，请检查：\n- 卡片信息\n- 账户余额")).Icon(kit.IconReceipt).Render(cx),
		kit.Alert("紧凑提示").Size(kit.AlertSizeXSmall).Icon(kit.IconNone).Description("无图标也保留等级标识。").Render(cx),
		kit.Alert("大号提示").Size(kit.AlertSizeLarge).Tone(kit.ToneSuccess).Render(cx), v.closable.Render(cx), el.Div().OnClick(func() { v.closable.SetVisible(true) }).P(8).Child(el.Text("恢复提示")),
		kit.Alert("提示 Info").Description("有 3 条新消息，稍后也可以查看。").Render(cx),
		kit.Alert("成功 Success").Description("订单 SO-123 已保存。").Tone(kit.ToneSuccess).Render(cx),
		kit.Alert("警告 Warning").Description("连接不稳定，请检查网络后重试。").Tone(kit.ToneWarning).Render(cx),
		kit.Alert("失败 Danger").Description("保存失败，输入内容已保留。").Tone(kit.ToneDanger).Render(cx),
		el.Div().W(el.Dp(180)).Child(kit.Alert("窄容器 123").Description("中英文 Mixed text should wrap naturally without clipping.").Render(cx)),
	)
}
