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
- 拖动表头右边缘调整列宽，调整后该列变为固定宽度。横向拖动表头主体可换序，指向目标列前半部插到其前，后半部插到其后；插入位置显示主题色细线，松手应用，移出表头区域或取消手势则放弃。拖动不会触发排序。停在非冻结表头区域左右边缘 32dp 内会持续横向滚动，越靠边速度越快（最高 360dp/s）；冻结列区域不触发滚动。松手或取消后停止，滚动期间会更新插入位置。
- 点击或用 ↑ ↓ Home End PageUp PageDown 选择，双击或回车激活。
- **回调、`Value`、`SetValue` 中的行号都是 `SetRows` 数据里的位置，不受排序影响。**
- `SetValue(i)` 会把选中行滚动到可见位置，表格不在屏幕上时会在显示后再滚动。
- `SetLoading(true)` 在行上显示加载动画，适合异步取数：数据到达后用 `core.Update` 调用 `SetRows` 和 `SetLoading(false)`。`Empty(text)` 设置无数据时的文字，默认用 locale 的"暂无数据"。
- 超出视口宽度时，用触控板或水平滚轮横向滚动；表头和数据同步移动，纵向仍按可见行构建。固定宽列保留设置的宽度，弹性列按比例填满剩余空间，每列至少 40dp；窗口不足时滚动。
- 调整列宽或窗口尺寸后会重新计算滚动范围。冻结列通过 `FrozenColumns` 配置。

Agent：角色 `table`，`value` 是行数（如"36 行"）；表头是 `columnheader`；每行是 `row`，名字是各列用" | "连接，`selected` 表示选中。

验证：`go run ./examples/components -section table`，加 `-theme dark` 检查深色。

表格构造时复制列配置，`SetRows` 复制二维数据，`Rows` / `Row` 返回副本。复用同一份列配置创建多个表格，拖动列宽不会互相影响。后续用 `SetRows` 更新数据，用 `SetColumnWidth(index, dp)` 修改某一张表的列宽；修改原始切片或构造时的 `ColumnSpec` 不会影响已创建的表格。

排序时虚拟行使用原始数据索引作为元素 key，自定义单元格中仍在构建范围内的输入/焦点状态随数据行移动。替换整个数据集后仍按新数据索引解释身份；跨数据集的业务状态应保存在应用模型中。

`FrozenColumns(left, right)` 固定开头 left 列与末尾 right 列，中间列横向滚动；表头、数据单元格、拖动列宽和点击区域保持一致。计数会限制在列数内，左侧优先，重复调用可调整或取消冻结。冻结列需固定宽度，原本的弹性列会转为 120dp，随后可用 `SetColumnWidth` 或拖动修改。窄视口装不下两侧时左侧优先，右侧被裁剪；中间列没有可用宽度时不绘制、不响应点击。纵向仍使用同一份虚拟列表，不复制数据行。

列管理使用源列索引（构造 `Table` 时的位置），移动或隐藏后不会改变这个索引：

- `MoveColumn(column, position)` 将源列移动到显示顺序中的 position，位置计入隐藏列；越界忽略。
- `OnColumnMove(func(column, from, to int))` 在用户拖动成功后回调；column 是源列索引，from/to 是含隐藏列的显示位置。程序 MoveColumn/SetLayoutState 不触发。
- `SetColumnVisible(column, visible)` 显示或隐藏源列，保留其位置和宽度。允许隐藏全部列，隐藏排序列不会取消排序。
- `LayoutState()` 返回独立的 `TableLayout` 快照，包含列顺序、宽度、弹性比例、隐藏状态和两侧冻结列数，可用 `encoding/json` 存到应用配置。
- `SetLayoutState(state)` 先检查列数、唯一索引、有限宽度和比例，再整体恢复；无效配置返回错误，保留原布局。布局仅适用于相同源列结构，应用应随业务数据结构管理配置版本。

冻结数量作用于当前可见顺序的两端；隐藏列后两侧不重叠，左侧优先。移动列、隐藏其他列时，仍在构建的自定义单元格保持元素身份；隐藏单元格本身会卸载，持久草稿仍需由应用模型保存。列操作不改变源数据、排序或行选择，也不触发行选择回调。示例提供移动、隐藏、保存和恢复按钮。

`MultiSelect()` 启用行多选：普通点击替换选区，Ctrl/Cmd 点击增减单行，Shift 点击或 Shift+方向键按当前排序扩展/缩小范围，Ctrl/Cmd+Shift 点击合并范围。`SelectedRows()` 返回按当前显示顺序排列的源行索引副本，`Value()` 表示活动行（可以已被取消选中）。`SetValue` 替换为单行选区，`SetSelectedRows` 替换多行选区，程序操作均不触发回调；`OnSelectionChange` 接收独立副本。数据缩短时移除越界选区，排序不改变选中的源行。

表格聚焦时 Ctrl/Cmd+A 全选多选表格的行，Ctrl/Cmd+C 复制选区。`SelectionText()` 返回相同的 TSV 内容：仅包含当前可见列，列和行都按显示顺序输出，单元格内的制表符、换行和双引号按 CSV 引号规则转义。输入框自行处理其文本选择和复制。

`CellSelect()` 切换到单元格模式并清空选区，`MultiSelect()` 可切回行模式。普通点击选单格，Ctrl/Cmd 点击增减单格，Shift 点击或方向键扩展矩形范围；方向键移动、Home/End 跳到行首/尾，Ctrl/Cmd+Home/End 跳到整表首/尾，PageUp/PageDown 移动八行，并滚动露出目标单元格。冻结列保持固定。

此模式点击表头选整列，Shift 点击选连续多列，Ctrl/Cmd 点击增减整列，双击表头排序。`TableCell{Row, Column}` 使用源数据索引，`SelectedCells()` 返回按显示顺序排列的副本，`SetSelectedCells` 程序赋值不触发 `OnCellSelectionChange`。`SetSelectedRows` 选中指定行的全部可见单元格；`SetValue` 选中该行首个可见单元格。隐藏列保留选区，复制时只输出参与选区的可见列；稀疏选区的未选交叉单元格输出空值。修改源数据长度后清理越界选区。

`SetFilter(func(row []string) bool)` 在排序前过滤源数据，传入的行是副本；`nil` 清除筛选。筛选条件变化后需重新调用此方法。`Len()` 是源行数，`VisibleLen()` 是筛选后行数；源数据索引和选区保留，筛选掉的行不参与复制或全选。

分页加载使用 `OnLoadMore(fn)` 和 `SetHasMore(true)`：距离底部两行以内自动请求，组件在回调前设为 loading，同一份数据最多自动请求一次。异步结果通过 `core.Update` 交付：`SetRows` 更新全部已加载数据，`SetLoading(false)` 结束请求，末页再 `SetHasMore(false)`。失败调用 `SetLoadError(message)`，停止自动重试，用户点击重试后重新调用 `OnLoadMore`。禁用或隐藏时不自动加载。筛选后内容不足一屏也会继续分页，应用必须正确标记末页。示例 `go run ./examples/components -section table_data` 模拟首次加载失败、重试和三页数据；状态提示始终位于视口内。

`RowMenu(func(row int) *MenuView)` / `CellMenu(func(row, column int) *MenuView)` 在右键请求时构造菜单，参数始终是源数据索引。单元格菜单优先，返回 nil 则使用行菜单；菜单贴着目标单元格弹出。右键已选中的行/格会保留现有多选，未选中的目标先成为选区。Shift+F10 为活动行/格打开菜单；Esc、外部点击或执行命令关闭，键盘打开后恢复原焦点。目标被筛掉、隐藏或禁用时关闭菜单。菜单复用 `Menu` 的子菜单与键盘导航。

性能：30 万行的表格在测试里建表、排序、跳到末尾不到一秒，滚动每帧约 1 毫秒（`TestTableThreeHundredThousandRows`、`BenchmarkTableFrame300k`）。

## 静态组合表格

少量数据或自定义布局可使用 `StaticTable`。它返回可直接设置样式的 `el.DivEl`，不创建数据表的选择、排序或虚拟列表状态。各部件在 Render 中构建；有状态的 Input、Button 等视图由应用持有，再把 Render 结果放入单元格。

```go
kit.StaticTable().Name("订单").Child(
    kit.TableHeader().Child(kit.TableRow().Child(
        kit.TableHead().Child(el.Text("单号")),
        kit.TableHead().Items(el.End).Child(el.Text("金额")),
    )),
    kit.TableBody().Child(kit.TableRow().Child(
        kit.TableDataCell().Child(el.Text("SO-001")),
        kit.TableDataCell().Items(el.End).Child(el.Text("¥250.00")),
    )),
    kit.TableFooter().Child(kit.TableRow().Decorate(nil).Child(
        kit.TableDataCell().Child(el.Text("合计")),
        kit.TableDataCell().Items(el.End).Child(el.Text("¥250.00")),
    )),
    kit.TableCaption().Child(el.Text("最近一笔订单")),
)
```

Header、Body、Footer 均可放任意数量的 Row；每个 Row 可放任意数量的单元格，也可插入完全自定义元素。Header/Footer 使用弱背景，Row 默认绘制底部分隔线，Footer 默认绘制顶线；`Decorate(nil)` 移除对应默认分隔线。根容器默认带边框、圆角和表面背景，可通过 Border/Rounded/Bg 覆盖。Caption 接受文字或富内容，默认在容器内显示，位置由 Child 顺序决定。

单元格默认平分行宽；`Flex(0).W(el.Dp(120)).NoShrink()` 固定列宽，`Flex(2)` 调整弹性比例。各行独立布局，表头、数据和汇总行应使用一致的列配置。单个单元格即可占满一整行；不做跨行合并或自动测量全表列宽。`Items(el.Center)` / `Items(el.End)` 设置内容水平对齐，P/Px/Py 调整留白；文字粗细等可在传入的 Text 上设置。需要长内容滚动时，在外层组合 ScrollX/ScrollY。

语义角色包括 table、rowgroup、row、columnheader、cell 和 caption。静态行没有默认选择、激活或键盘导航，子控件独立接收事件；需要整行操作时可显式配置 OnClick/Focusable/OnKey。原有 `TableCell` 是数据表选区坐标类型，因此静态单元格入口命名为 `TableDataCell`。

运行 `go run ./examples/components -section table_static` 查看交互、通栏备注及汇总示例。1×/2× 自动测试覆盖部分固定列的三段对齐、页尾/说明位置、子按钮事件和输入状态；本批未做真机视觉验收。

## 整列选择、列限制和密度

`ColumnSelect()` 进入独立的整列模式，清除原行/单元格选区。点击表头或数据单元格选择所属列，Ctrl/Cmd 点击增减，Shift 点击扩展连续列范围；左右键跳到相邻可选列，Home/End 跳到首尾，Shift 扩展，Ctrl/Cmd+A 选择全部可见且可选的列。双击表头继续排序。整列选择不触发行选择或行激活回调。

`SelectedColumns()` 返回显示顺序的源列索引副本，包含隐藏的已选列；`SetSelectedColumns([]int)` 静默赋值，`OnColumnSelectionChange` 接收用户变更。选区按列保存，SetRows 的新增行自然属于已选列，空表也可选择列。过滤不丢选区；复制只输出可见已选列和过滤后的行。`SelectedCells()` 可展开为当前过滤结果中的单元格快照，调用时成本随行数增长。`SetValue` / `SetSelectedRows` / `SetSelectedCells` 在整列模式下不改变选区，使用列接口设置。CellSelect/MultiSelect 可退出该模式。

列配置支持：

- `Selectable(false)`：禁止该列参与单元格/整列选择，范围选择、键盘、全选和程序选区赋值均跳过；不影响整行选择或列排序。
- `Resizable(false)`：移除用户拖动列宽的手柄。应用仍可用 SetColumnWidth 或布局恢复设置宽度。
- `Movable(false)`：MoveColumn 不移动该列，也不允许其他列跨过它改变其位置。显式 SetLayoutState 仍可恢复应用指定的完整布局。表头拖动遵守同样的锁定规则。

列配置在构造表格时复制。示例中单号列锁定宽度拖动和移动，表头和按钮均可移动其他列。

`Stripe(true)` 按过滤/排序后的可见行序交替填充弱背景；选区高亮优先。`RowHeight(dp)` 同步设置实际行高和虚拟列表尺寸，范围 24–256dp，含 1dp 分隔线，0 恢复 40dp。设置行高不改变单元格控件字号；自定义内容需适配行高。已有活动行会重新露出。

本批自动测试覆盖整列选择、范围跳过限制列、空表、数据追加、过滤/隐藏列复制、模式切换、禁用键盘、移动锁及 1×/2× 行高变化；真机拖动和视觉尚未验收。

表头拖动按实际绘制位置命中，支持已横向滚动的区域及左右冻结列；仅接受当前可见表头作为目标，不自动滚动到屏外列。隐藏列保留源索引并计入回调位置。禁用、隐藏列或程序恢复布局会取消正在进行的换序。1×/2× 事件测试覆盖换序、锁定、隐藏、冻结、滚动、取消和排序/列宽操作隔离，真机手感仍待验收。

整列选择不会把内嵌输入框的焦点强制移到表格。输入框可继续编辑，方向键由输入框处理；点击普通表头后，方向键仍用于列导航。自动测试覆盖单元格/整列两种模式、1×/2×、同帧与分帧点击，以及移动列后的文本保留。
