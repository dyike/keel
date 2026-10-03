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
)

func TestSwitchFocusRingVisibilityAndKeyboard(t *testing.T) {
	oldPalette, oldMotion := theme.Current(), theme.ReducedMotion
	core.Update(func() { theme.Apply(theme.Light()); theme.SetReducedMotion(true) })
	defer core.Update(func() { theme.Apply(oldPalette); theme.SetReducedMotion(oldMotion) })
	calls := 0
	s := kit.Switch("Setting", false).OnChange(func(bool) { calls++ })
	w := openTest(t, Options{Width: 300, Height: 120, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(s.Render(cx)) }))})
	count := func() int {
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
				if color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA) == theme.Primary {
					n++
				}
			}
		}
		return n
	}
	before := element(t, w, "Setting")
	if count() != 0 {
		t.Fatal("unfocused ring")
	}
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if count() == 0 {
		t.Fatal("keyboard focus ring missing")
	}
	s.FocusRing(false)
	if count() != 0 {
		t.Fatal("ring not hidden")
	}
	after := element(t, w, "Setting")
	if before.X != after.X || before.Y != after.Y || before.Width != after.Width || before.Height != after.Height {
		t.Fatal("focus changed layout")
	}
	if err := w.press("space"); err != nil {
		t.Fatal(err)
	}
	if !s.Value() || calls != 1 {
		t.Fatal("quiet focus lost keyboard activation")
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if s.Value() || calls != 2 {
		t.Fatal("enter did not toggle")
	}
	s.FocusRing(true)
	if count() == 0 {
		t.Fatal("focus ring not restored")
	}
	s.SetDisabled(true)
	if count() != 0 {
		t.Fatal("disabled focus ring")
	}
}
