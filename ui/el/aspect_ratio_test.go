package el

import (
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"math"
	"testing"
)

func TestAspectRatioResolvedWidthAndExplicitHeight(t *testing.T) {
	for _, scale := range []int{1, 2} {
		ratio := float32(2)
		height, maxWidth, maxHeight := Auto, Auto, Auto
		var box *DivEl
		root := Root(viewFunc(func(*Context) Element {
			box = Div().W(Dp(100)).H(height).MaxW(maxWidth).MaxH(maxHeight).AspectRatio(ratio).Child(Div().H(Dp(12)))
			return Div().Items(Start).Child(box)
		}))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(200*scale, 300*scale)
			root.Layout(gtx)
		})
		check := func(w, height int) {
			t.Helper()
			h.Frame()
			if box.n.size != image.Pt(w*scale, height*scale) {
				t.Fatal("aspect size", box.n.size, w, height)
			}
		}
		check(100, 50)
		height = Dp(30)
		check(100, 30)
		height = Auto
		maxWidth = Dp(60)
		check(60, 30)
		maxHeight = Dp(20)
		check(60, 20)
		maxHeight = Auto
		ratio = 0
		check(60, 12)
		ratio = float32(math.NaN())
		check(60, 12)
	}
}
