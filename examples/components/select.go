package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("select", "inputs", func() core.Widget {
		status := kit.Select(demoText("Status", "状态"), demoText("Awaiting payment", "待付款"), demoText("Paid", "已付款"), demoText("Shipped", "已发货"), demoText("Completed", "已完成")).Clearable(true).TitlePrefix(demoText("Status: ", "状态："))
		city := kit.Select(demoText("City", "城市 City"), demoText("Beijing", "北京"), demoText("Shanghai", "上海"), demoText("Guangzhou", "广州"), demoText("Shenzhen", "深圳"), demoText("Hangzhou", "杭州"), demoText("Chengdu", "成都"), "Chicago").Searchable()
		city.Clearable(true).MenuWidth(360).MenuMaxHeight(200).RowHeight(40).
			RenderItem(func(cx *el.Context, row kit.SelectItemContext) el.Element {
				return el.Div().Row().Gap(8).Child(kit.Icon(kit.IconInfo).Size(16).Render(cx), el.Text(row.Option.Label), el.Text(row.Option.Value))
			}).
			Empty(el.ViewFunc(func(*el.Context) el.Element {
				return el.Text(demoText("No matching cities; try Beijing or Chicago", "没有匹配城市，试试北京或 Chicago"))
			}))
		city.SetEntries(kit.SelectOption{Value: "bj", Label: demoText("Beijing", "北京"), Group: demoText("China", "中国")}, kit.SelectOption{Value: "sh", Label: demoText("Shanghai (unavailable)", "上海（暂不可用）"), Group: demoText("China", "中国"), Disabled: true}, kit.SelectOption{Value: "sz", Label: demoText("Shenzhen", "深圳"), Group: demoText("China", "中国")}, kit.SelectOption{Value: "chi", Label: "Chicago", Group: demoText("United States", "美国")})
		multiple := kit.Select(demoText("Multiselect · 10,000 options", "多选 · 一万条选项")).Multiple().Searchable()
		entries := make([]kit.SelectOption, 10000)
		for i := range entries {
			entries[i] = kit.SelectOption{Value: fmt.Sprintf("id-%d", i), Label: fmt.Sprintf(demoText("Option %05d", "选项 %05d"), i), Group: fmt.Sprintf(demoText("Group %d", "分组 %d"), i/1000+1), Disabled: i%100 == 5}
		}
		multiple.SetEntries(entries...)
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Items(el.Start).Child(el.Div().Gap(14).W(el.Dp(300)).MaxW(el.Full).Child(status.Render(cx), city.Render(cx), multiple.Render(cx)))
		}))
	})
}
