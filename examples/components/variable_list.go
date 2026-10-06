package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("variable_list", "data", func() core.Widget {
		keys := make([]string, 100000)
		for i := range keys {
			keys[i] = strconv.Itoa(i + 1)
		}
		older := 0
		horizontal := false
		expanded := map[string]bool{}
		var list *kit.VariableListView
		list = kit.VariableList(keys, 72, func(cx *el.Context, i int) el.Element {
			k := keys[i]
			number, _ := strconv.Atoi(k)
			text := strings.Repeat(demoText("Mixed text wraps to fit the available width.", "中文 mixed text 会随宽度自动换行。"), 1+number%3)
			if expanded[k] {
				text += strings.Repeat(demoText(" Expanded content increases the row height.", " 展开后的内容继续增加当前行高度。"), 5)
			}
			row := el.Div().Py(12).Px(16).Gap(6).Bg(theme.Surface).Border(1, theme.Border).Child(
				el.Text(demoText("Message ", "消息 ")+k).Bold(), el.Text(text),
				kit.Button(demoText("Expand / collapse", "展开 / 收起"), func() { expanded[k] = !expanded[k] }).Variant(kit.ButtonGhost).Render(cx))
			if horizontal {
				row.W(el.Dp(float32(180 + number%3*60)))
			}
			return row
		}).Height(460)
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Child(
				el.Text(demoText("Variable-height virtual list", "可变高度虚拟列表")).Bold().TextSize(24),
				el.Text(fmt.Sprintf(demoText("%d items; resize the window, expand a row, or prepend items to check the reading position.", "%d 条；调整窗口宽度、展开某行或在头部插入，检查阅读位置。"), len(keys))).TextColor(theme.Muted),
				el.Div().Row().Gap(8).Child(
					kit.Button(demoText("Toggle orientation", "切换横纵方向"), func() { horizontal = !horizontal; list.Horizontal(horizontal) }).Render(cx),
					kit.Button(demoText("Go to message 50000", "定位消息 50000"), func() { list.ScrollToKey(cx, "50000") }).Render(cx),
					kit.Button(demoText("Prepend 10 items", "头部插入 10 条"), func() {
						var add []string
						for range 10 {
							older++
							add = append(add, fmt.Sprintf(demoText("History-%d", "历史-%d"), older))
						}
						keys = append(add, keys...)
						list.SetKeys(keys)
					}).Variant(kit.ButtonSecondary).Render(cx)),
				list.Render(cx))
		}))
	})
}
