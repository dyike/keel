package el

import (
	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"math"
	"testing"
)

func TestMeasureTextMatchesLayout(t *testing.T) {
	for _, scale := range []float32{1, 1.6364, 2} {
		var measuredW, measuredH, actualW, actualH float32
		root := Root(ViewFunc(func(cx *Context) Element {
			measuredW, measuredH = cx.MeasureText("订单 Completed orders", 14, font.Bold)
			label := Text("订单 Completed orders").TextSize(14).Weight(font.Bold).MaxLines(1)
			return Div().Items(Start).Child(label).Decorate(func(gtx core.C, draw func()) {
				actualW, actualH = cx.LayoutSize(label)
				draw()
			})
		}))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: 1.25, PxPerSp: 1.25 * scale}
			gtx.Constraints.Max = image.Pt(1000, 400)
			root.Layout(gtx)
		})
		h.Frame()
		if math.Abs(float64(measuredW-actualW)) > 1 || math.Abs(float64(measuredH-actualH)) > 1 {
			t.Fatalf("scale %v: measured %vx%v, laid out %vx%v", scale, measuredW, measuredH, actualW, actualH)
		}
	}
}
