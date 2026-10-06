# CandlestickChart

English | [简体中文](candlestick_chart.zh-CN.md)

Displays the opening, high, low and closing values in the order they are passed in.

```go
chart := kit.CandlestickChart(
    kit.Candle{Label: "Monday", Open: 10, High: 14, Low: 8, Close: 12},
    kit.Candle{Label: "Tuesday", Open: 12, High: 15, Low: 9, Close: 10},
).Title("Price").Height(240)
```

Use the success-colored hollow real body when rising, use the dangerous-colored solid real body when falling, and display horizontal lines when flat. Direction is also expressed by shape. The original OHLC value can be read by hovering, and the data table can be read by keyboard operation and Agent.

`SetData` Copy data. The four values must be limited and meet the requirement that the lowest value is not higher than the opening and closing prices, and the highest value is not lower than the opening and closing prices; a gap is left for invalid data, and the table displays `—`. Negative values are supported; chart does not assume data comes from stocks. `Format(fn)` modifies the numerical format, `SetDisabled` disables interaction.

Dense data is bucketed by horizontal pixels, retaining the first valid opening, the last valid closing, the highest price and the lowest price. All invalid buckets are left empty; buckets containing some invalid data are aggregated according to valid data. Data tables and hovers still use the original data and are not affected by plot aggregation.

Agent: Outer layer `figure`, the name is the title, and each row of the table is arranged by category, open, high, low, and close. Frame column names follow the locale.

Verify: `go run ./examples/components -section candlestick_chart`, add `-theme dark` to check the dark theme. Tests cover data copying, illegal relationships, non-finite numbers, negative extreme value aggregation, hover, disable and Agent data tables.
