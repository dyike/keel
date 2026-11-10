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

// Use a distinctive border color to distinguish outlines from fills, text,
// shadows and antialiasing. Scope the palette so defaults and explicit styles
// are exercised through the same render path without changing other windows.
func TestComponentOutlinesAreOptIn(t *testing.T) {
	marker := color.NRGBA{R: 255, B: 255, A: 255}
	for _, base := range []theme.Palette{theme.Light(), theme.Dark()} {
		for _, tc := range []struct {
			name     string
			view     el.View
			explicit bool
		}{
			{"input", kit.Input("Name"), false},
			{"checkbox", kit.Checkbox("Check", false), false},
			{"radio", kit.Radio("Choice"), false},
			{"group", kit.GroupBox("Group").Child(kit.Label("Content")), false},
			{"collapsible", kit.Collapsible("Details", kit.Label("Content")), false},
			{"attachment", kit.Attachment("file.txt", 10), false},
			{"kbd", kit.Kbd("mod+s"), false},
			{"group-outline", kit.GroupBox("Group").Variant(kit.GroupBoxOutline).Child(kit.Label("Content")), true},
			{"accordion-outline", kit.Accordion().Bordered(true).Add("Item", kit.Label("Content")), true},
			{"element-border", el.ViewFunc(func(*el.Context) el.Element { return el.Div().Size(el.Dp(40)).Border(2, theme.Border) }), true},
			{"separator", el.ViewFunc(func(*el.Context) el.Element { return el.Div().W(el.Dp(80)).H(el.Dp(2)).Bg(theme.Border) }), true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				p := base
				p.Border = marker
				w := openTest(t, Options{Width: 360, Height: 180, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
					return cx.Themed(p, tc.view)
				}))})
				for _, enabled := range []bool{false, true, false} {
					p.Outline = color.NRGBA{}
					if enabled {
						p.Outline = marker
					}
					data, err := w.screenshot()
					if err != nil {
						t.Fatal(err)
					}
					im, err := png.Decode(bytes.NewReader(data))
					if err != nil {
						t.Fatal(err)
					}
					pixels := 0
					for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
						for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
							c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
							if int(c.R) > int(c.G)+40 && int(c.B) > int(c.G)+40 {
								pixels++
							}
						}
					}
					if want := enabled || tc.explicit; (pixels > 10) != want {
						t.Fatalf("enabled=%v explicit=%v: outline pixels=%d", enabled, tc.explicit, pixels)
					}
				}
			})
		}
	}
}
