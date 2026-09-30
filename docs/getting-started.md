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

    "github.com/dyike/keel/ui/layout"
    "github.com/dyike/keel/ui/widget"
    "github.com/dyike/keel/ui/window"
)

func main() {
    name := widget.Input("你的名字").Hint("例如：小明")
    result := widget.Text("等待输入")

    greet := func() {
        if v := strings.TrimSpace(name.Value()); v != "" {
            result.SetText("你好，" + v + "！")
        }
    }
    name.OnSubmit(func(string) { greet() }) // 回车也触发

    window.Open(window.Options{
        Title: "Hello", Width: 480, Height: 320,
        Content: layout.Column(
            widget.Heading("第一个窗口"),
            layout.Card(name, widget.Button("打招呼", greet), result),
        ),
    })
    window.Main()
}
```

这段代码的结构就是所有 Keel 程序的结构：

1. **先创建组件，拿到它们的指针。** `name`、`result` 是持有状态的对象，回调里通过它们读值、改值。
2. **用容器把组件拼成一棵树。** `layout.Column`、`layout.Card`、`layout.Row` 只负责摆放。
3. **`window.Open` 把树放进窗口，`window.Main` 进入事件循环。** `window.Main` 不会返回；最后一个窗口关闭时进程退出。

界面不需要手动刷新。回调执行完，所有窗口会自动重绘。

## 在你自己的项目里使用

Keel 还没发布版本。在你的项目里用本地替换引用它：

```sh
go mod edit -require=github.com/dyike/keel@v0.0.0
go mod edit -replace=github.com/dyike/keel=$HOME/Code/go/src/github.com/dyike/keel
go mod tidy
```

## 下一步

- 组件的完整 API：[组件与布局](widgets.md)
- 后台任务怎么更新界面：[架构 · 线程规则](architecture.md#线程规则)
- 多窗口和快捷键：[窗口与应用](app.md)
