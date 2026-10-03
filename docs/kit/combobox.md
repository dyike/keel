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

`Multiple()` 把选项显示为可移除的标签，输入区留给下一次搜索；`AllowCustom()` 可与它组合。`Values()` 返回选择顺序的副本，`SetValues` 去重后替换选择，不触发回调；`OnValuesChange` 接收用户增删后的副本。多选下应使用 `Values()`，`Value()` 保留最近的主值。`SetValue` 替换成单项选择，程序赋值会关闭结果列表并使旧查询失效。构造与 `SetOptions` 复制选项切片；更新搜索结果不移除已选标签。

`OnSearch(func(query string, token uint64))` 接管结果来源。每次输入、打开或重试都有新 token；异步完成后用 `core.Update` 调用 `SetResults(token, options...)` 或 `SetSearchError(token, message)`。过期、已关闭的结果返回 false，远程结果不做二次本地筛选。加载或失败期间不能执行旧结果，失败后提供重试。关闭、程序赋值或禁用会使请求失效；禁用导致的失焦不会提交草稿。

结果使用虚拟列表和过滤缓存。↑ ↓、PageUp/PageDown 会滚动露出高亮项，结果区高度受窗口限制。示例包含万条客户、自由输入多标签，以及可模拟失败的异步多选。


`DisableOption(value, true)` 禁止用户选择指定值，候选仍显示并带禁用语义。点击、上下方向键、PageUp/PageDown、Enter 和失焦提交均遵守禁用；AllowCustom 不能绕过同值的禁用。全部候选禁用时没有高亮项，也不能提交候选。传 false 恢复可选。

配置按字符串值保存，过滤、SetOptions 和异步 SetResults 不会清掉配置；同值重复候选一起禁用，尚未出现在候选中的值也可预先配置。禁用已选值不会自动删除或触发回调，多选标签仍可移除。SetValue/SetValues 保留应用主动赋值能力。动态禁用当前高亮项时改为下一可用项，全部禁用则清除高亮。


多选列表中，再次点击或按 Enter 确认已选候选会取消选择，弹层保持打开；搜索文字清空，异步模式为清空后的查询申请新 token。每次切换触发一次 OnValuesChange；移除最近主值时 Value 回退到剩余选择的最后一项，无剩余则为空。OnChange 只在该主值变化时触发，移除其他已选项不会重复通知主值。禁用候选不能从列表切换，已选标签仍可用移除按钮删除。
