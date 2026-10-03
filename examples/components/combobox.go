package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"strings"
	"time"
)

func init() {
	registerSection("combobox", "inputs", func() core.Widget {
		customer := kit.Combobox("客户", "华东物流", "北京百货", "深圳电子", "成都餐饮", "杭州茶业", "上海文具").Placeholder("输入筛选").Clearable(true).Size(28)
		tag := kit.Combobox("标签 Tag", "紧急", "VIP", "待跟进").AllowCustom().Multiple().Clearable(true).Size(48).CheckIcon(kit.Icon(kit.IconCheck)).Placeholder("可输入新标签；再选一次取消")
		country := kit.Combobox("国家（显示名称／实际值）").Clearable(true).Searchable(false)
		country.RowHeight(52).RenderItem(func(item kit.ComboboxItem, selected bool) el.View {
			return el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Child(el.Text(item.Label), el.Text("代码："+item.Value))
			})
		})
		confirmed := "尚未结束选择"
		country.OnConfirm(func(values []string) { confirmed = "确认值：" + strings.Join(values, ", ") })
		country.SetGroups(kit.ComboboxGroup{ID: "asia", Label: "亚洲", Items: []kit.ComboboxItem{{Value: "cn", Label: "中国"}, {Value: "jp", Label: "日本（暂不可选）", Disabled: true}}}, kit.ComboboxGroup{ID: "america", Label: "美洲", Items: []kit.ComboboxItem{{Value: "us", Label: "美国"}}})
		tag.Footer(kit.Button("添加示例标签", func() { tag.SetValues(append(tag.Values(), "新标签")) }).Variant(kit.ButtonGhost))
		large := make([]string, 10000)
		for i := range large {
			large[i] = fmt.Sprintf("客户 %05d", i)
		}
		customer.SetOptions(large...)
		customer.DisableOption("客户 00001", true).DisableOption("客户 00003", true)
		remote := kit.Combobox("异步搜索").Multiple().Placeholder("输入关键词；error 模拟失败")
		remote.OnSearch(func(query string, token uint64) {
			go func() {
				time.Sleep(250 * time.Millisecond)
				core.Update(func() {
					if strings.EqualFold(query, "error") {
						remote.SetSearchError(token, "查询失败，请重试或修改关键词")
						return
					}
					var options []string
					for i := range 20 {
						options = append(options, fmt.Sprintf("%s 结果 %02d", query, i+1))
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
