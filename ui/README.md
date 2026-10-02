# ui

界面模块集合。和 `native/` 一样，每个子目录是一个职责单一的模块。

| 模块 | 做什么 | 依赖 |
| --- | --- | --- |
| [core](core/) | 地基：`Widget` 接口、回调、线程规则（`Update`） | 无 |
| [theme](theme/) | 颜色、字号、字体 | 无 |
| [locale](locale/) | 框架文字，中英文切换 | 无 |
| [base](base/) | 组件行为，不含外观：键盘导航、首字母跳转、多选、打开状态 | 无 |
| [el](el/) | GPUI 风格：视图 + 链式样式元素 + flexbox | core、theme、locale |
| [kit](kit/) | 组件：按钮、表单、表格、浮层、应用外壳、图表 | base、el、core、theme、locale |
| [window](window/) | 窗口：`Open`、`Main`、快捷键、截图 | core、theme |
| [markdown](markdown/) | Markdown 渲染，针对 AI 流式输出优化 | el、core、theme、locale |

依赖只往下走，下层不知道上层存在：

```
kit ──────► el ──┐     kit ──► base
markdown ──► el  ├──► theme、locale
window ──────────┤
                 └──► core
```

`internal/` 下是只给这几个模块用的内部代码：`loop`（帧锁、更新队列）、`editorstyle`（输入框绘制）、`imageload`（图片加载）和 `uitest`（无界面测试工具），外部引用不到。

一个窗口通常要引入三个模块：

```go
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

window.Open(window.Options{Title: "Hello", Content: el.Root(&hello{name: kit.Input("你的名字")})})
window.Main()
```

后台 goroutine 改界面时再加 `core`（用 `core.Update`），改配色时加 `theme`。
