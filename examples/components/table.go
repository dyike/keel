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
			rows = append(rows, []string{fmt.Sprintf("SO-%05d", i+1), []string{demoText("East China Logistics", "华东物流"), demoText("Beijing Department Store", "北京百货"), "Shenzhen Tech"}[i%3], "¥" + strconv.Itoa(100+i*37%9000), []string{demoText("Awaiting shipment", "待发货"), demoText("Shipped", "已发货"), demoText("Completed", "已完成")}[i%7%3], demoText("East warehouse", "华东仓"), "2026-10-02"})
		}
		msg := demoText("5000 rows; first and last columns are frozen. Scroll horizontally for the middle columns. Click to sort, drag headers to reorder, drag column edges to resize, and press Enter or double-click to open.", "5000 行；首列和末列固定，横向滚动查看中间列。点击排序，拖动表头换序、列边缘调宽，回车或双击打开")
		var t *kit.TableView
		t = kit.Table(kit.Col(demoText("Order number", "单号")).Width(120).Movable(false).Resizable(false), kit.Col(demoText("Customer", "客户")).Width(180),
			kit.Col(demoText("Amount", "金额")).Width(120).Numeric(), kit.Col(demoText("Status", "状态")).Width(100).Cell(func(cx *el.Context, row int) el.Element {
				status := t.Row(row)[3]
				tone := map[string]kit.Tone{demoText("Awaiting shipment", "待发货"): kit.ToneWarning, demoText("Shipped", "已发货"): kit.ToneInfo, demoText("Completed", "已完成"): kit.ToneSuccess}[status]
				return kit.Tag(status).Tone(tone).Render(cx)
			}), kit.Col(demoText("Warehouse", "仓库")).Width(140), kit.Col(demoText("Date", "日期")).Width(140)).FrozenColumns(1, 1).MultiSelect().Stripe(true).RowHeight(36).Height(360).OnActivate(func(r int) { msg = demoText("Open ", "打开 ") + t.Row(r)[0] })
		t.RowMenu(func(row int) *kit.MenuView {
			return kit.Menu().Item(demoText("Open order", "打开订单"), "", func() { msg = demoText("Open ", "打开 ") + t.Row(row)[0] }).Item(demoText("Copy selection", "复制选区"), "mod+c", func() { el.WriteClipboard(t.SelectionText()) })
		})
		t.CellMenu(func(row, column int) *kit.MenuView {
			return kit.Menu().Item(demoText("Copy cell", "复制此单元格"), "", func() {
				values := t.Row(row)
				if column < len(values) {
					el.WriteClipboard(values[column])
				}
			}).Item(demoText("Open associated order", "打开所属订单"), "", func() { msg = demoText("Open ", "打开 ") + t.Row(row)[0] })
		})
		t.OnSelectionChange(func(rows []int) {
			msg = fmt.Sprintf(demoText("Selected %d rows; Ctrl/Cmd+C copies, Shift extends the range", "选中 %d 行；Ctrl/Cmd+C 复制，Shift 扩展范围"), len(rows))
		})
		t.SetRows(rows)
		saved := t.LayoutState()
		warehouseVisible := true
		cells := false
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).W(el.Dp(600)).MaxW(el.Full).Child(
				el.Div().Row().Wrap().Gap(8).Child(
					kit.Button(demoText("Select columns", "整列选择"), func() {
						t.ColumnSelect()
						msg = demoText("Column mode: arrows switch columns, Shift extends the selection, Cmd/Ctrl+A selects all", "整列模式：左右键切换，Shift 扩展，Cmd/Ctrl+A 全选")
					}).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Toggle row / cell selection", "切换行/单元格选择"), func() {
						cells = !cells
						if cells {
							t.CellSelect()
							msg = demoText("Cell mode: Shift extends the range, clicking a header selects the column, and double-clicking sorts.", "单元格模式：Shift 扩展范围，点击表头选整列，双击表头排序")
						} else {
							t.MultiSelect()
							msg = demoText("Multi-row selection mode", "行多选模式")
						}
					}).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Move customer after amount", "客户移到金额后"), func() { t.MoveColumn(1, 2) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Toggle warehouse column", "切换仓库列"), func() { warehouseVisible = !warehouseVisible; t.SetColumnVisible(4, warehouseVisible) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Save column layout", "保存列布局"), func() { saved = t.LayoutState(); msg = demoText("Column layout saved", "列布局已保存") }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Restore column layout", "恢复列布局"), func() {
						if err := t.SetLayoutState(saved); err != nil {
							msg = err.Error()
						} else {
							for _, c := range saved.Columns {
								if c.Column == 4 {
									warehouseVisible = !c.Hidden
								}
							}
							msg = demoText("Column layout restored", "列布局已恢复")
						}
					}).Variant(kit.ButtonSecondary).Size(28).Render(cx)),
				t.Render(cx), el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
