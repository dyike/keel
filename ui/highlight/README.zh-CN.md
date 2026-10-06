# ui/highlight

[English](README.md) | 简体中文

代码高亮，可选模块。CodeEditor、TextView 和 Markdown 代码块的语法着色都来自这里；不引入时代码显示为纯文本，其他功能不受影响。

```go
import _ "github.com/dyike/keel/ui/highlight"
```

- **为什么单独成包**：它用 [chroma](https://github.com/alecthomas/chroma)，内置几百种语言的规则，在启动时全部注册，链接器删不掉，约占 4 MB。不显示代码的应用不必带上。
- **依赖**：`ui/core`（实现 `core.Highlighter` 并在 `init` 里 `core.SetHighlighter`）和 chroma。
- 语言名支持别名和扩展名（`go`、`golang`、`Go`）；配色默认 `github` / `github-dark`，按代码背景深浅选择，Markdown 可用 `markdown.CodeStyle` 指定 chroma 风格名。

要换别的高亮实现，实现 `core.Highlighter` 并调用 `core.SetHighlighter`。
