package window

import (
	"github.com/dyike/keel/ui/core"
	"testing"
)

func TestWindowTextScaleOverride(t *testing.T) {
	for _, scale := range []float32{0, 1, 1.6364, 2} {
		var dp, sp float32
		w := openTest(t, Options{Width: 400, Height: 200, TextScale: scale, Content: core.Func(func(gtx core.C) core.D {
			dp, sp = gtx.Metric.PxPerDp, gtx.Metric.PxPerSp
			return core.D{}
		})})
		w.render()
		if dp == 0 {
			dp = 1
		}
		if sp == 0 {
			sp = 1
		}
		want := scale
		if want == 0 {
			want = 1
		}
		if sp/dp != want {
			t.Fatalf("scale=%v metric=%v/%v", scale, sp, dp)
		}
	}
}
