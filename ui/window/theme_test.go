package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/theme"
)

type paletteView struct{}

func (paletteView) Render(cx *el.Context) el.Element {
	return el.Div().Child(cx.Cache("panel", func() el.Element {
		return el.Div().H(el.Dp(100)).Bg(theme.Surface).Child(el.Text("主题内容"))
	}))
}

func TestThemeApplyUpdatesExistingWindows(t *testing.T) {
	original := theme.Current()
	defer func() { loop.Lock(); defer loop.Unlock(); theme.Apply(original) }()
	windows := []*Window{
		openTest(t, Options{Width: 160, Height: 160, Content: el.Root(paletteView{})}),
		openTest(t, Options{Width: 160, Height: 160, Content: el.Root(paletteView{})}),
	}
	for _, p := range []theme.Palette{theme.Dark(), theme.Light()} {
		core.Update(func() { theme.Apply(p) })
		for _, w := range windows {
			if e := element(t, w, "主题内容"); e.Role != "text" {
				t.Fatalf("theme lost content: %+v", e)
			}
			data, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			for _, sample := range []struct {
				y    int
				want color.NRGBA
			}{{80, p.Surface}, {140, p.Bg}} {
				if got := color.NRGBAModel.Convert(img.At(120, sample.y)); got != sample.want {
					t.Fatalf("theme pixel at y=%d: got %v want %v", sample.y, got, sample.want)
				}
			}
		}
	}
}
