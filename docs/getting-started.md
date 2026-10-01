# 快速开始

## 环境

- Go 1.26 或更新（`go.mod` 声明 `go 1.26.1`）。
- macOS：安装 Xcode Command Line Tools（`xcode-select --install`）。Gio 和 `native/` 都走 cgo。原生能力要求 macOS 14+。
- Windows、Linux：界面部分按 Gio 的要求准备环境（Linux 需要 Wayland 或 X11 的开发头文件）。`native/` 在这两个平台上只返回 `native.ErrUnsupported`。

## 跑示例

在仓库根目录执行：

```sh
go run ./examples/hello        # 输入框、复选框、按钮
go run ./examples/multiwindow  # 两个窗口，⌘+, 打开设置窗口
go run ./examples/hotkey       # 全局快捷键 ⌘⇧K，后台时钟每秒刷新
```

第一次编译要下载 Gio 并编译 cgo，大约几十秒；之后增量编译很快。macOS 链接时会打印 `ld: warning: ignoring duplicate libraries: '-lobjc'`，可以忽略，见[常见问题](troubleshooting.md#链接时出现--lobjc-警告)。

不想开窗口，也可以直接渲染成图片：

```sh
go run ./examples/hello -screenshot out.png
```

## 写第一个窗口

```go
package main

import (
    "strings"

    "github.com/dyike/keel/ui/el"
    "github.com/dyike/keel/ui/kit"
    "github.com/dyike/keel/ui/theme"
    "github.com/dyike/keel/ui/window"
)

type hello struct {
    name   *kit.InputView
    result string
}

func (h *hello) greet() {
    if v := strings.TrimSpace(h.name.Value()); v != "" {
        h.result = "你好，" + v + "！"
    }
}

func (h *hello) Render(cx *el.Context) el.Element {
    return el.Div().P(24).Gap(12).Child(
        el.Text("第一个窗口").TextSize(22).Bold(),
        h.name.Render(cx),
        el.Div().Row().Child(kit.Button("打招呼", h.greet).Render(cx)),
        el.Text(h.result).TextColor(theme.Muted),
    )
}

func main() {
    h := &hello{name: kit.Input("你的名字").Placeholder("例如：小明"), result: "等待输入"}
    h.name.OnSubmit(func(string) { h.greet() }) // 回车也触发
    window.Open(window.Options{Title: "Hello", Width: 480, Height: 320, Content: el.Root(h)})
    window.Main()
}
```

这段代码的结构就是所有 Keel 程序的结构：

1. **视图是一个 struct，保存状态。** 输入框这类有状态的 kit 组件创建一次，存在字段里；`result` 是普通字段。
2. **`Render` 按当前状态搭元素树。** 每帧调用，用 `el.Div` 排版，组件通过 `Render(cx)` 放进树里。回调只改字段。
3. **`window.Open` 把 `el.Root(view)` 放进窗口，`window.Main` 进入事件循环。** `window.Main` 不会返回；最后一个窗口关闭时进程退出。

界面不需要手动刷新。回调执行完，所有窗口会自动重绘。

## 在你自己的项目里使用

Keel 还没发布版本。在你的项目里用本地替换引用它：

```sh
go mod edit -require=github.com/dyike/keel@v0.0.0
go mod edit -replace=github.com/dyike/keel=$HOME/Code/go/src/github.com/dyike/keel
go mod tidy
```

## 下一步

- 写法：[元素与视图](el.md)；组件：[kit 组件](kit.md)
- 后台任务怎么更新界面：[架构 · 线程规则](architecture.md#线程规则)
- 多窗口和快捷键：[窗口与应用](app.md)
