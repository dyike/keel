package window

import (
	"bytes"
	"image/color"
	"image/png"
	"slices"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestToggleSegmentedFocus(t *testing.T) {
	calls := 0
	v := kit.ToggleGroup("A", "B", "C").Multiple().Segmented(true).Variant(kit.ToggleOutline).OnChange(func([]string) { calls++ })
	w := openTest(t, kitPage(v))
	w.press("Tab")
	w.press("Space")
	if !slices.Equal(v.Value(), []string{"A"}) {
		t.Fatal(v.Value())
	}
	for _, mode := range []bool{false, true, false, true} {
		v.Segmented(mode)
		w.snapshot()
		w.press("Space")
	}
	if !slices.Equal(v.Value(), []string{"A"}) || calls != 5 {
		t.Fatal("focus or callback changed", v.Value(), calls)
	}
	w.press("Tab")
	w.press("Space")
	if !slices.Equal(v.Value(), []string{"A", "B"}) {
		t.Fatal(v.Value())
	}
	v.Gap(8)
	w.snapshot()
	w.press("Space")
	if !slices.Equal(v.Value(), []string{"A"}) {
		t.Fatal("gap change lost focus", v.Value())
	}
	v.ResetGap().Size(kit.ToggleSizeLarge)
	w.snapshot()
	w.press("Space")
	if !slices.Equal(v.Value(), []string{"A", "B"}) {
		t.Fatal(v.Value())
	}
	v.SetDisabled(true)
	w.snapshot()
	w.press("Space")
	if calls != 8 {
		t.Fatal("disabled activated", calls)
	}
}

func TestToggleSegmentedCorners(t *testing.T) {
	old := theme.ReducedMotion
	core.Update(func() { theme.SetReducedMotion(true) })
	defer core.Update(func() { theme.SetReducedMotion(old) })
	bg := color.NRGBA{R: 120, G: 40, B: 80, A: 255}
	v := kit.ToggleGroup("Alpha", "Beta", "Gamma").Multiple().Segmented(true).Variant(kit.ToggleOutline)
	v.SetValue("Alpha", "Beta", "Gamma")
	w := openTest(t, Options{Width: 400, Height: 100, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(16).Bg(bg).Items(el.Start).Child(v.Render(cx)) }))})
	capture := func() func(int, int) color.NRGBA {
		t.Helper()
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return func(x, y int) color.NRGBA { return color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA) }
	}
	a, b, c := element(t, w, "Alpha"), element(t, w, "Beta"), element(t, w, "Gamma")
	pixel := capture()
	if pixel(a.X, a.Y) != bg || pixel(c.X+c.Width-1, c.Y) != bg {
		t.Fatal("outer corners are square")
	}
	if pixel(b.X, b.Y+3) == bg || pixel(a.X+a.Width-2, a.Y+3) == bg {
		t.Fatal("connected seam has rounded gaps")
	}
	v.Gap(8)
	pixel = capture()
	b = element(t, w, "Beta")
	if pixel(b.X, b.Y) != bg {
		t.Fatal("separated segment lacks corner")
	}
}
