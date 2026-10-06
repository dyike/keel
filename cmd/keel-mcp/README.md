# cmd/keel-mcp

English | [简体中文](README.zh-CN.md)

MCP server: Let Agent perform end-to-end testing of Keel applications. Launch the application (or connect the application you launched with `KEEL_AUTOMATION=1`), read the elements in the window, click, type, press keys, scroll, take screenshots.

```sh
go install github.com/dyike/keel/cmd/keel-mcp@latest
claude mcp add keel -- keel-mcp
```

- **Dependencies**: MCP Go SDK. It does not reference any Keel package or Gio, and only communicates with the application through the JSON protocol on the `KEEL_AUTOMATION` socket (see `ui/window/automation_server.go` for the protocol).
- **Application under test**: Any Keel program that calls `window.Main()` does not need to change the code.

| File | Responsibility |
| --- | --- |
| `main.go` | Start MCP server |
| `tools.go` | Tool definition, result formatting |
| `process.go` | Start, connect, stop the application under test, socket request, log buffering |
| `main_test.go` | Test launch and attach methods through MCP client |

For usage, tool list, principles and limitations, see [Agent End-to-End Test](../../docs/automation.md).
