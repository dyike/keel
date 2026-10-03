package main

import (
	"strconv"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("list", "data", func() core.Widget {
		var items []string
		for i := range 300 {
			items = append(items, "联系人 Contact "+strconv.Itoa(i+1))
		}
		msg := "Ctrl/Cmd 多选，Shift 连选；拖动条目重排，灰色条目不可操作"
		var l *kit.ListView
		l = kit.List(items...).Searchable(true).MultiSelect().Reorderable(func(from, to int) { msg = "已移动到第 " + strconv.Itoa(to+1) + " 项" }).Height(240).OnActivate(func(i int) { msg = "打开 " + l.Items()[i] })
		entries := l.Entries()
		for i := range entries {
			entries[i].Disabled = i%10 == 4
			entries[i].Group = "分组 " + strconv.Itoa(i/50+1)
			entries[i].Icon = kit.IconUser
			entries[i].Keywords = []string{"contact"}
		}
		l.SetEntries(entries...)
		l.RowHeight(36).RenderItem(func(cx *el.Context, item kit.ListItemContext) el.Element {
			return el.Div().Row().Grow().Items(el.Center).Gap(8).Child(kit.Icon(item.Item.Icon).Render(cx), el.Text(item.Item.Label).MaxLines(1).Grow(), kit.Button("详情", func() { msg = "详情：" + item.Item.Label }).Size(24).Variant(kit.ButtonGhost).Render(cx))
		})
		l.OnSelectionChange(func(values []int) { msg = "选中 " + strconv.Itoa(len(values)) + " 项" })
		inserted := 0
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).W(el.Dp(320)).MaxW(el.Full).Child(kit.Button("在开头插入条目", func() {
				inserted++
				entry := kit.ListItem{ID: "inserted-" + strconv.Itoa(inserted), Label: "新联系人 " + strconv.Itoa(inserted)}
				l.SetEntries(append([]kit.ListItem{entry}, l.Entries()...)...)
			}).Render(cx), l.Render(cx), el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
