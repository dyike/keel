package main

import (
	"fmt"
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

const sampleGo = `package main

import "fmt"

// greet returns a greeting for name.
func greet(name string) string {
	if name == "" {
		name = "世界"
	}
	return fmt.Sprintf("你好，%s！", name)
}

func main() {
	fmt.Println(greet("Keel"))
	undefinedCall()
}
`

func init() {
	registerSection("code_editor", "inputs", func() core.Widget {
		words := []kit.CodeCompletion{
			{Label: "fmt.Println", Detail: "func(a ...any)"}, {Label: "fmt.Sprintf", Detail: "func(format string, a ...any) string"},
			{Label: "greet", Detail: "func(name string) string"}, {Label: "strings.Builder", Detail: "type"},
		}
		status := ""
		ed := kit.CodeEditor(sampleGo).Language("go").Name("main.go").Height(380).
			OnComplete(func(line, col int, prefix string) []kit.CodeCompletion {
				var out []kit.CodeCompletion
				for _, w := range words {
					if strings.HasPrefix(strings.ToLower(w.Label), strings.ToLower(prefix)) {
						out = append(out, w)
					}
				}
				return out
			}).
			OnHover(func(line, col int) string {
				if line == 6 {
					return "func greet(name string) string"
				}
				return ""
			})
		ed.OnChange(func(string) { l, c := ed.Cursor(); status = fmt.Sprintf("第 %d 行，第 %d 列", l+1, c+1) })
		showWS := false
		ed.OnDefinition(func(line, col int) {
			if strings.HasPrefix(string([]rune(strings.Split(ed.Value(), "\n")[line])[col:]), "greet") {
				ed.SetCursor(5, 5) // the declaration of greet
			}
		})
		ed.SetDiagnostics([]kit.CodeDiagnostic{{Line: 14, Col: 1, EndLine: 14, EndCol: 14, Severity: kit.CodeSeverityError, Message: "undefined: undefinedCall"}})
		big := kit.CodeEditor("").Language("go").Name("large.go").Height(220)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).W(el.Dp(760)).MaxW(el.Full).Child(
				el.Text("输入时弹出补全（Ctrl+Space 手动触发），指针停在第 7 行或红色波浪线上看提示，按住 ⌘/Ctrl 点函数名跳转。⌘/Ctrl+F 查找替换，⌘/Ctrl+D 选下一处相同文字，⌘/Ctrl+Alt+↑↓ 或 Alt+点击加光标，Alt+Shift 拖动选列，行号左边的箭头折叠。Tab 缩进，先按 Esc 再按 Tab 离开编辑器。").TextColor(theme.Muted),
				el.Div().Row().Gap(8).Child(
					kit.Button("显示空白", func() { showWS = !showWS; ed.ShowWhitespace(showWS) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button("全部折叠", ed.FoldAll).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button("全部展开", ed.UnfoldAll).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button("查找替换", func() { ed.OpenSearch(true) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
				),
				ed.Render(cx),
				el.Text(status).TextSize(theme.TextSm).TextColor(theme.Muted),
				el.Div().Row().Gap(8).Items(el.Center).Child(
					kit.Button("载入 20 万行", func() {
						var sb strings.Builder
						for i := range 200000 {
							if i > 0 {
								sb.WriteByte('\n')
							}
							fmt.Fprintf(&sb, "x%d := %d // 第 %d 行", i, i*7, i+1)
						}
						big.SetValue(sb.String())
					}).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					el.Text(fmt.Sprintf("%d 行", big.Lines())).TextColor(theme.Muted),
				),
				big.Render(cx),
			)
		}))
	})
}
