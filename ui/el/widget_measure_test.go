package el

import (
	"image"
	"testing"

	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
)

func TestFixedWidgetAvoidsDiscardedDrawing(t *testing.T) {
	for _, full := range []bool{false, true} {
		var ops op.Ops
		measured, painted := 0, 0
		var paintSize image.Point
		root := Root(ViewFunc(func(*Context) Element {
			w := Widget(core.Func(func(gtx core.C) core.D {
				if gtx.Ops == &ops {
					painted++
					paintSize = gtx.Constraints.Max
				} else {
					measured++
				}
				return core.D{Size: gtx.Constraints.Constrain(image.Pt(12, 15))}
			}))
			if full {
				w.WFull().HFull()
			} else {
				w.W(Dp(100)).H(Dp(60))
			}
			return Div().WFull().HFull().Child(w)
		}))
		ctx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Exact(image.Pt(200, 100))}
		for range 2 {
			ops.Reset()
			root.Layout(ctx)
		}
		want := image.Pt(100, 60)
		if full {
			want = image.Pt(200, 100)
		}
		if measured != 0 || painted != 2 || paintSize != want {
			t.Fatalf("full=%v measured=%d painted=%d size=%v", full, measured, painted, paintSize)
		}
	}
}

func TestWidgetAutoHeightStillMeasures(t *testing.T) {
	var ops op.Ops
	measured, painted := 0, 0
	var paintSize image.Point
	root := Root(ViewFunc(func(*Context) Element {
		return Div().WFull().HFull().Child(Widget(core.Func(func(gtx core.C) core.D {
			if gtx.Ops == &ops {
				painted++
				paintSize = gtx.Constraints.Max
			} else {
				measured++
			}
			return core.D{Size: gtx.Constraints.Constrain(image.Pt(12, 15))}
		})).W(Dp(100)))
	}))
	root.Layout(core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Exact(image.Pt(200, 100))})
	if measured == 0 || painted != 1 || paintSize != image.Pt(100, 15) {
		t.Fatalf("measured=%d painted=%d size=%v", measured, painted, paintSize)
	}
}
