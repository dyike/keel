package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
	"time"
)

type progressControl interface {
	el.View
	SetValue(float32)
	SetIndeterminate(bool)
}

func TestProgressValueTransitions(t *testing.T) {
	old := theme.ReducedMotion
	defer core.Update(func() { theme.SetReducedMotion(old) })
	red := color.NRGBA{R: 255, A: 255}
	for _, p := range []progressControl{kit.Progress("Task").Color(red), kit.ProgressCircle("Task").Color(red).Size(64)} {
		core.Update(func() { theme.SetReducedMotion(false); p.SetValue(.2) })
		now := time.Unix(1000, 0)
		root := el.Embed(p)
		w := openTest(t, Options{Width: 300, Height: 180, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
		count := func() int {
			t.Helper()
			b, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			n := 0
			for y := 0; y < im.Bounds().Max.Y; y++ {
				for x := 0; x < im.Bounds().Max.X; x++ {
					c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
					if c.R > 240 && c.G < 15 && c.B < 15 {
						n++
					}
				}
			}
			return n
		}
		first := count()
		p.SetValue(.8)
		if count() != first {
			t.Fatal("value jumped at transition start")
		}
		if e := element(t, w, "Task"); e.Value != "80%" {
			t.Fatal("semantics should report target", e)
		}
		now = now.Add(kit.ProgressDuration / 2)
		middle := count()
		if middle <= first {
			t.Fatal("transition did not advance")
		}
		p.SetValue(.1)
		if count() != middle {
			t.Fatal("retarget jumped")
		}
		now = now.Add(kit.ProgressDuration / 2)
		reverse := count()
		if reverse >= middle {
			t.Fatal("retarget did not reverse")
		}
		now = now.Add(kit.ProgressDuration / 2)
		last := count()
		if last >= first {
			t.Fatal("transition did not reach smaller target")
		}
		core.Update(func() { theme.SetReducedMotion(true); p.SetValue(1) })
		full := count()
		if full <= middle {
			t.Fatal("reduced motion did not jump to target")
		}
		now = now.Add(kit.ProgressDuration)
		if count() != full {
			t.Fatal("reduced motion changed later")
		}
		core.Update(func() { theme.SetReducedMotion(false); p.SetIndeterminate(true) })
		count()
		p.SetValue(.2)
		if count() != first {
			t.Fatal("leaving indeterminate should initialize target")
		}
	}
}
