package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("empty", "controls", func() core.Widget { return el.Embed(emptyGallery{}) })
}

type emptyGallery struct{}

func (emptyGallery) Render(cx *el.Context) el.Element {
	return el.Div().W(el.Full).Gap(16).Child(
		kit.Empty("暂无订单").Description("创建订单后将在这里显示。").Icon(kit.IconInbox).Action(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(8).Name("新建订单").OnClick(func() {}).Child(el.Text("新建订单"))
		})).Render(cx),
		el.Div().W(el.Dp(180)).Child(kit.Empty("没有搜索结果 0").Description("请尝试其他关键词 Search again，或调整筛选条件。").Render(cx)),
		kit.Empty("暂无通知").Icon(kit.IconNone).Heading(kit.Tag("暂无通知")).DescriptionContent(el.ViewFunc(func(*el.Context) el.Element { return el.Text("通知将显示在这里").Bold() })).Footer(kit.Button("了解通知设置", func() {}).Variant(kit.ButtonLink)).PartStyle(kit.EmptyPartRoot, func(e *el.DivEl) { e.Bg(theme.Subtle).Rounded(theme.RadiusLg) }).Render(cx),
		kit.Empty("Alex 尚未加入").Media(kit.Avatar("Alex").Size(56)).Description("邀请成员一起协作。").Action(kit.Button("邀请成员", func() {})).Render(cx),
		kit.Empty("拖动文件到这里").Description("描边样式，适合上传区域。").Icon(kit.IconFile).Variant(kit.EmptyOutline).Render(cx),
		kit.Empty("这一周没有安排").Description("浅底样式，放在卡片或面板里。").Icon(kit.IconCalendar).Variant(kit.EmptyMuted).Render(cx),
	)
}
