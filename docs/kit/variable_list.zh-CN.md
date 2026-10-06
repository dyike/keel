# VariableList

[English](variable_list.md) | 简体中文

```go
list := kit.VariableList(messageIDs, 72, func(cx *el.Context, i int) el.Element {
    return el.Text(messages[i].Body) // 自动换行，自然高度
}).Height(400)
```

`VariableList` 的第二个参数是尚未测量行的估计高度（dp）。每帧只构建可视区及上下各一屏附近的行，布局后缓存实际高度，用前缀和索引查询行位置；普通滚动不扫描全部数据。

- key 必须非空且唯一。`SetKeys(ids)` 复制新的顺序，保留仍存在的 key 的尺寸缓存；重复或空 key 会在修改前 panic。
- 头部插入或上方行高度变化时，保持首个可见 key 的屏幕位置；该行被删除则选择原索引附近的行。窗口宽度或缩放变化自动清空尺寸缓存并重新测量。
- 可见行的高度变化自动检测；修改屏幕外的数据或字体后调用 `Invalidate(keys...)`，不传 key 则使全部尺寸缓存失效。
- `ScrollTo(cx, i)` 按当前列表的零基索引定位；`ScrollToKey(cx, key)` 按稳定 key 定位，适用于插入、删除或重排后的消息跳转，不存在的 key 不改变当前定位请求。两者只滚动到目标行可见，初次显示和测量改变行高后继续校正。`ScrollToAlign(cx, i, align)` / `ScrollToKeyAlign(cx, key, align)` 把目标行放到顶部、居中或底部（`kit.ScrollStart/ScrollCenter/ScrollEnd`），未测量的行先按估算高度定位，测量后自动校正。`Height`、`Fill`、`Count`、`ID` 和 `SetDisabled` 与等高列表用法一致。
- 稳定 key 保持构建范围内的元素身份；长期离开构建范围的输入值等业务状态仍应保存在应用模型中。
- 未访问行使用估计高度，因此滚动条长度会随测量调整。表格等行高已知场景继续使用等高列表。

验证：`go run ./examples/components -section variable_list`。点击“定位消息 50000”，再连续执行“头部插入 10 条”并重新定位，应始终看到“消息 50000”。同时检查展开行内容和窄窗口换行。


`Horizontal(true)` 启用横向变宽虚拟列表；构造时 estimate 参数作为未测量宽度。`Width(dp)` 设置视口宽度（横向默认 320dp），`Height(dp)` 设置高度；横向 `Fill()` 在 Row 父布局中填满主轴。项目渲染器可以返回不同自然宽度或显式宽度。

切换轴保留首个可见稳定 key，把项目内偏移归零，并清空旧轴测量；不会把行高当成列宽。待处理的 ScrollTo/ScrollToKey 保留，布局后继续修正。`ScrollToEnd(cx)` 定位末项。横向插入、删除、重排和尺寸改变同样保持阅读锚点；水平模式下交叉轴高度和缩放变化使缓存失效。可见项目每帧重新测量，修改离屏内容仍调用 Invalidate。

```go
strip := kit.VariableList(ids, 160, func(cx *el.Context, i int) el.Element {
    return el.Div().W(el.Dp(widths[i])).Child(el.Text(labels[i]))
}).Horizontal(true).Width(640).Height(100)
```

同一列表只沿一个轴虚拟化。Keel 使用可见项目实测与未访问项目估算，上游通过调用方提供全部尺寸；滚动条长度可能随测量调整，定位采用最小露出策略。
