# CodeEditor

代码编辑器：行号、语法高亮（chroma）、选择、撤销重做、剪贴板、输入法。只布局可见行：20 万行的文件每帧约 2 毫秒（`BenchmarkCodeEditorLargeFileFrame`），`TestCodeEditorLargeFileFrames` 检查在其中编辑和翻页。高亮在后台线程做，编辑后先保持原有颜色。

```go
ed := kit.CodeEditor(src).Language("go").Name("main.go").Height(400).
    OnChange(save).
    OnComplete(func(line, col int, prefix string) []kit.CodeCompletion { return lsp.Complete(line, col) }).
    OnHover(func(line, col int) string { return lsp.Hover(line, col) })
ed.SetDiagnostics(lsp.Diagnostics())
```

- 不内置语言服务器协议：应用把语言服务器的诊断、补全、悬停接到 `SetDiagnostics`、`OnComplete`、`OnHover`。
- 键盘：方向键、Home（在首个非空白列和第 0 列间切换）、End、PageUp/Down，Shift 扩展选区，Alt（Windows/Linux 为 Ctrl）按词移动，⌘/Ctrl+A/C/X/V/Z，⌘/Ctrl+Shift+Z 重做，Enter 保持缩进，Tab/Shift+Tab 缩进和反缩进，Ctrl+Space 补全。先按 Esc 再按 Tab 可以离开编辑器。
- 鼠标：单击定位，拖动选择，双击选词，三击选行，点行号选整行。
- `Value`/`SetValue`、`Cursor`/`SetCursor`、`Selection`、`Lines`、`SetReadOnly`、`SetDisabled`、`Focus`、`Fill`。

Agent：角色 `textbox`，名字是 `Name`，值是全文（超过 2000 行时是行数）；补全项是独立的 `option` 节点，包含选中状态，可按回车或点击接受；悬停提示是 `tooltip`。

验证：`go run ./examples/components -section code_editor`。

