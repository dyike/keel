# Keel 文档

Keel 用纯 Go 写桌面界面：界面由 [Gio](https://gioui.org) 绘制，原生能力（权限、截图、合成输入、全局快捷键）通过 cgo 调用 macOS API。应用代码里没有 HTML、CSS、JavaScript，也没有 WebView。

## 按你要做的事找文档

| 你要做的事 | 看这篇 |
| --- | --- |
| 第一次跑起来，写出第一个窗口 | [快速开始](getting-started.md) |
| 弄清模块怎么分、谁依赖谁、线程规则 | [架构](architecture.md)，以及每个模块目录下的 README |
| 开窗口、窗口快捷键、离屏截图 | [窗口与应用](app.md) |
| 用 GPUI 风格写界面：视图 + 链式样式 + flexbox | [元素与视图](el.md) |
| 渲染 AI 回答（流式 Markdown、代码高亮） | [Markdown](markdown.md) |
| 查某个组件的 API | [kit 组件规范与索引](kit.md) |
| 查看与 GPUI Kit 的实现进度和剩余缺口 | [组件进度对照](reports/gpui-progress-2026-10-02.md) |
| 申请权限、截屏、模拟键鼠、全局快捷键 | [原生能力](native.md) |
| 编译成 WebAssembly 在浏览器里运行 | [在浏览器里运行](web.md) |
| 新增组件或原生能力 | [扩展指南](extending.md) |
| 写测试、做截图对比 | [测试](testing.md) |
| 让 Agent 点击、输入、截图，跑端到端测试 | [Agent 端到端测试](automation.md) |
| 中文显示方框、快捷键不生效等问题 | [常见问题](troubleshooting.md) |
| 为什么选 Gio、为什么只有一把锁 | [设计决策](decisions.md) |

## 三分钟版本

```go
package main

import (
    "github.com/dyike/keel/ui/el"
    "github.com/dyike/keel/ui/kit"
    "github.com/dyike/keel/ui/window"
)

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

func main() {
    window.Open(window.Options{Title: "Hello", Content: el.Root(&hello{name: kit.Input("你的名字")})})
    window.Main()
}
```

视图是一个 struct，`Render` 按当前状态搭元素树；回调改字段，下一帧就画出来。写法见[元素与视图](el.md)，组件见 [kit](kit.md)。

只要记住一条规则：**回调里直接改组件；其他 goroutine 改组件要包进 `core.Update`**。原因见[架构](architecture.md#线程规则)。

## 维护这套文档

- 文档里的代码都应该能编译。改了 API，搜一遍 `docs/` 里的旧名字。
- 新增模块时：模块目录下写 README，更新 [架构](architecture.md#模块) 的模块图，在 `internal/deps` 的允许表里登记。
- 做了影响全局的取舍（换依赖、改线程模型），在 [设计决策](decisions.md) 里加一条。
