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

func TestSpinnerFrameTimeAndReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer core.Update(func() { theme.SetReducedMotion(old) })
	now := time.Unix(1000, 0)
	spinner := kit.Spinner().Size(32).Label("加载中")
	root := el.Embed(spinner)
	w := openTest(t, Options{Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
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
	now = now.Add(250 * time.Millisecond)
	b := capture()
	if bytes.Equal(a, b) {
		t.Fatal("frame time did not animate")
	}
	core.Update(func() { theme.SetReducedMotion(true) })
	a = capture()
	now = now.Add(250 * time.Millisecond)
	b = capture()
	if !bytes.Equal(a, b) {
		t.Fatal("reduced motion still animated")
	}
	core.Update(func() { spinner.Icon(kit.IconClock).Color(color.NRGBA{R: 255, A: 255}); theme.SetReducedMotion(false) })
	a = capture()
	now = now.Add(250 * time.Millisecond)
	b = capture()
	if bytes.Equal(a, b) {
		t.Fatal("custom icon did not rotate")
	}
	core.Update(func() { theme.SetReducedMotion(true) })
	a = capture()
	now = now.Add(250 * time.Millisecond)
	b = capture()
	if !bytes.Equal(a, b) {
		t.Fatal("custom icon ignored reduced motion")
	}
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	red := 0
	for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
			c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
			if c.R > 240 && c.G < 15 && c.B < 15 {
				red++
			}
		}
	}
	if red == 0 {
		t.Fatal("custom color missing")
	}
	core.Update(func() { spinner.Icon(kit.IconNone) })
	if bytes.Equal(b, capture()) {
		t.Fatal("ring not restored")
	}
	e := element(t, w, "加载中")
	if e.Role != "progressbar" || e.Value != "indeterminate" {
		t.Fatal(e)
	}
}
