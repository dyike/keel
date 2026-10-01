package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"testing"
	"time"
)

func TestSkeletonAnimationAndReducedMotion(t *testing.T) {
	for _, shimmer := range []bool{false, true} {
		t.Run(map[bool]string{false: "pulse", true: "shimmer"}[shimmer], func(t *testing.T) {
			old := theme.ReducedMotion
			defer core.Update(func() { theme.SetReducedMotion(old) })
			now := time.Unix(0, 0)
			v := kit.Skeleton().W(el.Dp(120)).H(el.Dp(30))
			if shimmer {
				v.Shimmer()
			}
			r := el.Embed(v)
			w := openTest(t, Options{Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return r.Layout(gtx) })})
			capture := func() []byte {
				t.Helper()
				b, e := w.screenshot()
				if e != nil {
					t.Fatal(e)
				}
				return b
			}
			core.Update(func() { theme.SetReducedMotion(false) })
			a := capture()
			now = now.Add(750 * time.Millisecond)
			b := capture()
			if bytes.Equal(a, b) {
				t.Fatal("animation did not change pixels")
			}
			core.Update(func() { theme.SetReducedMotion(true) })
			a = capture()
			now = now.Add(375 * time.Millisecond)
			b = capture()
			if !bytes.Equal(a, b) {
				t.Fatal("reduced motion moved")
			}
			if len(w.snapshot()) != 0 {
				t.Fatal("decorative skeleton exposed semantics")
			}
		})
	}
}
