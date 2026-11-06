# ui/highlight/golang

English | [简体中文](README.zh-CN.md)

Go-only syntax highlighting for applications whose snippets use Go:

```go
import _ "github.com/dyike/keel/ui/highlight/golang"
```

Supports Go (including aliases and `.go` filenames), Go text templates and plain text. Unknown languages fall back to uncolored text; automatic detection only considers these registered languages. All Chroma styles remain available. Use `ui/highlight` instead when arbitrary languages are required. Import one provider; if an application imports both, explicitly select one with `core.SetHighlighter(golang.Chroma{})` or `core.SetHighlighter(highlight.Chroma{})` before opening windows.

The Go lexer and two XML rules are adapted from Chroma v2.27.0 under its MIT [license](LICENSE). Registration, embed paths and keyed Go rule literals differ; the Go HTML template wrapper is omitted because it requires HTML and its nested languages. Tests compare supported-language tokens, colors and styles against upstream's full registry. Dependencies: Chroma, `ui/core`, `ui/internal/codehighlight`.
