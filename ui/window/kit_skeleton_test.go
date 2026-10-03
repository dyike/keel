package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"
)

func TestSkeletonAnimationAndReducedMotion(t *testing.T) {
	for _, mode := range []struct{ shimmer, secondary bool }{{false, false}, {true, false}, {false, true}, {true, true}} {
		shimmer := mode.shimmer
		t.Run(map[bool]string{false: "pulse", true: "shimmer"}[shimmer], func(t *testing.T) {
			old := theme.ReducedMotion
			defer core.Update(func() { theme.SetReducedMotion(old) })
			now := time.Unix(0, 0)
			v := kit.Skeleton().W(el.Dp(120)).H(el.Dp(30)).Secondary(mode.secondary).Rounded(12)
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

func TestSkeletonSecondaryRadiusAndTheme(t *testing.T) {
	oldMotion, oldPalette := theme.ReducedMotion, theme.Current()
	core.Update(func() { theme.SetReducedMotion(true) })
	defer core.Update(func() { theme.SetReducedMotion(oldMotion); theme.Apply(oldPalette) })
	v := kit.Skeleton().W(el.Dp(40)).H(el.Dp(40)).Rounded(0)
	w := openTest(t, Options{Width: 160, Height: 160, Content: el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Name("probe").Role("image").WFull().H(el.Dp(96)).Bg(theme.Surface).Child(v.Render(cx))
	}))})
	capture := func() image.Image {
		t.Helper()
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return im
	}
	probe := element(t, w, "probe")
	pixel := func(im image.Image, x, y int) color.NRGBA {
		return color.NRGBAModel.Convert(im.At(probe.X+x, probe.Y+y)).(color.NRGBA)
	}
	for _, palette := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(palette); v.Secondary(false).Rounded(0) })
		solid := capture()
		center := pixel(solid, 20, 20)
		if pixel(solid, 1, 1) != center {
			t.Fatal("square corner missing")
		}
		core.Update(func() { v.Secondary(true) })
		faint := capture()
		if pixel(faint, 20, 20) == center || pixel(faint, 20, 20) == pixel(faint, 60, 60) {
			t.Fatalf("secondary opacity missing: solid=%v faint=%v background=%v", center, pixel(faint, 20, 20), pixel(faint, 60, 60))
		}
		core.Update(func() { v.Secondary(false).Rounded(20) })
		rounded := capture()
		if pixel(rounded, 1, 1) != pixel(rounded, 60, 60) || pixel(rounded, 20, 20) != center {
			t.Fatal("radius clipping")
		}
		core.Update(func() { v.Rounded(1e30) })
		large := capture()
		if pixel(large, 1, 1) != pixel(rounded, 1, 1) || pixel(large, 20, 20) != center {
			t.Fatal("large radius not clamped")
		}
	}
	if len(w.snapshot()) != 1 {
		t.Fatal("decorative geometry gained semantics")
	}
}
