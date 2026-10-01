package main

import (
	"fmt"
	"strconv"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("table", "data", func() core.Widget {
		var rows [][]string
		for i := range 5000 {
			rows = append(rows, []string{fmt.Sprintf("SO-%05d", i+1), []string{"华东物流", "北京百货", "Shenzhen Tech"}[i%3], strconv.Itoa(100 + i*37%9000)})
		}
		msg := "5000 行；拖动表头右边缘调整列宽，回车或双击打开"
		var t *kit.TableView
		t = kit.Table(kit.Col("单号").Width(120), kit.Col("客户").Flex(2),
			kit.Col("金额").Numeric().Cell(func(cx *el.Context, row int) el.Element {
				return kit.Tag("¥" + t.Row(row)[2]).Tone(kit.ToneInfo).Render(cx)
			})).Height(360).OnActivate(func(r int) { msg = "打开 " + t.Row(r)[0] })
		t.SetRows(rows)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).W(el.Dp(600)).Child(t.Render(cx), el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
