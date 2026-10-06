package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() { registerSection("button_group", "controls", buttonGroupGallery) }

func buttonGroupGallery() core.Widget {
	// The selected view follows clicks.
	views := kit.ButtonGroup().Name(demoText("View", "视图"))
	for i, label := range []string{demoText("List", "列表"), demoText("Board", "看板"), demoText("Calendar", "日历")} {
		views.Add(kit.Button(label, func() {
			for j, b := range views.Buttons() {
				b.SetSelected(j == i)
			}
		}).Variant(kit.ButtonSecondary).Selected(i == 0))
	}
	pages := kit.ButtonGroup(kit.Button(demoText("Previous page", "上一页"), nil).Outline(true), kit.Button("1", nil).Outline(true).Selected(true),
		kit.Button("2", nil).Outline(true), kit.Button(demoText("Next page", "下一页"), nil).Outline(true)).Name(demoText("Pagination", "分页"))
	save := kit.ButtonGroup(kit.Button(demoText("Save", "保存"), nil), kit.Button("", nil).Name(demoText("More save options", "更多保存选项")).Icon(kit.IconChevronDown)).Name(demoText("Save", "保存"))
	order := kit.ButtonGroup(kit.Button(demoText("Pin", "置顶"), nil).Variant(kit.ButtonSecondary), kit.Button(demoText("Move up", "上移"), nil).Variant(kit.ButtonSecondary),
		kit.Button(demoText("Move down", "下移"), nil).Variant(kit.ButtonSecondary)).Vertical(true).Name(demoText("Sort", "排序"))
	return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(24).Gap(16).Items(el.Start).Child(
			el.Text(demoText("ButtonGroup", "ButtonGroup 按钮组")).TextSize(24).Bold(),
			el.Text(demoText("Only outer corners are rounded. Outline buttons share borders; solid buttons have narrow gaps.", "只圆外侧的角；描边按钮共用边框，实心按钮之间留细缝。")).TextColor(theme.Muted),
			views.Render(cx), pages.Render(cx), save.Render(cx), order.Render(cx))
	}))
}
