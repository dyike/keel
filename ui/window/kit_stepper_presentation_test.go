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

func TestCenteredStepperConnectorsFollowProgress(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	for _, palette := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(palette) })
		v := kit.Stepper("First", "Second", "Third").TextCenter(true).Navigation(kit.StepperNavigationAll)
		v.SetValue(1)
		w := openTest(t, Options{Width: 360, Height: 100, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return v.Render(cx) }))})
		w.render()
		w.render()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		for _, sample := range []struct {
			x    int
			want color.NRGBA
		}{{120, theme.Primary}, {240, theme.Border}} {
			got := color.NRGBAModel.Convert(im.At(sample.x, 12)).(color.NRGBA)
			if got != sample.want {
				t.Fatalf("connector at %d: %v want %v", sample.x, got, sample.want)
			}
		}
		w.click(element(t, w, "Third").center())
		if v.Value() != 2 {
			t.Fatal("centered step unavailable to agent")
		}
	}
}
