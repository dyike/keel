package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestKitKbdSnapshotAndTab(t *testing.T) {
	calls := 0
	w := openTest(t, Options{Content: el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(kit.Kbd("mod+shift+p").Render(cx), kit.Kbd("中文 123").Plain().Render(cx), kit.Button("next", func() { calls++ }).Render(cx))
	}))})
	for _, label := range []string{"mod+shift+p", "中文 123"} {
		e := element(t, w, label)
		if e.Role != "text" || e.Value != "" || e.Disabled {
			t.Fatalf("kbd semantics: %+v", e)
		}
	}
	if len(w.snapshot()) != 3 {
		t.Fatalf("duplicate semantics: %+v", w.snapshot())
	}
	w.press("tab")
	w.press("enter")
	if calls != 1 {
		t.Fatal("keycap entered Tab order")
	}
}

func TestKitKbdRuntimeTheme(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	core.Update(func() { theme.Apply(theme.Light()) })
	views := []*kit.KbdView{kit.Kbd("mod+s"), kit.Kbd("mod+s").Plain()}
	v := views[0]
	w := openTest(t, Options{Width: 200, Height: 100, Content: el.Embed(el.ViewFunc(func(cx *el.Context) el.Element { return v.Render(cx) }))})
	var width, height int
	for _, palette := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(palette) })
		for i, view := range views {
			v = view
			plain := i == 1
			e := element(t, w, "mod+s")
			data, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			border, text := 0, 0
			for y := e.Y; y < e.Y+e.Height; y++ {
				for x := e.X; x < e.X+e.Width; x++ {
					c := color.NRGBAModel.Convert(im.At(x, y))
					// Count outline coverage; a 1px stroke may be antialiased.
					if c != palette.Bg && (x < e.X+2 || x >= e.X+e.Width-2 || y < e.Y+2 || y >= e.Y+e.Height-2) {
						border++
					}
					if c == palette.Muted {
						text++
					}
				}
			}
			if text == 0 || (!plain && border == 0) || (plain && border != 0) {
				t.Fatalf("keycap colors plain=%v: border=%d text=%d", plain, border, text)
			}
			if width != 0 && (e.Width != width || e.Height != height) {
				t.Fatal("theme/Plain changed keycap dimensions")
			}
			width, height = e.Width, e.Height
		}
	}
}
