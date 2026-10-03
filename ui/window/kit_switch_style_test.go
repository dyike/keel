package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image/color"
	"image/png"
	"testing"
)

func TestSwitchCustomTrackColor(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	s := kit.Switch("Setting", true).Color(red)
	w := openTest(t, Options{Width: 300, Height: 120, Content: el.Root(s)})
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
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if c == red {
					n++
				}
			}
		}
		return n
	}
	large := count()
	if large == 0 {
		t.Fatal("missing custom track")
	}
	s.Size(kit.SwitchSmall).LabelSide(el.Left)
	if small := count(); small == 0 || small >= large {
		t.Fatal("small track", small, large)
	}
	s.SetDisabled(true)
	if count() != 0 {
		t.Fatal("disabled custom color not dimmed")
	}
	s.SetDisabled(false)
	s.SetValue(false)
	if count() != 0 {
		t.Fatal("custom color applied to unchecked track")
	}
	s.SetValue(true)
	s.ClearColor()
	if count() != 0 {
		t.Fatal("theme color not restored")
	}
	e := element(t, w, "Setting")
	if e.Role != "switch" || e.Checked == nil || !*e.Checked {
		t.Fatal(e)
	}
}
