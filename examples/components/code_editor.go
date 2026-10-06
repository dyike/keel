package main

import (
	"fmt"
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

var sampleGo = func() string {
	return demoText(`package main

import "fmt"

// greet returns a greeting for name. An empty name greets the whole world, so callers never need to check for one before asking.
func greet(name string) string {
	if name == "" {
		name = "world"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(greet("Keel"))
	undefinedCall()
}
`, `package main

import "fmt"

// greet returns a greeting for name. An empty name greets the whole world, so callers never need to check for one before asking.
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
`)
}

func init() {
	registerSection("code_editor", "inputs", func() core.Widget {
		words := []kit.CodeCompletion{
			{Label: "fmt.Println", Detail: "func(a ...any)"}, {Label: "fmt.Sprintf", Detail: "func(format string, a ...any) string"},
			{Label: "greet", Detail: "func(name string) string"}, {Label: "strings.Builder", Detail: "type"},
		}
		status := ""
		ed := kit.CodeEditor(sampleGo()).Language("go").Name("main.go").Height(380).
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
		ed.PasteReader(readComponentClipboard).OnPaste(func(data core.ClipboardData) bool {
			if len(data.Images)+len(data.Files) == 0 {
				return false
			}
			status = fmt.Sprintf(demoText("Code editor received %d images and %d file references", "代码编辑器接收 %d 张图片、%d 个文件引用"), len(data.Images), len(data.Files))
			return true
		}).OnPasteError(func(err error) {
			status = demoText("Rich clipboard read failed; falling back to text: ", "富剪贴板读取失败，回退文本：") + err.Error()
		})
		ed.OnChange(func(string) {
			l, c := ed.Cursor()
			status = fmt.Sprintf(demoText("Line %d, column %d", "第 %d 行，第 %d 列"), l+1, c+1)
		})
		showWS, wrap := false, false
		annotations := ed.Decorations(kit.CodeDecoration{Range: kit.CodeRange{Line: 5, Col: 5, EndLine: 5, EndCol: 10}, Style: kit.CodeDecorationFill},
			kit.CodeDecoration{Range: kit.CodeRange{Line: 6, Col: 1, EndLine: 8, EndCol: 2}, Style: kit.CodeDecorationFrame})
		query := kit.Input(demoText("Custom search", "自定义查找")).OnChange(func(value string) { ed.SetSearchQuery(value, kit.CodeSearchOptions{}) })
		customRules := false
		ed.OnDefinition(func(line, col int) {
			if strings.HasPrefix(string([]rune(strings.Split(ed.Value(), `
`)[line])[col:]), "greet") {
				ed.SetCursor(5, 5) // the declaration of greet
			}
		})
		ed.SetDiagnostics([]kit.CodeDiagnostic{{Line: 14, Col: 1, EndLine: 14, EndCol: 14, Severity: kit.CodeSeverityError, Message: "undefined: undefinedCall"}})
		big := kit.CodeEditor("").Language("go").Name("large.go").Height(220)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).W(el.Dp(760)).MaxW(el.Full).Child(
				el.Text(demoText("Typing opens completions (Ctrl+Space opens them manually). Hover over line 7 or the red underline for details; ⌘/Ctrl-click a function to jump to its definition. ⌘/Ctrl+F finds and replaces; ⌘/Ctrl+D selects the next match; ⌘/Ctrl+Alt+↑↓ or Alt-click adds a cursor; Alt+Shift-drag selects columns. Click gutter arrows to fold. Tab indents; press Esc then Tab to leave the editor.", "输入时弹出补全（Ctrl+Space 手动触发），指针停在第 7 行或红色波浪线上看提示，按住 ⌘/Ctrl 点函数名跳转。⌘/Ctrl+F 查找替换，⌘/Ctrl+D 选下一处相同文字，⌘/Ctrl+Alt+↑↓ 或 Alt+点击加光标，Alt+Shift 拖动选列，行号左边的箭头折叠。Tab 缩进，先按 Esc 再按 Tab 离开编辑器。")).TextColor(theme.Muted),
				el.Div().Row().Wrap().Gap(8).Child(
					kit.Button(demoText("Show whitespace", "显示空白"), func() { showWS = !showWS; ed.ShowWhitespace(showWS) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Soft wrap", "软换行"), func() { wrap = !wrap; ed.SoftWrap(wrap) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Collapse all", "全部折叠"), ed.FoldAll).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Expand all", "全部展开"), ed.UnfoldAll).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Find and replace", "查找替换"), func() { ed.OpenSearch(true) }).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Clear decorations", "清除装饰"), annotations.Clear).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					kit.Button(demoText("Toggle «» rules", "切换 «» 规则"), func() {
						customRules = !customRules
						if customRules {
							_ = ed.SetEditingRules(&kit.CodeLanguageRules{Brackets: []kit.CodePair{{Open: "{", Close: "}"}, {Open: "(", Close: ")"}, {Open: "[", Close: "]"}, {Open: "«", Close: "»"}}, AutoCloseBefore: ")]};"})
						} else {
							_ = ed.SetEditingRules(nil)
						}
					}).Variant(kit.ButtonSecondary).Size(28).Render(cx),
				),
				el.Div().Row().Gap(8).Items(el.Center).Child(query.Render(cx), kit.Button(demoText("Previous match", "上一处"), ed.PreviousSearchMatch).Render(cx), kit.Button(demoText("Next match", "下一处"), ed.NextSearchMatch).Render(cx), el.Text(fmt.Sprintf(demoText("%d matches", "%d 处"), len(ed.SearchSession().Matches)))),
				ed.Render(cx),
				el.Text(status).TextSize(theme.TextSm).TextColor(theme.Muted),
				el.Div().Row().Gap(8).Items(el.Center).Child(
					kit.Button(demoText("Load 200,000 lines", "载入 20 万行"), func() {
						var sb strings.Builder
						for i := range 200000 {
							if i > 0 {
								sb.WriteByte('\n')
							}
							fmt.Fprintf(&sb, demoText("x%d := %d // Line %d", "x%d := %d // 第 %d 行"), i, i*7, i+1)
						}
						big.SetValue(sb.String())
					}).Variant(kit.ButtonSecondary).Size(28).Render(cx),
					el.Text(fmt.Sprintf(demoText("%d rows", "%d 行"), big.Lines())).TextColor(theme.Muted),
				),
				big.Render(cx),
			)
		}))
	})
}
