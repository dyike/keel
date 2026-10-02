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
			rows = append(rows, []string{fmt.Sprintf("SO-%05d", i+1), []string{"华东物流", "北京百货", "Shenzhen Tech"}[i%3], "¥" + strconv.Itoa(100+i*37%9000), []string{"待发货", "已发货", "已完成"}[i%7%3], "华东仓", "2026-10-02"})
		}
		msg := "5000 行；首列和末列固定，横向滚动查看中间列。点击排序，拖动列边缘调整宽度，回车或双击打开"
		var t *kit.TableView
		t = kit.Table(kit.Col("单号").Width(120), kit.Col("客户").Width(180),
			kit.Col("金额").Width(120).Numeric(), kit.Col("状态").Width(100).Cell(func(cx *el.Context, row int) el.Element {
				status := t.Row(row)[3]
				tone := map[string]kit.Tone{"待发货": kit.ToneWarning, "已发货": kit.ToneInfo, "已完成": kit.ToneSuccess}[status]
				return kit.Tag(status).Tone(tone).Render(cx)
			}), kit.Col("仓库").Width(140), kit.Col("日期").Width(140)).FrozenColumns(1, 1).MultiSelect().Height(360).OnActivate(func(r int) { msg = "打开 " + t.Row(r)[0] })
		t.RowMenu(func(row int) *kit.MenuView {
			return kit.Menu().Item("打开订单", "", func() { msg = "打开 " + t.Row(row)[0] }).Item("复制选区", "mod+c", func() { el.WriteClipboard(t.SelectionText()) })
		})
		t.CellMenu(func(row, column int) *kit.MenuView {
			return kit.Menu().Item("复制此单元格", "", func() {
				values := t.Row(row)
				if column < len(values) {
					el.WriteClipboard(values[column])
				}
			}).Item("打开所属订单", "", func() { msg = "打开 " + t.Row(row)[0] })
		})
		t.OnSelectionChange(func(rows []int) {
			msg = fmt.Sprintf("选中 %d 行；Ctrl/Cmd+C 复制，Shift 扩展范围", len(rows))
		})
		t.SetRows(rows)
		saved := t.LayoutState()
		warehouseVisible := true
		cells := false
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).W(el.Dp(600)).MaxW(el.Full).Child(
				el.Div().Row().Wrap().Gap(8).Child(
					kit.Button("切换行/单元格选择", func() {
						cells = !cells
						if cells {
							t.CellSelect()
							msg = "单元格模式：Shift 扩展范围，点击表头选整列，双击表头排序"
						} else {
							t.MultiSelect()
							msg = "行多选模式"
						}
					}).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button("客户移到首列", func() { t.MoveColumn(1, 0) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button("切换仓库列", func() { warehouseVisible = !warehouseVisible; t.SetColumnVisible(4, warehouseVisible) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button("保存列布局", func() { saved = t.LayoutState(); msg = "列布局已保存" }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button("恢复列布局", func() {
						if err := t.SetLayoutState(saved); err != nil {
							msg = err.Error()
						} else {
							for _, c := range saved.Columns {
								if c.Column == 4 {
									warehouseVisible = !c.Hidden
								}
							}
							msg = "列布局已恢复"
						}
					}).Variant(kit.ButtonSecondary).Size(28).Render(cx)),
				t.Render(cx), el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
