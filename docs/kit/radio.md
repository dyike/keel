# Radio

选项分散在不同卡片、表格行里，用不了 RadioGroup 时，用单个 `kit.Radio(label)`：

```go
express, pickup := kit.Radio("快递配送"), kit.Radio("门店自提")
express.SetValue(true)
express.OnChange(func(bool) { pickup.SetValue(false) })
pickup.OnChange(func(bool) { express.SetValue(false) })
```

- 点击或空格选中；已选中时再点不会取消，所以互斥由应用在 OnChange 里处理。程序调用 `SetValue` 不触发回调。
- 外观和 RadioGroup 的选项一致：`Size`、`TextSize`，`Content` 用任意视图替换文字标签（构造参数仍是无障碍名称）。
- 禁用后点击无效；Agent 角色为 `radio`，选中时 `selected` 为 true。
- 一组普通的选项仍然推荐 RadioGroup：它自带方向键切换和 Tab 进入规则。

验证：`go run ./examples/components -section radio`。
