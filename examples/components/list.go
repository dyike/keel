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
		l = kit.List(items...).MultiSelect().Reorderable(func(from, to int) { msg = "已移动到第 " + strconv.Itoa(to+1) + " 项" }).Height(240).OnActivate(func(i int) { msg = "打开 " + l.Items()[i] })
		entries := l.Entries()
		for i := range entries {
			entries[i].Disabled = i%10 == 4
		}
		l.SetEntries(entries...)
		l.OnSelectionChange(func(values []int) { msg = "选中 " + strconv.Itoa(len(values)) + " 项" })
		inserted := 0
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).W(el.Dp(320)).Child(kit.Button("在开头插入条目", func() {
				inserted++
				entry := kit.ListItem{ID: "inserted-" + strconv.Itoa(inserted), Label: "新联系人 " + strconv.Itoa(inserted)}
				l.SetEntries(append([]kit.ListItem{entry}, l.Entries()...)...)
			}).Render(cx), l.Render(cx), el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
