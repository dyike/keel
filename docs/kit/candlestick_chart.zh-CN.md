# CandlestickChart

[English](candlestick_chart.md) | 简体中文

按传入顺序显示开盘、最高、最低和收盘值。

```go
chart := kit.CandlestickChart(
    kit.Candle{Label: "周一", Open: 10, High: 14, Low: 8, Close: 12},
    kit.Candle{Label: "周二", Open: 12, High: 15, Low: 9, Close: 10},
).Title("价格").Height(240)
```

上涨使用成功色空心实体，下跌使用危险色实心实体，平盘显示横线。方向同时由形状表达。悬停可读取原始 OHLC 数值，数据表可通过键盘操作及 Agent 读取。

`SetData` 复制数据。四个值必须有限，并满足最低值不高于开盘和收盘、最高值不低于开盘和收盘；无效数据留出缺口，表格显示 `—`。支持负值；图表不假定数据来自股票。`Format(fn)` 修改数值格式，`SetDisabled` 禁用交互。

密集数据按横向像素分桶，保留首个有效开盘、最后有效收盘、最高价和最低价。全部无效的桶留空；含部分无效数据的桶按有效数据聚合。数据表与悬停仍使用原始数据，不受绘图聚合影响。

Agent：外层 `figure`，名称是标题，表格每行按分类、开、高、低、收排列。框架列名跟随 locale。

验证：`go run ./examples/components -section candlestick_chart`，加 `-theme dark` 检查深色。测试覆盖数据复制、非法关系、非有限数、负值极值聚合、悬停、禁用和 Agent 数据表。
