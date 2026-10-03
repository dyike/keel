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
			text := strings.Repeat("中文 mixed text 会随宽度自动换行。", 1+number%3)
			if expanded[k] {
				text += strings.Repeat(" 展开后的内容继续增加当前行高度。", 5)
			}
			row := el.Div().Py(12).Px(16).Gap(6).Bg(theme.Surface).Border(1, theme.Border).Child(
				el.Text("消息 "+k).Bold(), el.Text(text),
				kit.Button("展开 / 收起", func() { expanded[k] = !expanded[k] }).Variant(kit.ButtonGhost).Render(cx))
			if horizontal {
				row.W(el.Dp(float32(180 + number%3*60)))
			}
			return row
		}).Height(460)
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Child(
				el.Text("可变高度虚拟列表").Bold().TextSize(24),
				el.Text(fmt.Sprintf("%d 条；调整窗口宽度、展开某行或在头部插入，检查阅读位置。", len(keys))).TextColor(theme.Muted),
				el.Div().Row().Gap(8).Child(
					kit.Button("切换横纵方向", func() { horizontal = !horizontal; list.Horizontal(horizontal) }).Render(cx),
					kit.Button("定位消息 50000", func() { list.ScrollToKey(cx, "50000") }).Render(cx),
					kit.Button("头部插入 10 条", func() {
						var add []string
						for range 10 {
							older++
							add = append(add, fmt.Sprintf("历史-%d", older))
						}
						keys = append(add, keys...)
						list.SetKeys(keys)
					}).Variant(kit.ButtonSecondary).Render(cx)),
				list.Render(cx))
		}))
	})
}
