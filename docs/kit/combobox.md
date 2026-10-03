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


`Footer(view)` 在候选滚动区下方显示持久操作区，加载、失败、空结果时也保留；`Footer(nil)` 移除。内容可含按钮或输入框，候选更新保留同一内容实例的焦点和状态；footer 操作不选中候选、不自动关闭，应用可调用 SetValue/SetValues 应用新选择并关闭。禁用触发字段或祖先会关闭整个弹层。

footer 最多占窗口扣除 80dp 后高度的三分之一，超高内容在自己的区域滚动；候选区为 footer 预留高度。首帧按上限预留，测量后下一帧收敛到实际高度。示例的“添加示例标签”按钮演示了主动更新选择。


`Clearable(true)` 在有已选值时显示独立清空按钮，单选／多选均可用。用户清空会关闭弹层、取消旧查询、清除草稿与错误，焦点返回输入框；OnChange 收到空字符串，多选 OnValuesChange 收到空切片，各通知一次。回调里重新 SetValue 的结果保留；回调参数描述本次清空事件。只剩搜索草稿而没有选择时不显示按钮，禁用状态遵循字段与祖先；`Clearable(false)` 隐藏按钮。程序 SetValue/SetValues 继续不触发回调。


`Size(dp)` 设置字段最小高度并同步缩放字号、间距、展开／清空按钮、多选标签和候选行，建议 28／36／48dp；0 恢复默认，负数和非有限值忽略。多选内容可换行，实际高度可能更高，标签保留 16dp 最小高度。footer 维持自身尺寸。打开期间修改尺寸保留输入焦点并重新露出高亮候选。

`CheckIcon(kit.Icon(kit.IconCheck))` 替换选中候选的图标，也接受 VectorIcon；保存图标配置副本，颜色和基础尺寸来自传入图标，随 Size 缩放。未选中候选保留同宽占位。`CheckIcon(nil)` 恢复默认；传入 IconNone 隐藏图案并保留尺寸。


`SetItems(...ComboboxItem)` 接受 `Value`、`Label`、`Disabled`，将稳定值与显示名称分离。空 Value 忽略，空 Label 使用 Value，重复 Value 保留首项，输入切片会复制。本地搜索同时匹配名称和值；回调、Value/Values 及程序赋值使用稳定值，候选和多选标签显示名称。候选语义的 name 是显示名称，value 是稳定值。

`SetItemResults(token, items...)` 用于结构化异步结果，和 SetResults 一样校验请求状态，不做本地二次筛选。替换候选不会删除选择；已选值不在新结果时保留原名称，返回候选后更新名称。SetOptions/SetResults 恢复字符串模式。条目 Disabled 与 DisableOption 取逻辑或，修改条目状态须重新 SetItems/SetItemResults。

提交输入文字时，精确值优先，其次是首个名称匹配；同名不同值建议从候选选择。未改动的单选显示文字在失焦时保持原值，避免名称碰巧等于另一条目的值时误选。正在编辑的草稿不会被名称更新覆盖。
