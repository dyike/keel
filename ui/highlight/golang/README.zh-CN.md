# ui/highlight/golang

[English](README.md) | 简体中文

仅需显示 Go 代码的应用可使用这个高亮入口：

```go
import _ "github.com/dyike/keel/ui/highlight/golang"
```

支持 Go（含别名和 `.go` 文件名）、Go 文本模板及纯文本。未知语言回退为无颜色文本；自动识别只考虑已注册的语言。所有 Chroma 配色仍可使用。需要任意语言时使用 `ui/highlight`。应用应选择一个入口；若同时导入两者，应在打开窗口前用 `core.SetHighlighter(golang.Chroma{})` 或 `core.SetHighlighter(highlight.Chroma{})` 明确选择。

Go lexer 和两个 XML 规则来自 Chroma v2.27.0，遵循其 MIT [许可证](LICENSE)。调整注册、嵌入路径和 Go 规则字段的具名初始化，省略会引入 HTML 及嵌套语言的 Go HTML 模板包装。测试与完整规则库对比受支持语言的 token、颜色和配色。依赖：Chroma、`ui/core`、`ui/internal/codehighlight`。
