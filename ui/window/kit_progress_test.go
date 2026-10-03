package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"math"
	"testing"
	"time"
)

func TestProgressStylesAnimationAndSemantics(t *testing.T) {
	old := theme.ReducedMotion
	defer core.Update(func() { theme.SetReducedMotion(old) })
	now := time.Unix(1000, 0)
	p := kit.Progress("Transfer").Height(20).Rounded(0).Color(color.NRGBA{R: 255, A: 255}).TrackStyle(func(e *el.DivEl) { e.Bg(color.NRGBA{B: 255, A: 255}).Border(1, color.NRGBA{G: 255, A: 255}) })
	p.SetValue(.5)
	root := el.Embed(p)
	w := openTest(t, Options{Width: 300, Height: 160, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
	capture := func() []byte {
		t.Helper()
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	e := element(t, w, "Transfer")
	if e.Role != "progressbar" || e.Value != "50%" {
		t.Fatal(e)
	}
	im, err := png.Decode(bytes.NewReader(capture()))
	if err != nil {
		t.Fatal(err)
	}
	y := e.Y + e.Height - 10
	filled := color.NRGBAModel.Convert(im.At(e.X+e.Width/4, y)).(color.NRGBA)
	track := color.NRGBAModel.Convert(im.At(e.X+e.Width*3/4, y)).(color.NRGBA)
	if filled.R < 240 || filled.G > 15 || track.B < 240 || track.R > 15 {
		t.Fatal("custom track/fill", filled, track, e)
	}
	p.Height(4).Rounded(2).TrackStyle(nil)
	if small := element(t, w, "Transfer"); small.Height >= e.Height {
		t.Fatal("height not applied", small, e)
	}
	core.Update(func() { p.SetIndeterminate(true); theme.SetReducedMotion(false) })
	a := capture()
	now = now.Add(500 * time.Millisecond)
	b := capture()
	if bytes.Equal(a, b) {
		t.Fatal("indeterminate did not animate")
	}
	if e := element(t, w, "Transfer"); e.Value != "indeterminate" {
		t.Fatal(e)
	}
	core.Update(func() { theme.SetReducedMotion(true) })
	a = capture()
	now = now.Add(500 * time.Millisecond)
	b = capture()
	if !bytes.Equal(a, b) {
		t.Fatal("reduced motion still animated")
	}
	p.SetValue(float32(math.NaN()))
	if e := element(t, w, "Transfer"); e.Value != "0%" || p.Value() != 0 {
		t.Fatal("invalid value", e)
	}
}
