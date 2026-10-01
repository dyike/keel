# Combobox

可输入、可筛选的下拉框。

```go
customer := kit.Combobox("客户", customers...).Placeholder("输入筛选")
tags := kit.Combobox("标签", "紧急", "VIP").AllowCustom()
```

- 输入时打开列表并筛选（不区分大小写，包含即匹配），点击选项选中。
- 回车的规则：
  - 输入的文字正好是某个选项，就选它；
  - 设置了 `AllowCustom` 时，保留输入的文字；
  - 否则选第一个匹配项。
- 没有设置 `AllowCustom` 时，离开输入框后文字如果不是选项，会恢复为上一次的选择。
- ↓ 打开列表并移动高亮，↑ 往回移动，回车选中高亮项。
- `Value()` / `SetValue`、`SetOptions`、`SetDisabled`、`SetError`。需要 `el.Root`。

Agent：容器角色 `combobox`，`value` 为当前选择；里面有 `textbox` 和展开按钮；列表是 `listbox` 和 `option`。

验证：`go run ./examples/components -section combobox`，加 `-theme dark` 检查深色。
