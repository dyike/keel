# ui

界面模块集合。和 `native/` 一样，每个子目录是一个职责单一的模块。

| 模块 | 做什么 | 依赖 |
| --- | --- | --- |
| [core](core/) | 地基：`Widget` 接口、回调、线程规则（`Update`） | 无 |
| [theme](theme/) | 颜色、字号、字体 | 无 |
| [layout](layout/) | 摆放组件：`Column`、`Row`、`Card` 等 | core、theme |
| [widget](widget/) | 交互组件：`Button`、`Input`、`Checkbox` 等 | core、theme、layout |
| [window](window/) | 窗口：`Open`、`Main`、快捷键、截图 | core、theme |
| [el](el/) | GPUI 风格：视图 + 链式样式元素 + flexbox，新界面优先用它 | core、theme |

依赖只往下走，下层不知道上层存在：

```
el ─────────────────┐
window ─────────────┤
widget ──► layout ──┼──► theme
   │         │      │
   └─────────┴──────┴──► core
```

`el` 是 `layout` + `widget` 的新写法，两者可以混用（`el.Widget` 嵌旧组件，`el.Embed` 放进旧布局），旧组件会逐个迁移到 `el`。

`internal/` 下是只给这几个模块用的内部代码：`loop`（帧锁、更新队列）和 `uitest`（无界面测试工具），外部引用不到。

一个窗口通常要引入三个模块：

```go
import (
    "github.com/dyike/keel/ui/layout"
    "github.com/dyike/keel/ui/widget"
    "github.com/dyike/keel/ui/window"
)

window.Open(window.Options{Title: "Hello", Content: layout.Card(
    widget.Input("你的名字"),
    widget.Button("好", ok),
)})
window.Main()
```

后台 goroutine 改界面时再加 `core`（用 `core.Update`），改配色时加 `theme`。
