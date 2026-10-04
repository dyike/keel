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

Enter 保持缩进，在 `{ ( [` 后多缩进一级（Python 另识别 `:`），在一对括号中间时把右括号放到下一行。Tab/Shift+Tab 缩进和反缩进所选行。没有选区时复制、剪切整行。粘贴的行数和光标数相同时，每个光标各得一行。先按 Esc 再按 Tab 可以离开编辑器。

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

## 自定义搜索会话

SetSearchQuery(query, CodeSearchOptions{MatchCase, WholeWord, Regex}) 启动搜索并绘制匹配，不打开内置面板、不抢焦点；Searchable(false) 只关闭内置入口，仍可使用应用搜索栏。SearchSession 返回查询、选项、面板状态、InvalidPattern、Truncated、Current 和 Matches 的副本；Current 从 0 开始，没有恰好选中的匹配时为 -1。

NextSearchMatch、PreviousSearchMatch 循环跳转，SelectSearchMatch(index) 跳到指定结果并展开折叠。ReplaceCurrentSearchMatch(text) 只替换恰好选中的匹配，ReplaceAllSearchMatches(text) 返回整份文档的替换数，一次操作对应一次撤销和一次 OnChange。只读或自身禁用时这两种替换返回 false/0。CloseSearch 结束高亮；只有关闭内置面板时才把焦点交回编辑区。

匹配使用逐行引擎，不支持跨行或零长度匹配。列表最多保留 10,000 条，多出的结果使 Truncated 为 true；全部替换仍遍历全部匹配。正则替换支持 Go regexp 的 `$1`/`${name}` 展开。大文档搜索和全部替换同步执行，应用应合并高频输入；此接口不是后台 LSP 搜索。

## 跟踪装饰集合

```go
marks := ed.Decorations(
    kit.CodeDecoration{Range: kit.CodeRange{Line: 2, Col: 0, EndLine: 2, EndCol: 8}, Style: kit.CodeDecorationFill},
    kit.CodeDecoration{Range: kit.CodeRange{Line: 4, Col: 0, EndLine: 4, EndCol: 6}, Style: kit.CodeDecorationFrame},
)
marks.Append(other)
tracked := marks.Get()
marks.Clear()   // 仍可复用
marks.Dispose() // 此后该集合的操作不再生效
```

CodeRange 使用从 0 开始的行号和 rune 列，半开区间；不同于上游的 UTF-8 字节偏移。每个集合独立，支持 Frame、Fill、Text（前景色）、Underline；Color 为 nil 时跟随 CodeText，填充使用低透明度。颜色与返回结果均复制，丢弃句柄不会移除装饰；保留句柄也会保留编辑器引用，用 Dispose 释放。

插入在两端不扩大范围，内部插入扩大范围；替换将内部锚点收敛到新范围，整段删除移除条目。撤销/重做同样变换当前位置，已删除的条目不会复活；应从语义数据重建需要恢复的装饰。SetValue 根据最长共同前缀/后缀变换中间改动，仍清空编辑历史。

可见行通过区间索引查询，折叠内容不绘制；更新/追加重建排序，编辑按已有顺序线性维护索引。填充在边框/下划线下，随后绘制选择与文字；同类后创建的集合和后追加的条目覆盖先前样式。装饰不预留空间、不截获输入。几何装饰按可见逻辑行分段绘制，空行没有字形区间时不画；没有连续跨行轮廓或软换行投影。Text 装饰改变颜色，不改变字重或字体度量。

## 语言编辑规则

```go
err := kit.SetCodeLanguageRules("template", kit.CodeLanguageRules{
    Brackets: []kit.CodePair{{Open: "{{", Close: "}}"}},
    AutoClosingPairs: []kit.CodePair{{Open: "{{", Close: "}}", NotIn: []kit.CodeSyntaxContext{kit.CodeSyntaxString, kit.CodeSyntaxComment}}},
    AutoCloseBefore: ";,}",
    Increase: `\{\{\s*$`, Decrease: `^\s*\}\}`,
})
ed.Language("template").AutoClose(true).SmartIndent(true)
```

SetCodeLanguageRules 按语言注册，可在同一事件里替换，下一次编辑立即使用。Chroma 能识别的名称归到其语言名，未知名称保留大小写；ClearCodeLanguageRules 恢复默认。SetEditingRules(&rules) 安装实例覆盖，nil 恢复注册表；非法正则、空/跨行/超过 64 rune 的分隔符拒绝整次配置，保留旧规则。

Brackets 控制 Enter 的结构缩进；AutoClosingPairs 为 nil 时使用 Brackets，非 nil 空切片关闭自动配对。输入支持多字符配对、跨越已有结束串、单行选区包裹和空配对 Backspace。AutoCloseBefore 限制后继字符，空白和行末始终允许。NotIn 默认用 Chroma 的 code/string/comment 分类，SyntaxContext 可由应用替换；未知语言没有语法分类时视为 Code。高亮继续由 Chroma 提供，不是 Tree-sitter。

Increase/Decrease 分别在 Enter 前后文本上匹配，不格式化现有行或粘贴。未提供规则时按结构括号缩进，默认 Python 另识别行尾冒号。AutoClose 和 SmartIndent 独立；关闭 SmartIndent 仍复制当前行前导空白，但不增加/拆分缩进。语言切换不重置这两个偏好。

`OnPaste(func(core.ClipboardData) bool)` 在文本插入前交付文本、编码图片和文件路径；true 表示应用已接收，false 继续普通文本粘贴。`PasteReader(core.ClipboardReader)` 配置异步富剪贴板读取；nil 使用 Gio 文本通路。原生读取失败时通过 `OnPasteError` 报告并回退文本，文本读取失败或超过 16MiB 则拒绝该次粘贴。

读取期间文档版本、任何选区或主光标改变，结果不会插入；禁用/只读会撤销等待中的请求，连续粘贴只接收最新请求。默认插入继续使用多光标粘贴规则及同一撤销记录，回调若自行修改文档或选区也不会再执行默认插入。回调在 UI 线程执行；平台读取完成由组件调度回 UI。

组件库已复用 Input 示例的 native/clipboard 适配，macOS / Windows / Linux X11 图片和文件交给示例回调，尚未支持的平台回退文本。自动测试覆盖多光标粘贴及整体撤销、文件消费、文档/选区变化拒绝、原生失败回退、只读与延迟到达的过期文本。macOS 原生桥接已实际读取图片快照；真实窗口的图片/文件粘贴仍待验收。
