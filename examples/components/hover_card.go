package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"time"
)

func init() {
	registerSection("hover_card", "overlays", func() core.Widget { return el.Root(newHoverCardGallery()) })
}

type hoverCardGallery struct{ card *kit.HoverCardView }

func newHoverCardGallery() *hoverCardGallery {
	return &hoverCardGallery{card: kit.HoverCard(
		kit.Button(demoText("@Alex", "@张三"), nil).Variant(kit.ButtonGhost),
		el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().Gap(12).Child(kit.Avatar(demoText("Alex Chen", "张三")).Size(40).Render(cx), el.Div().Gap(4).Grow().Child(
				el.Text(demoText("Alex Chen", "张三 Zhang San")).Bold(),
				el.Text(demoText("East China sales representative managing 12 customers.", "华东区销售，负责 12 个客户。")).TextSize(13).TextColor(theme.Muted),
				kit.Button(demoText("Send message", "发消息"), nil).Size(28).Variant(kit.ButtonSecondary).Render(cx),
			))
		}),
	).OpenDelay(200*time.Millisecond).CloseDelay(500*time.Millisecond).Placement(el.Bottom, el.Start).Offset(8)}
}

func (g *hoverCardGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text(demoText("HoverCard: hover for 0.2 seconds or focus to open. Moving onto the card keeps it open.", "HoverCard：悬停 0.2 秒或聚焦打开，移到卡片上保持打开")).Bold(),
		el.Div().Row().Gap(4).Child(el.Text(demoText("Owner: ", "负责人：")), g.card.Render(cx)),
	)
}
