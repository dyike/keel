# 无样式基础层：ui/base

[English](base.md) | 简体中文

`kit` 的组件外观是固定的。想要一个行为和 kit 一样、但长得完全不同的组件时，不用从头写键盘逻辑：`ui/base` 提供组件的行为，`el` 负责外观。两者组合起来，相当于 Radix、Headless UI 这类无样式组件库。

`base` 不绘制任何东西，也不依赖任何 Keel 模块，只是普通的状态和函数，可以单独测试。kit 自己的 List、Tree、Menu、Select、Command、Sidebar、Table 都建在它上面，所以自定义组件和 kit 组件的键盘手感一致。

## List：键盘导航

```go
nav := base.List{Count: len(items), Disabled: func(i int) bool { return items[i].Off }}
next, ok := nav.Key(e.Name, current) // ↑ ↓ Home End PageUp PageDown
```

- 跳过禁用项；没有可去的项时停在原地。
- 当前没有选中项（`-1`）时，↓ 到第一项，↑ 到最后一项。
- `Wrap: true` 时首尾相接，菜单用这个。`Page` 是翻页的步长，默认 10。
- `First`、`Last`、`Next(i, dir)` 单独可用，比如打开时聚焦第一个可用项。

## Typeahead：按首字母跳转

```go
var find base.Typeahead

if s, ok := base.Text(e.Name, e.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt) != 0); ok {
    if i, found := find.Find(time.Now(), s, current, nav, func(i int) string { return items[i].Label }); found {
        current = i
    }
    return true
}
```

- 连续输入 `r`、`e` 跳到 “Red”；停顿超过 1 秒（`TypeaheadPause`）重新开始。
- 重复按同一个字母，在以它开头的项之间循环。
- 不区分大小写，跳过禁用项。
- `Text` 把按键名转成输入的文字，排除 ↑、⏎ 这类用单个符号命名的功能键，以及带 Ctrl、Cmd、Alt 的快捷键。
- 浮层每次打开时调用 `Reset`，免得上次输入的字母影响这次。

kit 的 List、Tree、Menu、Select 都已经支持按首字母跳转。

## Selection：多选

```go
var sel base.Selection[string] // 按 ID 保存，排序、过滤后选择不丢

sel.Click(ids, i, mods.Contain(key.ModShift), mods.Contain(key.ModShortcut), disabled)
sel.Has(id)
sel.In(ids)      // 按 ids 的顺序列出选中的 ID
sel.Indexes(ids) // 选中项的下标
sel.Keep(func(id string) bool { return exists[id] }) // 数据变化后清掉不存在的
```

`Click` 按系统惯例处理：

| 操作 | 结果 |
| --- | --- |
| 单击 | 只选这一项，并把它设为范围起点 |
| Cmd 单击（其他平台 Ctrl） | 加入或移出这一项 |
| Shift 单击 | 从起点选到这一项，跳过禁用项 |
| Cmd + Shift 单击 | 把这段范围加到已有选择里 |

## Disclosure：打开/关闭

```go
d := base.Disclosure{OnChange: func(open bool) { ... }}
d.Toggle()     // 用户操作：状态变了才调 OnChange
d.Set(false)   // 程序设置：不调 OnChange
d.SetDisabled(true) // 禁用时同时关闭，之后打不开
```

区分“用户改的”和“程序改的”，和 kit 组件 `OnChange` 的约定一致：程序调用 `SetValue` 不触发回调。

## 例子：自定义外观的单选列表

```go
type swatches struct {
    colors []string
    nav    base.List
    find   base.Typeahead
    at     int
}

func (s *swatches) Render(cx *el.Context) el.Element {
    s.nav.Count = len(s.colors)
    box := el.Div().ID("swatches").Role("listbox").Name("颜色").Focusable(true).Row().Gap(8).
        OnKey(func(e el.KeyEvent) bool {
            if e.State != el.KeyPress {
                return true
            }
            if t, ok := base.Text(e.Name, false); ok {
                s.at, _ = s.find.Find(time.Now(), t, s.at, s.nav, func(i int) string { return s.colors[i] })
                return true
            }
            name := e.Name
            switch key.Name(name) { // 横向排列：← → 当作 ↑ ↓
            case key.NameLeftArrow:
                name = string(key.NameUpArrow)
            case key.NameRightArrow:
                name = string(key.NameDownArrow)
            }
            var ok bool
            s.at, ok = s.nav.Key(name, s.at)
            return ok
        })
    for i, c := range s.colors {
        box.Child(el.Div().Role("option").Name(c).Selected(i == s.at).Size(el.Dp(28)).Rounded(theme.RadiusFull).
            Border(2, theme.Border).When(i == s.at, func(d *el.DivEl) { d.Border(2, theme.Primary) }).
            OnClick(func() { s.at = i }))
    }
    return box
}
```

外观全部由 `el` 决定；键盘、跳转的规则与 kit 的列表一致。组件库示例的 Headless 页（`examples/components/headless.go`）有可运行的完整版本，还包括一个用 `Selection` 做的多选列表。
