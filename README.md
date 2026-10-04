<p><img src="docs/images/keel.svg" alt="Keel" width="80" height="80"></p>

# Keel

用纯 Go 写桌面界面，不需要 HTML、CSS、JavaScript 或 WebView。界面由 [Gio](https://gioui.org) 绘制；Gio 没有的原生能力（权限、截图、合成输入、全局快捷键、系统通知、富剪贴板）在 `native/` 下。支持 macOS、Windows、Linux，也能编译成 WebAssembly 在浏览器里运行。

```go
type hello struct {
    name   *kit.InputView
    result string
}

func (h *hello) Render(cx *el.Context) el.Element {
    return el.Div().P(24).Gap(12).Child(
        h.name.Render(cx),
        el.Div().Row().Child(kit.Button("打招呼", func() { h.result = "你好，" + h.name.Value() }).Render(cx)),
        el.Text(h.result),
    )
}

window.Open(window.Options{Title: "Hello", Content: el.Root(&hello{name: kit.Input("你的名字")})})
window.Main()
```

新建一个应用：

```sh
go install github.com/dyike/keel/cmd/keel@latest
keel new my-app && cd my-app
keel run            # 运行
keel build          # 打包当前平台（带图标的 .app / .exe / Linux 安装包）
```

或者在已有项目里 `go get github.com/dyike/keel@latest`。看示例和组件库：

```sh
go run ./examples/hello
go run ./examples/components   # 组件库：侧栏选择、搜索
go run ./examples/chat -sample=all
go test -race ./...
```

需要 Go 1.26+；macOS 上需要 Xcode Command Line Tools，Linux 上需要 Wayland/X11 的开发包（`keel doctor` 会检查）。在线文档和组件库：https://keel.dyike.com

## 模块

| 模块 | 做什么 |
| --- | --- |
| [ui/core](ui/core/) | 地基：`Widget` 接口、回调、线程规则 |
| [ui/theme](ui/theme/) | 颜色、字号、字体 |
| [ui/locale](ui/locale/) | 框架文字，中英文切换 |
| [ui/el](ui/el/) | GPUI 风格：视图 + 链式样式元素 + flexbox |
| [ui/kit](ui/kit/) | 组件：`Button`、`Input`、`Table`、`Dialog`、`Chart` … |
| [ui/base](ui/base/) | 无样式的组件行为：键盘导航、首字母跳转、多选 |
| [ui/window](ui/window/) | 窗口：`Open`、`Main`、快捷键、截图 |
| [ui/markdown](ui/markdown/) | Markdown 渲染，针对 AI 流式输出优化 |
| [ui/plot](ui/plot/) | 自定义图表的底层绘图 |
| [native/permission](native/permission/) | 检查、申请系统权限 |
| [native/screen](native/screen/) | 显示器列表、截图 |
| [native/input](native/input/) | 合成鼠标、键盘事件 |
| [native/hotkey](native/hotkey/) | 全局快捷键 |
| [native/notification](native/notification/) | 系统通知：投递、替换、撤回、点击回调 |
| [native/clipboard](native/clipboard/) | 读取剪贴板里的文本、图片、文件 |

两个命令行工具：

- [cmd/keel](cmd/keel/)：脚手架，新建项目、运行、按平台规范生成图标并打包，见 [脚手架与打包](docs/cli.md)。
- [cmd/keel-mcp](cmd/keel-mcp/)：MCP server，让 Agent 对应用做端到端测试（启动、读元素、点击、输入、滚动、截图），见 [Agent 端到端测试](docs/automation.md)。

依赖只往下走、同层不互相引用，`ui` 和 `native` 互不依赖，由 `go test ./...` 里的边界测试保证。每个模块目录下有 README。总览见 [ui/README.md](ui/README.md) 和 [native/README.md](native/README.md)。

## 文档

完整文档在 [docs/](docs/README.md)：

- [快速开始](docs/getting-started.md) · [脚手架与打包](docs/cli.md)
- [架构](docs/architecture.md)：模块划分、依赖方向、线程规则
- [元素与视图](docs/el.md)：GPUI 风格的写法
- [Markdown](docs/markdown.md)：AI 流式输出的渲染
- [窗口与应用](docs/app.md) · [kit 组件](docs/kit.md) · [原生能力](docs/native.md) · [在浏览器里运行](docs/web.md)
- [扩展指南](docs/extending.md)：新增组件、原生能力的步骤
- [测试](docs/testing.md) · [Agent 端到端测试](docs/automation.md) · [常见问题](docs/troubleshooting.md) · [设计决策](docs/decisions.md)

## 授权

Keel 采用 [AGPL-3.0](LICENSE) 与商业授权双授权。开源项目和个人自用按 AGPL 免费使用；在闭源软件里使用 Keel 需要购买商业授权。详见 [LICENSING.md](LICENSING.md)，参与贡献见 [CONTRIBUTING.md](CONTRIBUTING.md)。
