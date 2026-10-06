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
		kit.Empty(demoText("No orders yet", "暂无订单")).Description(demoText("New orders will appear here.", "创建订单后将在这里显示。")).Icon(kit.IconInbox).Action(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(8).Name(demoText("New order", "新建订单")).OnClick(func() {}).Child(el.Text(demoText("New order", "新建订单")))
		})).Render(cx),
		el.Div().W(el.Dp(180)).Child(kit.Empty(demoText("No search results 0", "没有搜索结果 0")).Description(demoText("Try another search query or adjust the filters.", "请尝试其他关键词 Search again，或调整筛选条件。")).Render(cx)),
		kit.Empty(demoText("No notifications", "暂无通知")).Icon(kit.IconNone).Heading(kit.Tag(demoText("No notifications", "暂无通知"))).DescriptionContent(el.ViewFunc(func(*el.Context) el.Element {
			return el.Text(demoText("Notifications appear here", "通知将显示在这里")).Bold()
		})).Footer(kit.Button(demoText("About notification settings", "了解通知设置"), func() {}).Variant(kit.ButtonLink)).PartStyle(kit.EmptyPartRoot, func(e *el.DivEl) { e.Bg(theme.Subtle).Rounded(theme.RadiusLg) }).Render(cx),
		kit.Empty(demoText("Alex has not joined yet", "Alex 尚未加入")).Media(kit.Avatar("Alex").Size(56)).Description(demoText("Invite members to collaborate.", "邀请成员一起协作。")).Action(kit.Button(demoText("Invite members", "邀请成员"), func() {})).Render(cx),
		kit.Empty(demoText("Drop files here", "拖动文件到这里")).Description(demoText("Outline style, suitable for upload areas.", "描边样式，适合上传区域。")).Icon(kit.IconFile).Variant(kit.EmptyOutline).Render(cx),
		kit.Empty(demoText("Nothing scheduled this week", "这一周没有安排")).Description(demoText("Soft background style for cards and panels.", "浅底样式，放在卡片或面板里。")).Icon(kit.IconCalendar).Variant(kit.EmptyMuted).Render(cx),
	)
}
