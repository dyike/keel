# Table

数据表格：排序、选择、调整列宽、自定义单元格，只构建可见的行。

```go
t := kit.Table(kit.Col("单号").Width(120), kit.Col("客户").Flex(2), kit.Col("金额").Numeric()).
    Height(360).OnActivate(open)
t.SetRows(rows)
```

- 列：
  - `Col(title)` 默认等分宽度；`Flex(w)` 按比例分配，`Width(dp)` 固定宽度；
  - `Numeric()` 右对齐并按数字排序；`NoSort()` 禁止按这一列排序；
  - `Cell(fn)` 用自定义元素渲染单元格（比如标签、按钮），Agent 看到的仍是行的文字。
- 点击表头排序，再点一次反向；`SortBy(col, desc)` 用程序排序，`col` 为 -1 时恢复原始顺序。
- 拖动表头右边缘调整列宽，调整后该列变为固定宽度。拖动不会触发排序。
- 点击或用 ↑ ↓ Home End PageUp PageDown 选择，双击或回车激活。
- **回调、`Value`、`SetValue` 中的行号都是 `SetRows` 数据里的位置，不受排序影响。**
- `SetValue(i)` 会把选中行滚动到可见位置，表格不在屏幕上时会在显示后再滚动。
- `SetLoading(true)` 在行上显示加载动画，适合异步取数：数据到达后用 `core.Update` 调用 `SetRows` 和 `SetLoading(false)`。`Empty(text)` 设置无数据时的文字，默认用 locale 的"暂无数据"。
- 超出视口宽度时，用触控板或水平滚轮横向滚动；表头和数据同步移动，纵向仍按可见行构建。固定宽列保留设置的宽度，弹性列按比例填满剩余空间，每列至少 40dp；窗口不足时滚动。
- 调整列宽或窗口尺寸后会重新计算滚动范围。冻结列另行实现。

Agent：角色 `table`，`value` 是行数（如"36 行"）；表头是 `columnheader`；每行是 `row`，名字是各列用" | "连接，`selected` 表示选中。

验证：`go run ./examples/components -section table`，加 `-theme dark` 检查深色。
