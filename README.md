# Keel

用纯 Go 写桌面界面，不需要 HTML、CSS、JavaScript 或 WebView。界面由 [Gio](https://gioui.org) 绘制；Gio 没有的原生能力（权限、截图、合成输入、全局快捷键）在 `native/` 下。

```go
name := widget.Input("你的名字")
result := widget.Text("")
window.Open(window.Options{Title: "Hello", Content: layout.Card(
    name,
    widget.Button("打招呼", func() { result.SetText("你好，" + name.Value()) }),
    result,
)})
window.Main()
```

```sh
go run ./examples/hello
go run ./examples/multiwindow
go run ./examples/hotkey
go run ./examples/chat -sample=all
go test -race ./...
```

需要 Go 1.26+；macOS 上需要 Xcode Command Line Tools。

## 模块

| 模块 | 做什么 |
| --- | --- |
| [ui/core](ui/core/) | 地基：`Widget` 接口、回调、线程规则 |
| [ui/theme](ui/theme/) | 颜色、字号、字体 |
| [ui/layout](ui/layout/) | 摆放组件：`Column`、`Row`、`Card` … |
| [ui/widget](ui/widget/) | 交互组件：`Button`、`Input`、`Checkbox` … |
| [ui/window](ui/window/) | 窗口：`Open`、`Main`、快捷键、截图 |
| [ui/el](ui/el/) | GPUI 风格：视图 + 链式样式元素 + flexbox，新界面优先用它 |
| [ui/markdown](ui/markdown/) | Markdown 渲染，针对 AI 流式输出优化 |
| [native/permission](native/permission/) | 检查、申请系统权限 |
| [native/screen](native/screen/) | 显示器列表、截图 |
| [native/input](native/input/) | 合成鼠标、键盘事件 |
| [native/hotkey](native/hotkey/) | 全局快捷键 |

另有工具 [cmd/keel-mcp](cmd/keel-mcp/)：MCP server，让 Agent 对应用做端到端测试（启动、读元素、点击、输入、滚动、截图），见 [Agent 端到端测试](docs/automation.md)。

依赖只往下走、同层不互相引用，`ui` 和 `native` 互不依赖，由 `go test ./...` 里的边界测试保证。每个模块目录下有 README。总览见 [ui/README.md](ui/README.md) 和 [native/README.md](native/README.md)。

## 文档

完整文档在 [docs/](docs/README.md)：

- [快速开始](docs/getting-started.md)
- [架构](docs/architecture.md)：模块划分、依赖方向、线程规则
- [元素与视图](docs/el.md)：GPUI 风格的写法
- [Markdown](docs/markdown.md)：AI 流式输出的渲染
- [窗口与应用](docs/app.md) · [组件与布局](docs/widgets.md) · [原生能力](docs/native.md)
- [扩展指南](docs/extending.md)：新增组件、容器、原生能力的步骤
- [测试](docs/testing.md) · [Agent 端到端测试](docs/automation.md) · [常见问题](docs/troubleshooting.md) · [设计决策](docs/decisions.md)
