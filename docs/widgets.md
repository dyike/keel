# 组件与布局

组件在 `ui/widget`，容器在 `ui/layout`，颜色字号在 `ui/theme`。它们都实现 `ui/core` 里的同一个接口：

```go
// package ui
type Widget interface {
    Layout(gtx core.C) core.D   // C = layout.Context, D = layout.Dimensions
}
```

所以任何组件都能放进任何容器，自己写的组件也一样。

组件构造函数返回指针，**把指针存下来**，之后在回调里通过它读值、改值。配置项用链式调用，在创建时一次写完：

```go
name := widget.Input("用户名").Hint("字母或数字").MaxLength(32).OnSubmit(login)
```

## 文字

| 构造 | 样式 |
| --- | --- |
| `widget.Text(s)` | 正文，15sp，`theme.Text` 色 |
| `widget.Heading(s)` | 标题，22sp，粗体 |
| `widget.Muted(s)` | 次要文字，13sp，`theme.Muted` 色 |

方法：`Text() string`、`SetText(s)`。文字按可用宽度自动换行，`\n` 换行。

## 按钮

```go
save := widget.Button("保存", onSave)                   // 主按钮，蓝底白字
widget.Button("取消", onCancel).Secondary()              // 浅灰底
widget.Button("删除", onDelete).Danger()                 // 红底
```

| 方法 | 说明 |
| --- | --- |
| `Secondary()`、`Danger()` | 改样式，返回自身，用于链式调用 |
| `SetText(s)` | 改按钮文字 |
| `SetDisabled(bool)` | 禁用后变灰，点击不触发回调 |
| `SetOnClick(fn)` | 替换回调 |

按钮宽度随文字，放进 `layout.Column` 也不会被拉满。

## 链接

```go
widget.Link("查看文档", openDocs)
```

主色文字，鼠标悬停显示手形光标。方法：`SetText(s)`。

## 输入框

```go
widget.Input("邮箱")        // 单行，回车触发 OnSubmit
widget.TextArea("备注")     // 多行，最小高度 72dp，回车换行
widget.Input("")            // 不要标签
```

| 链式配置 | 说明 |
| --- | --- |
| `Hint(s)` | 空内容时的灰色占位文字 |
| `Password()` | 用 `•` 遮盖内容 |
| `MaxLength(n)` | 最多 n 个字符（按 Unicode 码点计，一个汉字算 1） |
| `OnChange(func(string))` | 每次内容变化时调用，参数是新内容 |
| `OnSubmit(func(string))` | 单行输入框按回车时调用 |

| 方法 | 说明 |
| --- | --- |
| `Value() string` | 当前内容 |
| `SetValue(s)` | 替换内容；不会触发 `OnChange` |
| `SetReadOnly(bool)` | 只读：可以选中、复制，不能编辑 |

输入框获得焦点时边框变成 `theme.Primary` 色。输入校验（非空、格式）在回调里自己做：

```go
widget.Button("提交", func() {
    if strings.TrimSpace(email.Value()) == "" {
        errText.SetText("邮箱不能为空")
        return
    }
    ...
})
```

## 复选框

```go
agree := widget.Checkbox("同意条款", false).OnChange(func(v bool) { submit.SetDisabled(!v) })
```

方法：`Value() bool`、`SetValue(bool)`（不触发 `OnChange`）。点击图标或文字都能切换。

## 布局容器

| 函数 | 说明 |
| --- | --- |
| `layout.Column(children...)` | 纵向排列，间距 12dp。子组件宽度被拉满（按钮、复选框、链接除外，它们保持自身宽度） |
| `layout.Row(children...)` | 横向排列，间距 8dp，垂直居中 |
| `layout.Card(children...)` | 白底圆角卡片，1dp 边框，内边距 18dp，内部按 `Column` 排列 |
| `layout.Grow(w)` | 在 `Row` 里占满剩余宽度 |
| `layout.Divider()` | 通栏 1dp 分隔线 |
| `layout.Space(dp)` | 空白，在 `Column` 里是高度，在 `Row` 里是宽度 |

常见组合：

```go
// 输入框和按钮同一行，输入框吃掉剩余宽度
layout.Row(layout.Grow(widget.Input("")), widget.Button("搜索", search))

// 按钮靠右
layout.Row(layout.Grow(layout.Space(0)), widget.Button("取消", c).Secondary(), widget.Button("确定", ok))
```

`layout.Grow` 放在 `Column` 里没有意义：窗口的根视图可以滚动，纵向没有"剩余高度"。

`layout.Frame(gtx, bg, border, radius, inset, w)` 不是容器，是写组件时用的绘制工具：画圆角、边框、背景，再把内容放进内边距里。输入框和卡片都用它。

## 主题

在打开第一个窗口前修改：

```go
theme.Primary = theme.RGB(0x16a34a) // 换成绿色
```

| `theme` 的变量 | 默认值 | 用在哪里 |
| --- | --- | --- |
| `Bg` | `#f5f6f8` | 窗口背景 |
| `Surface` | `#ffffff` | 卡片、输入框底色 |
| `Border` | `#e3e5e8` | 边框、分隔线 |
| `Text` | `#1f2328` | 正文 |
| `Muted` | `#6b7280` | 次要文字、占位文字、未勾选图标 |
| `Primary` | `#2563eb` | 主按钮、链接、焦点边框、已勾选图标 |
| `Danger` | `#dc2626` | 危险按钮 |
| `Subtle` | `#eceef1` | 次要按钮底色 |
| `OnColor` | `#ffffff` | 主按钮、危险按钮上的文字 |

| 常量 | 值 | 说明 |
| --- | --- | --- |
| `BodySize` / `SmallSize` / `HeadingSize` | 15 / 13 / 22 sp | 字号 |
| `FontFace` | `PingFang SC, Hiragino Sans GB, Microsoft YaHei, Noto Sans CJK SC, Noto Sans SC, Go` | 字体优先级，逐字形回退 |
| `CJKNudge` | 2dp | 按钮、输入框里文字下移的量，见[常见问题](troubleshooting.md#按钮里的中文偏上) |

`theme.Material` 是底层的 Gio `material.Theme`，提供字形排版器和图标。写自定义组件时用它创建 `material.Label` 等。

## 直接写 Gio 代码

现成组件不够用、又不值得正式做成组件时，用 `core.Func` 把一段 Gio 布局代码当组件用：

```go
import (
    giolayout "gioui.org/layout" // 和 Keel 的 layout 同名，起别名
    "gioui.org/widget/material"

    "github.com/dyike/keel/ui/core"
    "github.com/dyike/keel/ui/layout"
    "github.com/dyike/keel/ui/theme"
    "github.com/dyike/keel/ui/widget"
)

badge := core.Func(func(gtx core.C) core.D {
    return giolayout.UniformInset(4).Layout(gtx, material.Caption(theme.Material, "beta").Layout)
})
layout.Row(widget.Heading("新功能"), badge)
```

`core.Func` 里有状态（比如 Gio 的 `widget.Clickable`）时，状态要放在闭包外面，否则每帧都会重建。需要复用第二次，就按[扩展指南](extending.md#新增组件)把它做成正式组件。
