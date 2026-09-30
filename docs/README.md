# Keel 文档

Keel 用纯 Go 写桌面界面：界面由 [Gio](https://gioui.org) 绘制，原生能力（权限、截图、合成输入、全局快捷键）通过 cgo 调用 macOS API。应用代码里没有 HTML、CSS、JavaScript，也没有 WebView。

## 按你要做的事找文档

| 你要做的事 | 看这篇 |
| --- | --- |
| 第一次跑起来，写出第一个窗口 | [快速开始](getting-started.md) |
| 弄清模块怎么分、谁依赖谁、线程规则 | [架构](architecture.md)，以及每个模块目录下的 README |
| 开窗口、窗口快捷键、离屏截图 | [窗口与应用](app.md) |
| 查某个组件或容器的 API | [组件与布局](widgets.md) |
| 申请权限、截屏、模拟键鼠、全局快捷键 | [原生能力](native.md) |
| 新增组件、容器或原生能力 | [扩展指南](extending.md) |
| 写测试、做截图对比 | [测试](testing.md) |
| 让 Agent 点击、输入、截图，跑端到端测试 | [Agent 端到端测试](automation.md) |
| 中文显示方框、快捷键不生效等问题 | [常见问题](troubleshooting.md) |
| 为什么选 Gio、为什么只有一把锁 | [设计决策](decisions.md) |

## 三分钟版本

```go
package main

import (
    "github.com/dyike/keel/ui/layout"
    "github.com/dyike/keel/ui/widget"
    "github.com/dyike/keel/ui/window"
)

func main() {
    name := widget.Input("你的名字")
    result := widget.Text("")
    window.Open(window.Options{Title: "Hello", Content: layout.Card(
        name,
        widget.Button("打招呼", func() { result.SetText("你好，" + name.Value()) }),
        result,
    )})
    window.Main()
}
```

只要记住一条规则：**回调里直接改组件；其他 goroutine 改组件要包进 `core.Update`**。原因见[架构](architecture.md#线程规则)。

## 维护这套文档

- 文档里的代码都应该能编译。改了 API，搜一遍 `docs/` 里的旧名字。
- 新增模块时：模块目录下写 README，更新 [架构](architecture.md#模块) 的模块图，在 `internal/deps` 的允许表里登记。
- 做了影响全局的取舍（换依赖、改线程模型），在 [设计决策](decisions.md) 里加一条。
