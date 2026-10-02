package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("select", "inputs", func() core.Widget {
		status := kit.Select("状态", "待付款", "已付款", "已发货", "已完成")
		city := kit.Select("城市 City", "北京", "上海", "广州", "深圳", "杭州", "成都", "Chicago").Searchable()
		city.SetEntries(kit.SelectOption{Value: "bj", Label: "北京", Group: "中国"}, kit.SelectOption{Value: "sh", Label: "上海（暂不可用）", Group: "中国", Disabled: true}, kit.SelectOption{Value: "sz", Label: "深圳", Group: "中国"}, kit.SelectOption{Value: "chi", Label: "Chicago", Group: "美国"})
		multiple := kit.Select("多选 · 一万条选项").Multiple().Searchable()
		entries := make([]kit.SelectOption, 10000)
		for i := range entries {
			entries[i] = kit.SelectOption{Value: fmt.Sprintf("id-%d", i), Label: fmt.Sprintf("选项 %05d", i), Group: fmt.Sprintf("分组 %d", i/1000+1), Disabled: i%100 == 5}
		}
		multiple.SetEntries(entries...)
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Items(el.Start).Child(el.Div().Gap(14).W(el.Dp(300)).MaxW(el.Full).Child(status.Render(cx), city.Render(cx), multiple.Render(cx)))
		}))
	})
}
