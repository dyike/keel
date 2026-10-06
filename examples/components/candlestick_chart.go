package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("candlestick_chart", "data", func() core.Widget {
		chart := kit.CandlestickChart(
			kit.Candle{Label: demoText("Monday", "周一"), Open: 100, High: 112, Low: 98, Close: 108},
			kit.Candle{Label: demoText("Tuesday", "周二"), Open: 108, High: 115, Low: 103, Close: 105},
			kit.Candle{Label: demoText("Wednesday", "周三"), Open: 105, High: 109, Low: 99, Close: 105},
			kit.Candle{Label: demoText("Thursday", "周四"), Open: 105, High: 121, Low: 104, Close: 118},
			kit.Candle{Label: demoText("Friday", "周五"), Open: 118, High: 120, Low: 106, Close: 109},
		).Title(demoText("OHLC · Hollow rising / filled falling", "开高低收 · 空心上涨 / 实心下跌")).Height(280)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).W(el.Dp(600)).MaxW(el.Full).Child(chart.Render(cx))
		}))
	})
}
