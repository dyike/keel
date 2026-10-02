# CodeEditor

代码编辑器：行号、语法高亮、多光标和列选择、查找替换、代码折叠、括号配对、撤销重做、剪贴板、输入法。只布局可见行，行按块存储：20 万行的文件每帧约 2 毫秒（`BenchmarkCodeEditorLargeFileFrame`），`TestCodeEditorLargeFileFrames` 检查在其中编辑和翻页。

```go
ed := kit.CodeEditor(src).Language("go").Name("main.go").Height(400).
    OnChange(save).
    OnComplete(func(line, col int, prefix string) []kit.CodeCompletion { return lsp.Complete(line, col) }).
    OnHover(func(line, col int) string { return lsp.Hover(line, col) }).
    OnDefinition(func(line, col int) { ed.SetCursor(lsp.Definition(line, col)) })
ed.SetDiagnostics(lsp.Diagnostics())
```

不内置语言服务器协议：应用把语言服务器的诊断、补全、悬停、跳转定义接到 `SetDiagnostics`、`OnComplete`、`OnHover`、`OnDefinition`。

## 高亮

chroma 按 `Language` 的名字选择语法。编辑后立即重新高亮改动附近的几十行，再在后台重算整个文件；后台结果回来前，其余行保持原有颜色。高亮同时标出字符串和注释，括号配对据此判断位置。没有用 Tree-sitter：它在 Go 里需要 cgo，会破坏浏览器版和交叉编译。

## 键盘

| 操作 | macOS | Windows / Linux |
| --- | --- | --- |
| 按词移动、删除 | Option+←/→、Option+⌫ | Ctrl+←/→、Ctrl+⌫ |
| 行首、行尾 | ⌘+←/→、Home/End（Home 在首个非空白列和第 0 列间切换） | Home/End |
| 文件首尾 | ⌘+↑/↓、⌘+Home/End | Ctrl+Home/End |
| 上下加光标 | ⌘+Option+↑/↓ 或 Option+Shift+↑/↓ | Ctrl+Alt+↑/↓ 或 Alt+Shift+↑/↓ |
| 选中下一处相同文字 | ⌘+D | Ctrl+D |
| 只留主光标 | Esc | Esc |
| 查找、查找替换 | ⌘+F、⌘+Option+F | Ctrl+F、Ctrl+H |
| 下一个、上一个匹配 | ⌘+G、⌘+Shift+G、F3、Shift+F3 | Ctrl+G、Ctrl+Shift+G、F3、Shift+F3 |
| 折叠、展开所在区域 | ⌘+Option+[、⌘+Option+] | Ctrl+Alt+[、Ctrl+Alt+] |
| 跳转定义 | F12 | F12 |
| 补全 | Ctrl+Space | Ctrl+Space |
| 撤销、重做 | ⌘+Z、⌘+Shift+Z | Ctrl+Z、Ctrl+Shift+Z |

Enter 保持缩进，在 `{ ( [ :` 后多缩进一级，在一对括号中间时把右括号放到下一行。Tab/Shift+Tab 缩进和反缩进所选行。没有选区时复制、剪切整行。粘贴的行数和光标数相同时，每个光标各得一行。先按 Esc 再按 Tab 可以离开编辑器。

## 鼠标

单击定位，拖动选择，双击选词，三击选行，点行号选整行（折叠时选整个区域）。Option/Alt+单击加光标，Option/Alt+Shift 拖动选一列。按住 ⌘/Ctrl 时指针下的标识符带下划线，单击跳转定义。行号左边的箭头折叠、展开。

## 选项与 API

- 括号配对 `AutoClose(bool)`，默认开：输入 `( [ { " ' ``` ` 自动补右侧；在右侧字符前输入同一个右括号时跳过它；退格删除空括号对；有选区时用括号包住选区；在空白行输入右括号时缩进退回到对应的左括号。字符串和注释里不配对括号，单词后的 `'` 不配对。
- 折叠按缩进：一行和它后面缩进更深的行组成一个区域，回到同级缩进的右括号行保持显示。`Fold`/`Unfold`/`FoldAll`/`UnfoldAll`/`Folded`；光标移进折叠区会自动展开；在区域上方编辑时折叠跟着移动。
- `TabSize(n, hard)` 设置制表位宽度和 Tab 插入制表符还是空格；默认宽 4，按文件现有缩进决定。`ShowWhitespace(bool)` 用点和箭头标出空格和制表符。
- 查找面板：区分大小写、全字匹配、正则表达式（替换里可以用 `$1`），匹配项在文中和滚动条旁标出，最多统计 10000 处。全部替换是一次撤销。`OpenSearch(replace)`、`CloseSearch`、`SearchMatches`、`Searchable(bool)`；只读编辑器只能查找。
- `Value`/`SetValue`、`Cursor`/`SetCursor`、`Cursors`、`Selection`、`Lines`、`SetReadOnly`、`SetDisabled`、`Focus`、`Fill`。多光标编辑整体算一次撤销。

Agent：编辑器角色 `textbox`，名字是 `Name`，值是全文（超过 2000 行时是行数）；补全项是独立的 `option`；悬停提示是 `tooltip`；查找面板角色 `search`，里面的输入框和按钮单独列出；折叠箭头是按钮，名字如"折叠 6"。

验证：`go run ./examples/components -section code_editor`。
