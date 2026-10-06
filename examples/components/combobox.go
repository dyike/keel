package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"strings"
	"time"
)

func init() {
	registerSection("combobox", "inputs", func() core.Widget {
		customer := kit.Combobox(demoText("Customer", "客户"), demoText("East China Logistics", "华东物流"), demoText("Beijing Department Store", "北京百货"), demoText("Shenzhen Electronics", "深圳电子"), demoText("Chengdu Catering", "成都餐饮"), demoText("Hangzhou Tea", "杭州茶业"), demoText("Shanghai Stationery", "上海文具")).Placeholder(demoText("Type to filter", "输入筛选")).Clearable(true).Size(28)
		tag := kit.Combobox(demoText("Tag", "标签 Tag"), demoText("Urgent", "紧急"), "VIP", demoText("Follow-up needed", "待跟进")).AllowCustom().Multiple().Clearable(true).Size(48).CheckIcon(kit.Icon(kit.IconCheck)).Placeholder(demoText("Enter custom tags; select again to remove", "可输入新标签；再选一次取消"))
		country := kit.Combobox(demoText("Country (display name / value)", "国家（显示名称／实际值）")).Clearable(true)
		country.RenderTrigger(func(state kit.ComboboxTriggerContext) el.View {
			label := demoText("Choose country", "选择国家")
			if len(state.Selection) > 0 {
				label = demoText("Country: ", "国家：") + state.Selection[0].Label
			}
			return el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().P(12).Rounded(theme.RadiusMd).Border(1, theme.Primary).Child(el.Text(label))
			})
		})
		country.RowHeight(52).RenderItem(func(item kit.ComboboxItem, selected bool) el.View {
			return el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Child(el.Text(item.Label), el.Text(demoText("Code: ", "代码：")+item.Value))
			})
		})
		confirmed := demoText("Selection not finished", "尚未结束选择")
		country.OnConfirm(func(values []string) {
			confirmed = demoText("Confirmed value: ", "确认值：") + strings.Join(values, ", ")
		})
		country.SetGroups(kit.ComboboxGroup{ID: "asia", Label: demoText("Asia", "亚洲"), Items: []kit.ComboboxItem{{Value: "cn", Label: demoText("China", "中国")}, {Value: "jp", Label: demoText("Japan (unavailable)", "日本（暂不可选）"), Disabled: true}}}, kit.ComboboxGroup{ID: "america", Label: demoText("Americas", "美洲"), Items: []kit.ComboboxItem{{Value: "us", Label: demoText("United States", "美国")}}})
		tag.Footer(kit.Button(demoText("Add sample tag", "添加示例标签"), func() { tag.SetValues(append(tag.Values(), demoText("New tag", "新标签"))) }).Variant(kit.ButtonGhost))
		large := make([]string, 10000)
		for i := range large {
			large[i] = fmt.Sprintf(demoText("Customer %05d", "客户 %05d"), i)
		}
		customer.SetOptions(large...)
		customer.DisableOption(demoText("Customer 00001", "客户 00001"), true).DisableOption(demoText("Customer 00003", "客户 00003"), true)
		remote := kit.Combobox(demoText("Async search", "异步搜索")).Multiple().Placeholder(demoText("Type a query; error simulates failure", "输入关键词；error 模拟失败"))
		remote.OnSearch(func(query string, token uint64) {
			go func() {
				time.Sleep(250 * time.Millisecond)
				core.Update(func() {
					if strings.EqualFold(query, "error") {
						remote.SetSearchError(token, demoText("Query failed. Retry or change the query", "查询失败，请重试或修改关键词"))
						return
					}
					var options []string
					for i := range 20 {
						options = append(options, fmt.Sprintf(demoText("%s result %02d", "%s 结果 %02d"), query, i+1))
					}
					remote.SetResults(token, options...)
				})
			}()
		})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Items(el.Start).Child(el.Div().Gap(14).W(el.Dp(300)).MaxW(el.Full).Child(customer.Render(cx), tag.Render(cx), country.Render(cx), el.Text(confirmed), remote.Render(cx)))
		}))
	})
}
