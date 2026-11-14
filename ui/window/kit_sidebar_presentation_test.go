package window

import (
	"bytes"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
)

func TestSidebarDividerPixelsAndSuffixAgent(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	for _, palette := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(palette) })
		calls := 0
		v := kit.Sidebar().Width(200).Height(180).BorderWidth(3).Collapsible(false).Section("", kit.SidebarItem{ID: "a", Label: "Alpha", Suffix: kit.Button("Refresh", func() { calls++ })})
		w := openTest(t, Options{Width: 300, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(v.Render(cx)) }))})
		for _, side := range []el.Side{el.Left, el.Right} {
			v.Side(side)
			w.render()
			data, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			x := 201
			if side == el.Right {
				x = 1
			}
			c := color.NRGBAModel.Convert(im.At(x, 80)).(color.NRGBA)
			if c != theme.Border {
				t.Fatalf("divider at %v: %v want %v", side, c, theme.Border)
			}
			action := element(t, w, "Refresh")
			if action.Role != "button" {
				t.Fatal("suffix not independently exposed", action)
			}
			w.click(action.center())
		}
		if calls != 2 || v.Value() != "" {
			t.Fatal("suffix activation selected navigation", calls, v.Value())
		}
	}
}

func TestSidebarResizeSeparatorAgent(t *testing.T) {
	v := kit.Sidebar().Width(200).Height(180).Resizable(160, 300).Section("Pages", kit.SidebarItem{ID: "a", Label: "Alpha"})
	w := openTest(t, Options{Width: 400, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Row().Items(el.Stretch).Child(v.Render(cx), el.Div().Grow())
	}))})
	w.render()
	handle := element(t, w, locale.Current().Resize)
	if handle.Role != "separator" || handle.Value != "200" {
		t.Fatalf("resize handle not exposed as a separator with its width: %+v", handle)
	}
	w.click(handle.center())
	if err := w.press(string(key.NameEnd)); err != nil {
		t.Fatal(err)
	}
	w.render()
	if got := element(t, w, locale.Current().Resize).Value; got != "300" || v.CurrentWidth() != 300 {
		t.Fatal("End did not resize to the maximum", got, v.CurrentWidth())
	}
}
