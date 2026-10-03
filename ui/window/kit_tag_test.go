package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
)

func TestKitTagSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Tag("已完成").Tone(kit.ToneSuccess))})
	if e := element(t, w, "已完成"); e.Role != "tag" || e.Value != "success" {
		t.Fatalf("invalid tag: %+v", e)
	}
}

func TestKitTagKeyboardRemovalAndSelection(t *testing.T) {
	removed := 0
	v := kit.Tag("待办").Selectable().OnRemove(func() { removed++ })
	w := openTest(t, Options{Content: el.Embed(v)})
	for _, s := range []string{"tab", "space", "tab", "backspace"} {
		if err := w.press(s); err != nil {
			t.Fatal(err)
		}
	}
	if !v.Value() || removed != 1 {
		t.Fatal("keyboard tag interaction failed")
	}
	e := element(t, w, "待办")
	if e.Selected == nil || !*e.Selected {
		t.Fatal("missing selected semantics")
	}
}

func TestKitTagCustomAppearancePixelsAndContent(t *testing.T) {
	blue := color.NRGBA{B: 255, A: 255}
	red := color.NRGBA{R: 255, A: 255}
	v := kit.Tag("Custom").Size(32).Rounded(0).Content(el.ViewFunc(func(*el.Context) el.Element { return el.Div().W(el.Dp(60)).H(el.Dp(20)) })).Appearance(func(a kit.TagAppearance) kit.TagAppearance {
		a.Background = blue
		a.SelectedBackground = red
		a.Border = color.NRGBA{}
		return a
	}).Selectable()
	w := openTest(t, kitPage(v))
	check := func(want color.NRGBA, equal bool) {
		t.Helper()
		e := element(t, w, "Custom")
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		got := color.NRGBAModel.Convert(img.At(e.X+2, e.Y+2)).(color.NRGBA)
		if (got == want) != equal {
			t.Fatalf("tag corner got=%v want=%v equal=%v", got, want, equal)
		}
	}
	check(blue, true)
	v.SetValue(true)
	check(red, true)
	v.Rounded(theme.RadiusFull)
	check(red, false)
	v.SetValue(false)
	v.Appearance(nil).Outline(true).Rounded(0)
	check(blue, false)
	if e := element(t, w, "Custom"); e.Role != "tag" || e.Selected == nil || *e.Selected {
		t.Fatalf("custom semantics %+v", e)
	}
	v.Content(nil)
	if roleOfName(w, "Custom") != "tag" {
		t.Fatal("content reset")
	}
}
