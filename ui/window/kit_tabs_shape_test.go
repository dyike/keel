package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestTabsKeepSelectedAppearanceOnFocus(t *testing.T) {
	for _, variant := range []kit.TabsVariant{kit.TabsPill, kit.TabsOutline, kit.TabsSegmented} {
		for _, p := range []theme.Palette{theme.Light(), theme.Dark()} {
			for _, closable := range []bool{false, true} {
				tabs := kit.Tabs().Variant(variant).Add("First", nil).Add("Second", nil)
				if closable {
					tabs.Closable(func(int) {})
				}
				w := openTest(t, Options{Width: 360, Height: 180, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
					return cx.Themed(p, el.ViewFunc(func(cx *el.Context) el.Element {
						return el.Div().WFull().HFull().Bg(p.Bg).Child(tabs.Render(cx))
					}))
				}))})
				check := func(name string) {
					t.Helper()
					e := element(t, w, name)
					data, err := w.screenshot()
					if err != nil {
						t.Fatal(err)
					}
					im, err := png.Decode(bytes.NewReader(data))
					if err != nil {
						t.Fatal(err)
					}
					// This point is inside a 6dp rounded rectangle but outside a capsule.
					got := color.NRGBAModel.Convert(im.At(e.X+4, e.Y+4)).(color.NRGBA)
					if variant == kit.TabsPill && got != p.Bg {
						t.Fatalf("%s closable=%v: focus squared the pill corner: got %v want %v", name, closable, got, p.Bg)
					}
					want := p.Highlight
					if variant == kit.TabsSegmented {
						want = p.Surface
					}
					fill := color.NRGBAModel.Convert(im.At(e.X+5, e.Y+e.Height/2)).(color.NRGBA)
					if fill != want {
						t.Fatalf("variant=%v %s closable=%v: focus changed selected fill: got %v want %v", variant, name, closable, fill, want)
					}
				}
				check("First")
				w.click(element(t, w, "Second").center())
				check("Second")
				if err := w.press("left"); err != nil {
					t.Fatal(err)
				}
				check("First")
			}
		}
	}
}
