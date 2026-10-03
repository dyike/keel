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

func TestSwitchThumbMotion(t *testing.T) {
	old := theme.ReducedMotion
	oldPalette := theme.Current()
	defer core.Update(func() { theme.SetReducedMotion(old); theme.Apply(oldPalette) })
	for _, size := range []kit.SwitchSize{kit.SwitchMedium, kit.SwitchSmall} {
		core.Update(func() { theme.SetReducedMotion(false); theme.Apply(theme.Light()) })
		now := time.Unix(1000, 0)
		s := kit.Switch("Task", false).Size(size)
		root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(s.Render(cx)) }))
		w := openTest(t, Options{Width: 200, Height: 100, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
		center := func() float64 {
			e := element(t, w, "Task")
			b, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			width, height := 36, 20
			if size == kit.SwitchSmall {
				width, height = 28, 16
			}
			count, sum := 0, 0
			for y := e.Y + e.Height/2 - height/2 + 2; y < e.Y+e.Height/2+height/2-2; y++ {
				for x := e.X + 2; x < e.X+width-2; x++ {
					if color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA) == theme.Surface {
						count++
						sum += x
					}
				}
			}
			if count == 0 {
				t.Fatal("thumb pixels missing", e, theme.Surface, im.At(e.X+10, e.Y+e.Height/2))
			}
			return float64(sum) / float64(count)
		}
		start := center()
		s.SetValue(true)
		if got := center(); got != start {
			t.Fatal("jumped at start", start, got, theme.ReducedMotion)
		}
		e := element(t, w, "Task")
		if e.Checked == nil || !*e.Checked {
			t.Fatal("delayed semantics")
		}
		now = now.Add(kit.SwitchDuration / 2)
		middle := center()
		if middle <= start {
			t.Fatal("thumb did not move")
		}
		s.SetValue(false)
		if center() != middle {
			t.Fatal("retarget jumped")
		}
		now = now.Add(kit.SwitchDuration)
		if center() != start {
			t.Fatal("reverse did not settle")
		}
		core.Update(func() { theme.SetReducedMotion(true); s.SetValue(true) })
		end := center()
		if end <= middle {
			t.Fatal("reduced motion did not snap")
		}
		now = now.Add(kit.SwitchDuration)
		if center() != end {
			t.Fatal("reduced motion still moving")
		}
	}
}
