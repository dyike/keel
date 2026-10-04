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
