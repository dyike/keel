# PieChart

饼图和环形图显示同一组数据的占比。

```go
chart := kit.PieChart(
    kit.PieSlice{Name: "线上", Value: 60},
    kit.PieSlice{Name: "门店", Value: 40},
).Title("收入来源").Donut(.6)
```

`Height(dp)` 设置绘图区高度；`Donut(fraction)` 设置中心孔半径比例，范围 0–0.9，默认 0。`Format(fn)` 设置数值格式，占比保留一位小数。`SetData` 复制数据并恢复图例可见状态。

图例可点击或用空格/Enter 切换，对剩余可见数据重新计算比例。悬停扇区显示名称、原始值和占比；中心孔不命中。颜色按原始索引分配，隐藏不会改变其余项颜色。`SetDisabled` 禁用图例、表格切换和悬停。

仅有限正数参与扇区；0 没有扇区，负数、NaN、Inf 在表格中显示为 `—`。全空、全部隐藏或全部无效时显示空状态。比例先按最大值缩放，两个最大浮点数也不会使总量溢出。数据表保留所有项，隐藏项占比为 0。

Agent：外层 `figure`，名字是标题；图例是带选中状态的 `toggle`。切换到数据表可读取名称、值和占比。

验证：`go run ./examples/components -section pie_chart`，加 `-theme dark` 检查深色。测试覆盖扇区与中心孔命中、图例重算、禁用、数据复制、空数据、异常值和数据表。
