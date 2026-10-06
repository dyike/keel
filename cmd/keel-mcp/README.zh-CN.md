# cmd/keel-mcp

[English](README.md) | 简体中文

MCP server：让 Agent 对 Keel 应用做端到端测试。启动应用（或连接你用 `KEEL_AUTOMATION=1` 启动的应用），读出窗口里的元素，点击、输入、按键、滚动、截图。

```sh
go install github.com/dyike/keel/cmd/keel-mcp@latest
claude mcp add keel -- keel-mcp
```

- **依赖**：MCP Go SDK。不引用任何 Keel 包，也不引用 Gio，只通过 `KEEL_AUTOMATION` socket 上的 JSON 协议和应用通信（协议见 `ui/window/automation_server.go`）。
- **被测应用**：任何调用 `window.Main()` 的 Keel 程序，不需要改代码。

| 文件 | 内容 |
| --- | --- |
| `main.go` | 启动 MCP server |
| `tools.go` | 工具定义、结果格式化 |
| `process.go` | 启动、连接、停止被测应用，socket 请求，日志缓冲 |
| `main_test.go` | 通过 MCP 客户端测试 launch 和 attach 两种方式 |

用法、工具列表、原理和局限见 [Agent 端到端测试](../../docs/automation.zh-CN.md)。
