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
	registerSection("table_data", "data", func() core.Widget {
		table := kit.Table(kit.Col(demoText("Order number", "单号")).Width(180), kit.Col(demoText("Customer", "客户")).Width(420)).Height(240)
		rows := make([][]string, 0, 60)
		for i := range 20 {
			rows = append(rows, []string{fmt.Sprintf("SO-%03d", i+1), []string{demoText("East China Logistics", "华东物流"), "Shenzhen Tech"}[i%2]})
		}
		table.SetRows(rows)
		table.SetHasMore(true)
		failNext := true
		table.OnLoadMore(func() {
			fail := failNext
			failNext = false
			go func() {
				time.Sleep(300 * time.Millisecond)
				core.Update(func() {
					if fail {
						table.SetLoadError(demoText("Simulated loading failure; click Retry to continue", "加载失败（示例模拟）；点重试继续"))
						return
					}
					end := min(len(rows)+20, 60)
					for i := len(rows); i < end; i++ {
						rows = append(rows, []string{fmt.Sprintf("SO-%03d", i+1), []string{demoText("East China Logistics", "华东物流"), "Shenzhen Tech"}[i%2]})
					}
					table.SetRows(rows)
					table.SetHasMore(len(rows) < 60)
					table.SetLoading(false)
				})
			}()
		})
		search := kit.Input(demoText("Filter loaded rows", "筛选已加载行")).OnChange(func(query string) {
			query = strings.ToLower(strings.TrimSpace(query))
			if query == "" {
				table.SetFilter(nil)
				return
			}
			table.SetFilter(func(row []string) bool { return strings.Contains(strings.ToLower(strings.Join(row, " ")), query) })
		})
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().W(el.Dp(540)).MaxW(el.Full).P(24).Gap(12).Child(search.Render(cx), table.Render(cx), el.Text(fmt.Sprintf(demoText("Loaded %d rows, %d after filtering. Scroll to the bottom to load the next page; 60 rows total.", "已加载 %d 行，筛选后 %d 行；滚动到底加载下一页，共 60 行"), table.Len(), table.VisibleLen())))
		}))
	})
}
